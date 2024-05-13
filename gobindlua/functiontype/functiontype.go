package functiontype

import (
	"bytes"
	"fmt"
	"go/ast"
	"io"
	"strings"
	"text/template"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
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
						LuaName:    "",
						DataType:   typ,
					}
					params = append(params, param)
					paramNum++
				} else {
					for _, name := range i.Names {
						typ := datatype.CreateDataTypeFromExpr(i.Type, packageSource)
						_, isEllipses := i.Type.(*ast.Ellipsis)
						luaName := gobindluautil.SnakeCase(name.Name)
						param := param.Param{
							IsEllipses: isEllipses,
							ParamNum:   paramNum,
							LuaName:    luaName,
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

func (f *FunctionType) GenerateLuaFunctionWrapper(out io.Writer, userDataCheckFnName string) {
	type FunctionGenerator struct {
		UserDataCheckFn string
		*FunctionType
	}

	templ := `
func {{ .FunctionType.SourceFnName }}(L *lua.LState) int {
	{{ if .FunctionType.Receiver -}}
		r := {{ .UserDataCheckFn }}(1, L)
	{{- end }}
	{{ range $idx, $param := .Params }}
		var p{{ $idx }} {{ $param.TemplateArg }}
	{{ end }}
	{{ range $idx, $param := .Params }}
		{
			{{ $param.ConvertLuaTypeToGo "ud" (printf "%s(%d)" $param.LuaParamType $param.ParamNum) $param.ParamNum }}
			p{{ $idx }} = {{ $param.ReferenceOrDereferenceForAssignmentToField }}ud
		}
	{{ end }}
	{{ .FunctionType.GenerateReturnValues "r" }} {{ if .FunctionType.Receiver -}}r.{{ end }}{{ .FunctionType.ActualFnName  }}({{ .FunctionType.GenerateParamValues "p" }})

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

	return {{ .FunctionType.NumReturns }}
}
`

	execTempl(out, FunctionGenerator{userDataCheckFnName, f}, templ)
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
---@param {{ $param.LuaName }} {{ $param.DataType.LuaType false }}
{{ end -}}
{{- end -}}
{{- end -}}
{{- range $idx,$ret := .Ret -}}
{{- if not $ret.IsError -}}
---@return {{ $ret.LuaType true }}
{{ end -}}
{{- end -}}
`

	execTempl(&out, f, templ)

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

func execTempl(out io.Writer, data any, templ string) {
	t := template.Must(template.New("").Parse(templ))
	err := t.Execute(out, data)

	if err != nil {
		panic(err)
	}
}
