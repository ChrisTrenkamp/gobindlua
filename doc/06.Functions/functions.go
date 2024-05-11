// gobindlua can be configured to only generate functions.
// If the go:generate directive is placed behind a package
// declaration, gobindlua will automatically generate the
// functions for that package.  Otherwise, you will need to
// pass in the -p parameter.

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
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
