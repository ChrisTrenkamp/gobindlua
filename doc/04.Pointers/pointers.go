package pointers

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type SomeStruct struct {
	A *map[*string]*map[string]*[]string
	B *Sub
	C Sub
	D *[]*[]*int
	E *[][]*Sub
	F []map[*Sub]*int
}

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Sub struct {
	Str *string
}

func NewSub(str *string) *Sub {
	return &Sub{
		Str: str,
	}
}
