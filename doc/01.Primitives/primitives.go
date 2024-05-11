package primitives

import (
	"fmt"

	primitivesubpackage "github.com/ChrisTrenkamp/gobindlua/doc/01.Primitives/primitive_subpackage"
)

/*
This gobindlua call will generate the file `lua_PrimitiveStruct.go`.
Projects should use "go run github.com/ChrisTrenkamp/gobindlua/gobindlua@version".
The version is left out of these examples for testing purposes.
*/

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua -s PrimitiveStruct
type PrimitiveStruct struct {
	// All exported fields will have bindings created for GopherLua
	MyBool  bool
	MyInt   int32
	MyInt64 int64
	// Notice how `lua_PrimitiveStruct.go` will properly cast the float64 to this named type
	MyFloat  primitivesubpackage.SomeFloat64
	MyString string
}

// Functions that return the type you're generating will be automatically bound
// to the metadata table.  The return type can be a pointer as well.
// If functions have a name that matches New(StructName)[OptionalQualifier],
// they will be added as a metatable field in the form of new[_optional_qualifier].
// Otherwise, the function will be added with the original name, but in snake_case form.
// This function will be added as "new".
func NewPrimitiveStruct() PrimitiveStruct {
	return PrimitiveStruct{}
}

// This method has a receiver on PrimitiveStruct, so gobindlua will create a binding for it.
func (p PrimitiveStruct) DivideMyInt(divisor float64) (float64, error) {
	if divisor == 0 {
		return 0, fmt.Errorf("divide by zero error")
	}

	return float64(p.MyInt) / divisor, nil
}
