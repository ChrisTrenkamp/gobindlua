package structgen

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strings"
	"text/template"
	"unicode"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/declaredinterface"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/functiontype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gblimports"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/param"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/structfield"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type StructGenerator struct {
	structToGenerate string
	wd               string
	pathToOutput     string
	dependantModules []string
	includeFunctions []string
	excludeFunctions []string

	packageSource         *packages.Package
	allDeclaredInterfaces []declaredinterface.DeclaredInterface
	structObject          types.Object

	StaticFunctions []functiontype.FunctionType
	UserDataMethods []functiontype.FunctionType
	Fields          []structfield.StructField
	imports         gblimports.Imports
}

func NewStructGenerator(structToGenerate, wd, pathToOutput string, dependantModules []string, includeFunctions, excludeFunctions []string) *StructGenerator {
	return &StructGenerator{
		structToGenerate: structToGenerate,
		wd:               wd,
		pathToOutput:     pathToOutput,
		dependantModules: dependantModules,
		includeFunctions: includeFunctions,
		excludeFunctions: excludeFunctions,
	}
}

func (g *StructGenerator) GenerateSourceCode() ([]byte, []byte, error) {
	if err := g.loadSourcePackage(); err != nil {
		return nil, nil, fmt.Errorf("failed to load imported go packages: %s", err)
	}

	g.StaticFunctions = g.gatherFunctionsToGenerate()
	g.UserDataMethods = g.gatherReceivers()
	g.Fields = g.gatherFields()
	g.imports = gblimports.NewImports(g.packageSource)

	g.imports.AddPackageFromFunctions(g.StaticFunctions)
	g.imports.AddPackageFromFunctions(g.UserDataMethods)

	for _, i := range g.Fields {
		g.imports.AddPackage(i.DataType)
	}

	goCode, gerr := g.generateGoCode()
	luaDef, lerr := g.generateLuaDefinitions()
	return goCode, luaDef, errors.Join(gerr, lerr)
}

func (g *StructGenerator) loadSourcePackage() error {
	var err error
	g.packageSource, g.allDeclaredInterfaces, err = gobindluautil.LoadSourcePackage(g.wd, g.dependantModules)

	if err != nil {
		return err
	}

	g.structObject = g.packageSource.Types.Scope().Lookup(g.structToGenerate)

	if g.structObject == nil {
		return fmt.Errorf("specified type not found")
	}

	if _, ok := g.structObject.Type().Underlying().(*types.Struct); !ok {
		return fmt.Errorf("specified type is not a struct")
	}

	return nil
}

func (g *StructGenerator) StructToGenerate() string {
	return g.structObject.Name()
}

func (g *StructGenerator) StructMetatableIdentifier() string {
	return gobindluautil.SnakeCase(g.StructToGenerate())
}

func (g *StructGenerator) StructMetatableFieldsIdentifier() string {
	return gobindluautil.StructFieldMetadataName(g.StructToGenerate())
}

func (g *StructGenerator) UserDataCheckFn() string {
	return "luaCheck" + g.StructToGenerate()
}

func (g *StructGenerator) SourceUserDataAccess() string {
	return "luaAccess" + g.StructToGenerate()
}

func (g *StructGenerator) SourceUserDataSet() string {
	return "luaSet" + g.StructToGenerate()
}

func (g *StructGenerator) gatherFunctionsToGenerate() []functiontype.FunctionType {
	return g.gatherConstructors()
}

func (g *StructGenerator) gatherConstructors() []functiontype.FunctionType {
	ret := make([]functiontype.FunctionType, 0)
	underylingStructType := g.structObject.Type().Underlying()
	constructorPrefix := "New" + g.structToGenerate

	for _, syn := range g.packageSource.Syntax {
		for _, dec := range syn.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok && fn.Type.Results != nil && fn.Recv == nil {
				for _, retType := range fn.Type.Results.List {
					retType := datatype.CreateDataTypeFromExpr(retType.Type, g.packageSource, g.allDeclaredInterfaces)

					if retType.Type.Underlying() == underylingStructType {
						fnName := fn.Name.Name

						if strings.HasPrefix(fnName, "luaCheck") {
							continue
						}

						if gobindluautil.HasFilters(g.includeFunctions, g.excludeFunctions) {
							if !gobindluautil.CheckInclude(fnName, g.includeFunctions, g.excludeFunctions) {
								continue
							}
						} else if !strings.HasPrefix(fnName, constructorPrefix) {
							continue
						}

						luaName := gobindluautil.SnakeCase(fnName)

						if strings.HasPrefix(fnName, constructorPrefix) {
							luaName = "New" + strings.TrimPrefix(fnName, constructorPrefix)
							luaName = gobindluautil.SnakeCase(luaName)
						}

						sourceCodeName := "luaConstructor" + g.StructToGenerate() + fnName
						ret = append(ret, functiontype.CreateFunction(fn, false, luaName, sourceCodeName, g.packageSource, g.allDeclaredInterfaces))
						break
					}
				}
			}
		}
	}

	return ret
}

