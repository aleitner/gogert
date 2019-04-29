package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
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

		err := generateStruct(fset, os.Stdout, name, structType.Fields)
		if err != nil {
			panic(err)
		}
		// Interpolate name into template string for the struct definition
		// `struct Type Name {}`
		// recursively Iterate over each field and interpolate name and respective types to c definitions

	}
}

func FromGoType(gotype string) string {
	switch gotype {
	case "string":
		return "uint8_t*"
	case "bool":
		return "bool"
	case "int8":
		return "int8_t"
	case "uint8":
		return "uint8_t"
	case "byte":
		return "uint8_t"
	case "int16":
		return "int16_t"
	case "uint16":
		return "uint16_t"
	case "int32":
		return "int32_t"
	case "rune":
		return "int32_t"
	case "uint32":
		return "uint32_t"
	case "int64":
		return "int64_t"
	case "uint64":
		return "uint64_t"
	case "int": //TODO: This is platform dependent
		return "int64_t"
	case "uint": //TODO: This is platform dependent
		return "uint64_t"
	case "uintptr": //TODO: This is platform dependent
		return "uint64_t"
	default:
		return gotype
	}
}

func generateStruct(fset *token.FileSet, w io.Writer, name string, fields *ast.FieldList) error {
	const (
		structBegin = "struct %s {\n"
		fieldFormat = "\t%s %s;\n"
		end         = "};\n\n"
	)

	_, err := fmt.Fprintf(w, structBegin, name)
	if err != nil {
		return err
	}
	for _, field := range fields.List {
		var typeNameBuf bytes.Buffer
		err := printer.Fprint(&typeNameBuf, fset, field.Type)
		if err != nil {
			return err
		}
		ctype := FromGoType(typeNameBuf.String())
		_, err = fmt.Fprintf(w, fieldFormat, ctype, field.Names[0].Name)
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
