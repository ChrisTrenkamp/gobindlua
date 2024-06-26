package gobindluautil

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"os/exec"
	"slices"
	"strings"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/declaredinterface"
	"github.com/gobeam/stringy"
	"golang.org/x/tools/go/packages"
)

func SnakeCase(str string) string {
	return stringy.New(str).SnakeCase().ToLower()
}

func StructOrInterfaceMetadataName(name string) string {
	return SnakeCase(name)
}

func StructFieldMetadataName(name string) string {
	return SnakeCase(name + "Fields")
}

func LoadSourcePackage(workingDir string, dependantModules []string) (*packages.Package, []declaredinterface.DeclaredInterface, error) {
	config := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedFiles | packages.NeedModule,
		Dir:  workingDir,
	}
	var pack []*packages.Package
	var err error

	pack, err = packages.Load(config, workingDir)

	if err != nil {
		return nil, nil, err
	}

	var ret *packages.Package

	if len(pack) > 1 {
		return nil, nil, fmt.Errorf("packages.Load returned more than one package")
	}

	ret = pack[0]

	if len(dependantModules) == 0 {
		return ret, nil, nil
	}

	expandedModules, err := expandGoModules(dependantModules)

	if err != nil {
		return nil, nil, err
	}

	pack, err = packages.Load(config, expandedModules...)

	if err != nil {
		return nil, nil, err
	}

	allDeclaredInterfaces := make([]declaredinterface.DeclaredInterface, 0)

	for _, i := range pack {
		for _, file := range i.Syntax {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)

				if !ok || len(genDecl.Specs) != 1 || !containsGobindLuaDirective(genDecl) {
					continue
				}

				typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)

				if !ok {
					continue
				}

				ifaceType, ok := typeSpec.Type.(*ast.InterfaceType)

				if !ok {
					continue
				}

				dataType := i.TypesInfo.Types[ifaceType].Type.(*types.Interface)
				allDeclaredInterfaces = append(allDeclaredInterfaces, declaredinterface.DeclaredInterface{
					Name:      typeSpec.Name.Name,
					Interface: dataType,
				})
			}
		}
	}

	return ret, allDeclaredInterfaces, nil
}

func expandGoModules(dependantModules []string) ([]string, error) {
	exp := slices.Clone(dependantModules)
	for i, val := range exp {
		if !strings.HasSuffix(val, "/") {
			val += "/"
		}
		val += "..."
		exp[i] = val
	}

	out := bytes.Buffer{}
	cmd := exec.Command("go", append([]string{"list"}, exp...)...)
	cmd.Stdout = &out
	err := cmd.Run()
	outStr := strings.TrimSpace(out.String())
	return strings.Split(outStr, "\n"), err
}

func containsGobindLuaDirective(genDecl *ast.GenDecl) bool {
	if genDecl.Doc == nil {
		return false
	}

	for _, i := range genDecl.Doc.List {
		if strings.Contains(i.Text, "//go:generate") && strings.Contains(i.Text, "github.com/ChrisTrenkamp/gobindlua/gobindlua") {
			return true
		}
	}

	return false
}

func HasFilters(includeFunctions, excludeFunctions []string) bool {
	return len(includeFunctions) > 0 || len(excludeFunctions) > 0
}

func CheckInclude(str string, includeFunctions, excludeFunctions []string) bool {
	if len(includeFunctions) > 0 {
		return slices.Contains(includeFunctions, str)
	}

	if len(excludeFunctions) > 0 {
		return !slices.Contains(excludeFunctions, str)
	}

	return true
}