func (g *StructGenerator) gatherReceivers() []functiontype.FunctionType {
	ret := make([]functiontype.FunctionType, 0)
	underylingStructType := g.structObject.Type().Underlying()

	for _, syn := range g.packageSource.Syntax {
		for _, dec := range syn.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok && fn.Recv != nil {
				fnName := fn.Name.Name

				if fnName == "LuaModuleName" || fnName == "LuaRegisterGlobalMetatable" || fnName == "LuaModuleLoader" || fnName == "LuaMetatableType" {
					continue
				}

				if gobindluautil.HasFilters(g.includeFunctions, g.excludeFunctions) {
					if !gobindluautil.CheckInclude(fnName, g.includeFunctions, g.excludeFunctions) {
						continue
					}
				} else if !unicode.IsUpper(rune(fnName[0])) {
					continue
				}

				for _, recType := range fn.Recv.List {
					recType := datatype.CreateDataTypeFromExpr(recType.Type, g.packageSource, g.allDeclaredInterfaces)

					if recType.Type.Underlying() == underylingStructType {
						luaName := gobindluautil.SnakeCase(fnName)
						sourceCodeName := "luaMethod" + g.StructToGenerate() + fnName
						ret = append(ret, functiontype.CreateFunction(fn, true, luaName, sourceCodeName, g.packageSource, g.allDeclaredInterfaces))
						break
					}
				}
			}
		}
	}

	return ret
}

func (g *StructGenerator) gatherFields() []structfield.StructField {
	ret := make([]structfield.StructField, 0)
	str := g.structObject.Type().Underlying().(*types.Struct)

	for i := 0; i < str.NumFields(); i++ {
		field := str.Field(i)
		tag := str.Tag(i)
		luaName := structfield.GetLuaNameFromTag(field, tag)

		if luaName != "" {
			ret = append(ret, structfield.CreateStructField(field, luaName, g.packageSource, g.allDeclaredInterfaces))
		}
	}

	return ret
}

func (g *StructGenerator) generateGoCode() ([]byte, error) {
	code := bytes.Buffer{}

	g.imports.GenerateHeader(&code)
	g.buildMetatableInitFunction(&code)
	g.buildMetatableFunctions(&code)
	g.builderUserDataFunctions(&code)

	originalCodeBytes := code.Bytes()
	formattedCode, err := imports.Process(g.pathToOutput, originalCodeBytes, nil)
	if err != nil {
		return originalCodeBytes, err
	}

	return formattedCode, nil
}

func (g *StructGenerator) buildMetatableInitFunction(out io.Writer) {
	templ := `
func (goType *{{ .StructToGenerate }}) LuaModuleName() string {
	return "{{ .StructMetatableIdentifier }}"
}

func (goType *{{ .StructToGenerate }}) LuaModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	{{ range $idx, $fn := .StaticFunctions -}}
		L.SetField(staticMethodsTable, "{{ $fn.LuaFnName }}", L.NewFunction({{ $fn.SourceFnName }}))
	{{ end }}
    L.Push(staticMethodsTable)

	return 1
}

func (goType *{{ .StructToGenerate }}) LuaRegisterGlobalMetatable(L *lua.LState) {
	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction({{ .SourceUserDataAccess }}))
	L.SetField(fieldsTable, "__newindex", L.NewFunction({{ .SourceUserDataSet }}))
}
`

	execTempl(out, g, templ)
}

func (g *StructGenerator) buildMetatableFunctions(out io.Writer) {
	for _, i := range g.StaticFunctions {
		i.GenerateLuaFunctionWrapper(out, g.UserDataCheckFn())
	}
}

func (g *StructGenerator) builderUserDataFunctions(out io.Writer) {
	g.generateUserDataPredefinitions(out)
	g.generateStructAccessFunction(out)
	g.generateStructSetFunction(out)
	g.generateStructMethods(out)
}

func (g *StructGenerator) generateUserDataPredefinitions(out io.Writer) {
	templ := `
func (r *{{ .StructToGenerate }}) LuaMetatableType() string {
	return "{{ .StructMetatableFieldsIdentifier }}"
}

func {{ .UserDataCheckFn }}(param int, L *lua.LState) *{{ .StructToGenerate }} {
	ud := L.CheckUserData(param)
	v, ok := ud.Value.(*{{ .StructToGenerate }})
	if !ok {
		L.ArgError(1, gobindlua.CastArgError("{{ .StructToGenerate }}", ud.Value))
	}
	return v
}
`

	execTempl(out, g, templ)
}

func (g *StructGenerator) generateStructAccessFunction(out io.Writer) {
	templ := `
func {{ .SourceUserDataAccess }}(L *lua.LState) int {
	{{- if gt (len .Fields) 0 }}
	p1 := {{ .UserDataCheckFn }}(1, L)
	{{- end }}
	p2 := L.CheckString(2)

	switch p2 {
		{{- range $idx, $field := .Fields }}
	case "{{ $field.LuaName }}":
		L.Push({{ $field.DataType.ConvertGoTypeToLua (printf "p1.%s" $field.FieldName) }})
		{{ end -}}

		{{- range $idx, $method := .UserDataMethods }}
	case "{{ $method.LuaFnName }}":
		L.Push(L.NewFunction({{ $method.SourceFnName }}))
		{{ end }}

	default:
		L.Push(lua.LNil)
	}

	return 1
}
`

	execTempl(out, g, templ)
}

