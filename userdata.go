package gobindlua

import lua "github.com/yuin/gopher-lua"

type LuaUserData interface {
	LuaMetatableType() string
}

func NewUserData(data LuaUserData, L *lua.LState) *lua.LUserData {
	return &lua.LUserData{
		Value:     data,
		Env:       L.Env,
		Metatable: L.GetTypeMetatable(data.LuaMetatableType()),
	}
}
