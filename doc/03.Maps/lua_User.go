// Code generated by gobindlua; DO NOT EDIT.
package maps

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType User) RegisterLuaType(L *lua.LState) {
	staticMethodsTable := L.NewTypeMetatable("user")
	L.SetGlobal("user", staticMethodsTable)
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorUserNewUser))

	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessUser))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetUser))
}

func luaConstructorUserNewUser(L *lua.LState) int {

	var p0 string

	var p1 int

	var p2 string

	{
		ud := string(L.CheckString(1))
		p0 = ud
	}

	{
		ud := int(L.CheckNumber(2))
		p1 = ud
	}

	{
		ud := string(L.CheckString(3))
		p2 = ud
	}

	r0 := NewUser(p0, p1, p2)

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *User) LuaMetatableType() string {
	return "user_fields"
}

func luaCheckUser(param int, L *lua.LState) *User {
	ud := L.CheckUserData(param)
	if v, ok := ud.Value.(*User); ok {
		return v
	}
	L.ArgError(1, "User expected")
	return nil
}

func luaAccessUser(L *lua.LState) int {
	p1 := luaCheckUser(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "name":
		L.Push((lua.LString)(p1.Name))

	case "age":
		L.Push((lua.LNumber)(p1.Age))

	case "email":
		L.Push((lua.LString)(p1.Email))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetUser(L *lua.LState) int {
	p1 := luaCheckUser(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "name":
		ud := string(L.CheckString(3))
		p1.Name = ud

	case "age":
		ud := int(L.CheckNumber(3))
		p1.Age = ud

	case "email":
		ud := string(L.CheckString(3))
		p1.Email = ud

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}
