package array

import (
	"fmt"

	slicessubpackage "github.com/ChrisTrenkamp/gobindlua/doc/02.Slices/slices_subpackage"
)

const ArrSize = 3

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type ArrayStruct struct {
	Elements [ArrSize]float32
}

//gobindlua:constructor
func NewArrayStruct(elems [3]float32) ArrayStruct {
	return ArrayStruct{Elements: elems}
}

//gobindlua:function
func (s *ArrayStruct) SetElements(j [3]float32) {
	s.Elements = j
}

//gobindlua:function
func (s *ArrayStruct) SetElementsFromSubpackage(j *slicessubpackage.AnArray) {
	s.Elements[0] = j[0]
	s.Elements[1] = j[1]
	s.Elements[2] = j[2]
}

//gobindlua:function
func (s ArrayStruct) String() string {
	return fmt.Sprintf("{%f, %f, %f}", s.Elements[0], s.Elements[1], s.Elements[2])
}
