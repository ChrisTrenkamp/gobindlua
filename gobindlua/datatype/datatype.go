package datatype

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type DataType struct {
	Type               types.Type
	PointerIndirection int
	packageSource      *packages.Package
}

func CreateDataTypeFromExpr(expr ast.Expr, packageSource *packages.Package) DataType {
	return CreateDataTypeFrom(packageSource.TypesInfo.Types[expr].Type, packageSource)
}

func CreateDataTypeFrom(t types.Type, packageSource *packages.Package) DataType {
	pointerIndirection := 0

	for {
		if p, ok := t.(*types.Pointer); ok {
			t = p.Elem()
			pointerIndirection++
		} else {
			break
		}
	}

	return DataType{
		Type:               t,
		PointerIndirection: pointerIndirection,
		packageSource:      packageSource,
	}
}

func (d *DataType) ConvertGoTypeToLua(variable string) string {
	return d.convertGoTypeToLua(variable, 0)
}

func (d *DataType) convertGoTypeToLua(variable string, level int) string {
	switch t := d.Type.Underlying().(type) {
	case *types.Basic:
		return fmt.Sprintf(`%s(%s)`, d.luaType(), variable)
	case *types.Slice:
		elem := CreateDataTypeFrom(t.Elem(), d.packageSource)
		indexCode := fmt.Sprintf("%[1]s[idx%[2]d]", variable, level)
		toLuaType := elem.convertGoTypeToLua(indexCode, level+1)
		toGoType := elem.convertLuaTypeToGo(fmt.Sprintf("t%d", level), fmt.Sprintf("val%d", level), 3, level+1)
		return fmt.Sprintf(`gobindlua.NewUserData(&gobindlua.LuaArray{
	Slice: %[1]s,
	Len:   func() int { return len(%[1]s) },
	Index: func(idx%[2]d int) lua.LValue { return %[6]s },
	SetIndex: func(idx%[2]d int, val%[2]d lua.LValue) {
		%[7]s

		%[5]s = %[4]s(t%[2]d)
	},
}, L)`, variable, level, elem.luaType(), elem.declaredGoType(), indexCode, toLuaType, toGoType)
	}

	return fmt.Sprintf("gobindlua.NewUserData(%s%s, L)", d.toSingleLevelPointer(), variable)
}

func (d *DataType) toSingleLevelPointer() string {
	if d.PointerIndirection < 0 {
		panic("Pointer indirection is less than 0 (how is this even possible?)")
	}

	if d.PointerIndirection == 0 {
		return "&"
	}

	ret := ""

	for i := d.PointerIndirection - 1; i > 0; i-- {
		ret += "*"
	}

	return ret
}

func (d *DataType) ReferenceOrDereferenceForAssignment() string {
	pointers := d.PointerIndirection

	switch d.Type.Underlying().(type) {
	case *types.Struct:
		pointers++
	}

	ret := ""

	for i := 0; i < pointers; i++ {
		ret += "*"
	}

	return ret
}

func (d *DataType) ConvertLuaTypeToGo(variable string, source string, paramNum int) string {
	return d.convertLuaTypeToGo(variable, source, paramNum, 0)
}

func (d *DataType) convertLuaTypeToGo(variableToCreate string, luaVariable string, paramNum, level int) string {
	switch t := d.Type.Underlying().(type) {
	case *types.Basic:
		if level == 0 {
			// Level 0 means the variable came from a L.Check*, which means it was already type-checked
			return fmt.Sprintf(`%s := %s(%s)`, variableToCreate, d.declaredGoType(), luaVariable)
		} else {
			return fmt.Sprintf(`%[1]s, ok := %[3]s.(%[2]s)
if !ok {
	L.ArgError(%[4]d, "argument not a %[5]s instance")
}
`, variableToCreate, d.luaType(), luaVariable, paramNum, d.declaredGoType())
		}
	case *types.Slice:
		return d.convertLuaTypeToGoSlice(t, variableToCreate, luaVariable, paramNum, level)
	case *types.Struct:
		return fmt.Sprintf(`%[1]s, ok := %[2]s.Value.(*%[3]s)

if !ok {
	L.ArgError(3, "%[3]s expected")
}
`, variableToCreate, luaVariable, d.declaredGoType())
	}

	return "CANNOT_CONVERT_LUA_TYPE_TO_GO"
}

