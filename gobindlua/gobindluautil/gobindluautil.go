package gobindluautil

import (
	"fmt"
	"slices"

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

func LoadSourcePackage(workingDir string) (*packages.Package, error) {
	config := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  workingDir,
	}
	var pack []*packages.Package
	var err error

	pack, err = packages.Load(config, workingDir)

	if err != nil {
		return nil, err
	}

	if len(pack) == 1 {
		return pack[0], nil
	}

	return nil, fmt.Errorf("packages.Load returned more than one package")
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
