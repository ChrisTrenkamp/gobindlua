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

func UserDataEqual(l, r LuaUserData, L *lua.LState) bool {
	eqFn := L.GetGlobal(l.LuaMetatableType()).(*lua.LTable).RawGet(lua.LString("__eq")).(*lua.LFunction)

	err := L.CallByParam(lua.P{
		Fn:   eqFn,
		NRet: 1,
	}, NewUserData(l, L), NewUserData(r, L))

	if err != nil {
		L.RaiseError("failed to compare objects: %s", err)
	}

	ret := L.Get(-1).(lua.LBool)
	L.Pop(1)

	return bool(ret)
}
