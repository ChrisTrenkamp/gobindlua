package gobindlua

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func CastArgError(exp string, got any) string {
	return fmt.Sprintf("expected %s, received %T", reduceUserData(exp), reduceUserData(got))
}

func TableElementCastError(exp string, got any, level int) string {
	return fmt.Sprintf("inner table assignment (level %d) expected %s, received %T", level, reduceUserData(exp), reduceUserData(got))
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
