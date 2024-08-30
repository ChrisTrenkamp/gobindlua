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

type SpecializedInt uint32

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
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

// Use the gobindlua:constructor directive to declare a function as a
// constructor in the Lua bindings.  If there's multiple return values, the first value
// MUST be the struct you're creating bindings for.  The return type can be a pointer as well.
// If constructors have a name that matches New(StructName)[OptionalQualifier],
// they will be added as a metatable field in the form of new[_optional_qualifier].
// Otherwise, the function will be added with the original name, but in snake_case form.
// This function will be added as "new".
//
//gobindlua:constructor
func NewPrimitiveStruct() PrimitiveStruct {
	return PrimitiveStruct{}
}

// Use the gobindlua:function directive to bind a method in Lua.
//
//gobindlua:function
func (p PrimitiveStruct) DivideMyInt(divisor float64) (float64, error) {
	if divisor == 0 {
		return 0, fmt.Errorf("divide by zero error")
	}

	return float64(p.MyInt) / divisor, nil
}

// There's no gobindlua directive, so this method has been excluded.
func (p PrimitiveStruct) ExcludedMethod() {
	fmt.Println("I've been excluded from gobindlua.")
}

//gobindlua:function
func (p *PrimitiveStruct) SetSpecializedInt(i SpecializedInt) {
	p.MySpecializedInt = i
}
