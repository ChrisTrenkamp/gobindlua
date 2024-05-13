// gobindlua can be configured to only generate functions.
// If the go:generate directive is placed behind a package
// declaration, gobindlua will automatically generate the
// functions for that package.  Otherwise, you will need to
// pass in the -package option.

// The -i option is used to explicitly declare which functions
// or methods you want to include in the bindings.

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua -i PrintMe -i Split
package functions

import (
	"fmt"
	"strings"
)

func PrintMe(args ...any) {
	fmt.Println(args...)
}

func Split(s string, spl string) []string {
	return strings.Split(s, spl)
}

// Since the -i option was specified, this function was not included.
func NotIncluded() {
	fmt.Println("this function is not included")
}
