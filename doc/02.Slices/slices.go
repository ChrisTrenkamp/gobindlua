package slices

import (
	"fmt"
	"strings"
)

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Vector struct {
	Elements []float64
}

func NewVectorFrom(elems []float64) Vector {
	return Vector{Elements: elems}
}

func NewVectorVariadic(elems ...float64) Vector {
	return Vector{Elements: elems}
}

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

func (v Vector) OuterProduct(o Vector) (Matrix, error) {
	ret := make([][]float64, 0)

	for i := 0; i < len(v.Elements); i++ {
		row := make([]float64, len(o.Elements))

		for j := 0; j < len(row); j++ {
			row[j] = v.Elements[i] * o.Elements[j]
		}

		ret = append(ret, row)
	}

	return Matrix{Elements: ret}, nil
}

const ArrSize = 3

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type ArrayStruct struct {
	Elements [ArrSize]float32
}

func NewArrayStruct(elems [3]float32) ArrayStruct {
	return ArrayStruct{Elements: elems}
}

func (s *ArrayStruct) SetElements(j [3]float32) {
	s.Elements = j
}

func (s ArrayStruct) String() string {
	return fmt.Sprintf("{%f, %f, %f}", s.Elements[0], s.Elements[1], s.Elements[2])
}

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Matrix struct {
	Elements [][]float64
}

func NewMatrixFrom(elems [][]float64) Matrix {
	return Matrix{Elements: elems}
}

func (m Matrix) String() string {
	ret := ""

	for i := 0; i < len(m.Elements); i++ {
		line := ""

		for j := 0; j < len(m.Elements[i]); j++ {
			line += fmt.Sprintf("%.2f ", m.Elements[i][j])
		}

		ret += strings.TrimSpace(line) + "\n"
	}

	return strings.TrimSpace(ret)
}
