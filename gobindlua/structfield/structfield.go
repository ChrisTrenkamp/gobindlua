package structfield

import (
	"go/types"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"golang.org/x/tools/go/packages"
)

type StructField struct {
	FieldName string
	LuaName   string
	datatype.DataType
}

func CreateStructField(field *types.Var, packageSource *packages.Package) StructField {
	name := field.Name()
	typ := datatype.CreateDataTypeFrom(field.Type(), packageSource)

	return StructField{
		FieldName: name,
		LuaName:   gobindluautil.SnakeCase(name),
		DataType:  typ,
	}
}
