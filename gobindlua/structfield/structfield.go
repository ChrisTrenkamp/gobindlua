package structfield

import (
	"go/types"
	"regexp"
	"strings"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/declaredinterface"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"golang.org/x/tools/go/packages"
)

type StructField struct {
	FieldName string
	LuaName   string
	datatype.DataType
}

func CreateStructField(field *types.Var, luaName string, packageSource *packages.Package, allDeclaredInterfaces []declaredinterface.DeclaredInterface) StructField {
	name := field.Name()
	typ := datatype.CreateDataTypeFrom(field.Type(), packageSource, allDeclaredInterfaces)

	return StructField{
		FieldName: name,
		LuaName:   luaName,
		DataType:  typ,
	}
}

var TAG_REGEX = regexp.MustCompile(`gobindlua:"([^"]+)"`)

func GetLuaNameFromTag(field *types.Var, tag string) string {
	res := TAG_REGEX.FindAllStringSubmatch(tag, -1)

	if len(res) != 1 {
		return fieldName(field)
	}

	sub := res[0]

	if len(sub) != 2 {
		return fieldName(field)
	}

	tagName := strings.TrimSpace(sub[1])

	if tagName == "-" {
		return ""
	}

	return tagName
}

func fieldName(field *types.Var) string {
	if field.Exported() {
		return gobindluautil.SnakeCase(field.Name())
	}

	return ""
}
