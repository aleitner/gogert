package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"regexp"
	"sort"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "samples/main.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	structNames := getStructNames(file)
	structs := findStructs(file, structNames)
	cStructs := fromGoStructs(structs, fset)

	sort.Slice(cStructs, func(i, j int) bool {
		return len(cStructs[i].dependencies) < len(cStructs[j].dependencies)
	})

	cStructsJSON, err := json.MarshalIndent(cStructs, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("cStructs: %s\n", cStructsJSON)
}

type set map[string]struct{}

type structTypeMap map[string]*ast.StructType

type CStructMeta struct {
	name              string
	dependencies      []string
	structDeclaration bytes.Buffer
}

func generateStruct(fset *token.FileSet, name string, structs map[string]*ast.StructType, fields *ast.FieldList) (cstruct *CStructMeta, err error) {
	const (
		structBegin = "#ifndef CSTRUCTS_%s\n\tstruct %s {\n"
		fieldFormat = "\t\t%s %s; // gotype: %s\n"
		end         = "\t};\n#endif\n\n"
	)

	cstruct = &CStructMeta{name: name}

	_, err = fmt.Fprintf(&cstruct.structDeclaration, structBegin, name, name)
	if err != nil {
		return cstruct, err
	}

	for _, field := range fields.List {
		var typeNameBuf bytes.Buffer
		err := printer.Fprint(&typeNameBuf, fset, field.Type)
		if err != nil {
			return cstruct, err
		}

		ctype := fromGoType(typeNameBuf.String())
		if strings.Contains(ctype, "struct") {
			cstruct.dependencies = append(cstruct.dependencies, typeNameBuf.String())
		}
		_, err = fmt.Fprintf(&cstruct.structDeclaration, fieldFormat, ctype, field.Names[0].Name, typeNameBuf.String())
		if err != nil {
			return cstruct, err
		}
	}

	_, err = fmt.Fprintf(&cstruct.structDeclaration, end)
	if err != nil {
		return cstruct, err
	}

	return cstruct, nil
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
		cstruct, err := generateStruct(fset, name, structs, structType.Fields)
		if err != nil {
			panic(err)
		}

		fmt.Println(cstruct.structDeclaration.String())
		cStructs = append(cStructs, cstruct)
	}
	return cStructs
}
