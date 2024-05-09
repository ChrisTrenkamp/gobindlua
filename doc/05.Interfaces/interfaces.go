package interfaces

import "github.com/ChrisTrenkamp/gobindlua"

type Mammal interface {
	Sound() string
	gobindlua.LuaUserData
}

//go:generate gobindlua -s Dog
type Dog struct{}

func NewDog() Dog {
	return Dog{}
}

func (d Dog) Sound() string {
	return "bark"
}

//go:generate gobindlua -s Lion
type Lion struct{}

func NewLion() Lion {
	return Lion{}
}

func (c Lion) Sound() string {
	return "rawr"
}

//go:generate gobindlua -s Human
type Human struct{}

func NewHuman() Human {
	return Human{}
}

func (h Human) Sound() string {
	return "burp"
}

//go:generate gobindlua -s MammalList
type MammalList struct {
	Pet     Mammal
	NonPets []Mammal
}

func NewMammalList() MammalList {
	return MammalList{}
}
