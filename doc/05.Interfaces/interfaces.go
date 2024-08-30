package interfaces

import "github.com/ChrisTrenkamp/gobindlua"

// You can generate Lua definitions for interfaces by attaching a //go:generate directive
// to an interface.

// In order to pass around interfaces, they must implement gobindlua.LuaUserData.
// The interface implementation doesn't necessarily need to be generated with
// gobindlua, but its metadata table must be globally available.  Otherwise, it
// will not work.

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Mammal interface {
	Sound() string
	gobindlua.LuaUserData
}

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

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Lion struct{}

//gobindlua:constructor
func NewLion() Lion {
	return Lion{}
}

//gobindlua:function
func (c Lion) Sound() string {
	return "rawr"
}

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

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type MammalList struct {
	Pet     Mammal
	NonPets []Mammal
}

//gobindlua:constructor
func NewMammalList() MammalList {
	return MammalList{}
}
