// Code generated by gobindlua; DO NOT EDIT.
package functions

import (
	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func FunctionsModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "print_me", L.NewFunction(luaFunctionPrintMe))
	L.SetField(staticMethodsTable, "split", L.NewFunction(luaFunctionSplit))

	L.Push(staticMethodsTable)

	return 1
}

func luaFunctionPrintMe(L *lua.LState) int {

	var p0 []any

	{

		ud, err := gobindlua.MapVariadicArgsToGoSlice[any](1, L, func(val0 lua.LValue) any {

			v0 := gobindlua.UnwrapLValueToAny(val0)

			return (any)(v0)
		})

		if err != nil {
			L.ArgError(1, err.Error())
		}

		p0 = ud
	}

	PrintMe(p0...)

	return 0
}

func luaFunctionSplit(L *lua.LState) int {

	var p0 string

	var p1 string

	{
		ud := string(L.CheckString(1))
		p0 = ud
	}

	{
		ud := string(L.CheckString(2))
		p1 = ud
	}

	r0 := Split(p0, p1)

	L.Push(gobindlua.NewUserData(&gobindlua.LuaArray{
		Slice: r0,
		Len:   func() int { return len(r0) },
		Index: func(idx0 int) lua.LValue { return (lua.LString)((r0)[idx0]) },
		SetIndex: func(idx0 int, val0 lua.LValue) {

			t0, ok := val0.(lua.LString)

			if !ok {
				L.ArgError(3, "argument not a string instance")
			}

			(r0)[idx0] = (string)(t0)
		},
	}, L))

	return 1
}
