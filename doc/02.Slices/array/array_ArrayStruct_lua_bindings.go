// Code generated by gobindlua; DO NOT EDIT.
package array

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	slicessubpackage "github.com/ChrisTrenkamp/gobindlua/doc/02.Slices/slices_subpackage"
	lua "github.com/yuin/gopher-lua"
)

func (goType *ArrayStruct) LuaModuleName() string {
	return "array_struct"
}

func (goType *ArrayStruct) LuaModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorArrayStructNewArrayStruct))

	L.Push(staticMethodsTable)

	return 1
}

func (goType *ArrayStruct) LuaRegisterGlobalMetatable(L *lua.LState) {
	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessArrayStruct))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetArrayStruct))
}

func luaConstructorArrayStructNewArrayStruct(L *lua.LState) int {

	var p0 [3]float32

	{

		udsl, err := gobindlua.MapLuaArrayOrTableToGoSlice[float32](L.CheckAny(1), 0, func(val0 lua.LValue) float32 {

			v0_n, ok := val0.(lua.LNumber)

			if !ok {
				gobindlua.TableElemCastError(L, 1, "float32", val0)
			}

			v0 := float32(v0_n)

			return v0
		})

		if err != nil {
			L.ArgError(1, err.Error())
		}

		ud := (*[3]float32)(udsl)

		p0 = *ud
	}

	r0 := NewArrayStruct(p0)

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *ArrayStruct) LuaMetatableType() string {
	return "array_struct_fields"
}

func luaCheckArrayStruct(param int, L *lua.LState) *ArrayStruct {
	ud := L.CheckUserData(param)
	v, ok := ud.Value.(*ArrayStruct)
	if !ok {
		gobindlua.CastArgError(L, 1, "ArrayStruct", ud.Value)
	}
	return v
}

func luaAccessArrayStruct(L *lua.LState) int {
	p1 := luaCheckArrayStruct(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "elements":
		L.Push(gobindlua.NewUserData(&gobindlua.LuaArray{
			Slice: p1.Elements,
			Len:   func() int { return len(p1.Elements) },
			Index: func(idx0 int) lua.LValue { return (lua.LNumber)((p1.Elements)[idx0]) },
			SetIndex: func(idx0 int, val0 lua.LValue) {

				t0_n, ok := val0.(lua.LNumber)

				if !ok {
					gobindlua.TableElemCastError(L, 1, "float32", val0)
				}

				t0 := float32(t0_n)

				(p1.Elements)[idx0] = t0
			},
		}, L))

	case "set_elements":
		L.Push(L.NewFunction(luaMethodArrayStructSetElements))

	case "set_elements_from_subpackage":
		L.Push(L.NewFunction(luaMethodArrayStructSetElementsFromSubpackage))

	case "string":
		L.Push(L.NewFunction(luaMethodArrayStructString))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetArrayStruct(L *lua.LState) int {
	p1 := luaCheckArrayStruct(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "elements":

		udsl, err := gobindlua.MapLuaArrayOrTableToGoSlice[float32](L.CheckAny(3), 0, func(val0 lua.LValue) float32 {

			v0_n, ok := val0.(lua.LNumber)

			if !ok {
				gobindlua.TableElemCastError(L, 1, "float32", val0)
			}

			v0 := float32(v0_n)

			return v0
		})

		if err != nil {
			L.ArgError(3, err.Error())
		}

		ud := (*[3]float32)(udsl)

		p1.Elements = *ud

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}

func luaMethodArrayStructSetElements(L *lua.LState) int {
	r := luaCheckArrayStruct(1, L)

	var p0 [3]float32

	{

		udsl, err := gobindlua.MapLuaArrayOrTableToGoSlice[float32](L.CheckAny(2), 0, func(val0 lua.LValue) float32 {

			v0_n, ok := val0.(lua.LNumber)

			if !ok {
				gobindlua.TableElemCastError(L, 1, "float32", val0)
			}

			v0 := float32(v0_n)

			return v0
		})

		if err != nil {
			L.ArgError(2, err.Error())
		}

		ud := (*[3]float32)(udsl)

		p0 = *ud
	}

	r.SetElements(p0)

	return 0
}

func luaMethodArrayStructSetElementsFromSubpackage(L *lua.LState) int {
	r := luaCheckArrayStruct(1, L)

	var p0 *slicessubpackage.AnArray

	{

		udsl, err := gobindlua.MapLuaArrayOrTableToGoSlice[float32](L.CheckAny(2), 0, func(val0 lua.LValue) float32 {

			v0_n, ok := val0.(lua.LNumber)

			if !ok {
				gobindlua.TableElemCastError(L, 1, "float32", val0)
			}

			v0 := float32(v0_n)

			return v0
		})

		if err != nil {
			L.ArgError(2, err.Error())
		}

		ud := (*slicessubpackage.AnArray)(udsl)

		p0 = ud
	}

	r.SetElementsFromSubpackage(p0)

	return 0
}

func luaMethodArrayStructString(L *lua.LState) int {
	r := luaCheckArrayStruct(1, L)

	r0 := r.String()

	L.Push((lua.LString)(r0))

	return 1
}
