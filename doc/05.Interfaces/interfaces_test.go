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
local Dog = require "Dog"
local Lion = require "Lion"
local Human = require "Human"
local MammalList = require "MammalList"

local mammals = MammalList.NewMammalList()
mammals.Pet = Dog.NewDog()
mammals.NonPets = { Lion.NewLion(), Human.NewHuman() }

print("My pet says: " .. mammals.Pet:Sound())
print("The other mammals say:")
for i=1,#mammals.NonPets,1 do
	print(mammals.NonPets[i]:Sound())
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
