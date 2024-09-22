// gobindlua can be configured to only generate functions.
// If the go:generate directive is placed behind a package
// declaration. gobindlua will generate bindings for functions
// that have a gobindlua:function directive.

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
package functions

import (
	"fmt"
	"strings"
)

//gobindlua:function
func PrintMe(args ...any) {
	fmt.Println(args...)
}

//gobindlua:function
func Split(s string, spl string) []string {
	return strings.Split(s, spl)
}

func NotIncluded() {
	fmt.Println("this function is not included")
}

//gobindlua:function
func GoLeftPad(str string, pad int) string {
	return strings.Repeat("go", pad) + str
}

//gobindlua:function
func DoFunc(fn func(string, int) string) {
	fmt.Println(`Result of fn("foo", 3) call:`, fn("foo", 3))
}
