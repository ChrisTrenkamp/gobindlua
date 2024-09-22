// Code generated by gobindlua; DO NOT EDIT.
package human

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType *Human) LuaModuleName() string {
	return "human"
}

func (goType *Human) LuaModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorHumanNewHuman))

	L.Push(staticMethodsTable)

	return 1
}

func (goType *Human) LuaRegisterGlobalMetatable(L *lua.LState) {
	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessHuman))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetHuman))
}

func luaConstructorHumanNewHuman(L *lua.LState) int {

	r0 := NewHuman()

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *Human) LuaMetatableType() string {
	return "human_fields"
}

func luaCheckHuman(param int, L *lua.LState) *Human {
	ud := L.CheckUserData(param)
	v, ok := ud.Value.(*Human)
	if !ok {
		gobindlua.CastArgError(L, 1, "Human", ud.Value)
	}
	return v
}

func luaAccessHuman(L *lua.LState) int {
	p2 := L.CheckString(2)

	switch p2 {
	case "sound":
		L.Push(L.NewFunction(luaMethodHumanSound))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetHuman(L *lua.LState) int {
	p2 := L.CheckString(2)

	switch p2 {

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}

func luaMethodHumanSound(L *lua.LState) int {
	r := luaCheckHuman(1, L)

	r0 := r.Sound()

	L.Push((lua.LString)(r0))

	return 1
}
