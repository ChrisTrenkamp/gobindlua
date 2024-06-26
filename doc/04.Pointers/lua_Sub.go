// Code generated by gobindlua; DO NOT EDIT.
package pointers

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType Sub) RegisterLuaType(L *lua.LState) {
	staticMethodsTable := L.NewTypeMetatable("sub")
	L.SetGlobal("sub", staticMethodsTable)
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorSubNewSub))

	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessSub))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetSub))
}

func luaConstructorSubNewSub(L *lua.LState) int {

	var p0 *string

	{
		ud := string(L.CheckString(1))
		p0 = &ud
	}

	r0 := NewSub(p0)

	L.Push(gobindlua.NewUserData(r0, L))

	return 1
}

func (r *Sub) LuaMetatableType() string {
	return "sub_fields"
}

func luaCheckSub(param int, L *lua.LState) *Sub {
	ud := L.CheckUserData(param)
	if v, ok := ud.Value.(*Sub); ok {
		return v
	}
	L.ArgError(1, "Sub expected")
	return nil
}

func luaAccessSub(L *lua.LState) int {
	p1 := luaCheckSub(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "str":
		L.Push((lua.LString)(*p1.Str))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetSub(L *lua.LState) int {
	p1 := luaCheckSub(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "str":
		ud := string(L.CheckString(3))
		p1.Str = &ud

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}
