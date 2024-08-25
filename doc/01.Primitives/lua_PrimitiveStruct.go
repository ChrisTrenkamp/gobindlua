// Code generated by gobindlua; DO NOT EDIT.
package primitives

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	primitivesubpackage "github.com/ChrisTrenkamp/gobindlua/doc/01.Primitives/primitive_subpackage"
	lua "github.com/yuin/gopher-lua"
)

func (goType *PrimitiveStruct) LuaModuleName() string {
	return "primitive_struct"
}

func (goType *PrimitiveStruct) LuaModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorPrimitiveStructNewPrimitiveStruct))

	L.Push(staticMethodsTable)

	return 1
}

func (goType *PrimitiveStruct) LuaRegisterGlobalMetatable(L *lua.LState) {
	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessPrimitiveStruct))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetPrimitiveStruct))
}

func luaConstructorPrimitiveStructNewPrimitiveStruct(L *lua.LState) int {

	r0 := NewPrimitiveStruct()

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *PrimitiveStruct) LuaMetatableType() string {
	return "primitive_struct_fields"
}

func luaCheckPrimitiveStruct(param int, L *lua.LState) *PrimitiveStruct {
	ud := L.CheckUserData(param)
	if v, ok := ud.Value.(*PrimitiveStruct); ok {
		return v
	}
	L.ArgError(1, "PrimitiveStruct expected")
	return nil
}

func luaAccessPrimitiveStruct(L *lua.LState) int {
	p1 := luaCheckPrimitiveStruct(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "my_bool":
		L.Push((lua.LBool)(p1.MyBool))

	case "my_int":
		L.Push((lua.LNumber)(p1.MyInt))

	case "my_int64":
		L.Push((lua.LNumber)(p1.MyInt64))

	case "my_float":
		L.Push((lua.LNumber)(p1.MyFloat))

	case "my_string":
		L.Push((lua.LString)(p1.SomeString))

	case "my_specialized_int":
		L.Push((lua.LNumber)(p1.MySpecializedInt))

	case "divide_my_int":
		L.Push(L.NewFunction(luaMethodPrimitiveStructDivideMyInt))

	case "set_specialized_int":
		L.Push(L.NewFunction(luaMethodPrimitiveStructSetSpecializedInt))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetPrimitiveStruct(L *lua.LState) int {
	p1 := luaCheckPrimitiveStruct(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "my_bool":
		ud := bool(L.CheckBool(3))
		p1.MyBool = ud

	case "my_int":
		ud := int32(L.CheckNumber(3))
		p1.MyInt = ud

	case "my_int64":
		ud := int64(L.CheckNumber(3))
		p1.MyInt64 = ud

	case "my_float":
		ud := primitivesubpackage.SomeFloat64(L.CheckNumber(3))
		p1.MyFloat = ud

	case "my_string":
		ud := string(L.CheckString(3))
		p1.SomeString = ud

	case "my_specialized_int":
		ud := SpecializedInt(L.CheckNumber(3))
		p1.MySpecializedInt = ud

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}

func luaMethodPrimitiveStructDivideMyInt(L *lua.LState) int {
	r := luaCheckPrimitiveStruct(1, L)

	var p0 float64

	{
		ud := float64(L.CheckNumber(2))
		p0 = ud
	}

	r0, r1 := r.DivideMyInt(p0)

	if r1 != nil {
		L.Error(lua.LString(r1.Error()), 1)
	}

	L.Push((lua.LNumber)(r0))

	return 1
}

func luaMethodPrimitiveStructSetSpecializedInt(L *lua.LState) int {
	r := luaCheckPrimitiveStruct(1, L)

	var p0 SpecializedInt

	{
		ud := SpecializedInt(L.CheckNumber(2))
		p0 = ud
	}

	r.SetSpecializedInt(p0)

	return 0
}
