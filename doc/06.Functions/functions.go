package functions

import (
	"fmt"
	"strings"
)

// gobindlua can be configured to only generate functions.

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua -p functions

func PrintMe(args ...any) {
	fmt.Println(args...)
}

func Split(s string, spl string) []string {
	return strings.Split(s, spl)
}
