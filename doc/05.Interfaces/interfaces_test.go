package interfaces

import (
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	"github.com/ChrisTrenkamp/gobindlua/doc/05.Interfaces/dog"
	"github.com/ChrisTrenkamp/gobindlua/doc/05.Interfaces/human"
	"github.com/ChrisTrenkamp/gobindlua/doc/05.Interfaces/lion"
	"github.com/ChrisTrenkamp/gobindlua/doc/05.Interfaces/mammallist"
	lua "github.com/yuin/gopher-lua"
)

const script = `
local dog = require "dog"
local lion = require "lion"
local human = require "human"
local mammal_list = require "mammal_list"

local mammals = mammal_list.new()
mammals.pet = dog.new()
mammals.non_pets = { lion.new(), human.new() }

print("My pet says: " .. mammals.pet:sound())
print("The other mammals say:")
for i=1,#mammals.non_pets,1 do
	print(mammals.non_pets[i]:sound())
end
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	gobindlua.Register(L, &dog.Dog{}, &lion.Lion{}, &human.Human{}, &mammallist.MammalList{})

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	// Output: My pet says: bark
	// The other mammals say:
	// rawr
	// burp
}
