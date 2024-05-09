// Code generated by gobindlua; DO NOT EDIT.
package interfaces

import (
	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType Dog) RegisterLuaType(L *lua.LState) {
	staticMethodsTable := L.NewTypeMetatable("dog")
	L.SetGlobal("dog", staticMethodsTable)
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorDogNewDog))

	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessDog))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetDog))
}

func luaConstructorDogNewDog(L *lua.LState) int {

	r0 := NewDog()

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *Dog) LuaMetatableType() string {
	return "dog_fields"
}

func luaCheckDog(param int, L *lua.LState) *Dog {
	ud := L.CheckUserData(param)
	if v, ok := ud.Value.(*Dog); ok {
		return v
	}
	L.ArgError(1, "Dog expected")
	return nil
}

func luaAccessDog(L *lua.LState) int {
	p2 := L.CheckString(2)

	switch p2 {
	case "sound":
		L.Push(L.NewFunction(luaMethodDogSound))
	}

	return 1
}

func luaSetDog(L *lua.LState) int {
	p2 := L.CheckString(2)

	switch p2 {
	}

	return 1
}

func luaMethodDogSound(L *lua.LState) int {
	r := luaCheckDog(1, L)

	r0 := r.Sound()

	L.Push((lua.LString)(r0))

	return 1
}