package functions

import (
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
functions.print_me(functions.split("foo_bar", "_"), functions.split("eggs&ham", "&"))
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	RegisterFunctionsLuaType(L)
	gobindlua.RegisterLuaArray(L)

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	// Output:
	//[foo bar] [eggs ham]
}
