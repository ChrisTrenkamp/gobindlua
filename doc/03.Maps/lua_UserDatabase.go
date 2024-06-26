// Code generated by gobindlua; DO NOT EDIT.
package maps

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType UserDatabase) RegisterLuaType(L *lua.LState) {
	staticMethodsTable := L.NewTypeMetatable("user_database")
	L.SetGlobal("user_database", staticMethodsTable)
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorUserDatabaseNewUserDatabase))
	L.SetField(staticMethodsTable, "new_from", L.NewFunction(luaConstructorUserDatabaseNewUserDatabaseFrom))

	fieldsTable := L.NewTypeMetatable(goType.LuaMetatableType())
	L.SetGlobal(goType.LuaMetatableType(), fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(luaAccessUserDatabase))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(luaSetUserDatabase))
}

func luaConstructorUserDatabaseNewUserDatabase(L *lua.LState) int {

	r0 := NewUserDatabase()

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func luaConstructorUserDatabaseNewUserDatabaseFrom(L *lua.LState) int {

	var p0 map[int]User

	{

		ud, err := gobindlua.MapLuaArrayOrTableToGoMap[int, User](L.CheckAny(1), func(key0, val0 lua.LValue) (int, User) {

			k0, ok := key0.(lua.LNumber)

			if !ok {
				L.ArgError(1, "argument not a int instance")
			}

			v0_ud, ok := val0.(*lua.LUserData)

			if !ok {
				L.ArgError(1, "UserData expected")
			}

			v0, ok := v0_ud.Value.(*User)

			if !ok {
				L.ArgError(3, "User expected")
			}

			return (int)(k0), (User)(*v0)
		})

		if err != nil {
			L.ArgError(1, err.Error())
		}

		p0 = ud
	}

	r0 := NewUserDatabaseFrom(p0)

	L.Push(gobindlua.NewUserData(&r0, L))

	return 1
}

func (r *UserDatabase) LuaMetatableType() string {
	return "user_database_fields"
}

func luaCheckUserDatabase(param int, L *lua.LState) *UserDatabase {
	ud := L.CheckUserData(param)
	if v, ok := ud.Value.(*UserDatabase); ok {
		return v
	}
	L.ArgError(1, "UserDatabase expected")
	return nil
}

func luaAccessUserDatabase(L *lua.LState) int {
	p1 := luaCheckUserDatabase(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "users":
		L.Push(gobindlua.NewUserData(&gobindlua.LuaMap{
			Map: p1.Users,
			Len: func() int { return len(p1.Users) },
			GetValue: func(key0 lua.LValue) lua.LValue {

				keyVal0, ok := key0.(lua.LNumber)

				if !ok {
					L.ArgError(3, "argument not a int instance")
				}

				ret0 := (p1.Users)[(int)(keyVal0)]
				return gobindlua.NewUserData(&ret0, L)
			},
			SetValue: func(key0 lua.LValue, val0 lua.LValue) {

				keyVal0, ok := key0.(lua.LNumber)

				if !ok {
					L.ArgError(3, "argument not a int instance")
				}

				valVal0_ud, ok := val0.(*lua.LUserData)

				if !ok {
					L.ArgError(3, "UserData expected")
				}

				valVal0, ok := valVal0_ud.Value.(*User)

				if !ok {
					L.ArgError(3, "User expected")
				}

				(p1.Users)[(int)(keyVal0)] = (User)(*valVal0)
			},
			ForEach: func(f0 func(k0, v0 lua.LValue)) {
				for k0_iter, v0_iter := range p1.Users {
					retKey0 := k0_iter
					ret0 := v0_iter
					key0 := (lua.LNumber)(retKey0)
					val0 := gobindlua.NewUserData(&ret0, L)
					f0(key0, val0)
				}
			},
		}, L))

	default:
		L.Push(lua.LNil)
	}

	return 1
}

func luaSetUserDatabase(L *lua.LState) int {
	p1 := luaCheckUserDatabase(1, L)
	p2 := L.CheckString(2)

	switch p2 {
	case "users":

		ud, err := gobindlua.MapLuaArrayOrTableToGoMap[int, User](L.CheckAny(3), func(key0, val0 lua.LValue) (int, User) {

			k0, ok := key0.(lua.LNumber)

			if !ok {
				L.ArgError(3, "argument not a int instance")
			}

			v0_ud, ok := val0.(*lua.LUserData)

			if !ok {
				L.ArgError(3, "UserData expected")
			}

			v0, ok := v0_ud.Value.(*User)

			if !ok {
				L.ArgError(3, "User expected")
			}

			return (int)(k0), (User)(*v0)
		})

		if err != nil {
			L.ArgError(3, err.Error())
		}

		p1.Users = ud

	default:
		L.ArgError(2, fmt.Sprintf("unknown field %s", p2))
	}

	return 0
}
