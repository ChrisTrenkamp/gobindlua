package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/go/packages"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/datatype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/functiontype"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/structfield"
)

type flagArray []string

func (i *flagArray) String() string {
	return ""
}

func (i *flagArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var errStructOrPackageUnspecified = fmt.Errorf("-s or -p must be specified")
var errIncorrectGoGeneratePlacement = fmt.Errorf("go:generate gobindlua directives must be placed behind a struct or package declaration")

func determineFromGoLine(structToGenerate, packageToGenerate, interfaceToGenerate *string) error {
	lineStr := os.Getenv("GOLINE")
	gofile := os.Getenv("GOFILE")

	if lineStr == "" || gofile == "" {
		return errStructOrPackageUnspecified
	}

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return err
	}

	f, err := os.ReadFile(gofile)
	if err != nil {
		return err
	}

	spl := bytes.Split(f, []byte("\n"))
	var splLine []byte

	for {
		if len(spl) < line {
			return errIncorrectGoGeneratePlacement
		}

		splLine = spl[line]
		line++

		splLine = bytes.TrimSpace(splLine)

		if len(splLine) == 0 || bytes.HasPrefix(splLine, []byte("//")) {
			continue
		}

		break
	}

	norm := regexp.MustCompile(`\s+`).ReplaceAllString(string(splLine), " ")
	normSpl := strings.Split(norm, " ")

	if len(normSpl) < 2 {
		return errIncorrectGoGeneratePlacement
	}

	switch normSpl[0] {
	case "type":
		if len(normSpl) < 3 {
			return errIncorrectGoGeneratePlacement
		}

		if strings.HasPrefix(normSpl[2], "struct") {
			*structToGenerate = normSpl[1]
			return nil
		}

		if strings.HasPrefix(normSpl[2], "interface") {
			*interfaceToGenerate = normSpl[1]
			return nil
		}
	case "package":
		*packageToGenerate = normSpl[1]
		return nil
	}

	return errIncorrectGoGeneratePlacement
}

func numNotEmpty(str ...*string) int {
	ret := 0

	for _, i := range str {
		if *i != "" {
			ret++
		}
	}

	return ret
}

