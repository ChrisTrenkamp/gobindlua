package datatype

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"text/template"

	"golang.org/x/tools/go/packages"
)

type DataType struct {
	Type               types.Type
	PointerIndirection int
	packageSource      *packages.Package
}

func CreateDataTypeFromExpr(expr ast.Expr, packageSource *packages.Package) DataType {
	return CreateDataTypeFrom(packageSource.TypesInfo.Types[expr].Type, packageSource)
}

func CreateDataTypeFrom(t types.Type, packageSource *packages.Package) DataType {
	pointerIndirection := 0

	for {
		if p, ok := t.(*types.Pointer); ok {
			t = p.Elem()
			pointerIndirection++
		} else {
			break
		}
	}

	return DataType{
		Type:               t,
		PointerIndirection: pointerIndirection,
		packageSource:      packageSource,
	}
}

func (d *DataType) ConvertGoTypeToLua(variable string) string {
	return convertGoTypeToLua(variable, d, 0)
}

func convertGoTypeToLua(variable string, variableType *DataType, level int) string {
	switch t := variableType.Type.Underlying().(type) {
	case *types.Basic:
		return fmt.Sprintf(`(%s%s)(%s)`, variableType.dereference(), variableType.luaType(), variable)
	case *types.Slice:
		return convertGoTypeToLuaSlice(t, variableType, variable, level)
	case *types.Map:
		return convertGoTypeToLuaMap(t, variableType, variable, level)
	}

	return fmt.Sprintf("gobindlua.NewUserData(%s%s, L)", variableType.referenceOrDereferenceUserDataForAssignment(), variable)
}

func convertGoTypeToLuaSlice(t *types.Slice, variableType *DataType, variable string, level int) string {
	elem := CreateDataTypeFrom(t.Elem(), variableType.packageSource)
	indexCode := fmt.Sprintf("(%s%s)[idx%d]", variableType.dereference(), variable, level)

	toLuaType := convertGoTypeToLua(indexCode, &elem, level+1)
	indexReturn := toLuaType

	if _, ok := elem.Type.Underlying().(*types.Basic); ok && elem.PointerIndirection > 0 {
		primIndex := fmt.Sprintf("*(%s%s)[idx%d]", variableType.dereference(), variable, level)
		derefElem := elem
		derefElem.PointerIndirection = 0
		indexReturn = convertGoTypeToLua(primIndex, &derefElem, level+1)
	}

	toGoType := elem.convertLuaTypeToGo(fmt.Sprintf("t%d", level), fmt.Sprintf("val%d", level), 3, level+1)
	pointerIndirection := elem.ReferenceOrDereferenceForAssignmentToField()

	args := struct {
		Variable            string
		Level               int
		LuaType             string
		GoType              string
		IndexCode           string
		DeclaredGoType      string
		PointerIndirection  string
		TemplateArg         string
		VariableDereference string
		IndexReturn         string
	}{
		Variable:            variable,
		Level:               level,
		LuaType:             toLuaType,
		GoType:              toGoType,
		IndexCode:           indexCode,
		DeclaredGoType:      elem.declaredGoType(),
		PointerIndirection:  pointerIndirection,
		TemplateArg:         elem.TemplateArg(),
		VariableDereference: variableType.dereference(),
		IndexReturn:         indexReturn,
	}

	templ := `gobindlua.NewUserData(&gobindlua.LuaArray{
	Slice: {{ .Variable }},
	Len:   func() int { return len({{ .VariableDereference }}{{ .Variable }}) },
	Index: func(idx{{ .Level }} int) lua.LValue { return {{ .IndexReturn }} },
	SetIndex: func(idx{{ .Level }} int, val{{ .Level }} lua.LValue) {
		{{ .GoType }}

		{{ .IndexCode }} = ({{ .TemplateArg }})({{ .PointerIndirection }}t{{ .Level }})
	},
}, L)`

	return execTempl(templ, args)
}

