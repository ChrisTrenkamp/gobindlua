// Code generated by gobindlua; DO NOT EDIT.
package interfaces

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType *MammalList) RegisterLuaType(L *lua.LState) {
	staticMethodsTable := L.NewTypeMetatable("mammal_list")
	L.SetGlobal("mammal_list", staticMethodsTable)
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorMammalListNewMammalList))

	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessMammalList))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetMammalList))
}

func luaConstructorMammalListNewMammalList(L *lua.LState) int {

	r0 := NewMammalList()

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *MammalList) LuaMetatableType() string {
	return "mammal_list_fields"
}

func luaCheckMammalList(param int, L *lua.LState) *MammalList {
	ud := L.CheckUserData(param)
	if v, ok := ud.Value.(*MammalList); ok {
		return v
	}
	L.ArgError(1, "MammalList expected")
	return nil
}

func luaAccessMammalList(L *lua.LState) int {
	p1 := luaCheckMammalList(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "pet":
		L.Push(gobindlua.NewUserData(p1.Pet, L))

	case "non_pets":
		L.Push(gobindlua.NewUserData(&gobindlua.LuaArray{
			Slice: p1.NonPets,
			Len:   func() int { return len(p1.NonPets) },
			Index: func(idx0 int) lua.LValue { return gobindlua.NewUserData((p1.NonPets)[idx0], L) },
			SetIndex: func(idx0 int, val0 lua.LValue) {

				t0_ud, ok := val0.(*lua.LUserData)

				if !ok {
					L.ArgError(3, "UserData expected")
				}

				t0, ok := t0_ud.Value.(Mammal)

				if !ok {
					L.ArgError(3, "Mammal expected")
				}

				(p1.NonPets)[idx0] = (Mammal)(t0)
			},
		}, L))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetMammalList(L *lua.LState) int {
	p1 := luaCheckMammalList(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "pet":

		ud, ok := L.CheckUserData(3).Value.(Mammal)

		if !ok {
			L.ArgError(3, "Mammal expected")
		}

		p1.Pet = ud

	case "non_pets":

		ud, err := gobindlua.MapLuaArrayOrTableToGoSlice[Mammal](L.CheckAny(3), func(val0 lua.LValue) Mammal {

			v0_ud, ok := val0.(*lua.LUserData)

			if !ok {
				L.ArgError(3, "UserData expected")
			}

			v0, ok := v0_ud.Value.(Mammal)

			if !ok {
				L.ArgError(3, "Mammal expected")
			}

			return (Mammal)(v0)
		})

		if err != nil {
			L.ArgError(3, err.Error())
		}

		p1.NonPets = ud

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}
