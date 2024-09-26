// Code generated by gobindlua; DO NOT EDIT.
package lion

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType *Lion) LuaModuleName() string {
	return "Lion"
}

func (goType *Lion) LuaModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "NewLion", L.NewFunction(luaConstructorLionNewLion))

	L.Push(staticMethodsTable)

	return 1
}

func (goType *Lion) LuaRegisterGlobalMetatable(L *lua.LState) {
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
	return "LionTable"
}

func luaCheckLion(param int, L *lua.LState) *Lion {
	ud := L.CheckUserData(param)
	v, ok := ud.Value.(*Lion)
	if !ok {
		gobindlua.CastArgError(L, 1, "Lion", ud.Value)
	}
	return v
}

func luaAccessLion(L *lua.LState) int {
	p2 := L.CheckString(2)

	switch p2 {
	case "Sound":
		L.Push(L.NewFunction(luaMethodLionSound))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetLion(L *lua.LState) int {
	p2 := L.CheckString(2)

	switch p2 {

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}

func luaMethodLionSound(L *lua.LState) int {
	r := luaCheckLion(1, L)

	r0 := r.Sound()

	L.Push((lua.LString)(r0))

	return 1
}
