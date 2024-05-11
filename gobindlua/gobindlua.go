package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/functiontype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/structfield"
)

type flagArray []string

func (i *flagArray) String() string {
	return ""
}

func (i *flagArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var errStructOrPackageUnspecified = fmt.Errorf("-s or -p must be specified")
var errIncorrectGoGeneratePlacement = fmt.Errorf("go:generate gobindlua directives must be placed behind a struct or package declaration")

func determineFromGoLine(structToGenerate, packageToGenerate *string) error {
	lineStr := os.Getenv("GOLINE")
	gofile := os.Getenv("GOFILE")

	if lineStr == "" || gofile == "" {
		return errStructOrPackageUnspecified
	}

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return err
	}

	f, err := os.ReadFile(gofile)
	if err != nil {
		return err
	}

	spl := bytes.Split(f, []byte("\n"))
	var splLine []byte

	for {
		if len(spl) < line {
			return errIncorrectGoGeneratePlacement
		}

		splLine = spl[line]
		line++

		splLine = bytes.TrimSpace(splLine)

		if len(splLine) == 0 || bytes.HasPrefix(splLine, []byte("//")) {
			continue
		}

		break
	}

	norm := regexp.MustCompile(`\s+`).ReplaceAllString(string(splLine), " ")
	normSpl := strings.Split(norm, " ")

	if len(normSpl) < 2 {
		return errIncorrectGoGeneratePlacement
	}

	switch normSpl[0] {
	case "type":
		if len(normSpl) < 3 {
			return errIncorrectGoGeneratePlacement
		}

		if strings.HasPrefix(normSpl[2], "struct") {
			*structToGenerate = normSpl[1]
			return nil
		}
	case "package":
		*packageToGenerate = normSpl[1]
		return nil
	}

	return errIncorrectGoGeneratePlacement
}

