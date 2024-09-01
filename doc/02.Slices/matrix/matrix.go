package matrix

import (
	"fmt"
	"strings"
)

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Matrix struct {
	Elements [][]float64
}

//gobindlua:constructor
func NewMatrixFrom(elems [][]float64) Matrix {
	return Matrix{Elements: elems}
}

//gobindlua:function
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
