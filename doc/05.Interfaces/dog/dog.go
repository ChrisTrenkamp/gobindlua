package dog

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Dog struct{}

//gobindlua:constructor
func NewDog() Dog {
	return Dog{}
}

//gobindlua:function
func (d Dog) Sound() string {
	return "bark"
}
