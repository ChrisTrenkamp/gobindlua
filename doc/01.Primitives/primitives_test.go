package primitives

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
--[[ You can call the constructor functions and methods and access the fields from Lua. ]]
data = primitive_struct.new()
data.my_bool = true
data.my_int = 42
data.my_int64 = 0xDEADBEEF
data.my_float = 3.14
data.my_string = "all your lua are belong to us"

print("MyBool: " .. tostring(data.my_bool))
print("MyInt: " .. tostring(data.my_int))
print("MyInt64: " .. tostring(data.my_int64))
print("MyFloat: " .. tostring(data.my_float))
print("MyString: " .. tostring(data.my_string))

print("DivideMyInt: " .. tostring(data:divide_my_int(2)))
local _, err = pcall(function () data:divide_my_int(0) end)
print("DivideMyInt error: " .. err)
`

func ExamplePrimitiveStruct() {
	L := lua.NewState()
	defer L.Close()

	gobindlua.Register(L, PrimitiveStruct{})

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
	// DivideMyInt: 21
	// DivideMyInt error: <string>:17: divide by zero error
	// {
	// 	"MyBool": true,
	// 	"MyInt": 42,
	// 	"MyInt64": 3735928559,
	// 	"MyFloat": 3.14,
	// 	"MyString": "all your lua are belong to us"
	// }
}
