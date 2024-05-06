package gobindluautil

import "github.com/gobeam/stringy"

func SnakeCase(str string) string {
	return stringy.New(str).SnakeCase().ToLower()
}