func main() {
	includeFunctions := make(flagArray, 0)
	excludeFunctions := make(flagArray, 0)
	workingDir := flag.String("d", "", "The Go source directory to generate the bindings from. Uses the current working directory if empty")
	structToGenerate := flag.String("s", "", "Generate the GopherLua bindings for the given struct.")
	packageToGenerate := flag.String("p", "", "Generate the GopherLua bindings from the functions in the -d parameter.")
	flag.Var(&includeFunctions, "i", "Only include the given function or method names.")
	flag.Var(&excludeFunctions, "x", "Exclude the given function or method names.")
	metatableName := flag.String("t", "", "Generate a LuaMetatableType method that returns the given value.  Takes the snake_case form of -s if unspecified.")
	outFile := flag.String("o", "", "The output file.  Defaults to lua_structname.go if empty.")

	flag.Parse()

	if flag.NArg() != 0 {
		log.Fatal("gobindlua does not accept arguments")
	}

	if *structToGenerate == "" && *packageToGenerate == "" {
		if err := determineFromGoLine(structToGenerate, packageToGenerate); err != nil {
			log.Fatal(err.Error())
		}
	}

	if *structToGenerate != "" && *packageToGenerate != "" {
		log.Fatal("only one of -s or -p may be specified")
	}

	if len(includeFunctions) > 0 && len(excludeFunctions) > 0 {
		log.Fatal("only one of -i or -x may be specified")
	}

	if *metatableName == "" && *structToGenerate != "" {
		*metatableName = gobindluautil.SnakeCase(*structToGenerate)
	}

	if *workingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("error getting working directory: %s", err)
		}
		*workingDir = wd
	}

	if *outFile == "" {
		if *structToGenerate != "" {
			*outFile = "lua_" + *structToGenerate + ".go"
		} else {
			*outFile = "lua_" + filepath.Base(*packageToGenerate) + ".go"
		}
	}

	pathToOutput := filepath.Join(*workingDir, *outFile)
	gen := NewGenerator(*structToGenerate, *packageToGenerate, *workingDir, *metatableName, includeFunctions, excludeFunctions)
	outBytes, err := gen.GenerateSourceCode(pathToOutput)

	if len(outBytes) > 0 {
		if werr := os.WriteFile(pathToOutput, outBytes, 0644); werr != nil {
			log.Fatal(werr)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}

type Generator struct {
	structToGenerate  string
	packageToGenerate string
	wd                string
	metatableName     string
	includeFunctions  []string
	excludeFunctions  []string

	packageSource *packages.Package
	structObject  types.Object

	StaticFunctions []functiontype.FunctionType
	UserDataMethods []functiontype.FunctionType

	imports               map[string]string
	metatableInitFunction bytes.Buffer
	metatableFunctions    bytes.Buffer
	userDataFunctions     bytes.Buffer
}

func NewGenerator(structToGenerate, packageToGenerate, wd string, metatableName string, includeFunctions, excludeFunctions []string) *Generator {
	return &Generator{
		structToGenerate:  structToGenerate,
		packageToGenerate: packageToGenerate,
		wd:                wd,
		metatableName:     metatableName,
		includeFunctions:  includeFunctions,
		excludeFunctions:  excludeFunctions,
	}
}

func (g *Generator) GenerateSourceCode(pathToOutput string) ([]byte, error) {
	if err := g.loadSourcePackage(); err != nil {
		return nil, fmt.Errorf("failed to load imported go packages: %s", err)
	}

	g.StaticFunctions = g.gatherFunctionsToGenerate()
	g.UserDataMethods = g.gatherReceivers()

	g.imports = make(map[string]string)
	g.imports["github.com/yuin/gopher-lua"] = "lua"
	g.imports["github.com/ChrisTrenkamp/gobindlua"] = ""

	g.addPackageFromFunctions(g.StaticFunctions)
	g.addPackageFromFunctions(g.UserDataMethods)

	if g.structToGenerate != "" {
		for _, i := range g.GatherFields() {
			g.addPackage(i.DataType)
		}
	}

	g.buildMetatableInitFunction()
	g.buildMetatableFunctions()
	g.builderUserDataFunctions()

	code := bytes.Buffer{}
	fmt.Fprintf(&code, "// Code generated by gobindlua; DO NOT EDIT.\n")
	fmt.Fprintf(&code, "package %s\n\nimport (\n", g.packageSource.Types.Name())
	for pkg, name := range g.imports {
		fmt.Fprintf(&code, "\t%s \"%s\"\n", name, pkg)
	}
	fmt.Fprintf(&code, ")\n")
	io.Copy(&code, &g.metatableInitFunction)
	io.Copy(&code, &g.metatableFunctions)
	io.Copy(&code, &g.userDataFunctions)

	originalCodeBytes := code.Bytes()
	formattedCode, err := imports.Process(pathToOutput, originalCodeBytes, nil)
	if err != nil {
		return originalCodeBytes, err
	}

	return formattedCode, nil
}

func (g *Generator) addPackageFromFunctions(f []functiontype.FunctionType) {
	for _, i := range f {
		for _, p := range i.Params {
			g.addPackage(p.DataType)
		}

		for _, p := range i.Ret {
			g.addPackage(p)
		}
	}
}

func (g *Generator) addPackage(d datatype.DataType) {
	if p := d.Package(); p != "" && p != g.packageSource.ID {
		g.imports[p] = ""
	}
}

func (g *Generator) loadSourcePackage() error {
	config := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  g.wd,
	}
	var pack []*packages.Package
	var err error

	pack, err = packages.Load(config, g.wd)

	if err != nil {
		return err
	}

	if len(pack) == 1 {
		g.packageSource = pack[0]
	} else {
		return fmt.Errorf("packages.Load returned more than one package")
	}

	if g.structToGenerate != "" {
		g.structObject = g.packageSource.Types.Scope().Lookup(g.structToGenerate)

		if g.structObject == nil {
			return fmt.Errorf("specified type not found")
		}

		if _, ok := g.structObject.Type().Underlying().(*types.Struct); !ok {
			return fmt.Errorf("specified type is not a struct")
		}
	}

	return nil
}

func (g *Generator) StructToGenerate() string {
	return g.structObject.Name()
}

func (g *Generator) PackageToGenerateFunctionName() string {
	caser := cases.Title(language.English)
	return "Register" + caser.String(g.packageToGenerate) + "LuaType"
}

func (g *Generator) PackageToGenerateMetatableName() string {
	return g.packageToGenerate
}

func (g *Generator) StructMetatableIdentifier() string {
	return gobindluautil.SnakeCase(g.StructToGenerate())
}

func (g *Generator) StructMetatableFieldsIdentifier() string {
	return gobindluautil.SnakeCase(g.StructToGenerate() + "Fields")
}

