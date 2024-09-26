package datatype

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strings"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/declaredinterface"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"golang.org/x/tools/go/packages"
)

type FunctionType struct {
	ActualFnName string
	LuaFnName    string
	SourceFnName string
	Receiver     bool
	Params       []Param
	Ret          []DataType
}

func CreateFunctionFromExpr(fn *ast.FuncDecl, luaFnName, goBindFnName string, packageSource *packages.Package, allDeclaredInterfaces []declaredinterface.DeclaredInterface) FunctionType {
	actualName := fn.Name.Name
	fnTyp := packageSource.TypesInfo.ObjectOf(fn.Name).Type().(*types.Signature)
	return CreateFunction(fnTyp, actualName, luaFnName, goBindFnName, packageSource, allDeclaredInterfaces)
}

func CreateFunction(typ *types.Signature, actualFnName, luaFnName, goBindFnName string, packageSource *packages.Package, allDeclaredInterfaces []declaredinterface.DeclaredInterface) FunctionType {
	params := make([]Param, 0)
	ret := make([]DataType, 0)

	paramNum := 1

	if typ.Recv() != nil {
		paramNum = 2
	}

	for i := 0; i < typ.Params().Len(); i++ {
		param := typ.Params().At(i)
		paramType := CreateDataTypeFrom(param.Type(), packageSource, allDeclaredInterfaces)
		luaName := gobindluautil.SnakeCase(param.Name())
		isEllipses := i == typ.Params().Len()-1 && typ.Variadic()
		params = append(params, Param{
			IsEllipses: isEllipses,
			ParamNum:   paramNum,
			LuaName:    luaName,
			DataType:   paramType,
		})
		paramNum++
	}

	for i := 0; i < typ.Results().Len(); i++ {
		res := typ.Results().At(i)
		resTyp := CreateDataTypeFrom(res.Type(), packageSource, allDeclaredInterfaces)
		ret = append(ret, resTyp)
	}

	return FunctionType{
		ActualFnName: actualFnName,
		LuaFnName:    luaFnName,
		SourceFnName: goBindFnName,
		Receiver:     typ.Recv() != nil,
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

func (f *FunctionType) GenerateLuaFunctionWrapper(out io.Writer, userDataCheckFnName string) {
	type FunctionGenerator struct {
		UserDataCheckFn string
		*FunctionType
	}

	templ := `
func {{ .SourceFnName }}(L *lua.LState) int {
	{{ if .Receiver -}}
		r := {{ .UserDataCheckFn }}(1, L)
	{{- end }}
	{{ range $idx, $param := .Params }}
		var p{{ $idx }} {{ $param.ActualTemplateArg }}
	{{ end }}
	{{ range $idx, $param := .Params }}
		{
			{{ $param.ConvertLuaTypeToGo "ud" (printf "%s(%d)" $param.LuaParamType $param.ParamNum) $param.ParamNum }}
			p{{ $idx }} = {{ $param.ReferenceOrDereferenceForAssignmentToField }}ud
		}
	{{ end }}
	{{ .GenerateReturnValues "r" }} {{ if .Receiver -}}r.{{ end }}{{ .ActualFnName  }}({{ .GenerateParamValues "p" }})

	{{ range $idx, $ret := .Ret -}}
		{{- if $ret.IsError -}}
			if r{{ $idx }} != nil {
				L.Error(lua.LString(r{{ $idx }}.Error()), 1)
			}
		{{- end -}}
	{{- end }}

	{{ range $idx, $ret := .Ret -}}
		{{- if not $ret.IsError -}}
			{{ $name := printf "r%d" $idx }}
			L.Push({{$ret.ConvertGoTypeToLua $name}})
		{{- end -}}
	{{- end }}

	return {{ .NumReturns }}
}
`

	execTemplString(out, FunctionGenerator{userDataCheckFnName, f}, templ)
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

func (f *FunctionType) GenerateLuaFunctionParamRetDefinitions() string {
	out := bytes.Buffer{}

	templ := `
{{ range $idx,$param := .Params -}}
{{ if $param.IsEllipses -}}
---@param ... {{ $param.LuaType false }}
{{ else -}}
{{ if ne $param.LuaName "" -}}
---@param {{ $param.LuaName }} {{ $param.LuaType false }}
{{ end -}}
{{- end -}}
{{- end -}}
{{- range $idx,$ret := .Ret -}}
{{- if not $ret.IsError -}}
---@return {{ $ret.LuaType true }}
{{ end -}}
{{- end -}}
`

	execTemplString(&out, f, templ)

	return out.String()
}

func (f *FunctionType) GenerateLuaFunctionParamStubs() string {
	ret := make([]string, 0)

	for _, p := range f.Params {
		if p.IsEllipses {
			ret = append(ret, "...")
		} else {
			ret = append(ret, p.LuaName)
		}
	}

	return strings.Join(ret, ", ")
}
