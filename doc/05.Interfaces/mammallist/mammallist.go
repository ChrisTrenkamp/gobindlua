package mammallist

import "github.com/ChrisTrenkamp/gobindlua/doc/05.Interfaces/mammal"

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type MammalList struct {
	Pet     mammal.Mammal
	NonPets []mammal.Mammal
}

//gobindlua:constructor
func NewMammalList() MammalList {
	return MammalList{}
}
