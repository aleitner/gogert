package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type set map[string]struct{}

type structTypeMap map[string]*ast.StructType

type CStructMeta struct {
	name              string
	dependencies      []string
	structDeclaration bytes.Buffer
}

func main() {
	input := flag.String("input", ".", "input directory")
	output := flag.String("output", "stdout", "output directory")

	flag.Parse()

	fi, err := os.Stat(*input)
	if os.IsNotExist(err) {
		log.Fatal(err)
	}

	if !fi.IsDir() {
		log.Fatal(fmt.Errorf("Specified Input Path is not a directory: %s", *output))
	}

	var outputFile *os.File
	if *output == "stdout" {
		outputFile = os.Stdout
	} else {
		outputPath := path.Join(*output, "/cstructs.h")

		fi, err := os.Stat(*output)
		if os.IsNotExist(err) {
			log.Fatal(err)
		}

		if !fi.IsDir() {
			log.Fatal(fmt.Errorf("Specified Output Path is not a directory: %s", *output))
		}

		outputFile, err = os.Create(outputPath)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := outputFile.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	w := bufio.NewWriter(outputFile)
	if err := Parse(w, *input); err != nil {
		log.Fatal(err)
	}
	w.Flush()
}

// Parse through a directory for exported go structs to be converted
func Parse(w io.Writer, path string) error {
	var cStructs []*CStructMeta

	// Walk through all directories
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				return nil
			}

			fset := token.NewFileSet()
			astPkgs, err := parser.ParseDir(fset, path, func(info os.FileInfo) bool {
				name := info.Name()
				return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
			}, parser.ParseComments)
			if err != nil {
				return err
			}

			// Loop through every file and convert the exported go structs
			for _, pkg := range astPkgs {
				for _, file := range pkg.Files {
					structNames := getStructNames(file)
					structs := findStructs(file, structNames)
					cStructs = append(cStructs, fromGoStructs(structs, fset)...)
				}
			}

			return nil
		})
	if err != nil {
		return err
	}

	// organize the cstructs
	sort.Slice(cStructs, func(i, j int) bool {
		return len(cStructs[i].dependencies) < len(cStructs[j].dependencies)
	})

	for _, cstruct := range cStructs {
		fmt.Fprintf(w, cstruct.structDeclaration.String())
	}

	return nil
}

// generateStructRecursivewill generate a cstruct and any cstruct dependencies
func generateStructRecursive(fset *token.FileSet, name string, fields *ast.FieldList) (cstructs []*CStructMeta, err error) {
	const (
		structBegin = "#ifndef CSTRUCTS_%s\n\tstruct %s {\n"
		fieldFormat = "\t\t%s %s; // gotype: %s\n"
		end         = "\t};\n#endif\n\n"
	)

	cstructs = []*CStructMeta{}
	cstruct := &CStructMeta{name: name}
	cstructs = append(cstructs, cstruct)

	_, err = fmt.Fprintf(&cstruct.structDeclaration, structBegin, name, name)
	if err != nil {
		return cstructs, err
	}

	for _, field := range fields.List {
		var typeNameBuf bytes.Buffer
		err := printer.Fprint(&typeNameBuf, fset, field.Type)
		if err != nil {
			return cstructs, err
		}

		ctype, dependencies := fromGoType(typeNameBuf.String())
		cstructs = append(cstructs, dependencies...)
		if strings.Contains(ctype, "struct") {
			cstruct.dependencies = append(cstruct.dependencies, ctype)
		}
		_, err = fmt.Fprintf(&cstruct.structDeclaration, fieldFormat, ctype, field.Names[0].Name, typeNameBuf.String())
		if err != nil {
			return cstructs, err
		}
	}

	_, err = fmt.Fprintf(&cstruct.structDeclaration, end)
	if err != nil {
		return cstructs, err
	}

	return cstructs, nil
}

func getStructNames(file *ast.File) set {
	// Parse comments for exported structs
	structNames := make(set)
	ast.Inspect(file, func(n ast.Node) bool {
		// collect comments
		c, ok := n.(*ast.CommentGroup)
		if ok {
			re, _ := regexp.Compile(`(?:\n|^)CExport\s+(\w+)`)
			matches := re.FindStringSubmatch(c.Text())
			if len(matches) > 1 {
				structNames[strings.TrimSpace(matches[1])] = struct{}{}
			}
		}
		return true
	})
	return structNames
}

func findStructs(file *ast.File, structNames set) structTypeMap {
	structs := make(structTypeMap)
	ast.Inspect(file, func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if t.Type == nil {
			return true
		}

		structName := t.Name.Name

		_, ok = structNames[structName]
		if !ok {
			return true
		}

		x, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structs[structName] = x

		return true
	})
	return structs
}

func fromGoStructs(structs structTypeMap, fset *token.FileSet) []*CStructMeta {
	var cStructs []*CStructMeta
	for name, structType := range structs {
		cstructRecursive, err := generateStructRecursive(fset, name, structType.Fields)
		if err != nil {
			log.Fatal(err)
		}

		cStructs = append(cStructs, cstructRecursive...)
	}
	return cStructs
}