func (g *StructGenerator) generateStructSetFunction(out io.Writer) {
	templ := `
func {{ .SourceUserDataSet }}(L *lua.LState) int {
	{{- if gt (len .Fields) 0 }}
	p1 := {{ .UserDataCheckFn }}(1, L)
	{{- end }}
	p2 := L.CheckString(2)

	switch p2 {
		{{- range $idx, $field := .Fields }}
	case "{{ $field.LuaName }}":
		{{ $field.DataType.ConvertLuaTypeToGo "ud" (printf "%s(3)" $field.DataType.LuaParamType) 3 }}
		p1.{{ $field.FieldName }} = {{ $field.DataType.ReferenceOrDereferenceForAssignmentToField }}ud
		{{ end }}

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}
`

	execTempl(out, g, templ)
}

func (g *StructGenerator) generateStructMethods(out io.Writer) {
	for _, i := range g.UserDataMethods {
		i.GenerateLuaFunctionWrapper(out, g.UserDataCheckFn())
	}
}

func (g *StructGenerator) generateLuaDefinitions() ([]byte, error) {
	ret := bytes.Buffer{}

	templ := `---Code generated by gobindlua.  DO NOT EDIT.
---@meta {{ .StructMetatableIdentifier }}
{{- $gen := . }}

local {{ $gen.StructMetatableIdentifier }} = {}
{{ range $fidx,$staticFunc := .StaticFunctions -}}
{{ $staticFunc.GenerateLuaFunctionParamRetDefinitions -}}
function {{ $gen.StructMetatableIdentifier }}.{{ $staticFunc.LuaFnName }}({{ $staticFunc.GenerateLuaFunctionParamStubs }}) end
{{ end -}}

{{- $fieldIdent := .StructMetatableFieldsIdentifier }}
---@class {{ $fieldIdent }}{{ .GenerateInterfaceDeclarations }}
{{- range $fidx,$field := .Fields }}
---@field public {{ $field.LuaName }} {{ $field.DataType.LuaType true }}
{{- end }}
local {{ $fieldIdent }} = {}
{{- range $midx,$methodFunc := .UserDataMethods }}
{{ $methodFunc.GenerateLuaFunctionParamRetDefinitions -}}
function {{ $fieldIdent }}:{{ $methodFunc.LuaFnName }}({{ $methodFunc.GenerateLuaFunctionParamStubs }}) end
{{- end }}

return {{ $gen.StructMetatableIdentifier }}
`

	execTempl(&ret, g, templ)

	return ret.Bytes(), nil
}

func (g *StructGenerator) GenerateInterfaceDeclarations() string {
	ret := make([]string, 0)

	for _, i := range g.allDeclaredInterfaces {
		if g.implementsInterface(i.Interface) {
			ret = append(ret, gobindluautil.StructOrInterfaceMetadataName(i.Name))
		}
	}

	if len(ret) == 0 {
		return ""
	}

	return " : " + strings.Join(ret, ", ")
}

func (s *StructGenerator) implementsInterface(iface *types.Interface) bool {
	numIfaceMethods := iface.NumMethods()
	matchingMethods := 0

	for i := 0; i < iface.NumMethods(); i++ {
		ifaceMethod := iface.Method(i)
		if ifaceMethod.Name() == "LuaMetatableType" {
			numIfaceMethods--
			continue
		}

		ifaceFunc := ifaceMethod.Type().(*types.Signature)

		matches := false

		for _, structMethod := range s.UserDataMethods {
			if structMethod.ActualFnName == ifaceMethod.Name() {
				if s.methodParamsMatch(structMethod.Params, ifaceFunc.Params()) &&
					s.methodReturnsMatch(structMethod.Ret, ifaceFunc.Results()) {
					matches = true
				}
			}
		}

		if matches {
			matchingMethods++
		}
	}

	return matchingMethods == numIfaceMethods
}

func (s *StructGenerator) methodParamsMatch(structParams []param.Param, interfaceParams *types.Tuple) bool {
	if len(structParams) != interfaceParams.Len() {
		return false
	}

	for i := 0; i < len(structParams); i++ {
		sp := structParams[i].Type
		ip := interfaceParams.At(i).Type()

		if !types.Identical(sp, ip) {
			return false
		}
	}

	return true
}

func (s *StructGenerator) methodReturnsMatch(structRet []datatype.DataType, interfaceRet *types.Tuple) bool {
	if len(structRet) != interfaceRet.Len() {
		return false
	}

	for i := 0; i < len(structRet); i++ {
		sp := structRet[i].Type
		ip := interfaceRet.At(i).Type()

		if !types.Identical(sp, ip) {
			return false
		}
	}

	return true
}

func execTempl(out io.Writer, data any, templ string) {
	t := template.Must(template.New("").Parse(templ))
	err := t.Execute(out, data)

	if err != nil {
		panic(err)
	}
}
