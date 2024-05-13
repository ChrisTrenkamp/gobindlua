package main

import (
	"bytes"
	"io"
	"strings"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
)

func (g *StructGenerator) generateLuaDefinitions() ([]byte, error) {
	ret := bytes.Buffer{}

	if g.structToGenerate != "" {
		g.generateLuaStructDefinition(&ret)
	} else {
		g.generateLuaPackageDefinition(&ret)
	}

	return ret.Bytes(), nil
}

func (g *StructGenerator) generateLuaStructDefinition(w io.Writer) {
	templ := `---Code generated by gobindlua.  DO NOT EDIT.
---@meta {{ .StructMetatableIdentifier }}
{{- $gen := . }}

local {{ $gen.StructMetatableIdentifier }} = {}
{{ range $fidx,$staticFunc := .StaticFunctions -}}
{{ $staticFunc.GenerateLuaFunctionParamRetDefinitions -}}
function {{ $gen.StructMetatableIdentifier }}.{{ $staticFunc.LuaFnName }}({{ $staticFunc.GenerateLuaFunctionParamStubs }}) end
{{ end -}}

{{- $fieldIdent := .StructMetatableFieldsIdentifier }}
---@class {{ $fieldIdent }}{{ .GenerateInterfaceDeclarations }}
{{- range $fidx,$field := .Fields }}
---@field public {{ $field.LuaName }} {{ $field.DataType.LuaType true }}
{{- end }}
local {{ $fieldIdent }} = {}
{{- range $midx,$methodFunc := .UserDataMethods }}
{{ $methodFunc.GenerateLuaFunctionParamRetDefinitions -}}
function {{ $fieldIdent }}:{{ $methodFunc.LuaFnName }}({{ $methodFunc.GenerateLuaFunctionParamStubs }}) end
{{- end }}

return {{ $gen.StructMetatableIdentifier }}
`

	execTempl(w, g, templ)
}

func (g *StructGenerator) GenerateInterfaceDeclarations() string {
	if len(g.implementsDeclarations) == 0 {
		return ""
	}

	ret := make([]string, 0)

	for _, i := range g.implementsDeclarations {
		ret = append(ret, gobindluautil.StructOrInterfaceMetadataName(i))
	}

	return " : " + strings.Join(ret, ", ")
}

func (g *StructGenerator) generateLuaPackageDefinition(w io.Writer) {
	templ := `---Code generated by gobindlua.  DO NOT EDIT.
---@meta {{ .PackageToGenerateMetatableName }}
{{- $gen := . }}

local {{ $gen.PackageToGenerateMetatableName }} = {}

{{- range $fidx,$staticFunc := .StaticFunctions -}}
{{ $staticFunc.GenerateLuaFunctionParamRetDefinitions -}}
function {{ $gen.PackageToGenerateMetatableName }}.{{ $staticFunc.LuaFnName }}({{ $staticFunc.GenerateLuaFunctionParamStubs }}) end
{{ end }}
return {{ $gen.PackageToGenerateMetatableName }}
`

	execTempl(w, g, templ)
}
