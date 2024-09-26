// Code generated by gobindlua; DO NOT EDIT.
package functions

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType *FnContainer) LuaModuleName() string {
	return "fn_container"
}

func (goType *FnContainer) LuaModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorFnContainerNewFnContainer))

	L.Push(staticMethodsTable)

	return 1
}

func (goType *FnContainer) LuaRegisterGlobalMetatable(L *lua.LState) {
	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessFnContainer))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetFnContainer))
}

func luaConstructorFnContainerNewFnContainer(L *lua.LState) int {

	var p0 func(string, int) string

	{

		ud_lf, ok := L.CheckAny(1).(*lua.LFunction)

		if !ok {
			gobindlua.CastArgError(L, 1, "func(string, int) string", L.CheckAny(1))
		}

		ud := func(p0 string, p1 int) string {
			L.Push(ud_lf)

			L.Push((lua.LString)(p0))

			L.Push((lua.LNumber)(p1))

			L.Call(2, 1)

			r0l_n, ok := L.Get(-1).(lua.LString)

			if !ok {
				gobindlua.FuncResCastError(L, 1, "string", L.Get(-1))
			}

			r0l := string(r0l_n)

			L.Pop(1)

			return r0l
		}

		p0 = ud
	}

	r0 := NewFnContainer(p0)

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *FnContainer) LuaMetatableType() string {
	return "fn_container_fields"
}

func luaCheckFnContainer(param int, L *lua.LState) *FnContainer {
	ud := L.CheckUserData(param)
	v, ok := ud.Value.(*FnContainer)
	if !ok {
		gobindlua.CastArgError(L, 1, "FnContainer", ud.Value)
	}
	return v
}

func luaAccessFnContainer(L *lua.LState) int {
	recv := luaCheckFnContainer(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "fn":
		L.Push(L.NewFunction(func(L *lua.LState) int {

			var p0 string

			var p1 int

			{
				ud := string(L.CheckString(1))
				p0 = ud
			}

			{
				ud := int(L.CheckNumber(2))
				p1 = ud
			}

			r0 := recv.Fn(p0, p1)

			L.Push((lua.LString)(r0))

			return 1
		}))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetFnContainer(L *lua.LState) int {
	recv := luaCheckFnContainer(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "fn":

		ud_lf, ok := L.CheckAny(3).(*lua.LFunction)

		if !ok {
			gobindlua.CastArgError(L, 3, "func(string, int) string", L.CheckAny(3))
		}

		ud := func(p0 string, p1 int) string {
			L.Push(ud_lf)

			L.Push((lua.LString)(p0))

			L.Push((lua.LNumber)(p1))

			L.Call(2, 1)

			r0l_n, ok := L.Get(-1).(lua.LString)

			if !ok {
				gobindlua.FuncResCastError(L, 1, "string", L.Get(-1))
			}

			r0l := string(r0l_n)

			L.Pop(1)

			return r0l
		}

		recv.Fn = ud

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}