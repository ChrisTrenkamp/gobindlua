package interfaces

import (
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
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

	gobindlua.Register(L, Dog{}, Lion{}, Human{}, MammalList{})

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	// Output: My pet says: bark
	// The other mammals say:
	// rawr
	// burp
}
