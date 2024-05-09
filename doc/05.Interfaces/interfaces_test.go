package interfaces

import (
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
--[[
Implement gobindlua.LuaUserData in order to pass interfaces
around in GopherLua. The interface implementation doesn't necessarily
need to be generated with gobindlua, but its metadata table must be
globally available.  Otherwise, it will not work.
]]
local mammals = mammal_list.new()
mammals.pet = dog.new()
mammals.non_pets = { lion.new(), human.new() }

print("My pet says: " .. mammals.pet:sound())
print("The other mammals say:")
for i=1,#mammals.non_pets,1 do
	print(mammals.non_pets[i]:sound())
end
`

func ExampleMammalList() {
	L := lua.NewState()
	defer L.Close()

	Dog{}.RegisterLuaType(L)
	Lion{}.RegisterLuaType(L)
	Human{}.RegisterLuaType(L)
	MammalList{}.RegisterLuaType(L)
	gobindlua.RegisterLuaArray(L)

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	// Output: My pet says: bark
	// The other mammals say:
	// rawr
	// burp
}
