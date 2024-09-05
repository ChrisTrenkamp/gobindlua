package human

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Human struct{}

//gobindlua:constructor
func NewHuman() Human {
	return Human{}
}

//gobindlua:function
func (h Human) Sound() string {
	return "burp"
}
