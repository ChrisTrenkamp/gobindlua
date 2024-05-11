package gobindlua

import lua "github.com/yuin/gopher-lua"

type LuaRegister interface {
	RegisterLuaType(L *lua.LState)
}

func Register(L *lua.LState, r ...LuaRegister) {
	RegisterLuaArray(L)
	RegisterLuaMap(L)

	for _, i := range r {
		i.RegisterLuaType(L)
	}
}

func Funcs(r func(*lua.LState)) LuaRegister {
	return &internalRegistrar{
		reg: r,
	}
}

type internalRegistrar struct {
	reg func(l *lua.LState)
}

func (i *internalRegistrar) RegisterLuaType(L *lua.LState) {
	i.reg(L)
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
