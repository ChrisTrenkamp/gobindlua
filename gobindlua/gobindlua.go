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

	"github.com/ChrisTrenkamp/gobindlua/gobindlua/gobindluautil"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/interfacegen"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/packagegen"
	"github.com/ChrisTrenkamp/gobindlua/gobindlua/structgen"
)

var errStructOrPackageUnspecified = fmt.Errorf("-s or -p must be specified")
var errIncorrectGoGeneratePlacement = fmt.Errorf("go:generate gobindlua directives must be placed behind a struct or package declaration")
var gofile = os.Getenv("GOFILE")

func determineFromGoLine(structToGenerate, packageToGenerate, interfaceToGenerate *string) error {
	lineStr := os.Getenv("GOLINE")

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

		if bytes.HasPrefix(splLine, []byte("//go:generate")) {
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
	workingDir := flag.String("d", "", "The Go source directory to generate the bindings from. Uses the current working directory if empty.")
	structToGenerate := flag.String("struct", "", "Generate GopherLua bindings and Lua definitions for the given struct.")
	packageToGenerate := flag.String("package", "", "Generate GopherLua bindings and Lua definitions for the given package.")
	interfaceToGenerate := flag.String("interface", "", "Generate Lua definitions for the given interface.")

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

	if *workingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("error getting working directory: %s", err)
		}
		*workingDir = wd
	}

	dependantModules, err := findGobindLuaConf(*workingDir)

	outFile := strings.TrimSuffix(filepath.Base(gofile), ".go")

	if *structToGenerate != "" {
		outFile += "_" + *structToGenerate
	} else if *interfaceToGenerate != "" {
		outFile += "_" + *interfaceToGenerate
	}

	outFile += "_lua_bindings"

	basePathToOutput := filepath.Join(*workingDir, outFile)

	var goBytes []byte
	var luaDefBytes []byte

	if *structToGenerate != "" {
		gen := structgen.NewStructGenerator(
			*structToGenerate,
			*workingDir,
			basePathToOutput+".go",
			dependantModules,
		)
		goBytes, luaDefBytes, err = gen.GenerateSourceCode()
	} else if *packageToGenerate != "" {
		gen := packagegen.NewPackageGenerator(
			*packageToGenerate,
			*workingDir,
			basePathToOutput+".go", dependantModules,
		)
		goBytes, luaDefBytes, err = gen.GenerateSourceCode()
	} else if *interfaceToGenerate != "" {
		gen := interfacegen.NewInterfaceGenerator(*interfaceToGenerate, *workingDir, dependantModules)
		luaDefBytes, err = gen.GenerateSourceCode()
	}

	if len(goBytes) > 0 {
		outPath := basePathToOutput + ".go"
		if werr := writeToFile(outPath, goBytes); werr != nil {
			log.Fatal(werr)
		}
	}

	if len(luaDefBytes) > 0 {
		outPath := basePathToOutput + ".lua"
		if werr := writeToFile(outPath, luaDefBytes); werr != nil {
			log.Fatal(werr)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}

func writeToFile(outPath string, content []byte) error {
	origContents, err := os.ReadFile(outPath)

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	str := string(bytes.Split(origContents, []byte("\n"))[0])

	if str != "" && !strings.Contains(str, gobindluautil.GEN_HEADER) {
		return fmt.Errorf("%s does not have the gobindlua header and will not overwrite", outPath)
	}

	return os.WriteFile(outPath, content, 0644)
}

func findGobindLuaConf(wd string) ([]string, error) {
	for {
		modFile := filepath.Join(wd, "go.mod")
		_, err := os.Stat(modFile)

		if os.IsNotExist(err) {
			wd = filepath.Dir(wd)
			continue
		}

		if err != nil {
			return nil, err
		}

		break
	}

	confFile := filepath.Join(wd, "gobindlua-conf.txt")
	fileBytes, err := os.ReadFile(confFile)

	if os.IsNotExist(err) {
		return nil, nil
	}

	fileString := strings.Replace(string(fileBytes), "\r", "", -1)
	fileString = strings.TrimSpace(fileString)
	ret := strings.Split(fileString, "\n")
	return ret, err
}
