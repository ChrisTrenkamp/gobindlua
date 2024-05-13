package gobindluautil

import (
	"fmt"

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
