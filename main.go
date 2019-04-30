package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"regexp"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "samples/sample.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Parse comments for exported structs
	var structNames = map[string](struct{}){}
	ast.Inspect(file, func(n ast.Node) bool {
		// collect comments
		c, ok := n.(*ast.CommentGroup)
		if ok {
			re, _ := regexp.Compile(`(?:\n|^)CExport\s+(\w+)`)
			matches := re.FindStringSubmatch(c.Text())
			if len(matches) > 1 {
				structNames[strings.TrimSpace(matches[1])] = (struct{}{})
			}

		}

		return true
	})

	// Find structs
	structs := make(map[string]*ast.StructType, len(structNames))
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

	for name, structType := range structs {

		cstruct, err := generateStruct(fset, name, structType.Fields)
		fmt.Println(cstruct.structDeclaration.String())
		if err != nil {
			panic(err)
		}

	}
}

type CstructMeta struct {
	name              string
	dependencies      []string
	structDeclaration bytes.Buffer
}

func generateStruct(fset *token.FileSet, name string, fields *ast.FieldList) (cstruct *CstructMeta, err error) {
	const (
		structBegin = "#ifndef CSTRUCTS_%s\n\tstruct %s {\n"
		fieldFormat = "\t\t%s %s; // gotype: %s\n"
		end         = "\t};\n#endif\n\n"
	)

	cstruct = &CstructMeta{name: name}

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
