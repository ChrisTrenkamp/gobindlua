package gobindlua

import lua "github.com/yuin/gopher-lua"

type LuaRegistrar interface {
	LuaModuleName() string
	LuaModuleLoader(L *lua.LState) int
	LuaRegisterGlobalMetatable(L *lua.LState)
}

func Register(L *lua.LState, toRegister ...LuaRegistrar) {
	LuaArrayModuleLoader(L)
	L.PreloadModule(ARRAY_MODULES_NAME, LuaArrayModuleLoader)
	LuaArrayRegisterGlobalMetatable(L)
	L.PreloadModule(MAP_MODULES_NAME, LuaMapModuleLoader)
	LuaMapRegisterGlobalMetatable(L)

	for _, i := range toRegister {
		L.PreloadModule(i.LuaModuleName(), i.LuaModuleLoader)
		i.LuaRegisterGlobalMetatable(L)
	}
}

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
