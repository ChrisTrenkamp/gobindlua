package functions

import (
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
local functions = require "functions"
local FnContainer = require "FnContainer"

function lua_left_pad(str, pad)
	local ret = ""

	for i=1,pad,1 do
		ret = "lua" .. ret
	end

	return ret .. str
end

functions.PrintMe(functions.Split("foo_bar", "_"), functions.Split("eggs&ham", "&"))

print("NotIncluded was excluded from the bindings: " .. tostring(functions.NotIncluded == nil))

--[[ You can seamlessly pass Lua and Go functions as parameters. ]]
functions.DoFunc(lua_left_pad)
functions.DoFunc(functions.GoLeftPad)

--[[ You can also assign methods to struct fields. ]]
container = FnContainer.NewFnContainer(lua_left_pad)
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	// For pure functions, we use the LuaPreloadModule function instead of gobindlua.Register.
	LuaPreloadModule(L)

	gobindlua.Register(L, &FnContainer{})

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	container := L.GetGlobal("container").(*lua.LUserData).Value.(*FnContainer)
	str := " hi lua from go!"
	fmt.Println(container.Fn(&str, 2))

	// Output:
	//[foo bar] [eggs ham]
	//NotIncluded was excluded from the bindings: true
	//Result of fn("foo", 3) call: lualualuafoo
	//Result of fn("foo", 3) call: gogogofoo
	//lualua hi lua from go!
}
