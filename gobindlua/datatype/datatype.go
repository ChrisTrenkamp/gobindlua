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
		pointerIndirection := elem.ReferenceOrDereferenceForAssignmentToField()

		return fmt.Sprintf(`gobindlua.NewUserData(&gobindlua.LuaArray{
	Slice: %[1]s,
	Len:   func() int { return len(%[1]s) },
	Index: func(idx%[2]d int) lua.LValue { return %[6]s },
	SetIndex: func(idx%[2]d int, val%[2]d lua.LValue) {
		%[7]s

		%[5]s = %[4]s(%[8]st%[2]d)
	},
}, L)`, variable, level, elem.luaType(), elem.declaredGoType(), indexCode, toLuaType, toGoType, pointerIndirection)
	case *types.Map:
		key := CreateDataTypeFrom(t.Key(), d.packageSource)
		keyIndex := fmt.Sprintf("%[1]s[key%[2]d]", variable, level)
		keyLuaType := key.convertGoTypeToLua(keyIndex, level+1)
		keyGoType := key.convertLuaTypeToGo(fmt.Sprintf("keyVal%d", level), fmt.Sprintf("key%d", level), 3, level+1)
		keyPointerIndirection := key.ReferenceOrDereferenceForAssignmentToField()

		val := CreateDataTypeFrom(t.Elem(), d.packageSource)
		valLuaType := val.convertGoTypeToLua(fmt.Sprintf("ret%d", level), level+1)
		valGoType := val.convertLuaTypeToGo(fmt.Sprintf("valVal%d", level), fmt.Sprintf("val%d", level), 3, level+1)
		valPointerIndirection := val.ReferenceOrDereferenceForAssignmentToField()

		return fmt.Sprintf(`gobindlua.NewUserData(&gobindlua.LuaMap{
	Map: %[1]s,
	Len:   func() int { return len(%[1]s) },
	GetValue: func(key%[2]d lua.LValue) lua.LValue {
		%[6]s
		ret%[2]d := %[1]s[%[3]s(keyVal%[2]d)]
		return %[7]s
	},
	SetValue: func(key%[2]d lua.LValue, val%[2]d lua.LValue) {
		%[6]s
		%[8]s
		%[1]s[%[3]s(%[11]skeyVal%[2]d)] = %[9]s(%[10]svalVal%[2]d)
	},
}, L)`, variable, level, key.declaredGoType(), keyIndex, keyLuaType, keyGoType, valLuaType, valGoType, val.declaredGoType(), valPointerIndirection, keyPointerIndirection)
	}

	return fmt.Sprintf("gobindlua.NewUserData(%s%s, L)", d.referenceOrDereferenceUserDataForAssignment(), variable)
}

func (d *DataType) ReferenceOrDereferenceForAssignmentToField() string {
	goPointerLevel := d.PointerIndirection
	luaPointerLevel := 0

	if _, ok := d.Type.Underlying().(*types.Struct); ok {
		luaPointerLevel++
	}

	if luaPointerLevel < goPointerLevel {
		return "&"
	}

	ret := ""

	for luaPointerLevel > goPointerLevel {
		ret += "*"
		goPointerLevel++
	}

	return ret
}

