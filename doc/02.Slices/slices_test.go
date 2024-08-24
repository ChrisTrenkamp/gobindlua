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
local a = vector.new_from({3,2,1})
for i=1,#a.elements,1 do
	print("Go slice element index " .. tostring(i) .. ": " .. a.elements[i])
end

a.elements[1] = 1
a.elements[3] = 3

--[[ You can also convert the slice back to a table. ]]
local a_table = gbl_array.to_table(a.elements)
print("a_table type: " .. type(a_table))
for i=1,#a_table,1 do
	print("Element index " .. tostring(i) .. ": " .. a_table[i])
end

--[[ gobindlua can also handle variadic arguments. ]]
local b = vector.new_variadic(4,5,6)

print("Inner product: " .. tostring(a:inner_product(b)))

m.elements = a:outer_product(b).elements
print("Outer product:")
print(m:string())

--[[ Arrays are a separate type, which can be generated with gobindlua ]]
local an_array = array_struct.new({1, 2, 3})
print("Array elements before:")
print(an_array:string())

an_array:set_elements({4, 5, 6})
print("Array elements after:")
print(an_array:string())

an_array:set_elements_from_subpackage({3, 2, 1})
print("Array elements set_elements_from_subpackage:")
print(an_array:string())

an_array.elements = {7, 8, 9}
print("Array elements directly set:")
print(an_array:string())

local direct_array_access = an_array.elements
direct_array_access[1] = 10
print("Direct array modification:")
print(an_array:string())

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

func Example() {
	L := lua.NewState()
	defer L.Close()

	gobindlua.Register(L, Vector{}, Matrix{}, ArrayStruct{})

	matrix := Matrix{}
	L.SetGlobal("m", gobindlua.NewUserData(&matrix, L))

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	jsonBytes, err := json.Marshal(matrix.Elements)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Outer product result in Go:", string(jsonBytes))

	// Output:
	//Go slice element index 1: 3
	//Go slice element index 2: 2
	//Go slice element index 3: 1
	// a_table type: table
	// Element index 1: 1
	// Element index 2: 2
	// Element index 3: 3
	// Inner product: 32
	// Outer product:
	// 4.00 5.00 6.00
	// 8.00 10.00 12.00
	// 12.00 15.00 18.00
	// Array elements before:
	// {1.000000, 2.000000, 3.000000}
	// Array elements after:
	// {4.000000, 5.000000, 6.000000}
	// Array elements set_elements_from_subpackage:
	// {3.000000, 2.000000, 1.000000}
	// Array elements directly set:
	// {7.000000, 8.000000, 9.000000}
	// Direct array modification:
	// {10.000000, 8.000000, 9.000000}
	// Identity matrix:
	// 1.00 0.00 0.00
	// 0.00 1.00 0.00
	// 0.00 0.00 1.00
	// Outer product result in Go: [[4,5,6],[8,10,12],[12,15,18]]
}
