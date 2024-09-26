package packagegen

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"io"
	"text/template"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/declaredinterface"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gblimports"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type PackageGenerator struct {
	Header            string
	packageToGenerate string
	wd                string
	pathToOutput      string
	dependantModules  []string

	packageSource         *packages.Package
	allDeclaredInterfaces []declaredinterface.DeclaredInterface

	StaticFunctions []datatype.FunctionType
	imports         gblimports.Imports
}

func NewPackageGenerator(packageToGenerate, wd, pathToOutput string, dependantModules []string) *PackageGenerator {
	return &PackageGenerator{
		Header:            gobindluautil.GEN_HEADER,
		packageToGenerate: packageToGenerate,
		wd:                wd,
		pathToOutput:      pathToOutput,
		dependantModules:  dependantModules,
	}
}

func (g *PackageGenerator) GenerateSourceCode() ([]byte, []byte, error) {
	if err := g.loadSourcePackage(); err != nil {
		return nil, nil, fmt.Errorf("failed to load imported go packages: %s", err)
	}

	g.StaticFunctions = g.gatherAllFunctions()
	g.imports = gblimports.NewImports(g.packageSource)

	g.imports.AddPackageFromFunctions(g.StaticFunctions)

	goCode, gerr := g.generateGoCode()
	luaDef, lerr := g.generateLuaDefinitions()
	return goCode, luaDef, errors.Join(gerr, lerr)
}

func (g *PackageGenerator) loadSourcePackage() error {
	var err error
	g.packageSource, g.allDeclaredInterfaces, err = gobindluautil.LoadSourcePackage(g.wd, g.dependantModules)
	return err
}

func (g *PackageGenerator) PackageToGenerateFunctionName() string {
	caser := cases.Title(language.English)
	return caser.String(g.packageToGenerate) + "ModuleLoader"
}

func (g *PackageGenerator) PackageToGenerateMetatableName() string {
	return g.packageToGenerate
}

func (g *PackageGenerator) gatherAllFunctions() []datatype.FunctionType {
	ret := make([]datatype.FunctionType, 0)

	for _, syn := range g.packageSource.Syntax {
		for _, dec := range syn.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok && fn.Recv == nil {
				fnName := fn.Name.Name

				if !gobindluautil.HasGoBindLuaDirective(fn, "function") {
					continue
				}

				luaName := gobindluautil.SnakeCase(fnName)
				sourceCodeName := "luaFunction" + fnName
				ret = append(ret, datatype.CreateFunctionFromExpr(fn, luaName, sourceCodeName, g.packageSource, g.allDeclaredInterfaces))
			}
		}
	}

	return ret
}

func (g *PackageGenerator) generateGoCode() ([]byte, error) {
	code := bytes.Buffer{}

	g.imports.GenerateHeader(&code)
	g.buildMetatableInitFunction(&code)
	g.buildMetatableFunctions(&code)

	originalCodeBytes := code.Bytes()
	formattedCode, err := imports.Process(g.pathToOutput, originalCodeBytes, nil)
	if err != nil {
		return originalCodeBytes, err
	}

	return formattedCode, nil
}

func (g *PackageGenerator) buildMetatableInitFunction(out io.Writer) {
	templ := `
func {{ .PackageToGenerateFunctionName }}(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	{{ range $idx, $fn := .StaticFunctions -}}
		L.SetField(staticMethodsTable, "{{ $fn.LuaFnName }}", L.NewFunction({{ $fn.SourceFnName }}))
	{{ end }}
    L.Push(staticMethodsTable)

	return 1
}
`

	execTempl(out, g, templ)
}

func (g *PackageGenerator) buildMetatableFunctions(out io.Writer) {
	for _, i := range g.StaticFunctions {
		i.GenerateLuaFunctionWrapper(out, "")
	}
}

func (g *PackageGenerator) generateLuaDefinitions() ([]byte, error) {
	ret := bytes.Buffer{}

	g.generateLuaPackageDefinition(&ret)

	return ret.Bytes(), nil
}

func (g *PackageGenerator) generateLuaPackageDefinition(w io.Writer) {
	templ := `---{{ .Header }}
---@meta {{ .PackageToGenerateMetatableName }}
{{- $gen := . }}

local {{ $gen.PackageToGenerateMetatableName }} = {}

{{- range $fidx,$staticFunc := .StaticFunctions -}}
{{ $staticFunc.GenerateLuaFunctionParamRetDefinitions -}}
function {{ $gen.PackageToGenerateMetatableName }}.{{ $staticFunc.LuaFnName }}({{ $staticFunc.GenerateLuaFunctionParamStubs }}) end
{{ end }}
return {{ $gen.PackageToGenerateMetatableName }}
`

	execTempl(w, g, templ)
}

func execTempl(out io.Writer, data any, templ string) {
	t := template.Must(template.New("").Parse(templ))
	err := t.Execute(out, data)

	if err != nil {
		panic(err)
	}
}
