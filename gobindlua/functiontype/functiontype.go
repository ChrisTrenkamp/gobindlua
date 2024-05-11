package functiontype

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/param"
	"golang.org/x/tools/go/packages"
)

type FunctionType struct {
	ActualFnName string
	LuaFnName    string
	SourceFnName string
	Receiver     bool
	Params       []param.Param
	Ret          []datatype.DataType
}

func CreateFunction(fn *ast.FuncDecl, receiver bool, luaName, sourceCodeName string, packageSource *packages.Package) FunctionType {
	params := make([]param.Param, 0)
	ret := make([]datatype.DataType, 0)

	if fn.Type != nil {
		if fn.Type.Params != nil {
			paramNum := 1

			if receiver {
				paramNum++
			}

			for _, i := range fn.Type.Params.List {
				if len(i.Names) == 0 {
					typ := datatype.CreateDataTypeFromExpr(i.Type, packageSource)
					_, isEllipses := i.Type.(*ast.Ellipsis)
					param := param.Param{
						IsEllipses: isEllipses,
						ParamNum:   paramNum,
						DataType:   typ,
					}
					params = append(params, param)
					paramNum++
				} else {
					for range i.Names {
						typ := datatype.CreateDataTypeFromExpr(i.Type, packageSource)
						_, isEllipses := i.Type.(*ast.Ellipsis)
						param := param.Param{
							IsEllipses: isEllipses,
							ParamNum:   paramNum,
							DataType:   typ,
						}
						params = append(params, param)
						paramNum++
					}
				}
			}
		}

		if fn.Type.Results != nil {
			for _, i := range fn.Type.Results.List {
				if len(i.Names) == 0 {
					ret = append(ret, datatype.CreateDataTypeFromExpr(i.Type, packageSource))
				} else {
					for range i.Names {
						ret = append(ret, datatype.CreateDataTypeFromExpr(i.Type, packageSource))
					}
				}
			}
		}
	}

	return FunctionType{
		ActualFnName: fn.Name.Name,
		LuaFnName:    luaName,
		SourceFnName: sourceCodeName,
		Receiver:     receiver,
		Params:       params,
		Ret:          ret,
	}
}

func (f *FunctionType) NumReturns() int {
	ret := len(f.Ret)

	for _, i := range f.Ret {
		if i.IsError() {
			ret--
		}
	}

	return ret
}

func (f *FunctionType) GenerateReturnValues(prefix string) string {
	if len(f.Ret) == 0 {
		return ""
	}

	ret := make([]string, 0)

	for i := range f.Ret {
		ret = append(ret, fmt.Sprintf("%s%d", prefix, i))
	}

	return strings.Join(ret, ",") + " := "
}

func (f *FunctionType) GenerateParamValues(prefix string) string {
	ret := make([]string, 0)

	for i := range f.Params {
		if f.Params[i].IsEllipses {
			ret = append(ret, fmt.Sprintf("%s%d...", prefix, i))
		} else {
			ret = append(ret, fmt.Sprintf("%s%d", prefix, i))
		}
	}

	return strings.Join(ret, ",")
}
