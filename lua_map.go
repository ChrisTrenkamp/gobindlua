package gobindlua

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

const MAP_METATABLE_NAME = "gobindluamap"

func RegisterLuaMap(L *lua.LState) {
	if L.GetGlobal(MAP_METATABLE_NAME) != lua.LNil {
		return
	}

	mt := L.NewTypeMetatable(MAP_METATABLE_NAME)
	L.SetGlobal(MAP_METATABLE_NAME, mt)

	mt.RawSetString("__index", L.NewFunction(mapIndex))
	mt.RawSetString("__newindex", L.NewFunction(mapNewIndex))
	mt.RawSetString("__len", L.NewFunction(mapLen))
}

type LuaMap struct {
	Map      interface{}
	Len      func() int
	GetValue func(idx lua.LValue) lua.LValue
	SetValue func(idx lua.LValue, val lua.LValue)
}

func (*LuaMap) LuaMetatableType() string {
	return MAP_METATABLE_NAME
}

func checkMap(param int, L *lua.LState) *LuaMap {
	m, ok := L.CheckUserData(param).Value.(*LuaMap)

	if !ok {
		L.ArgError(1, "LuaMap type expected")
	}

	return m
}

func mapIndex(L *lua.LState) int {
	m := checkMap(1, L)
	L.Push(m.GetValue(L.CheckAny(2)))
	return 1
}

func mapNewIndex(L *lua.LState) int {
	m := checkMap(1, L)
	m.SetValue(L.CheckAny(2), L.CheckAny(3))
	return 0
}

func mapLen(L *lua.LState) int {
	L.Push(lua.LNumber(checkMap(1, L).Len()))
	return 1
}

func MapLuaArrayOrTableToGoMap[K comparable, V any](p lua.LValue, mapper func(k, v lua.LValue) (K, V)) (map[K]V, error) {
	switch t := p.(type) {
	case *lua.LUserData:
		ar, ok := t.Value.(*LuaMap)

		if !ok {
			return nil, fmt.Errorf("incorrect user type.  expected LuaMap, got %T", ar)
		}

		m, ok := ar.Map.(map[K]V)

		if !ok {
			return nil, fmt.Errorf("incorrect array type in LuaMap.  expected %T, got %T", m, t.Value)
		}

		return m, nil
	case *lua.LTable:
		ret := make(map[K]V)

		t.ForEach(func(k, v lua.LValue) {
			ck, cv := mapper(k, v)
			ret[ck] = cv
		})

		return ret, nil
	}

	return nil, fmt.Errorf("expected LuaArray or table")
}
