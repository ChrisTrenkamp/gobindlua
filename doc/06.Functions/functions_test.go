package functions

import (
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
functions.print_me(functions.split("foo_bar", "_"), functions.split("eggs&ham", "&"))

print("NotIncluded was excluded from the bindings: " .. tostring(functions.not_included == nil))
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	gobindlua.Register(L, gobindlua.Funcs(RegisterFunctionsLuaType))

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	// Output:
	//[foo bar] [eggs ham]
	//NotIncluded was excluded from the bindings: true
}