func (d *DataType) convertLuaTypeToGoSlice(t *types.Slice, variableToCreate string, luaVariable string, paramNum, level int) string {
	elem := CreateDataTypeFrom(t.Elem(), d.packageSource)
	toGoType := elem.convertLuaTypeToGo(fmt.Sprintf("v%d", level), fmt.Sprintf("val%d", level), paramNum, level+1)
	return fmt.Sprintf(`%[5]s, err := gobindlua.MapLuaArrayOrTableToGoSlice[%[2]s](%[1]s, func(val%[6]d lua.LValue) %[2]s {
%[7]s
return %[2]s(v%[6]d)
})

if err != nil {
L.ArgError(%[4]d, err.Error())
}
`, luaVariable, elem.ActualGoType(), elem.luaType(), paramNum, variableToCreate, level, toGoType)
}

func (d *DataType) ConvertLuaTypeToGoSliceEllipses(t *types.Slice, variableToCreate string, luaVariable string, paramNum int) string {
	level := 0
	elem := CreateDataTypeFrom(t.Elem(), d.packageSource)
	toGoType := elem.convertLuaTypeToGo(fmt.Sprintf("v%d", level), fmt.Sprintf("val%d", level), paramNum, level+1)
	return fmt.Sprintf(`%[5]s, err := gobindlua.MapVariadicArgsToGoSlice[%[2]s](%[4]d, L, func(val%[6]d lua.LValue) %[2]s {
%[7]s
return %[2]s(v%[6]d)
})

if err != nil {
L.ArgError(%[4]d, err.Error())
}
`, luaVariable, elem.ActualGoType(), elem.luaType(), paramNum, variableToCreate, level, toGoType)
}

func (d *DataType) declaredGoType() string {
	if n, ok := d.Type.(*types.Named); ok && !d.IsError() {
		pkgName := n.Obj().Pkg().Name()
		name := n.Obj().Name()

		if n.Obj().Pkg().Path() == d.packageSource.ID {
			return name
		}

		return pkgName + "." + name
	}

	return d.Type.String()
}

func (d *DataType) Package() string {
	if n, ok := d.Type.(*types.Named); ok && !d.IsError() {
		pkg := n.Obj().Pkg()

		if pkg != nil {
			return pkg.Path()
		}
	}

	return ""
}

func (d *DataType) ActualGoType() string {
	switch d.Type.Underlying().(type) {
	case *types.Basic, *types.Slice:
		return d.Type.Underlying().String()
	}

	return d.declaredGoType()
}

func (d *DataType) IsError() bool {
	return d.Type.String() == "error"
}

func (d *DataType) luaType() string {
	switch d.Type.Underlying().(type) {
	case *types.Basic:
		switch d.ActualGoType() {
		case "bool":
			return "lua.LBool"
		case "string":
			return "lua.LString"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "byte", "uint16", "uint32", "uint64", "float32", "float64":
			return "lua.LNumber"
		}
	case *types.Slice:
		return "*gobindlua.LuaArray"
	case *types.Struct:
		return "lua.LUserData"
	}

	return "UNSUPPORTED_TYPE"
}

func (d *DataType) LuaParamType() string {
	switch d.Type.Underlying().(type) {
	case *types.Basic:
		switch d.ActualGoType() {
		case "bool":
			return "L.CheckBool"
		case "string":
			return "L.CheckString"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "byte", "uint16", "uint32", "uint64", "float32", "float64":
			return "L.CheckNumber"
		}
	case *types.Slice:
		return "L.CheckAny"
	case *types.Struct:
		return "L.CheckUserData"
	}

	return "UNSUPPORTED_TYPE"
}
