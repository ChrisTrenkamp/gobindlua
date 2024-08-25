package gobindlua

import (
	lua "github.com/yuin/gopher-lua"
)

const ARRAY_MODULES_NAME = "gbl_array"
const ARRAY_METATABLE_NAME = "gbl_array_fields"

func LuaArrayModuleLoader(L *lua.LState) int {
	staticMethodsTable := L.NewTable()
	L.SetField(staticMethodsTable, "to_table", L.NewFunction(arrayToTable))

	L.Push(staticMethodsTable)

	return 1
}

func LuaArrayRegisterGlobalMetatable(L *lua.LState) {
	fieldsTable := L.NewTypeMetatable(ARRAY_METATABLE_NAME)
	L.SetGlobal(ARRAY_METATABLE_NAME, fieldsTable)
	L.SetField(fieldsTable, "__index", L.NewFunction(arrayIndex))
	L.SetField(fieldsTable, "__newindex", L.NewFunction(arrayNewIndex))
	L.SetField(fieldsTable, "__len", L.NewFunction(arrayLen))
}

type LuaArray struct {
	Slice    interface{}
	Len      func() int
	Index    func(idx int) lua.LValue
	SetIndex func(idx int, val lua.LValue)
}

func (*LuaArray) LuaMetatableType() string {
	return ARRAY_METATABLE_NAME
}

func checkArray(param int, L *lua.LState) *LuaArray {
	slice, ok := L.CheckUserData(param).Value.(*LuaArray)

	if !ok {
		L.ArgError(1, "LuaArray type expected")
	}

	return slice
}

func arrayIndex(L *lua.LState) int {
	slice := checkArray(1, L)
	idx := int(L.CheckNumber(2) - 1)

	if idx < 0 || idx >= slice.Len() {
		L.ArgError(2, "out of bounds array index")
	}

	L.Push(slice.Index(idx))
	return 1
}

func arrayNewIndex(L *lua.LState) int {
	slice := checkArray(1, L)
	idx := int(L.CheckNumber(2) - 1)

	if idx < 0 || idx >= slice.Len() {
		L.ArgError(1, "out of bounds array index")
	}

	n := L.CheckAny(3)

	slice.SetIndex(idx, n)

	return 0
}

func arrayLen(L *lua.LState) int {
	L.Push(lua.LNumber(checkArray(1, L).Len()))
	return 1
}

func arrayToTable(L *lua.LState) int {
	slice := checkArray(1, L)
	ret := L.NewTable()
	len := slice.Len()

	for i := len - 1; i >= 0; i-- {
		ret.RawSetInt(i+1, slice.Index(i))
	}

	L.Push(ret)
	return 1
}

func MapLuaArrayOrTableToGoSlice[T any](p lua.LValue, level int, mapper func(val lua.LValue) T) ([]T, error) {
	var ret []T

	switch t := p.(type) {
	case *lua.LUserData:
		ar, ok := t.Value.(*LuaArray)

		if !ok {
			return nil, badArrayOrTableCast(ret, ar, level)
		}

		ret, ok = ar.Slice.([]T)

		if !ok {
			return nil, badArrayOrTableCast(ret, ar.Slice, level)
		}

		return ret, nil
	case *lua.LTable:
		ret = make([]T, t.MaxN())

		for i := 1; i <= t.MaxN(); i++ {
			ret[i-1] = mapper(t.RawGetInt(i))
		}

		return ret, nil
	case *lua.LNilType:
		return nil, nil
	default:
		return nil, badArrayOrTableCast(ret, t, level)
	}
}

func MapVariadicArgsToGoSlice[T any](start int, L *lua.LState, mapper func(val lua.LValue) T) ([]T, error) {
	ret := make([]T, 0)

	for i := start; i <= L.GetTop(); i++ {
		val := L.CheckAny(i)

		ret = append(ret, mapper(val))
	}

	return ret, nil
}