func main() {
	includeFunctions := make(flagArray, 0)
	excludeFunctions := make(flagArray, 0)
	implementsDeclarations := make(flagArray, 0)
	workingDir := flag.String("d", "", "The Go source directory to generate the bindings from. Uses the current working directory if empty.")
	structToGenerate := flag.String("struct", "", "Generate GopherLua bindings and Lua definitions for the given struct.")
	packageToGenerate := flag.String("package", "", "Generate GopherLua bindings and Lua definitions for the given package.")
	interfaceToGenerate := flag.String("interface", "", "Generate Lua definitions for the given interface.")
	flag.Var(&includeFunctions, "i", "Only include the given function or method names.")
	flag.Var(&excludeFunctions, "x", "Exclude the given function or method names.")
	flag.Var(&implementsDeclarations, "im", "Declares the given struct implements an interface.")

	flag.Parse()

	if flag.NArg() != 0 {
		log.Fatal("gobindlua does not accept arguments")
	}

	if *structToGenerate == "" && *packageToGenerate == "" && *interfaceToGenerate == "" {
		if err := determineFromGoLine(structToGenerate, packageToGenerate, interfaceToGenerate); err != nil {
			log.Fatal(err.Error())
		}
	}

	if numNotEmpty(structToGenerate, packageToGenerate, interfaceToGenerate) != 1 {
		log.Fatal("only one of -struct, -package, or -interface may be specified")
	}

	if len(includeFunctions) > 0 && len(excludeFunctions) > 0 {
		log.Fatal("only one of -i or -x may be specified")
	}

	if *workingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("error getting working directory: %s", err)
		}
		*workingDir = wd
	}

	outFile := ""

	if *structToGenerate != "" {
		outFile = "lua_" + *structToGenerate
	} else if *packageToGenerate != "" {
		outFile = "lua_" + filepath.Base(*packageToGenerate)
	} else {
		outFile = "lua_" + *interfaceToGenerate
	}

	basePathToOutput := filepath.Join(*workingDir, outFile)

	var goBytes []byte
	var luaDefBytes []byte
	var err error

	if *structToGenerate != "" || *packageToGenerate != "" {
		gen := NewStructGenerator(
			*structToGenerate,
			*packageToGenerate,
			*workingDir,
			basePathToOutput+".go",
			includeFunctions,
			excludeFunctions,
			implementsDeclarations,
		)
		goBytes, luaDefBytes, err = gen.GenerateSourceCode()
	} else if *interfaceToGenerate != "" {
		gen := NewInterfaceGenerator(*interfaceToGenerate, *workingDir)
		luaDefBytes, err = gen.GenerateSourceCode()
	}

	if len(goBytes) > 0 {
		outPath := basePathToOutput + ".go"
		if werr := os.WriteFile(outPath, goBytes, 0644); werr != nil {
			log.Fatal(werr)
		}
	}

	if len(luaDefBytes) > 0 {
		outPath := basePathToOutput + "_definitions.lua"
		if werr := os.WriteFile(outPath, luaDefBytes, 0644); werr != nil {
			log.Fatal(werr)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}

type StructGenerator struct {
	structToGenerate       string
	packageToGenerate      string
	wd                     string
	pathToOutput           string
	includeFunctions       []string
	excludeFunctions       []string
	implementsDeclarations []string

	packageSource *packages.Package
	structObject  types.Object

	StaticFunctions []functiontype.FunctionType
	UserDataMethods []functiontype.FunctionType
	Fields          []structfield.StructField

	imports               map[string]string
	metatableInitFunction bytes.Buffer
	metatableFunctions    bytes.Buffer
	userDataFunctions     bytes.Buffer
}

func NewStructGenerator(structToGenerate, packageToGenerate, wd, pathToOutput string, includeFunctions, excludeFunctions, implementsDeclarations []string) *StructGenerator {
	return &StructGenerator{
		structToGenerate:       structToGenerate,
		packageToGenerate:      packageToGenerate,
		wd:                     wd,
		pathToOutput:           pathToOutput,
		includeFunctions:       includeFunctions,
		excludeFunctions:       excludeFunctions,
		implementsDeclarations: implementsDeclarations,
	}
}

func (g *StructGenerator) GenerateSourceCode() ([]byte, []byte, error) {
	if err := g.loadSourcePackage(); err != nil {
		return nil, nil, fmt.Errorf("failed to load imported go packages: %s", err)
	}

	g.StaticFunctions = g.gatherFunctionsToGenerate()
	g.UserDataMethods = g.gatherReceivers()
	g.Fields = g.gatherFields()

	g.imports = make(map[string]string)
	g.imports["github.com/yuin/gopher-lua"] = "lua"
	g.imports["github.com/ChrisTrenkamp/gobindlua"] = ""

	g.addPackageFromFunctions(g.StaticFunctions)
	g.addPackageFromFunctions(g.UserDataMethods)

	if g.IsGeneratingStruct() {
		for _, i := range g.Fields {
			g.addPackage(i.DataType)
		}
	}

	g.buildMetatableInitFunction()
	g.buildMetatableFunctions()
	g.builderUserDataFunctions()

	goCode, gerr := g.generateGoCode()
	luaDef, lerr := g.generateLuaDefinitions()
	return goCode, luaDef, errors.Join(gerr, lerr)
}

func (g *StructGenerator) addPackageFromFunctions(f []functiontype.FunctionType) {
	for _, i := range f {
		for _, p := range i.Params {
			g.addPackage(p.DataType)
		}

		for _, p := range i.Ret {
			g.addPackage(p)
		}
	}
}

func (g *StructGenerator) addPackage(d datatype.DataType) {
	if p := d.Package(); p != "" && p != g.packageSource.ID {
		g.imports[p] = ""
	}
}

func (g *StructGenerator) loadSourcePackage() error {
	var err error
	g.packageSource, err = gobindluautil.LoadSourcePackage(g.wd)

	if err != nil {
		return err
	}

	if g.IsGeneratingStruct() {
		g.structObject = g.packageSource.Types.Scope().Lookup(g.structToGenerate)

		if g.structObject == nil {
			return fmt.Errorf("specified type not found")
		}

		if _, ok := g.structObject.Type().Underlying().(*types.Struct); !ok {
			return fmt.Errorf("specified type is not a struct")
		}
	}

	return nil
}

func (g *StructGenerator) StructToGenerate() string {
	return g.structObject.Name()
}

func (g *StructGenerator) PackageToGenerateFunctionName() string {
	caser := cases.Title(language.English)
	return "Register" + caser.String(g.packageToGenerate) + "LuaType"
}

func (g *StructGenerator) PackageToGenerateMetatableName() string {
	return g.packageToGenerate
}

func (g *StructGenerator) StructMetatableIdentifier() string {
	return gobindluautil.SnakeCase(g.StructToGenerate())
}

func (g *StructGenerator) StructMetatableFieldsIdentifier() string {
	return gobindluautil.StructFieldMetadataName(g.StructToGenerate())
}

func (g *StructGenerator) UserDataCheckFn() string {
	return "luaCheck" + g.StructToGenerate()
}

func (g *StructGenerator) SourceUserDataAccess() string {
	return "luaAccess" + g.StructToGenerate()
}

func (g *StructGenerator) SourceUserDataSet() string {
	return "luaSet" + g.StructToGenerate()
}

func (g *StructGenerator) IsGeneratingStruct() bool {
	return g.structToGenerate != ""
}

func (g *StructGenerator) gatherFunctionsToGenerate() []functiontype.FunctionType {
	if g.IsGeneratingStruct() {
		return g.gatherConstructors()
	}

	return g.gatherAllFunctions()
}

func (g *StructGenerator) hasFilters() bool {
	return len(g.includeFunctions) > 0 || len(g.excludeFunctions) > 0
}

func (g *StructGenerator) checkInclude(str string) bool {
	if len(g.includeFunctions) > 0 {
		return slices.Contains(g.includeFunctions, str)
	}

	if len(g.excludeFunctions) > 0 {
		return !slices.Contains(g.excludeFunctions, str)
	}

	return true
}

func (g *StructGenerator) gatherConstructors() []functiontype.FunctionType {
	ret := make([]functiontype.FunctionType, 0)
	underylingStructType := g.structObject.Type().Underlying()
	constructorPrefix := "New" + g.structToGenerate

	for _, syn := range g.packageSource.Syntax {
		for _, dec := range syn.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok && fn.Type.Results != nil {
				for _, retType := range fn.Type.Results.List {
					retType := datatype.CreateDataTypeFromExpr(retType.Type, g.packageSource)

					if retType.Type.Underlying() == underylingStructType {
						fnName := fn.Name.Name

						if strings.HasPrefix(fnName, "luaCheck") {
							continue
						}

						if g.hasFilters() {
							if !g.checkInclude(fnName) {
								continue
							}
						} else if !strings.HasPrefix(fnName, constructorPrefix) {
							continue
						}

						luaName := gobindluautil.SnakeCase(fnName)

						if strings.HasPrefix(fnName, constructorPrefix) {
							luaName = "New" + strings.TrimPrefix(fnName, constructorPrefix)
							luaName = gobindluautil.SnakeCase(luaName)
						}

						sourceCodeName := "luaConstructor" + g.StructToGenerate() + fnName
						ret = append(ret, functiontype.CreateFunction(fn, false, luaName, sourceCodeName, g.packageSource))
						break
					}
				}
			}
		}
	}

	return ret
}