func (g *Generator) UserDataCheckFn() string {
	return "luaCheck" + g.StructToGenerate()
}

func (g *Generator) SourceUserDataAccess() string {
	return "luaAccess" + g.StructToGenerate()
}

func (g *Generator) SourceUserDataSet() string {
	return "luaSet" + g.StructToGenerate()
}

func (g *Generator) buildMetatableInitFunction() {
	if g.structToGenerate == "" {
		templ := `
func {{ .PackageToGenerateFunctionName }}(L *lua.LState) {
	staticMethodsTable := L.NewTypeMetatable("{{ .PackageToGenerateMetatableName }}")
	L.SetGlobal("{{ .PackageToGenerateMetatableName }}", staticMethodsTable)
	{{ range $idx, $fn := .StaticFunctions -}}
		L.SetField(staticMethodsTable, "{{ $fn.LuaFnName }}", L.NewFunction({{ $fn.SourceFnName }}))
	{{ end }}
}
`

		execTempl(&g.metatableInitFunction, g, templ)
		return
	}

	templ := `
func (goType {{ .StructToGenerate }}) RegisterLuaType(L *lua.LState) {
	staticMethodsTable := L.NewTypeMetatable("{{ .StructMetatableIdentifier }}")
	L.SetGlobal("{{ .StructMetatableIdentifier }}", staticMethodsTable)
	{{ range $idx, $fn := .StaticFunctions -}}
		L.SetField(staticMethodsTable, "{{ $fn.LuaFnName }}", L.NewFunction({{ $fn.SourceFnName }}))
	{{ end }}
	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction({{ .SourceUserDataAccess }}))
	L.SetField(fieldsTable, "__newindex", L.NewFunction({{ .SourceUserDataSet }}))
}
`

	execTempl(&g.metatableInitFunction, g, templ)
}

func (g *Generator) buildMetatableFunctions() {
	for _, i := range g.StaticFunctions {
		generateLuaFunctionWrapper(&g.metatableFunctions, g, i)
	}
}

func (g *Generator) builderUserDataFunctions() {
	if g.structToGenerate == "" {
		return
	}

	g.generateUserDataPredefinitions()
	g.generateStructAccessFunction()
	g.generateStructSetFunction()
	g.generateStructMethods()
}

func (g *Generator) generateUserDataPredefinitions() {
	templ := `
func (r *{{ .StructToGenerate }}) LuaMetatableType() string {
	return "{{ .StructMetatableFieldsIdentifier }}"
}

func {{ .UserDataCheckFn }}(param int, L *lua.LState) *{{ .StructToGenerate }} {
	ud := L.CheckUserData(param)
	if v, ok := ud.Value.(*{{ .StructToGenerate }}); ok {
		return v
	}
	L.ArgError(1, "{{ .StructToGenerate }} expected")
	return nil
}
`

	execTempl(&g.userDataFunctions, g, templ)
}

