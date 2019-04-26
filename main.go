package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
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

		err := generateStruct(os.Stdout, name, structType.Fields)
		if err != nil {
			panic(err)
		}
		// Interpolate name into template string for the struct definition
		// `struct Type Name {}`
		// recursively Iterate over each field and interpolate name and respective types to c definitions

	}
}

func generateStruct(w io.Writer, name string, fields *ast.FieldList) error {
	const (
		structBegin = "struct %s {\n"
		fieldFormat = "\t%s %s\n"
		end         = "}\n\n"
	)

	_, err := fmt.Fprintf(w, structBegin, name)
	if err != nil {
		return err
	}
	for _, field := range fields.List {
		_, err = fmt.Fprintf(w, fieldFormat, field.Names[0].Name, "butts")
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(w, end)
	if err != nil {
		return err
	}

	return nil
}
