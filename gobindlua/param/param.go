package param

import (
	"go/types"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
)

type Param struct {
	IsEllipses bool
	ParamNum   int
	datatype.DataType
}

func (p *Param) ConvertLuaTypeToGo(variableToCreate string, luaVariable string, paramNum int) string {
	if p.IsEllipses {
		if t, ok := p.DataType.Type.Underlying().(*types.Slice); ok {
			return p.DataType.ConvertLuaTypeToGoSliceEllipses(t, variableToCreate, luaVariable, paramNum)
		}
	}

	return p.DataType.ConvertLuaTypeToGo(variableToCreate, luaVariable, paramNum)
}
