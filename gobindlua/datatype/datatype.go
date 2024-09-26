package datatype

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strconv"
	"strings"
	"text/template"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/declaredinterface"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"golang.org/x/tools/go/packages"
)

type DataType struct {
	Type                  types.Type
	PointerIndirection    int
	packageSource         *packages.Package
	allDeclaredInterfaces []declaredinterface.DeclaredInterface
}

func CreateDataTypeFromExpr(expr ast.Expr, packageSource *packages.Package, allDeclaredInterfaces []declaredinterface.DeclaredInterface) DataType {
	return CreateDataTypeFrom(packageSource.TypesInfo.Types[expr].Type, packageSource, allDeclaredInterfaces)
}

func CreateDataTypeFrom(t types.Type, packageSource *packages.Package, allDeclaredInterfaces []declaredinterface.DeclaredInterface) DataType {
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
		Type:                  t,
		PointerIndirection:    pointerIndirection,
		packageSource:         packageSource,
		allDeclaredInterfaces: allDeclaredInterfaces,
	}
}

func (d *DataType) ConvertGoTypeToLua(variable string) string {
	return convertGoTypeToLua(variable, d, 0)
}

func (d *DataType) ConvertGoTypeToLuaWithTableLevel(variable string, tableLevel int) string {
	return convertGoTypeToLua(variable, d, tableLevel)
}

func convertGoTypeToLua(variable string, variableType *DataType, tableLevel int) string {
	switch t := variableType.Type.Underlying().(type) {
	case *types.Basic:
		return fmt.Sprintf(`(%s)(%s%s)`, variableType.luaType(), variableType.dereference(), variable)
	case *types.Array:
		return convertGoTypeToLuaSlice(t.Elem(), variableType, variable, tableLevel)
	case *types.Slice:
		return convertGoTypeToLuaSlice(t.Elem(), variableType, variable, tableLevel)
	case *types.Map:
		return convertGoTypeToLuaMap(t.Key(), t.Elem(), variableType, variable, tableLevel)
	case *types.Interface:
		return fmt.Sprintf("gobindlua.NewUserData(%s, L)", variable)
	case *types.Signature:
		return convertGoTypeToFunc(t, variableType, variable)
	}

	return fmt.Sprintf("gobindlua.NewUserData(%s%s, L)", variableType.referenceOrDereferenceUserDataForAssignment(), variable)
}

func convertGoTypeToLuaSlice(typ types.Type, variableType *DataType, variable string, tableLevel int) string {
	elem := CreateDataTypeFrom(typ, variableType.packageSource, variableType.allDeclaredInterfaces)
	indexCode := fmt.Sprintf("(%s%s)[idx%d]", variableType.dereference(), variable, tableLevel)

	toLuaType := elem.ConvertGoTypeToLuaWithTableLevel(indexCode, tableLevel+1)
	indexReturn := toLuaType

	if _, ok := elem.Type.Underlying().(*types.Basic); ok && elem.PointerIndirection > 0 {
		primIndex := fmt.Sprintf("*(%s%s)[idx%d]", variableType.dereference(), variable, tableLevel)
		derefElem := elem
		derefElem.PointerIndirection = 0
		indexReturn = derefElem.ConvertGoTypeToLuaWithTableLevel(primIndex, tableLevel+1)
	}

	toGoType := elem.ConvertLuaTypeToGoWithTableLevel(fmt.Sprintf("t%d", tableLevel), fmt.Sprintf("val%d", tableLevel), 3, tableLevel+1)
	pointerIndirection := elem.ReferenceOrDereferenceForAssignmentToField()

	args := struct {
		Variable            string
		TableLevel          int
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
		TableLevel:          tableLevel,
		LuaType:             toLuaType,
		GoType:              toGoType,
		IndexCode:           indexCode,
		DeclaredGoType:      elem.declaredGoType(),
		PointerIndirection:  pointerIndirection,
		TemplateArg:         elem.TemplateArg(),
		VariableDereference: variableType.dereference(),
		IndexReturn:         indexReturn,
	}

	templ := `gobindlua.NewUserData(&gobindlua.GblSlice{
	Slice: {{ .Variable }},
	Len:   func() int { return len({{ .VariableDereference }}{{ .Variable }}) },
	Index: func(idx{{ .TableLevel }} int) lua.LValue { return {{ .IndexReturn }} },
	SetIndex: func(idx{{ .TableLevel }} int, val{{ .TableLevel }} lua.LValue) {
		{{ .GoType }}

		{{ .IndexCode }} = {{ .PointerIndirection }}t{{ .TableLevel }}
	},
}, L)`

	return execTempl(templ, args)
}

