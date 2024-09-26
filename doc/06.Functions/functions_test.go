package functions

import (
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
local functions = require "functions"
local fn_container = require "fn_container"

function lua_left_pad(str, pad)
	local ret = ""

	for i=1,pad,1 do
		ret = "lua" .. ret
	end

	return ret .. str
end

functions.print_me(functions.split("foo_bar", "_"), functions.split("eggs&ham", "&"))

print("NotIncluded was excluded from the bindings: " .. tostring(functions.not_included == nil))

--[[ You can seamlessly pass Lua and Go functions as parameters. ]]
functions.do_func(lua_left_pad)
functions.do_func(functions.go_left_pad)

--[[ You can also assign methods to struct fields. ]]
container = fn_container.new(lua_left_pad)
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	// For pure functions, we use the PreloadModule function instead of gobindlua.Register.
	L.PreloadModule("functions", FunctionsModuleLoader)

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