func convertGoTypeToLuaMap(t *types.Map, variableType *DataType, variable string, level int) string {
	key := CreateDataTypeFrom(t.Key(), variableType.packageSource)
	keyGoType := key.convertLuaTypeToGo(fmt.Sprintf("keyVal%d", level), fmt.Sprintf("key%d", level), 3, level+1)
	keyPointerIndirection := key.ReferenceOrDereferenceForAssignmentToField()

	val := CreateDataTypeFrom(t.Elem(), variableType.packageSource)
	valLuaType := convertGoTypeToLua(fmt.Sprintf("ret%d", level), &val, level+1)
	valGoType := val.convertLuaTypeToGo(fmt.Sprintf("valVal%d", level), fmt.Sprintf("val%d", level), 3, level+1)
	valPointerIndirection := val.ReferenceOrDereferenceForAssignmentToField()

	indexReturn := valLuaType

	if _, ok := val.Type.Underlying().(*types.Basic); ok && val.PointerIndirection > 0 {
		derefVal := val
		derefVal.PointerIndirection = 0
		indexReturn = convertGoTypeToLua(fmt.Sprintf("*(ret%d)", level), &derefVal, level+1)
	}

	args := struct {
		Variable              string
		Level                 int
		KeyGoType             string
		KeyDeclaredGoType     string
		KeyPointerIndirection string
		KeyTemplateArg        string
		ValLuaType            string
		ValGoType             string
		ValDeclaredGoType     string
		ValPointerIndirection string
		ValTemplateArg        string
		VariableDereference   string
		IndexReturn           string
	}{
		Variable:              variable,
		Level:                 level,
		KeyGoType:             keyGoType,
		KeyDeclaredGoType:     key.declaredGoType(),
		KeyPointerIndirection: keyPointerIndirection,
		KeyTemplateArg:        key.TemplateArg(),
		ValLuaType:            valLuaType,
		ValGoType:             valGoType,
		ValDeclaredGoType:     val.declaredGoType(),
		ValPointerIndirection: valPointerIndirection,
		ValTemplateArg:        val.TemplateArg(),
		VariableDereference:   variableType.dereference(),
		IndexReturn:           indexReturn,
	}

	templ := `gobindlua.NewUserData(&gobindlua.LuaMap{
Map: {{ .Variable }},
Len:   func() int { return len({{ .VariableDereference }}{{ .Variable }}) },
GetValue: func(key{{ .Level }} lua.LValue) lua.LValue {
	{{ .KeyGoType }}
	ret{{ .Level }} := ({{ .VariableDereference }}{{ .Variable }})[({{ .KeyTemplateArg }})({{ .KeyPointerIndirection }}keyVal{{ .Level }})]
	return {{ .IndexReturn }}
},
SetValue: func(key{{ .Level }} lua.LValue, val{{ .Level }} lua.LValue) {
	{{ .KeyGoType }}
	{{ .ValGoType }}
	({{ .VariableDereference }}{{ .Variable }})[({{ .KeyTemplateArg }})({{ .KeyPointerIndirection }}keyVal{{ .Level }})] = ({{ .ValTemplateArg }})({{ .ValPointerIndirection }}valVal{{ .Level }})
},
}, L)`

	return execTempl(templ, args)
}

func (d *DataType) ReferenceOrDereferenceForAssignmentToField() string {
	goPointerLevel := d.PointerIndirection
	luaPointerLevel := 0

	if _, ok := d.Type.Underlying().(*types.Struct); ok {
		luaPointerLevel++
	}

	if luaPointerLevel < goPointerLevel {
		return "&"
	}

	ret := ""

	for luaPointerLevel > goPointerLevel {
		ret += "*"
		goPointerLevel++
	}

	return ret
}

func (d *DataType) referenceOrDereferenceUserDataForAssignment() string {
	if d.PointerIndirection < 0 {
		panic("Pointer indirection is less than 0 (how is this even possible?)")
	}

	if d.PointerIndirection == 0 {
		return "&"
	}

	ret := ""

	for i := d.PointerIndirection - 1; i > 0; i-- {
		ret += "*"
	}

	return ret
}

func (d *DataType) ConvertLuaTypeToGo(variable string, source string, paramNum int) string {
	return d.convertLuaTypeToGo(variable, source, paramNum, 0)
}

