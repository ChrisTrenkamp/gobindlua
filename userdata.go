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

func UnwrapLValueToAny(l lua.LValue) any {
	switch t := l.(type) {
	case *lua.LUserData:
		switch s := t.Value.(type) {
		case *LuaArray:
			return s.Slice
		case *LuaMap:
			return s.Map
		default:
			return s
		}
	default:
		return l
	}
}