func (d *DataType) referenceOrDereferenceUserDataForAssignment() string {
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

func (d *DataType) ConvertLuaTypeToGo(variable string, source string, paramNum int) string {
	return d.convertLuaTypeToGo(variable, source, paramNum, 0)
}

func (d *DataType) convertLuaTypeToGo(variableToCreate string, luaVariable string, paramNum, level int) string {
	switch t := d.Type.Underlying().(type) {
	case *types.Basic:
		return d.convertLuaTypeToGoPrimitive(variableToCreate, luaVariable, paramNum, level)
	case *types.Slice:
		return d.convertLuaTypeToGoSlice(t, variableToCreate, luaVariable, paramNum, level)
	case *types.Map:
		return d.convertLuaTypeToGoMap(t, variableToCreate, luaVariable, paramNum, level)
	case *types.Struct:
		return d.convertLuaTypeToStruct(variableToCreate, luaVariable, paramNum, level)
	}

	return "CANNOT_CONVERT_LUA_TYPE_TO_GO"
}

func (d *DataType) convertLuaTypeToGoPrimitive(variableToCreate string, luaVariable string, paramNum, level int) string {
	if level == 0 {
		// Level 0 means the variable came from a L.Check*, which means it was already type-checked
		return fmt.Sprintf(`%s := %s(%s)`, variableToCreate, d.declaredGoType(), luaVariable)
	}

	return fmt.Sprintf(`%[1]s, ok := %[3]s.(%[2]s)
if !ok {
L.ArgError(%[4]d, "argument not a %[5]s instance")
}
`, variableToCreate, d.luaType(), luaVariable, paramNum, d.declaredGoType())
}

func (d *DataType) convertLuaTypeToGoSlice(t *types.Slice, variableToCreate string, luaVariable string, paramNum, level int) string {
	elem := CreateDataTypeFrom(t.Elem(), d.packageSource)
	toGoType := elem.convertLuaTypeToGo(fmt.Sprintf("v%d", level), fmt.Sprintf("val%d", level), paramNum, level+1)
	pointerIndirection := elem.ReferenceOrDereferenceForAssignmentToField()
	return fmt.Sprintf(`%[5]s, err := gobindlua.MapLuaArrayOrTableToGoSlice[%[2]s](%[1]s, func(val%[6]d lua.LValue) %[2]s {
%[7]s
return %[2]s(%[8]sv%[6]d)
})

if err != nil {
L.ArgError(%[4]d, err.Error())
}
`, luaVariable, elem.ActualGoType(), elem.luaType(), paramNum, variableToCreate, level, toGoType, pointerIndirection)
}

func (d *DataType) convertLuaTypeToGoMap(t *types.Map, variableToCreate string, luaVariable string, paramNum, level int) string {
	k := CreateDataTypeFrom(t.Key(), d.packageSource)
	v := CreateDataTypeFrom(t.Elem(), d.packageSource)
	keyGoType := k.convertLuaTypeToGo(fmt.Sprintf("k%d", level), fmt.Sprintf("key%d", level), paramNum, level+1)
	valGoType := v.convertLuaTypeToGo(fmt.Sprintf("v%d", level), fmt.Sprintf("val%d", level), paramNum, level+1)
	keyPointerIndirection := k.ReferenceOrDereferenceForAssignmentToField()
	valPointerIndirection := v.ReferenceOrDereferenceForAssignmentToField()

	return fmt.Sprintf(`%[5]s, err := gobindlua.MapLuaArrayOrTableToGoMap[%[2]s, %[7]s](%[1]s, func(key%[6]d, val%[6]d lua.LValue) (%[2]s, %[7]s) {
%[8]s
%[9]s
return %[2]s(%[10]sk%[6]d), %[7]s(%[11]sv%[6]d)
})

if err != nil {
L.ArgError(%[4]d, err.Error())
}
`, luaVariable, k.ActualGoType(), k.luaType(), paramNum, variableToCreate, level, v.ActualGoType(), keyGoType, valGoType, keyPointerIndirection, valPointerIndirection)
}

func (d *DataType) convertLuaTypeToStruct(variableToCreate string, luaVariable string, paramNum, level int) string {
	if level == 0 {
		return fmt.Sprintf(`%[1]s, ok := %[2]s.Value.(*%[3]s)

if !ok {
	L.ArgError(3, "%[3]s expected")
}
`, variableToCreate, luaVariable, d.declaredGoType())
	}

	return fmt.Sprintf(`%[1]s_ud, ok := %[2]s.(*lua.LUserData)

if !ok {
	L.ArgError(%[4]d, "UserData expected")
}

%[1]s, ok := %[1]s_ud.Value.(*%[3]s)

if !ok {
	L.ArgError(3, "%[3]s expected")
}
`, variableToCreate, luaVariable, d.declaredGoType(), paramNum)
}

func (d *DataType) ConvertLuaTypeToGoSliceEllipses(t *types.Slice, variableToCreate string, luaVariable string, paramNum int) string {
	level := 0
	elem := CreateDataTypeFrom(t.Elem(), d.packageSource)
	toGoType := elem.convertLuaTypeToGo(fmt.Sprintf("v%d", level), fmt.Sprintf("val%d", level), paramNum, level+1)
	pointerIndirection := elem.ReferenceOrDereferenceForAssignmentToField()

	return fmt.Sprintf(`%[5]s, err := gobindlua.MapVariadicArgsToGoSlice[%[2]s](%[4]d, L, func(val%[6]d lua.LValue) %[2]s {
%[7]s
return %[2]s(%[8]sv%[6]d)
})

if err != nil {
L.ArgError(%[4]d, err.Error())
}
`, luaVariable, elem.ActualGoType(), elem.luaType(), paramNum, variableToCreate, level, toGoType, pointerIndirection)
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
	switch t := d.Type.Underlying().(type) {
	case *types.Basic, *types.Slice:
		return d.Type.Underlying().String()
	case *types.Map:
		l := CreateDataTypeFrom(t.Key(), d.packageSource)
		v := CreateDataTypeFrom(t.Elem(), d.packageSource)

		return fmt.Sprintf("map[%s]%s", l.ActualGoType(), v.ActualGoType())
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
	case *types.Map:
		return "*gobindlua.LuaMap"
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
	case *types.Map:
		return "L.CheckAny"
	case *types.Struct:
		return "L.CheckUserData"
	}

	return "UNSUPPORTED_TYPE"
}