func (d *DataType) convertLuaTypeToGo(variableToCreate string, luaVariable string, paramNum, level int) string {
	switch t := d.Type.Underlying().(type) {
	case *types.Basic:
		return d.convertLuaTypeToGoPrimitive(variableToCreate, luaVariable, paramNum, level)
	case *types.Slice:
		return d.convertLuaTypeToGoSlice(t, variableToCreate, luaVariable, paramNum, level)
	case *types.Map:
		return d.convertLuaTypeToGoMap(t, variableToCreate, luaVariable, paramNum, level)
	case *types.Struct:
		return d.convertLuaTypeToStruct(variableToCreate, luaVariable, paramNum, level)
	}

	return "CANNOT_CONVERT_LUA_TYPE_TO_GO"
}

func execTempl(templ string, data any) string {
	out := bytes.Buffer{}

	t := template.Must(template.New("").Parse(templ))
	err := t.Execute(&out, data)

	if err != nil {
		panic(err)
	}

	return out.String()
}

func (d *DataType) convertLuaTypeToGoPrimitive(variableToCreate string, luaVariable string, paramNum, level int) string {
	if level == 0 {
		// Level 0 means the variable came from a L.Check*, which means it was already type-checked
		return fmt.Sprintf(`%s := %s(%s)`, variableToCreate, d.declaredGoType(), luaVariable)
	}

	args := struct {
		VariableToCreate string
		LuaVariable      string
		LuaType          string
		ParamNum         int
		DeclaredGoType   string
	}{
		VariableToCreate: variableToCreate,
		LuaVariable:      luaVariable,
		LuaType:          d.luaType(),
		ParamNum:         paramNum,
		DeclaredGoType:   d.declaredGoType(),
	}

	if d.PointerIndirection == 0 {
		templ := `
{{ .VariableToCreate }}, ok := {{ .LuaVariable }}.({{ .LuaType }})

if !ok {
	L.ArgError({{ .ParamNum }}, "argument not a {{ .DeclaredGoType }} instance")
}
`
		return execTempl(templ, args)
	}

	templ := `
{{ .VariableToCreate }}_n, ok := {{ .LuaVariable }}.({{ .LuaType }})

if !ok {
	L.ArgError({{ .ParamNum }}, "argument not a {{ .DeclaredGoType }} instance")
}

{{ .VariableToCreate }} := {{ .DeclaredGoType }}({{ .VariableToCreate }}_n)
`
	return execTempl(templ, args)
}

func (d *DataType) convertLuaTypeToGoSlice(t *types.Slice, variableToCreate string, luaVariable string, paramNum, level int) string {
	elem := CreateDataTypeFrom(t.Elem(), d.packageSource)
	toGoType := elem.convertLuaTypeToGo(fmt.Sprintf("v%d", level), fmt.Sprintf("val%d", level), paramNum, level+1)
	pointerIndirection := elem.ReferenceOrDereferenceForAssignmentToField()

	args := struct {
		VariableToCreate   string
		ActualGoType       string
		LuaVariable        string
		Level              int
		ToGoType           string
		PointerIndirection string
		ParamNum           int
		TemplateArg        string
	}{
		VariableToCreate:   variableToCreate,
		ActualGoType:       elem.ActualGoType(),
		LuaVariable:        luaVariable,
		Level:              level,
		ToGoType:           toGoType,
		PointerIndirection: pointerIndirection,
		ParamNum:           paramNum,
		TemplateArg:        elem.TemplateArg(),
	}
	templ := `
{{ .VariableToCreate }}, err := gobindlua.MapLuaArrayOrTableToGoSlice[{{ .TemplateArg }}]({{ .LuaVariable }}, func(val{{ .Level }} lua.LValue) {{ .TemplateArg }} {
	{{ .ToGoType }}
	return ({{ .TemplateArg }})({{ .PointerIndirection }}v{{ .Level }})
})

if err != nil {
	L.ArgError({{ .ParamNum }}, err.Error())
}		
`
	return execTempl(templ, args)
}

