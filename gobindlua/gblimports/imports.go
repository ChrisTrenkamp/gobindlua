package gblimports

import (
	"fmt"
	"io"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/functiontype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"golang.org/x/tools/go/packages"
)

type Imports struct {
	imports       map[string]string
	packageSource *packages.Package
}

func NewImports(packageSource *packages.Package) Imports {
	ret := Imports{
		imports:       make(map[string]string),
		packageSource: packageSource,
	}

	ret.imports["github.com/yuin/gopher-lua"] = "lua"
	ret.imports["github.com/ChrisTrenkamp/gobindlua"] = ""
	return ret
}

func (g Imports) AddPackageFromFunctions(f []functiontype.FunctionType) {
	for _, i := range f {
		for _, p := range i.Params {
			g.AddPackage(p.DataType)
		}

		for _, p := range i.Ret {
			g.AddPackage(p)
		}
	}
}

func (g Imports) AddPackage(d datatype.DataType) {
	if p := d.Package(); p != "" && p != g.packageSource.ID {
		g.imports[p] = ""
	}
}

func (g Imports) GenerateHeader(code io.Writer) {
	fmt.Fprintf(code, "// %s\n", gobindluautil.GEN_HEADER)
	fmt.Fprintf(code, "package %s\n\nimport (\n", g.packageSource.Types.Name())

	for pkg, name := range g.imports {
		fmt.Fprintf(code, "\t%s \"%s\"\n", name, pkg)
	}

	fmt.Fprintf(code, ")\n")
}
