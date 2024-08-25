package functions

import (
	"log"

	lua "github.com/yuin/gopher-lua"
)

const script = `
local functions = require "functions"

functions.print_me(functions.split("foo_bar", "_"), functions.split("eggs&ham", "&"))

print("NotIncluded was excluded from the bindings: " .. tostring(functions.not_included == nil))
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	// For pure functions, we use the PreloadModule function instead of gobindlua.Register.
	L.PreloadModule("functions", FunctionsModuleLoader)

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	// Output:
	//[foo bar] [eggs ham]
	//NotIncluded was excluded from the bindings: true
}
