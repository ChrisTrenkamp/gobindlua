package primitives

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
local PrimitiveStruct = require "PrimitiveStruct"

--[[ You can call the constructor functions and methods and access the fields from Lua. ]]
data = PrimitiveStruct.NewPrimitiveStruct()
data.MyBool = true
data.MyInt = 42
data.MyInt64 = 0xDEADBEEF
data.MyFloat = 3.14
data.MyString = "all your lua are belong to us"

print("MyBool: " .. tostring(data.MyBool))
print("MyInt: " .. tostring(data.MyInt))
print("MyInt64: " .. tostring(data.MyInt64))
print("MyFloat: " .. tostring(data.MyFloat))
print("MyString: " .. tostring(data.MyString))
print("WillBeExcluded has been excluded: " .. tostring(data.WillBeExcluded == nil))

data:SetSpecializedInt(9001)
print("MySpecializedInt: " .. tostring(data.MySpecializedInt))

print("DivideMyInt: " .. tostring(data:DivideMyInt(2)))
local _, err = pcall(function () data:DivideMyInt(0) end)
print("DivideMyInt error: " .. err)

print("ExcludedMethod has been excluded: " .. tostring(data.ExcludedMethod == nil))
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	var registrarPointer *PrimitiveStruct
	gobindlua.Register(L, registrarPointer)
	// For ease of use, you can pass in &PrimitiveStruct{} to the Register function as well.
	// The Register function does not require the pointer to contain a value.  It's simply
	// for registering the metadata tables.

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	data := L.GetGlobal("data").(*lua.LUserData).Value.(*PrimitiveStruct)
	jsonBytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jsonBytes))

	// Output: MyBool: true
	// MyInt: 42
	// MyInt64: 3735928559
	// MyFloat: 3.14
	// MyString: all your lua are belong to us
	// WillBeExcluded has been excluded: true
	// MySpecializedInt: 9001
	// DivideMyInt: 21
	// DivideMyInt error: <string>:23: divide by zero error
	// ExcludedMethod has been excluded: true
	//{
	//	"MyBool": true,
	//	"MyInt": 42,
	//	"MyInt64": 3735928559,
	//	"MyFloat": 3.14,
	//	"SomeString": "all your lua are belong to us",
	//	"WillBeExcluded": "",
	//	"MySpecializedInt": 9001
	//}
}