func convertGoTypeToLuaMap(keyType, valType types.Type, variableType *DataType, variable string, level int) string {
	key := CreateDataTypeFrom(keyType, variableType.packageSource, variableType.allDeclaredInterfaces)
	keyLuaType := key.ConvertGoTypeToLuaWithTableLevel(fmt.Sprintf("retKey%d", level), level+1)
	keyGoType := key.ConvertLuaTypeToGoWithTableLevel(fmt.Sprintf("keyVal%d", level), fmt.Sprintf("key%d", level), 3, level+1)
	keyPointerIndirection := key.ReferenceOrDereferenceForAssignmentToField()

	val := CreateDataTypeFrom(valType, variableType.packageSource, variableType.allDeclaredInterfaces)
	valLuaType := val.ConvertGoTypeToLuaWithTableLevel(fmt.Sprintf("ret%d", level), level+1)
	valGoType := val.ConvertLuaTypeToGoWithTableLevel(fmt.Sprintf("valVal%d", level), fmt.Sprintf("val%d", level), 3, level+1)
	valPointerIndirection := val.ReferenceOrDereferenceForAssignmentToField()

	indexReturn := valLuaType

	if _, ok := val.Type.Underlying().(*types.Basic); ok && val.PointerIndirection > 0 {
		derefVal := val
		derefVal.PointerIndirection = 0
		indexReturn = derefVal.ConvertGoTypeToLuaWithTableLevel(fmt.Sprintf("*(ret%d)", level), level+1)
	}

	args := struct {
		Variable              string
		TableLevel            int
		KeyLuaType            string
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
		TableLevel:            level,
		KeyLuaType:            keyLuaType,
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

	templ := `gobindlua.NewUserData(&gobindlua.GblMap{
Map: {{ .Variable }},
Len:   func() int { return len({{ .VariableDereference }}{{ .Variable }}) },
GetValue: func(key{{ .TableLevel }} lua.LValue) lua.LValue {
	{{ .KeyGoType }}
	ret{{ .TableLevel }} := ({{ .VariableDereference }}{{ .Variable }})[({{ .KeyTemplateArg }})({{ .KeyPointerIndirection }}keyVal{{ .TableLevel }})]
	return {{ .IndexReturn }}
},
SetValue: func(key{{ .TableLevel }} lua.LValue, val{{ .TableLevel }} lua.LValue) {
	{{ .KeyGoType }}
	{{ .ValGoType }}
	({{ .VariableDereference }}{{ .Variable }})[({{ .KeyTemplateArg }})({{ .KeyPointerIndirection }}keyVal{{ .TableLevel }})] = ({{ .ValTemplateArg }})({{ .ValPointerIndirection }}valVal{{ .TableLevel }})
},
ForEach: func(f{{ .TableLevel }} func(k{{ .TableLevel }}, v{{ .TableLevel }} lua.LValue)) {
	for k{{ .TableLevel }}_iter,v{{ .TableLevel }}_iter := range {{ .VariableDereference }}{{ .Variable }} {
		retKey{{ .TableLevel }} := k{{ .TableLevel }}_iter
		ret{{ .TableLevel }} := v{{ .TableLevel }}_iter
		key{{ .TableLevel }} := {{ .KeyLuaType }}
		val{{ .TableLevel }} := {{ .ValLuaType }}
		f{{ .TableLevel }}(key{{ .TableLevel }}, val{{ .TableLevel }})
	}
},
}, L)`

	return execTempl(templ, args)
}

func convertGoTypeToFunc(typ *types.Signature, variableType *DataType, variable string) string {
	fn := CreateFunction(typ, variable, "", "", variableType.packageSource, variableType.allDeclaredInterfaces)
	buf := bytes.Buffer{}
	fn.GenerateLuaFunctionWrapper(&buf, "")
	return fmt.Sprintf("L.NewFunction(%s)", strings.TrimSpace(buf.String()))
}

func (d *DataType) ReferenceOrDereferenceForAssignmentToField() string {
	goPointerLevel := d.PointerIndirection
	luaPointerLevel := 0

	if _, ok := d.Type.Underlying().(*types.Struct); ok {
		luaPointerLevel++
	}

	if _, ok := d.Type.Underlying().(*types.Array); ok {
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
	return d.convertLuaTypeToGo(variable, source, paramNum, 0, 0)
}

func (d *DataType) ConvertLuaTypeToGoWithTableLevel(variable string, source string, paramNum, tableLevel int) string {
	return d.convertLuaTypeToGo(variable, source, paramNum, tableLevel, 0)
}

func (d *DataType) ConvertLuaTypeToGoWithFunctionParam(variable string, source string, paramNum, functionParam int) string {
	return d.convertLuaTypeToGo(variable, source, paramNum, 0, functionParam)
}

func (d *DataType) convertLuaTypeToGo(variableToCreate string, luaVariable string, paramNum, tableLevel, functionParam int) string {
	switch t := d.Type.Underlying().(type) {
	case *types.Basic:
		return d.convertLuaTypeToGoPrimitive(variableToCreate, luaVariable, paramNum, tableLevel, functionParam)
	case *types.Array:
		return d.convertLuaTypeToGoSlice(t.Elem(), variableToCreate, luaVariable, paramNum, tableLevel, true)
	case *types.Slice:
		return d.convertLuaTypeToGoSlice(t.Elem(), variableToCreate, luaVariable, paramNum, tableLevel, false)
	case *types.Map:
		return d.convertLuaTypeToGoMap(t.Key(), t.Elem(), variableToCreate, luaVariable, paramNum, tableLevel)
	case *types.Struct:
		return d.convertLuaTypeToStruct(variableToCreate, luaVariable, paramNum, tableLevel)
	case *types.Interface:
		return d.convertLuaTypeToInterface(variableToCreate, luaVariable, paramNum, tableLevel, functionParam)
	case *types.Signature:
		return d.convertLuaTypeToFunc(t, variableToCreate, luaVariable, paramNum, tableLevel, functionParam)
	}

	return "CANNOT_CONVERT_LUA_TYPE_TO_GO"
}

func (d *DataType) convertLuaTypeToGoPrimitive(variableToCreate string, luaVariable string, paramNum, tableLevel, functionParam int) string {
	if tableLevel == 0 && functionParam == 0 {
		// Level 0 means the variable came from a L.Check*, which means it was already type-checked
		return fmt.Sprintf(`%s := %s(%s)`, variableToCreate, d.declaredGoType(), luaVariable)
	}

	args := struct {
		VariableToCreate string
		LuaVariable      string
		LuaType          string
		ParamNum         int
		DeclaredGoType   string
		TableLevel       int
		FunctionParam    int
		GenUtil          GenUtil
	}{
		VariableToCreate: variableToCreate,
		LuaVariable:      luaVariable,
		LuaType:          d.luaType(),
		ParamNum:         paramNum,
		DeclaredGoType:   d.declaredGoType(),
		TableLevel:       tableLevel,
		FunctionParam:    functionParam,
	}

	templ := `
{{ .VariableToCreate }}_n, ok := {{ .LuaVariable }}.({{ .LuaType }})

if !ok {
	{{ .GenUtil.GenerateCastError .FunctionParam .TableLevel .ParamNum  .DeclaredGoType .LuaVariable }}
}

{{ .VariableToCreate }} := {{ .DeclaredGoType }}({{ .VariableToCreate }}_n)
`
	return execTempl(templ, args)
}

func (d *DataType) convertLuaTypeToGoSlice(typ types.Type, variableToCreate string, luaVariable string, paramNum, tableLevel int, isArray bool) string {
	elem := CreateDataTypeFrom(typ, d.packageSource, d.allDeclaredInterfaces)
	toGoType := elem.ConvertLuaTypeToGoWithTableLevel(fmt.Sprintf("v%d", tableLevel), fmt.Sprintf("val%d", tableLevel), paramNum, tableLevel+1)
	pointerIndirection := elem.ReferenceOrDereferenceForAssignmentToField()

	args := struct {
		OriginalArrType    string
		VariableToCreate   string
		ActualGoType       string
		LuaVariable        string
		TableLevel         int
		ToGoType           string
		PointerIndirection string
		ParamNum           int
		TemplateArg        string
	}{
		OriginalArrType:    strings.TrimPrefix(d.ActualTemplateArg(), "*"),
		VariableToCreate:   variableToCreate,
		ActualGoType:       elem.ActualGoType(),
		LuaVariable:        luaVariable,
		TableLevel:         tableLevel,
		ToGoType:           toGoType,
		PointerIndirection: pointerIndirection,
		ParamNum:           paramNum,
		TemplateArg:        elem.TemplateArg(),
	}

	templ := ""

	if isArray {
		templ = `
{{ .VariableToCreate }}sl, err := gobindlua.MapLuaArrayOrTableToGoSlice[{{ .TemplateArg }}]({{ .LuaVariable }}, {{ .TableLevel }}, func(val{{ .TableLevel }} lua.LValue) {{ .TemplateArg }} {
	{{ .ToGoType }}
	return {{ .PointerIndirection }}v{{ .TableLevel }}
})

if err != nil {
	L.ArgError({{ .ParamNum }}, err.Error())
}

{{ .VariableToCreate }} := (*{{ .OriginalArrType }})({{ .VariableToCreate }}sl)
`
	} else {
		templ = `
{{ .VariableToCreate }}, err := gobindlua.MapLuaArrayOrTableToGoSlice[{{ .TemplateArg }}]({{ .LuaVariable }}, {{ .TableLevel }}, func(val{{ .TableLevel }} lua.LValue) {{ .TemplateArg }} {
	{{ .ToGoType }}
	return {{ .PointerIndirection }}v{{ .TableLevel }}
})

if err != nil {
	L.ArgError({{ .ParamNum }}, err.Error())
}
`
	}

	return execTempl(templ, args)
}

func (d *DataType) convertLuaTypeToGoMap(keyType, valueType types.Type, variableToCreate string, luaVariable string, paramNum, tableLevel int) string {
	k := CreateDataTypeFrom(keyType, d.packageSource, d.allDeclaredInterfaces)
	v := CreateDataTypeFrom(valueType, d.packageSource, d.allDeclaredInterfaces)
	keyGoType := k.ConvertLuaTypeToGoWithTableLevel(fmt.Sprintf("k%d", tableLevel), fmt.Sprintf("key%d", tableLevel), paramNum, tableLevel+1)
	valGoType := v.ConvertLuaTypeToGoWithTableLevel(fmt.Sprintf("v%d", tableLevel), fmt.Sprintf("val%d", tableLevel), paramNum, tableLevel+1)
	keyPointerIndirection := k.ReferenceOrDereferenceForAssignmentToField()
	valPointerIndirection := v.ReferenceOrDereferenceForAssignmentToField()

	args := struct {
		VariableToCreate      string
		KeyActualGoType       string
		ValActualGoType       string
		LuaVariable           string
		TableLevel            int
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
		TableLevel:            tableLevel,
		KeyGoType:             keyGoType,
		ValGoType:             valGoType,
		KeyPointerIndirection: keyPointerIndirection,
		ValPointerIndirection: valPointerIndirection,
		ParamNum:              paramNum,
		KeyTemplateArg:        k.TemplateArg(),
		ValTemplateArg:        v.TemplateArg(),
	}

	templ := `
{{ .VariableToCreate }}, err := gobindlua.MapLuaArrayOrTableToGoMap[{{ .KeyTemplateArg }}, {{ .ValTemplateArg }}]({{ .LuaVariable }}, {{ .TableLevel }}, func(key{{ .TableLevel }}, val{{ .TableLevel }} lua.LValue) ({{ .KeyTemplateArg }}, {{ .ValTemplateArg }}) {
	{{ .KeyGoType }}
	{{ .ValGoType }}
	return {{ .KeyPointerIndirection }}k{{ .TableLevel }}, {{ .ValPointerIndirection }}v{{ .TableLevel }}
})

if err != nil {
	L.ArgError({{ .ParamNum }}, err.Error())
}
`

	return execTempl(templ, args)
}

func (d *DataType) convertLuaTypeToStruct(variableToCreate string, luaVariable string, paramNum, tableLevel int) string {
	args := struct {
		VariableToCreate string
		LuaVariable      string
		DeclaredGoType   string
		ParamNum         int
		TableLevel       int
		GenUtil          GenUtil
	}{
		VariableToCreate: variableToCreate,
		LuaVariable:      luaVariable,
		DeclaredGoType:   d.declaredGoType(),
		ParamNum:         paramNum,
		TableLevel:       tableLevel,
	}

	if tableLevel == 0 {
		templ := `
{{ .VariableToCreate }}, ok := {{ .LuaVariable }}.Value.(*{{ .DeclaredGoType }})

if !ok {
	{{ .GenUtil.GenerateCastError 0 0 3  .DeclaredGoType .LuaVariable }}
}
`

		return execTempl(templ, args)
	}

	templ := `
{{ .VariableToCreate }}_ud, ok := {{ .LuaVariable }}.(*lua.LUserData)

if !ok {
	{{ .GenUtil.GenerateCastError 0 .TableLevel .ParamNum  .DeclaredGoType .LuaVariable }}
}

{{ .VariableToCreate }}, ok := {{ .VariableToCreate }}_ud.Value.(*{{ .DeclaredGoType }})

if !ok {
	{{ .GenUtil.GenerateCastError 0 .TableLevel 3  .DeclaredGoType .LuaVariable }}
}
`

	return execTempl(templ, args)
}

func (d *DataType) isEmptyInterface() bool {
	i, ok := d.Type.Underlying().(*types.Interface)

	if !ok {
		return false
	}

	return i.NumEmbeddeds() == 0
}

func (d *DataType) convertLuaTypeToInterface(variableToCreate string, luaVariable string, paramNum, tableLevel, functionParam int) string {
	args := struct {
		VariableToCreate string
		LuaVariable      string
		DeclaredGoType   string
		ParamNum         int
		TableLevel       int
		FunctionParam    int
		GenUtil          GenUtil
	}{
		VariableToCreate: variableToCreate,
		LuaVariable:      luaVariable,
		DeclaredGoType:   d.declaredGoType(),
		ParamNum:         paramNum,
		TableLevel:       tableLevel,
		FunctionParam:    functionParam,
	}

	if d.isEmptyInterface() {
		templ := `
{{ .VariableToCreate }} := gobindlua.UnwrapLValueToAny({{ .LuaVariable }})
`

		return execTempl(templ, args)
	}

	if tableLevel == 0 {
		templ := `
{{ .VariableToCreate }}, ok := {{ .LuaVariable }}.Value.({{ .DeclaredGoType }})

if !ok {
	{{ .GenUtil.GenerateCastError .FunctionParam .TableLevel .ParamNum  .DeclaredGoType .LuaVariable }}
}
`

		return execTempl(templ, args)
	}

	templ := `
{{ .VariableToCreate }}_ud, ok := {{ .LuaVariable }}.(*lua.LUserData)

if !ok {
	{{ .GenUtil.GenerateCastError .FunctionParam .TableLevel .ParamNum  .DeclaredGoType .LuaVariable }}
}

{{ .VariableToCreate }}, ok := {{ .VariableToCreate }}_ud.Value.({{ .DeclaredGoType }})

if !ok {
	{{ .GenUtil.GenerateCastError .FunctionParam .TableLevel .ParamNum  .DeclaredGoType .LuaVariable }}
}
`

	return execTempl(templ, args)
}

func (d *DataType) convertLuaTypeToFunc(typ *types.Signature, variableToCreate string, luaVariable string, paramNum int, tableLevel, functionParam int) string {
	type funcDataType struct {
		GoName  string
		LuaName string
		GetCall string
		GetPos  int
		*DataType
	}

	params := make([]*funcDataType, 0)
	results := make([]*funcDataType, 0)

	for i := 0; i < typ.Params().Len(); i++ {
		d := d.createDataTypeFrom(typ.Params().At(i).Type())
		goName := "p" + strconv.Itoa(i)
		luaName := goName + "l"
		params = append(params, &funcDataType{GoName: goName, LuaName: luaName, DataType: &d})
	}

	for i := 0; i < typ.Results().Len(); i++ {
		d := d.createDataTypeFrom(typ.Results().At(i).Type())
		goName := "r" + strconv.Itoa(i)
		luaName := goName + "l"
		getCall := fmt.Sprintf("L.Get(%d)", -(i + 1))
		getPos := i + 1
		results = append(results, &funcDataType{GoName: goName, LuaName: luaName, GetCall: getCall, GetPos: getPos, DataType: &d})
	}

	sig := "func("

	for i, a := range params {
		sig += a.GoName + " " + a.dereference() + a.ActualGoType()

		if i != len(params)-1 {
			sig += ", "
		}
	}

	sig += ") ("

	for i, r := range results {
		sig += r.ActualGoType()

		if i != len(results)-1 {
			sig += ", "
		}
	}

	sig += ")"

	args := struct {
		VariableToCreate string
		LuaVariable      string
		LuaType          string
		ParamNum         int
		DeclaredGoType   string
		TableLevel       int
		FunctionParam    int
		Signature        string
		FuncVariable     string
		Parameters       []*funcDataType
		Results          []*funcDataType
		GenUtil          GenUtil
	}{
		VariableToCreate: variableToCreate,
		LuaVariable:      luaVariable,
		LuaType:          d.luaType(),
		ParamNum:         paramNum,
		DeclaredGoType:   d.declaredGoType(),
		TableLevel:       tableLevel,
		FunctionParam:    functionParam,
		Signature:        sig,
		FuncVariable:     variableToCreate + "_lf",
		Parameters:       params,
		Results:          results,
	}

	templ := `
{{ $paramNum := .ParamNum }}
{{ .FuncVariable }}, ok := {{ .LuaVariable }}.({{ .LuaType }})

if !ok {
	{{ .GenUtil.GenerateCastError .FunctionParam .TableLevel .ParamNum  .DeclaredGoType .LuaVariable }}
}

{{ .VariableToCreate }} := {{ .Signature }} {
	L.Push({{ .FuncVariable }})

	{{ range $p := .Parameters }}
		L.Push({{ $p.ConvertGoTypeToLua $p.GoName }})
	{{ end }}

	L.Call({{ len .Parameters }}, {{ len .Results }})

	{{ range $i, $r := .Results }}
		{{ $r.ConvertLuaTypeToGoWithFunctionParam $r.LuaName $r.GetCall $paramNum $r.GetPos }}
	{{ end }}
	
	L.Pop({{ len .Results }})

	return {{ range $r := .Results }} {{ $r.LuaName }} {{ end }}
}
`

	return execTempl(templ, args)
}

func (d *DataType) ConvertLuaTypeToGoSliceEllipses(t *types.Slice, variableToCreate string, luaVariable string, paramNum int) string {
	level := 0
	elem := CreateDataTypeFrom(t.Elem(), d.packageSource, d.allDeclaredInterfaces)
	toGoType := elem.ConvertLuaTypeToGoWithTableLevel(fmt.Sprintf("v%d", level), fmt.Sprintf("val%d", level), paramNum, level+1)
	pointerIndirection := elem.ReferenceOrDereferenceForAssignmentToField()

	args := struct {
		VariableToCreate   string
		ActualGoType       string
		ParamNum           int
		TableLevel         int
		GoType             string
		PointerIndirection string
		TemplateArg        string
	}{
		VariableToCreate:   variableToCreate,
		ActualGoType:       elem.ActualGoType(),
		ParamNum:           paramNum,
		TableLevel:         level,
		GoType:             toGoType,
		PointerIndirection: pointerIndirection,
		TemplateArg:        elem.TemplateArg(),
	}

	templ := `
{{ .VariableToCreate }}, err := gobindlua.MapVariadicArgsToGoSlice[{{ .TemplateArg }}]({{ .ParamNum }}, L, func(val{{ .TableLevel }} lua.LValue) {{ .TemplateArg }} {
	{{ .GoType }}
	return {{ .PointerIndirection }}v{{ .TableLevel }}
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

func (d *DataType) ActualTemplateArg() string {
	indir := d.dereference()

	if _, ok := d.Type.(*types.Named); ok && !d.IsError() {
		return indir + d.declaredGoType()
	}

	switch t := d.Type.Underlying().(type) {
	case *types.Array:
		sl := CreateDataTypeFrom(t.Elem(), d.packageSource, d.allDeclaredInterfaces)
		return fmt.Sprintf("[%d]%s", t.Len(), sl.ActualTemplateArg())
	case *types.Slice:
		sl := CreateDataTypeFrom(t.Elem(), d.packageSource, d.allDeclaredInterfaces)
		return indir + "[]" + sl.ActualTemplateArg()
	case *types.Map:
		k := CreateDataTypeFrom(t.Key(), d.packageSource, d.allDeclaredInterfaces)
		v := CreateDataTypeFrom(t.Elem(), d.packageSource, d.allDeclaredInterfaces)
		return indir + "map[" + k.ActualTemplateArg() + "]" + v.ActualTemplateArg()
	}

	return indir + d.declaredGoType()
}

func (d *DataType) TemplateArg() string {
	indir := d.dereference()

	switch t := d.Type.Underlying().(type) {
	case *types.Array:
		sl := CreateDataTypeFrom(t.Elem(), d.packageSource, d.allDeclaredInterfaces)
		return indir + "[]" + sl.TemplateArg()
	case *types.Slice:
		sl := CreateDataTypeFrom(t.Elem(), d.packageSource, d.allDeclaredInterfaces)
		return indir + "[]" + sl.TemplateArg()
	case *types.Map:
		k := CreateDataTypeFrom(t.Key(), d.packageSource, d.allDeclaredInterfaces)
		v := CreateDataTypeFrom(t.Elem(), d.packageSource, d.allDeclaredInterfaces)
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
	case *types.Basic, *types.Slice, *types.Array:
		return d.Type.Underlying().String()
	case *types.Map:
		l := CreateDataTypeFrom(t.Key(), d.packageSource, d.allDeclaredInterfaces)
		v := CreateDataTypeFrom(t.Elem(), d.packageSource, d.allDeclaredInterfaces)

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
	case *types.Slice, *types.Array:
		return "*gobindlua.GblSlice"
	case *types.Map:
		return "*gobindlua.GblMap"
	case *types.Struct, *types.Interface:
		return "lua.LUserData"
	case *types.Signature:
		return "*lua.LFunction"
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
	case *types.Slice, *types.Array, *types.Map, *types.Signature:
		return "L.CheckAny"
	case *types.Struct, *types.Interface:
		return "L.CheckUserData"
	}

	return "UNSUPPORTED_TYPE"
}

func (d *DataType) LuaType(isFunctionReturn bool) string {
	switch t := d.Type.Underlying().(type) {
	case *types.Basic:
		switch d.ActualGoType() {
		case "bool":
			return "boolean"
		case "string":
			return "string"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "byte", "uint16", "uint32", "uint64", "float32", "float64":
			return "number"
		}
	case *types.Array:
		elem := d.createDataTypeFrom(t.Elem())
		return fmt.Sprintf("%s[]", elem.LuaType(isFunctionReturn))
	case *types.Slice:
		elem := d.createDataTypeFrom(t.Elem())

		/*
			TODO: https://github.com/LuaLS/lua-language-server/issues/1861
			Generic classes are not supported by LuaLS.  For now, just declare gbl_array's and gbl_map's as arrays/dictionaries.
			The problem with this approach is the language server and the actual implementation will start differing when you start
			passing gbl_array's and gbl_map's into non-gobindlua functions.  If it proves to be a big problem, we could change gobindlua
			to directly convert all Go slices and maps into tables everywhere.
			if isFunctionReturn {
				return fmt.Sprintf("%s<%s>", gobindlua.ARRAY_METATABLE_NAME, elem.LuaType(isFunctionReturn))
			} else {
				return fmt.Sprintf("(%[1]s<%[2]s> | %[2]s[])", gobindlua.ARRAY_METATABLE_NAME, elem.LuaType(isFunctionReturn))
			}
		*/
		return fmt.Sprintf("%s[]", elem.LuaType(isFunctionReturn))
	case *types.Map:
		key := d.createDataTypeFrom(t.Key())
		val := d.createDataTypeFrom(t.Elem())

		/*
			TODO: https://github.com/LuaLS/lua-language-server/issues/1861
			if isFunctionReturn {
				return fmt.Sprintf("%s<%s,%s>", gobindlua.MAP_METATABLE_NAME, key.LuaType(isFunctionReturn), val.LuaType(isFunctionReturn))
			} else {
				return fmt.Sprintf("(%[1]s<%[2]s,%[3]s> | table<%[2]s,%[2]s>)", gobindlua.MAP_METATABLE_NAME, key.LuaType(isFunctionReturn), val.LuaType(isFunctionReturn))
			}
		*/
		return fmt.Sprintf("table<%s,%s>", key.LuaType(isFunctionReturn), val.LuaType(isFunctionReturn))
	case *types.Struct:
		return gobindluautil.StructFieldMetadataName(d.declaredGoTypeWithoutPackage())
	case *types.Interface:
		for _, i := range d.allDeclaredInterfaces {
			if types.Identical(t, i.Interface) {
				return gobindluautil.LookupCustomName(d.declaredGoTypeWithoutPackage())
			}
		}

		return "any"
	case *types.Signature:
		return "function"
	}

	return "UNSUPPORTED_TYPE"
}

func (d *DataType) declaredGoTypeWithoutPackage() string {
	pkg := d.declaredGoType()
	spl := strings.SplitN(pkg, ".", 2)

	if len(spl) == 2 {
		return spl[1]
	}

	return pkg
}

func (d *DataType) createDataTypeFrom(t types.Type) DataType {
	return CreateDataTypeFrom(t, d.packageSource, d.allDeclaredInterfaces)
}

type GenUtil struct{}

func (GenUtil) GenerateCastError(funcParam, tableLevel, assignNum int, declaredGoType, luaVariable string) string {
	if funcParam > 0 {
		return fmt.Sprintf(`gobindlua.FuncResCastError(L, %d, "%s", %s)`, funcParam, declaredGoType, luaVariable)
	} else if tableLevel > 0 {
		return fmt.Sprintf(`gobindlua.TableElemCastError(L, %d, "%s", %s)`, tableLevel, declaredGoType, luaVariable)
	}

	return fmt.Sprintf(`gobindlua.CastArgError(L, %d, "%s", %s)`, assignNum, declaredGoType, luaVariable)
}

func (GenUtil) Concat(l, r string) string {
	return l + r
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

func execTemplString(out io.Writer, data any, templ string) {
	t := template.Must(template.New("").Parse(templ))
	err := t.Execute(out, data)

	if err != nil {
		panic(err)
	}
}