func (d *DataType) convertLuaTypeToGoMap(t *types.Map, variableToCreate string, luaVariable string, paramNum, level int) string {
	k := CreateDataTypeFrom(t.Key(), d.packageSource)
	v := CreateDataTypeFrom(t.Elem(), d.packageSource)
	keyGoType := k.convertLuaTypeToGo(fmt.Sprintf("k%d", level), fmt.Sprintf("key%d", level), paramNum, level+1)
	valGoType := v.convertLuaTypeToGo(fmt.Sprintf("v%d", level), fmt.Sprintf("val%d", level), paramNum, level+1)
	keyPointerIndirection := k.ReferenceOrDereferenceForAssignmentToField()
	valPointerIndirection := v.ReferenceOrDereferenceForAssignmentToField()

	args := struct {
		VariableToCreate      string
		KeyActualGoType       string
		ValActualGoType       string
		LuaVariable           string
		Level                 int
		KeyGoType             string
		ValGoType             string
		KeyPointerIndirection string
		ValPointerIndirection string
		ParamNum              int
		KeyTemplateArg        string
		ValTemplateArg        string
	}{
		VariableToCreate:      variableToCreate,
		KeyActualGoType:       k.ActualGoType(),
		ValActualGoType:       v.ActualGoType(),
		LuaVariable:           luaVariable,
		Level:                 level,
		KeyGoType:             keyGoType,
		ValGoType:             valGoType,
		KeyPointerIndirection: keyPointerIndirection,
		ValPointerIndirection: valPointerIndirection,
		ParamNum:              paramNum,
		KeyTemplateArg:        k.TemplateArg(),
		ValTemplateArg:        v.TemplateArg(),
	}

	templ := `
{{ .VariableToCreate }}, err := gobindlua.MapLuaArrayOrTableToGoMap[{{ .KeyTemplateArg }}, {{ .ValTemplateArg }}]({{ .LuaVariable }}, func(key{{ .Level }}, val{{ .Level }} lua.LValue) ({{ .KeyTemplateArg }}, {{ .ValTemplateArg }}) {
	{{ .KeyGoType }}
	{{ .ValGoType }}
	return ({{ .KeyTemplateArg }})({{ .KeyPointerIndirection }}k{{ .Level }}), ({{ .ValTemplateArg }})({{ .ValPointerIndirection }}v{{ .Level }})
})

if err != nil {
	L.ArgError({{ .ParamNum }}, err.Error())
}
`

	return execTempl(templ, args)
}

func (d *DataType) convertLuaTypeToStruct(variableToCreate string, luaVariable string, paramNum, level int) string {
	args := struct {
		VariableToCreate string
		LuaVariable      string
		DeclaredGoType   string
		ParamNum         int
	}{
		VariableToCreate: variableToCreate,
		LuaVariable:      luaVariable,
		DeclaredGoType:   d.declaredGoType(),
		ParamNum:         paramNum,
	}

	if level == 0 {
		templ := `
{{ .VariableToCreate }}, ok := {{ .LuaVariable }}.Value.(*{{ .DeclaredGoType }})

if !ok {
	L.ArgError(3, "{{ .DeclaredGoType }} expected")
}
`

		return execTempl(templ, args)
	}

	templ := `
{{ .VariableToCreate }}_ud, ok := {{ .LuaVariable }}.(*lua.LUserData)

if !ok {
	L.ArgError({{ .ParamNum }}, "UserData expected")
}

{{ .VariableToCreate }}, ok := {{ .VariableToCreate }}_ud.Value.(*{{ .DeclaredGoType }})

if !ok {
	L.ArgError(3, "{{ .DeclaredGoType }} expected")
}
`

	return execTempl(templ, args)
}

