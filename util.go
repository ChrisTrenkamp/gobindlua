package gobindlua

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func FuncResCastError(L *lua.LState, res int, exp string, got any) {
	L.RaiseError("function result number %d expects %s, received %T", res, reduceUserData(exp), reduceUserData(got))
}

func TableElemCastError(L *lua.LState, level int, exp string, got any) {
	L.RaiseError("inner table assignment, level %d, expects %s, received %T", level, reduceUserData(exp), reduceUserData(got))
}

func CastArgError(L *lua.LState, arg int, exp string, got any) {
	L.ArgError(arg, fmt.Sprintf("expected %s, received %T", reduceUserData(exp), reduceUserData(got)))
}

func badArrayOrTableCast(exp, got any, level int) error {
	exp = reduceUserData(exp)
	got = reduceUserData(got)

	if level == 0 {
		return fmt.Errorf("expected %T, received %T", exp, got)
	}

	return fmt.Errorf("inner table assignment (level %d) expected %T, received %T", level+1, exp, got)
}

func reduceUserData(d any) any {
	if ud, ok := d.(*lua.LUserData); ud != nil && ok {
		return ud.Value
	}

	return d
}
