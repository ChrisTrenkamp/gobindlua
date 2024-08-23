package primitives

import (
	"fmt"

	primitivesubpackage "github.com/ChrisTrenkamp/gobindlua/doc/01.Primitives/primitive_subpackage"
)

// The following go:generate directive will generate the file `lua_PrimitiveStruct.go`.
// Projects should use "go run github.com/ChrisTrenkamp/gobindlua/gobindlua@version".
// The version is left out of these examples for testing purposes.
// If the go:generate directive is placed behind a struct declaration, gobindlua will
// generate the bindings for that struct.

// The -x option is used to exclude functions and methods.  You can also use the -i
// option to selectively include which functions and methods you want.

type SpecializedInt uint32

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua -x ExcludedMethod
type PrimitiveStruct struct {
	// All exported fields will have bindings created for GopherLua
	MyBool  bool
	MyInt   int32
	MyInt64 int64
	// Notice how `lua_PrimitiveStruct.go` will properly cast the float64 to this named type
	MyFloat primitivesubpackage.SomeFloat64
	// You can use tags to explicitly declare the Lua name of the field.
	SomeString string `gobindlua:"my_string"`
	// You can also exclude fields
	WillBeExcluded   string `gobindlua:"-"`
	MySpecializedInt SpecializedInt
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

// The -x parameter prevented this method from being included in the bindings.
func (p PrimitiveStruct) ExcludedMethod() {
	fmt.Println("I've been excluded from gobindlua.")
}

func (p *PrimitiveStruct) SetSpecializedInt(i SpecializedInt) {
	p.MySpecializedInt = i
}
