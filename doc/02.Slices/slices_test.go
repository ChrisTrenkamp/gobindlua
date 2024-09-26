package slices

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	"github.com/ChrisTrenkamp/gobindlua/doc/02.Slices/array"
	"github.com/ChrisTrenkamp/gobindlua/doc/02.Slices/matrix"
	"github.com/ChrisTrenkamp/gobindlua/doc/02.Slices/vector"
	lua "github.com/yuin/gopher-lua"
)

const script = `
local Vector = require "Vector"
local ArrayStruct = require "ArrayStruct"
local Matrix = require "Matrix"
local GblSlice = require "GblSlice"

--[[ Notice you can use lua tables as parameters for slices. ]]
local a = Vector.NewVectorFrom({3,2,1})
for i=1,#a.Elements,1 do
	print("Go slice element index " .. tostring(i) .. ": " .. a.Elements[i])
end

a.Elements[1] = 1
a.Elements[3] = 3

--[[ You can also convert the slice back to a table. ]]
local a_table = GblSlice.ToTable(a.Elements)
print("a_table type: " .. type(a_table))
for i=1,#a_table,1 do
	print("Element index " .. tostring(i) .. ": " .. a_table[i])
end

--[[ gobindlua can also handle variadic arguments. ]]
local b = Vector.NewVectorVariadic(4,5,6)

print("Inner product: " .. tostring(a:InnerProduct(b)))

m.Elements = a:OuterProduct(b).Elements
print("Outer product:")
print(m:String())

--[[ Arrays are a separate type, which can be generated with gobindlua ]]
local an_array = ArrayStruct.NewArrayStruct({1, 2, 3})
print("Array Elements before:")
print(an_array:String())

an_array:SetElements({4, 5, 6})
print("Array Elements after:")
print(an_array:String())

an_array:SetElementsFromSubpackage({3, 2, 1})
print("Array Elements SetElementsFromSubpackage:")
print(an_array:String())

an_array.Elements = {7, 8, 9}
print("Array Elements directly set:")
print(an_array:String())

local direct_array_access = an_array.Elements
direct_array_access[1] = 10
print("Direct array modification:")
print(an_array:String())

local identity_matrix = Matrix.NewMatrixFrom(
	{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1}
	}
)
print("Identity matrix:")
print(identity_matrix:String())
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	gobindlua.Register(L, &vector.Vector{}, &array.ArrayStruct{}, &matrix.Matrix{})

	matrix := matrix.Matrix{}
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
	// Array Elements before:
	// {1.000000, 2.000000, 3.000000}
	// Array Elements after:
	// {4.000000, 5.000000, 6.000000}
	// Array Elements SetElementsFromSubpackage:
	// {3.000000, 2.000000, 1.000000}
	// Array Elements directly set:
	// {7.000000, 8.000000, 9.000000}
	// Direct array modification:
	// {10.000000, 8.000000, 9.000000}
	// Identity matrix:
	// 1.00 0.00 0.00
	// 0.00 1.00 0.00
	// 0.00 0.00 1.00
	// Outer product result in Go: [[4,5,6],[8,10,12],[12,15,18]]
}
