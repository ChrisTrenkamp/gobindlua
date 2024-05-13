package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/interfacegen"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/packagegen"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/structgen"
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

	if *structToGenerate != "" {
		gen := structgen.NewStructGenerator(
			*structToGenerate,
			*workingDir,
			basePathToOutput+".go",
			includeFunctions,
			excludeFunctions,
			implementsDeclarations,
		)
		goBytes, luaDefBytes, err = gen.GenerateSourceCode()
	} else if *packageToGenerate != "" {
		gen := packagegen.NewPackageGenerator(
			*packageToGenerate,
			*workingDir,
			basePathToOutput+".go",
			includeFunctions,
			excludeFunctions,
		)
		goBytes, luaDefBytes, err = gen.GenerateSourceCode()
	} else if *interfaceToGenerate != "" {
		gen := interfacegen.NewInterfaceGenerator(*interfaceToGenerate, *workingDir)
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
