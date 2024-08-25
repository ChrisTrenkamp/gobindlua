package gobindlua

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

const MAP_MODULES_NAME = "gbl_map"
const MAP_METATABLE_NAME = "gbl_map_fields"

func LuaMapModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "to_table", L.NewFunction(mapToTable))

	L.Push(staticMethodsTable)

	return 1
}

func LuaMapRegisterGlobalMetatable(L *lua.LState) {
	fieldsTable := L.NewTypeMetatable(MAP_METATABLE_NAME)
	L.SetGlobal(MAP_METATABLE_NAME, fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(mapIndex))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(mapNewIndex))
	L.SetField(fieldsTable, "__len", L.NewFunction(mapLen))
}

type LuaMap struct {
	Map      interface{}
	Len      func() int
	GetValue func(idx lua.LValue) lua.LValue
	SetValue func(idx lua.LValue, val lua.LValue)
	ForEach  func(f func(k, v lua.LValue))
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

func mapToTable(L *lua.LState) int {
	m := checkMap(1, L)
	ret := L.NewTable()

	m.ForEach(func(k, v lua.LValue) {
		ret.RawSet(k, v)
	})

	L.Push(ret)
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
	case *lua.LNilType:
		return nil, nil
	}

	return nil, fmt.Errorf("expected LuaArray or table")
}
