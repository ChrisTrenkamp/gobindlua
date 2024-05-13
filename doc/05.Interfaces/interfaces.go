package interfaces

import "github.com/ChrisTrenkamp/gobindlua"

// You can generate Lua definitions for interfaces by attaching a go:generate directive
// on an interface, or with the -interface options.

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Mammal interface {
	Sound() string
	gobindlua.LuaUserData
}

// You can declare that a struct implements an interface in the Lua definitions by passing
// in the -im flag

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua -im Mammal
type Dog struct{}

func NewDog() Dog {
	return Dog{}
}

func (d Dog) Sound() string {
	return "bark"
}

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua -im Mammal
type Lion struct{}

func NewLion() Lion {
	return Lion{}
}

func (c Lion) Sound() string {
	return "rawr"
}

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua -im Mammal
type Human struct{}

func NewHuman() Human {
	return Human{}
}

func (h Human) Sound() string {
	return "burp"
}

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type MammalList struct {
	Pet     Mammal
	NonPets []Mammal
}

func NewMammalList() MammalList {
	return MammalList{}
}