func (g *StructGenerator) gatherAllFunctions() []functiontype.FunctionType {
	ret := make([]functiontype.FunctionType, 0)

	for _, syn := range g.packageSource.Syntax {
		for _, dec := range syn.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok {
				fnName := fn.Name.Name

				if g.hasFilters() {
					if !g.checkInclude(fnName) {
						continue
					}
				} else if !unicode.IsUpper(rune(fnName[0])) || g.PackageToGenerateFunctionName() == fnName {
					continue
				}

				luaName := gobindluautil.SnakeCase(fnName)
				sourceCodeName := "luaFunction" + fnName
				ret = append(ret, functiontype.CreateFunction(fn, false, luaName, sourceCodeName, g.packageSource))
			}
		}
	}

	return ret
}

func (g *StructGenerator) gatherReceivers() []functiontype.FunctionType {
	if !g.IsGeneratingStruct() {
		return nil
	}

	ret := make([]functiontype.FunctionType, 0)
	underylingStructType := g.structObject.Type().Underlying()

	for _, syn := range g.packageSource.Syntax {
		for _, dec := range syn.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok && fn.Recv != nil {
				fnName := fn.Name.Name

				if fnName == "RegisterLuaType" || fnName == "LuaMetatableType" {
					continue
				}

				if g.hasFilters() {
					if !g.checkInclude(fnName) {
						continue
					}
				} else if !unicode.IsUpper(rune(fnName[0])) {
					continue
				}

				for _, recType := range fn.Recv.List {
					recType := datatype.CreateDataTypeFromExpr(recType.Type, g.packageSource)

					if recType.Type.Underlying() == underylingStructType {
						luaName := gobindluautil.SnakeCase(fnName)
						sourceCodeName := "luaMethod" + g.StructToGenerate() + fnName
						ret = append(ret, functiontype.CreateFunction(fn, true, luaName, sourceCodeName, g.packageSource))
						break
					}
				}
			}
		}
	}

	return ret
}

func (g *StructGenerator) gatherFields() []structfield.StructField {
	if !g.IsGeneratingStruct() {
		return nil
	}

	ret := make([]structfield.StructField, 0)
	str := g.structObject.Type().Underlying().(*types.Struct)

	for i := 0; i < str.NumFields(); i++ {
		field := str.Field(i)
		tag := str.Tag(i)
		luaName := structfield.GetLuaNameFromTag(field, tag)

		if luaName != "" {
			ret = append(ret, structfield.CreateStructField(field, luaName, g.packageSource))
		}
	}

	return ret
}
