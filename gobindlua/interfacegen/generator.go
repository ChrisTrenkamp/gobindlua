package interfacegen

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"strconv"
	"text/template"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/declaredinterface"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/functiontype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/param"
	"golang.org/x/tools/go/packages"
)

type InterfaceGenerator struct {
	Header              string
	interfaceToGenerate string
	workingDir          string
	dependantModules    []string

	packageSource         *packages.Package
	allDeclaredInterfaces []declaredinterface.DeclaredInterface
	interfaceObject       types.Object

	InterfaceMethods []functiontype.FunctionType
}

func NewInterfaceGenerator(interfaceToGenerate, workingDir string, dependantModules []string) InterfaceGenerator {
	return InterfaceGenerator{
		Header:              gobindluautil.GEN_HEADER,
		interfaceToGenerate: interfaceToGenerate,
		workingDir:          workingDir,
		dependantModules:    dependantModules,
	}
}

func (g *InterfaceGenerator) GenerateSourceCode() ([]byte, error) {
	if err := g.loadSourcePackage(); err != nil {
		return nil, fmt.Errorf("failed to load imported go packages: %s", err)
	}

	luaBytes := bytes.Buffer{}
	g.InterfaceMethods = g.gatherInterfaceMethods()
	g.generateLuaPackageDefinition(&luaBytes)

	return luaBytes.Bytes(), nil
}

func (g *InterfaceGenerator) loadSourcePackage() error {
	var err error
	g.packageSource, g.allDeclaredInterfaces, err = gobindluautil.LoadSourcePackage(g.workingDir, g.dependantModules)

	if err != nil {
		return err
	}

	g.interfaceObject = g.packageSource.Types.Scope().Lookup(g.interfaceToGenerate)

	if g.interfaceObject == nil {
		return fmt.Errorf("specified type not found")
	}

	if _, ok := g.interfaceObject.Type().Underlying().(*types.Interface); !ok {
		return fmt.Errorf("specified type is not a struct")
	}

	return nil
}

func (g *InterfaceGenerator) gatherInterfaceMethods() []functiontype.FunctionType {
	ret := make([]functiontype.FunctionType, 0)
	und := g.interfaceObject.Type().Underlying().(*types.Interface)

	for i := 0; i < und.NumMethods(); i++ {
		fn := und.Method(i)
		f := g.createFuncDecl(fn)

		if f.ActualFnName == "LuaMetatableType" {
			continue
		}

		ret = append(ret, f)
	}

	return ret
}

func (g *InterfaceGenerator) createFuncDecl(fn *types.Func) functiontype.FunctionType {
	name := fn.Name()
	luaName := gobindluautil.SnakeCase(name)
	var params []param.Param
	var ret []datatype.DataType
	typ := fn.Type().(*types.Signature)

	for i := 0; i < typ.Params().Len(); i++ {
		p := typ.Params().At(i)
		typ := datatype.CreateDataTypeFrom(p.Type(), g.packageSource, g.allDeclaredInterfaces)
		luaName := gobindluautil.SnakeCase(p.Name())

		if luaName == "" {
			luaName = "_" + strconv.Itoa(i)
		}

		// TODO: Ellipses are simply reported as slices.  While this is technically correct, it would be nice
		// to properly detect ellipses.
		params = append(params, param.Param{
			IsEllipses: false,
			ParamNum:   i,
			LuaName:    luaName,
			DataType:   typ,
		})
	}

	for i := 0; i < typ.Results().Len(); i++ {
		r := typ.Results().At(i)
		ret = append(ret, datatype.CreateDataTypeFrom(r.Type(), g.packageSource, g.allDeclaredInterfaces))
	}

	return functiontype.FunctionType{
		ActualFnName: name,
		LuaFnName:    luaName,
		SourceFnName: "",
		Receiver:     true,
		Params:       params,
		Ret:          ret,
	}
}

func (g *InterfaceGenerator) InterfaceToGenerate() string {
	return gobindluautil.SnakeCase(g.interfaceToGenerate)
}

func (g *InterfaceGenerator) generateLuaPackageDefinition(w io.Writer) {
	templ := `---{{ .Header }}
---@meta {{ .InterfaceToGenerate }}
{{- $gen := . }}

---@class {{ .InterfaceToGenerate }}
local {{ $gen.InterfaceToGenerate }} = {}
{{ range $fidx,$method := .InterfaceMethods -}}
{{ $method.GenerateLuaFunctionParamRetDefinitions -}}
function {{ $gen.InterfaceToGenerate }}.{{ $method.LuaFnName }}({{ $method.GenerateLuaFunctionParamStubs }}) end
{{ end }}
return {{ $gen.InterfaceToGenerate }}
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
