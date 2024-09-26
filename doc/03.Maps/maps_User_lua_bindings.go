// Code generated by gobindlua; DO NOT EDIT.
package maps

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType *User) LuaModuleName() string {
	return "user"
}

func (goType *User) LuaModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorUserNewUser))

	L.Push(staticMethodsTable)

	return 1
}

func (goType *User) LuaRegisterGlobalMetatable(L *lua.LState) {
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
	v, ok := ud.Value.(*User)
	if !ok {
		gobindlua.CastArgError(L, 1, "User", ud.Value)
	}
	return v
}

func luaAccessUser(L *lua.LState) int {
	recv := luaCheckUser(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "name":
		L.Push((lua.LString)(recv.Name))

	case "age":
		L.Push((lua.LNumber)(recv.Age))

	case "email":
		L.Push((lua.LString)(recv.Email))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetUser(L *lua.LState) int {
	recv := luaCheckUser(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "name":
		ud := string(L.CheckString(3))
		recv.Name = ud

	case "age":
		ud := int(L.CheckNumber(3))
		recv.Age = ud

	case "email":
		ud := string(L.CheckString(3))
		recv.Email = ud

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}