package pointers

import (
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
local sub = require "sub"

--[[ gobindlua will even work with pointers. ]]
s.a = {
	["a"]={
		["b"]={
			"c"
		}
	}
}
s.b = sub.new("d")
s.c = sub.new("e")
s.d = {
	{
		1,2
	},
	{
		3,4
	}
}
s.e = {
	{
		sub.new("f"),sub.new("g")
	},
	{
		sub.new("h"),sub.new("i")
	}
}
s.f = {
	{
		[sub.new("j")]=5
	},
	{
		[sub.new("k")]=6
	}
}
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	gobindlua.Register(L, &Sub{}, &SomeStruct{})

	someStruct := SomeStruct{}
	L.SetGlobal("s", gobindlua.NewUserData(&someStruct, L))

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	fmt.Println("A={")
	for k1, v1 := range *someStruct.A {
		fmt.Printf("\t%s={\n", *k1)
		for k2, v2 := range *v1 {
			fmt.Printf("\t\t%s={\n", k2)
			for _, v3 := range *v2 {
				fmt.Printf("\t\t\t%s\n", v3)
			}
			fmt.Println("\t\t}")
		}
		fmt.Println("\t}")
	}
	fmt.Println("}")
	fmt.Printf("B=%s\n", *someStruct.B.Str)
	fmt.Printf("C=%s\n", *someStruct.C.Str)

	fmt.Println("D={")
	for _, v1 := range *someStruct.D {
		fmt.Println("\t{")
		for _, v2 := range *v1 {
			fmt.Printf("\t\t%d\n", *v2)
		}
		fmt.Println("\t}")
	}
	fmt.Println("}")

	fmt.Println("E={")
	for _, v1 := range *someStruct.E {
		fmt.Println("\t{")
		for _, v2 := range v1 {
			fmt.Printf("\t\t%s\n", *v2.Str)
		}
		fmt.Println("\t}")
	}
	fmt.Println("}")

	fmt.Println("F={")
	for _, v1 := range someStruct.F {
		fmt.Println("\t{")
		for k2, v2 := range v1 {
			fmt.Printf("\t\t%s:%d\n", *k2.Str, *v2)
		}
		fmt.Println("\t}")
	}
	fmt.Println("}")

	// Output:
	//A={
	//	a={
	//		b={
	//			c
	//		}
	//	}
	//}
	//B=d
	//C=e
	//D={
	//	{
	//		1
	//		2
	//	}
	//	{
	//		3
	//		4
	//	}
	//}
	//E={
	//	{
	//		f
	//		g
	//	}
	//	{
	//		h
	//		i
	//	}
	//}
	//F={
	//	{
	//		j:5
	//	}
	//	{
	//		k:6
	//	}
	//}
}
