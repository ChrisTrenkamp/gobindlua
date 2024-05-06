package slices

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
--[[ Notice you can use lua tables as parameters for slices. ]]
local a = vector.new_from({1,2,3})
local b = vector.new_variadic(4,5,6)
print("Inner product: " .. tostring(a:inner_product(b)))
m.elements = a:outer_product(b).elements
print("Outer product:")
print(m:string())
local identity_matrix = matrix.new_from(
	{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1}
	}
)
print("Identity matrix:")
print(identity_matrix:string())
`

func ExampleVector() {
	L := lua.NewState()
	defer L.Close()

	Vector{}.RegisterLuaType(L)
	Matrix{}.RegisterLuaType(L)
	gobindlua.RegisterLuaArray(L)

	matrix := Matrix{}
	L.SetGlobal("m", gobindlua.NewUserData(&matrix, L))

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	jsonBytes, err := json.Marshal(matrix.Elements)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jsonBytes))

	// Output:
	// Inner product: 32
	// Outer product:
	// 4.00 5.00 6.00
	// 8.00 10.00 12.00
	// 12.00 15.00 18.00
	// Identity matrix:
	// 1.00 0.00 0.00
	// 0.00 1.00 0.00
	// 0.00 0.00 1.00
	// [[4,5,6],[8,10,12],[12,15,18]]
}