func (g *Generator) generateStructAccessFunction() {
	templ := `
func {{ .SourceUserDataAccess }}(L *lua.LState) int {
	{{- if gt (len .GatherFields) 0 }}
	p1 := {{ .UserDataCheckFn }}(1, L)
	{{- end }}
	p2 := L.CheckString(2)

	switch p2 {
		{{- range $idx, $field := .GatherFields }}
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

	execTempl(&g.userDataFunctions, g, templ)
}

func (g *Generator) generateStructSetFunction() {
	templ := `
func {{ .SourceUserDataSet }}(L *lua.LState) int {
	{{- if gt (len .GatherFields) 0 }}
	p1 := {{ .UserDataCheckFn }}(1, L)
	{{- end }}
	p2 := L.CheckString(2)

	switch p2 {
		{{- range $idx, $field := .GatherFields }}
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

	execTempl(&g.userDataFunctions, g, templ)
}

func (g *Generator) generateStructMethods() {
	for _, i := range g.UserDataMethods {
		generateLuaFunctionWrapper(&g.userDataFunctions, g, i)
	}
}

func generateLuaFunctionWrapper(out io.Writer, g *Generator, f functiontype.FunctionType) {
	type FunctionGenerator struct {
		*Generator
		*functiontype.FunctionType
	}

	templ := `
func {{ .FunctionType.SourceFnName }}(L *lua.LState) int {
	{{ if .FunctionType.Receiver -}}
		r := {{ .Generator.UserDataCheckFn }}(1, L)
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

	execTempl(out, FunctionGenerator{g, &f}, templ)
}

func execTempl(out io.Writer, data any, templ string) {
	t := template.Must(template.New("").Parse(templ))
	err := t.Execute(out, data)

	if err != nil {
		panic(err)
	}
}

func (g *Generator) gatherFunctionsToGenerate() []functiontype.FunctionType {
	if g.structToGenerate != "" {
		return g.gatherConstructors()
	}

	return g.gatherAllFunctions()
}

func (g *Generator) hasFilters() bool {
	return len(g.includeFunctions) > 0 || len(g.excludeFunctions) > 0
}

func (g *Generator) checkInclude(str string) bool {
	if len(g.includeFunctions) > 0 {
		return slices.Contains(g.includeFunctions, str)
	}

	if len(g.excludeFunctions) > 0 {
		return !slices.Contains(g.excludeFunctions, str)
	}

	return true
}

func (g *Generator) gatherConstructors() []functiontype.FunctionType {
	ret := make([]functiontype.FunctionType, 0)
	underylingStructType := g.structObject.Type().Underlying()
	constructorPrefix := "New" + g.structToGenerate

	for _, syn := range g.packageSource.Syntax {
		for _, dec := range syn.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok && fn.Type.Results != nil {
				for _, retType := range fn.Type.Results.List {
					retType := datatype.CreateDataTypeFromExpr(retType.Type, g.packageSource)

					if retType.Type.Underlying() == underylingStructType {
						fnName := fn.Name.Name

						if strings.HasPrefix(fnName, "luaCheck") {
							continue
						}

						if g.hasFilters() {
							if !g.checkInclude(fnName) {
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
						ret = append(ret, functiontype.CreateFunction(fn, false, luaName, sourceCodeName, g.packageSource))
						break
					}
				}
			}
		}
	}

	return ret
}

func (g *Generator) gatherAllFunctions() []functiontype.FunctionType {
	ret := make([]functiontype.FunctionType, 0)

	for _, syn := range g.packageSource.Syntax {
		for _, dec := range syn.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok {
				fnName := fn.Name.Name

				if g.hasFilters() {
					if !g.checkInclude(fnName) {
						continue
					}
				} else if !unicode.IsUpper(rune(fnName[0])) || g.PackageToGenerateFunctionName() == fnName {
					continue
				}

				luaName := gobindluautil.SnakeCase(fnName)
				sourceCodeName := "luaFunction" + fnName
				ret = append(ret, functiontype.CreateFunction(fn, false, luaName, sourceCodeName, g.packageSource))
			}
		}
	}

	return ret
}

func (g *Generator) gatherReceivers() []functiontype.FunctionType {
	if g.structToGenerate == "" {
		return nil
	}

	ret := make([]functiontype.FunctionType, 0)
	underylingStructType := g.structObject.Type().Underlying()

	for _, syn := range g.packageSource.Syntax {
		for _, dec := range syn.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok && fn.Recv != nil {
				fnName := fn.Name.Name

				if fnName == "RegisterLuaType" || fnName == "LuaMetatableType" {
					continue
				}

				if g.hasFilters() {
					if !g.checkInclude(fnName) {
						continue
					}
				} else if !unicode.IsUpper(rune(fnName[0])) {
					continue
				}

				for _, recType := range fn.Recv.List {
					recType := datatype.CreateDataTypeFromExpr(recType.Type, g.packageSource)

					if recType.Type.Underlying() == underylingStructType {
						luaName := gobindluautil.SnakeCase(fnName)
						sourceCodeName := "luaMethod" + g.StructToGenerate() + fnName
						ret = append(ret, functiontype.CreateFunction(fn, true, luaName, sourceCodeName, g.packageSource))
						break
					}
				}
			}
		}
	}

	return ret
}

func (g *Generator) GatherFields() []structfield.StructField {
	ret := make([]structfield.StructField, 0)
	str := g.structObject.Type().Underlying().(*types.Struct)

	for i := 0; i < str.NumFields(); i++ {
		field := str.Field(i)
		tag := str.Tag(i)
		luaName := structfield.GetLuaNameFromTag(field, tag)

		if luaName != "" {
			ret = append(ret, structfield.CreateStructField(field, luaName, g.packageSource))
		}
	}

	return ret
}
