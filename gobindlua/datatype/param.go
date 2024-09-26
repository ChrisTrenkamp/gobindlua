package datatype

import (
	"go/types"
)

type Param struct {
	IsEllipses bool
	ParamNum   int
	LuaName    string
	DataType
}

func (p *Param) ConvertLuaTypeToGo(variableToCreate string, luaVariable string, paramNum int) string {
	if p.IsEllipses {
		if t, ok := p.DataType.Type.Underlying().(*types.Slice); ok {
			return p.DataType.ConvertLuaTypeToGoSliceEllipses(t, variableToCreate, luaVariable, paramNum)
		}
	}

	return p.DataType.ConvertLuaTypeToGo(variableToCreate, luaVariable, paramNum)
}

func (p *Param) LuaType(isFunctionReturn bool) string {
	if p.IsEllipses {
		if t, ok := p.DataType.Type.Underlying().(*types.Slice); ok {
			elem := p.DataType.createDataTypeFrom(t.Elem())
			return elem.LuaType(isFunctionReturn)
		}
	}

	return p.DataType.LuaType(isFunctionReturn)
}