func (d *DataType) ConvertLuaTypeToGoSliceEllipses(t *types.Slice, variableToCreate string, luaVariable string, paramNum int) string {
	level := 0
	elem := CreateDataTypeFrom(t.Elem(), d.packageSource)
	toGoType := elem.convertLuaTypeToGo(fmt.Sprintf("v%d", level), fmt.Sprintf("val%d", level), paramNum, level+1)
	pointerIndirection := elem.ReferenceOrDereferenceForAssignmentToField()

	args := struct {
		VariableToCreate   string
		ActualGoType       string
		ParamNum           int
		Level              int
		GoType             string
		PointerIndirection string
		TemplateArg        string
	}{
		VariableToCreate:   variableToCreate,
		ActualGoType:       elem.ActualGoType(),
		ParamNum:           paramNum,
		Level:              level,
		GoType:             toGoType,
		PointerIndirection: pointerIndirection,
		TemplateArg:        elem.TemplateArg(),
	}

	templ := `
{{ .VariableToCreate }}, err := gobindlua.MapVariadicArgsToGoSlice[{{ .TemplateArg }}]({{ .ParamNum }}, L, func(val{{ .Level }} lua.LValue) {{ .TemplateArg }} {
	{{ .GoType }}
	return ({{ .TemplateArg }})({{ .PointerIndirection }}v{{ .Level }})
})

if err != nil {
	L.ArgError({{ .ParamNum }}, err.Error())
}
`

	return execTempl(templ, args)
}

func (d *DataType) declaredGoType() string {
	if n, ok := d.Type.(*types.Named); ok && !d.IsError() {
		pkgName := n.Obj().Pkg().Name()
		name := n.Obj().Name()

		if n.Obj().Pkg().Path() == d.packageSource.ID {
			return name
		}

		return pkgName + "." + name
	}

	return d.Type.String()
}

func (d *DataType) Package() string {
	if n, ok := d.Type.(*types.Named); ok && !d.IsError() {
		pkg := n.Obj().Pkg()

		if pkg != nil {
			return pkg.Path()
		}
	}

	return ""
}

func (d *DataType) TemplateArg() string {
	indir := d.dereference()

	switch t := d.Type.Underlying().(type) {
	case *types.Slice:
		sl := CreateDataTypeFrom(t.Elem(), d.packageSource)
		return indir + "[]" + sl.TemplateArg()
	case *types.Map:
		k := CreateDataTypeFrom(t.Key(), d.packageSource)
		v := CreateDataTypeFrom(t.Elem(), d.packageSource)
		return indir + "map[" + k.TemplateArg() + "]" + v.TemplateArg()
	}

	return indir + d.ActualGoType()
}

func (d *DataType) dereference() string {
	indir := ""

	for i := 0; i < d.PointerIndirection; i++ {
		indir += "*"
	}

	return indir
}

func (d *DataType) ActualGoType() string {
	switch t := d.Type.Underlying().(type) {
	case *types.Basic, *types.Slice:
		return d.Type.Underlying().String()
	case *types.Map:
		l := CreateDataTypeFrom(t.Key(), d.packageSource)
		v := CreateDataTypeFrom(t.Elem(), d.packageSource)

		return fmt.Sprintf("map[%s]%s", l.ActualGoType(), v.ActualGoType())
	}

	return d.declaredGoType()
}

func (d *DataType) IsError() bool {
	return d.Type.String() == "error"
}

func (d *DataType) luaType() string {
	switch d.Type.Underlying().(type) {
	case *types.Basic:
		switch d.ActualGoType() {
		case "bool":
			return "lua.LBool"
		case "string":
			return "lua.LString"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "byte", "uint16", "uint32", "uint64", "float32", "float64":
			return "lua.LNumber"
		}
	case *types.Slice:
		return "*gobindlua.LuaArray"
	case *types.Map:
		return "*gobindlua.LuaMap"
	case *types.Struct:
		return "lua.LUserData"
	}

	return "UNSUPPORTED_TYPE"
}

func (d *DataType) LuaParamType() string {
	switch d.Type.Underlying().(type) {
	case *types.Basic:
		switch d.ActualGoType() {
		case "bool":
			return "L.CheckBool"
		case "string":
			return "L.CheckString"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "byte", "uint16", "uint32", "uint64", "float32", "float64":
			return "L.CheckNumber"
		}
	case *types.Slice:
		return "L.CheckAny"
	case *types.Map:
		return "L.CheckAny"
	case *types.Struct:
		return "L.CheckUserData"
	}

	return "UNSUPPORTED_TYPE"
}
