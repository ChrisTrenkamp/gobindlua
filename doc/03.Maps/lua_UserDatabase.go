// Code generated by gobindlua; DO NOT EDIT.
package maps

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

func (goType *UserDatabase) LuaModuleName() string {
	return "user_database"
}

func (goType *UserDatabase) LuaModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "new", L.NewFunction(luaConstructorUserDatabaseNewUserDatabase))
	L.SetField(staticMethodsTable, "new_from", L.NewFunction(luaConstructorUserDatabaseNewUserDatabaseFrom))

	L.Push(staticMethodsTable)

	return 1
}

func (goType *UserDatabase) LuaRegisterGlobalMetatable(L *lua.LState) {
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

		ud, err := gobindlua.MapLuaArrayOrTableToGoMap[int, User](L.CheckAny(1), 0, func(key0, val0 lua.LValue) (int, User) {

			k0, ok := key0.(lua.LNumber)

			if !ok {
				L.ArgError(1, gobindlua.CastArgError("int", key0))
			}

			v0_ud, ok := val0.(*lua.LUserData)

			if !ok {
				L.ArgError(1, gobindlua.TableElementCastError("User", val0, 1))
			}

			v0, ok := v0_ud.Value.(*User)

			if !ok {
				L.ArgError(3, gobindlua.TableElementCastError("User", v0_ud.Value, 1))
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
	v, ok := ud.Value.(*UserDatabase)
	if !ok {
		L.ArgError(1, gobindlua.CastArgError("UserDatabase", ud.Value))
	}
	return v
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
					L.ArgError(3, gobindlua.CastArgError("int", key0))
				}

				ret0 := (p1.Users)[(int)(keyVal0)]
				return gobindlua.NewUserData(&ret0, L)
			},
			SetValue: func(key0 lua.LValue, val0 lua.LValue) {

				keyVal0, ok := key0.(lua.LNumber)

				if !ok {
					L.ArgError(3, gobindlua.CastArgError("int", key0))
				}

				valVal0_ud, ok := val0.(*lua.LUserData)

				if !ok {
					L.ArgError(3, gobindlua.TableElementCastError("User", val0, 1))
				}

				valVal0, ok := valVal0_ud.Value.(*User)

				if !ok {
					L.ArgError(3, gobindlua.TableElementCastError("User", valVal0_ud.Value, 1))
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

		ud, err := gobindlua.MapLuaArrayOrTableToGoMap[int, User](L.CheckAny(3), 0, func(key0, val0 lua.LValue) (int, User) {

			k0, ok := key0.(lua.LNumber)

			if !ok {
				L.ArgError(3, gobindlua.CastArgError("int", key0))
			}

			v0_ud, ok := val0.(*lua.LUserData)

			if !ok {
				L.ArgError(3, gobindlua.TableElementCastError("User", val0, 1))
			}

			v0, ok := v0_ud.Value.(*User)

			if !ok {
				L.ArgError(3, gobindlua.TableElementCastError("User", v0_ud.Value, 1))
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
