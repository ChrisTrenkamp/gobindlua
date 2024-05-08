package gobindlua

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

const SLICE_METATABLE_NAME = "gobindluaslice"

func RegisterLuaArray(L *lua.LState) {
	if L.GetGlobal(SLICE_METATABLE_NAME) != lua.LNil {
		return
	}

	mt := L.NewTypeMetatable(SLICE_METATABLE_NAME)
	L.SetGlobal(SLICE_METATABLE_NAME, mt)

	mt.RawSetString("__index", L.NewFunction(arrayIndex))
	mt.RawSetString("__newindex", L.NewFunction(arrayNewIndex))
	mt.RawSetString("__len", L.NewFunction(arrayLen))
}

type LuaArray struct {
	Slice    interface{}
	Len      func() int
	Index    func(idx int) lua.LValue
	SetIndex func(idx int, val lua.LValue)
}

func (*LuaArray) LuaMetatableType() string {
	return SLICE_METATABLE_NAME
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

func MapLuaArrayOrTableToGoSlice[T any](p lua.LValue, mapper func(val lua.LValue) T) ([]T, error) {
	switch t := p.(type) {
	case *lua.LUserData:
		ar, ok := t.Value.(*LuaArray)

		if !ok {
			return nil, fmt.Errorf("incorrect user type.  expected LuaArray, got %T", ar)
		}

		sl, ok := ar.Slice.([]T)

		if !ok {
			return nil, fmt.Errorf("incorrect array type in LuaArray.  expected %T, got %T", sl, t.Value)
		}

		return sl, nil
	case *lua.LTable:
		ret := make([]T, t.MaxN())

		for i := 1; i <= t.MaxN(); i++ {
			ret[i-1] = mapper(t.RawGetInt(i))
		}

		return ret, nil
	}

	return nil, fmt.Errorf("expected LuaArray or table")
}

func MapVariadicArgsToGoSlice[T any](start int, L *lua.LState, mapper func(val lua.LValue) T) ([]T, error) {
	ret := make([]T, 0)

	for i := start; i <= L.GetTop(); i++ {
		val := L.CheckAny(i)

		switch val.(type) {
		case *lua.LUserData:
		case *lua.LTable:
			sl, err := MapLuaArrayOrTableToGoSlice(val, mapper)

			if err != nil {
				return nil, err
			}

			ret = append(ret, sl...)
		default:
			ret = append(ret, mapper(val))
		}
	}

	return ret, nil
}
