// Code generated by gobindlua; DO NOT EDIT.
package slices

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType Vector) RegisterLuaType(L *lua.LState) {
	staticMethodsTable := L.NewTypeMetatable("vector")
	L.SetGlobal("vector", staticMethodsTable)
	L.SetField(staticMethodsTable, "new_from", L.NewFunction(luaConstructorVectorNewVectorFrom))
	L.SetField(staticMethodsTable, "new_variadic", L.NewFunction(luaConstructorVectorNewVectorVariadic))

	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessVector))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetVector))
}

func luaConstructorVectorNewVectorFrom(L *lua.LState) int {

	var p0 []float64

	{

		ud, err := gobindlua.MapLuaArrayOrTableToGoSlice[float64](L.CheckAny(1), func(val0 lua.LValue) float64 {

			v0, ok := val0.(lua.LNumber)

			if !ok {
				L.ArgError(1, "argument not a float64 instance")
			}

			return (float64)(v0)
		})

		if err != nil {
			L.ArgError(1, err.Error())
		}

		p0 = ud
	}

	r0 := NewVectorFrom(p0)

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func luaConstructorVectorNewVectorVariadic(L *lua.LState) int {

	var p0 []float64

	{

		ud, err := gobindlua.MapVariadicArgsToGoSlice[float64](1, L, func(val0 lua.LValue) float64 {

			v0, ok := val0.(lua.LNumber)

			if !ok {
				L.ArgError(1, "argument not a float64 instance")
			}

			return (float64)(v0)
		})

		if err != nil {
			L.ArgError(1, err.Error())
		}

		p0 = ud
	}

	r0 := NewVectorVariadic(p0...)

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *Vector) LuaMetatableType() string {
	return "vector_fields"
}

func luaCheckVector(param int, L *lua.LState) *Vector {
	ud := L.CheckUserData(param)
	if v, ok := ud.Value.(*Vector); ok {
		return v
	}
	L.ArgError(1, "Vector expected")
	return nil
}

func luaAccessVector(L *lua.LState) int {
	p1 := luaCheckVector(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "elements":
		L.Push(gobindlua.NewUserData(&gobindlua.LuaArray{
			Slice: p1.Elements,
			Len:   func() int { return len(p1.Elements) },
			Index: func(idx0 int) lua.LValue { return (lua.LNumber)((p1.Elements)[idx0]) },
			SetIndex: func(idx0 int, val0 lua.LValue) {

				t0, ok := val0.(lua.LNumber)

				if !ok {
					L.ArgError(3, "argument not a float64 instance")
				}

				(p1.Elements)[idx0] = (float64)(t0)
			},
		}, L))

	case "inner_product":
		L.Push(L.NewFunction(luaMethodVectorInnerProduct))

	case "outer_product":
		L.Push(L.NewFunction(luaMethodVectorOuterProduct))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetVector(L *lua.LState) int {
	p1 := luaCheckVector(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "elements":

		ud, err := gobindlua.MapLuaArrayOrTableToGoSlice[float64](L.CheckAny(3), func(val0 lua.LValue) float64 {

			v0, ok := val0.(lua.LNumber)

			if !ok {
				L.ArgError(3, "argument not a float64 instance")
			}

			return (float64)(v0)
		})

		if err != nil {
			L.ArgError(3, err.Error())
		}

		p1.Elements = ud

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}

func luaMethodVectorInnerProduct(L *lua.LState) int {
	r := luaCheckVector(1, L)

	var p0 Vector

	{

		ud, ok := L.CheckUserData(2).Value.(*Vector)

		if !ok {
			L.ArgError(3, "Vector expected")
		}

		p0 = *ud
	}

	r0, r1 := r.InnerProduct(p0)

	if r1 != nil {
		L.Error(lua.LString(r1.Error()), 1)
	}

	L.Push((lua.LNumber)(r0))

	return 1
}

func luaMethodVectorOuterProduct(L *lua.LState) int {
	r := luaCheckVector(1, L)

	var p0 Vector

	{

		ud, ok := L.CheckUserData(2).Value.(*Vector)

		if !ok {
			L.ArgError(3, "Vector expected")
		}

		p0 = *ud
	}

	r0, r1 := r.OuterProduct(p0)

	if r1 != nil {
		L.Error(lua.LString(r1.Error()), 1)
	}

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}
