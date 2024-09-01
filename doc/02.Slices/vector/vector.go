package vector

import (
	"fmt"

	"github.com/ChrisTrenkamp/gobindlua/doc/02.Slices/matrix"
)

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Vector struct {
	Elements []float64
}

//gobindlua:constructor
func NewVectorFrom(elems []float64) Vector {
	return Vector{Elements: elems}
}

//gobindlua:constructor
func NewVectorVariadic(elems ...float64) Vector {
	return Vector{Elements: elems}
}

//gobindlua:function
func (v Vector) InnerProduct(o Vector) (float64, error) {
	if len(v.Elements) != len(o.Elements) {
		return 0, fmt.Errorf("vector lengths not equal")
	}

	ret := float64(0)

	for i := 0; i < len(v.Elements); i++ {
		ret += v.Elements[i] * o.Elements[i]
	}

	return ret, nil
}

//gobindlua:function
func (v Vector) OuterProduct(o Vector) (matrix.Matrix, error) {
	ret := make([][]float64, 0)

	for i := 0; i < len(v.Elements); i++ {
		row := make([]float64, len(o.Elements))

		for j := 0; j < len(row); j++ {
			row[j] = v.Elements[i] * o.Elements[j]
		}

		ret = append(ret, row)
	}

	return matrix.Matrix{Elements: ret}, nil
}
