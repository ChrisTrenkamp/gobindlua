// Code generated by gobindlua; DO NOT EDIT.
package interfaces

import (
	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType Lion) RegisterLuaType(L *lua.LState) {
	staticMethodsTable := L.NewTypeMetatable("lion")
	L.SetGlobal("lion", staticMethodsTable)
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorLionNewLion))

	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessLion))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetLion))
}

func luaConstructorLionNewLion(L *lua.LState) int {

	r0 := NewLion()

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *Lion) LuaMetatableType() string {
	return "lion_fields"
}

func luaCheckLion(param int, L *lua.LState) *Lion {
	ud := L.CheckUserData(param)
	if v, ok := ud.Value.(*Lion); ok {
		return v
	}
	L.ArgError(1, "Lion expected")
	return nil
}

func luaAccessLion(L *lua.LState) int {
	p2 := L.CheckString(2)

	switch p2 {
	case "sound":
		L.Push(L.NewFunction(luaMethodLionSound))
	}

	return 1
}

func luaSetLion(L *lua.LState) int {
	p2 := L.CheckString(2)

	switch p2 {
	}

	return 1
}

func luaMethodLionSound(L *lua.LState) int {
	r := luaCheckLion(1, L)

	r0 := r.Sound()

	L.Push((lua.LString)(r0))

	return 1
}
