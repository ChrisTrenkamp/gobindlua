package pointers

//go:generate gobindlua -s SomeStruct
type SomeStruct struct {
	A *map[*string]*map[string]*[]string
	B *Sub
	C Sub
	D *[]*[]*int
	E *[][]*Sub
	F []map[*Sub]*int
}

//go:generate gobindlua -s Sub
type Sub struct {
	Str *string
}

func NewSub(str *string) *Sub {
	return &Sub{
		Str: str,
	}
}
