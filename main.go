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

func fromGoType(gotype string) string {
	var ctype string

	gotypeWithoutPtr, ptrRef := separatePtr(gotype)

	if strings.HasPrefix(gotypeWithoutPtr, "[]") {
		return fromSliceType(gotypeWithoutPtr) + ptrRef
	}

	if strings.HasPrefix(gotypeWithoutPtr, "map") {
		return fromMapType(gotypeWithoutPtr) + ptrRef
	}

	switch gotypeWithoutPtr {
	case "string":
		ctype = "uint8_t"
		break
	case "bool":
		ctype = "bool"
		break
	case "int8":
		ctype = "int8_t"
		break
	case "uint8":
		ctype = "uint8_t"
		break
	case "byte":
		ctype = "uint8_t"
		break
	case "int16":
		ctype = "int16_t"
		break
	case "uint16":
		ctype = "uint16_t"
		break
	case "int32":
		ctype = "int32_t"
		break
	case "rune":
		ctype = "int32_t"
		break
	case "uint32":
		ctype = "uint32_t"
		break
	case "int64":
		ctype = "int64_t"
		break
	case "uint64":
		ctype = "uint64_t"
		break
	case "int": //TODO: This is platform dependent
		ctype = "int64_t"
		break
	case "uint": //TODO: This is platform dependent
		ctype = "uint64_t"
		break
	case "uintptr": //TODO: This is platform dependent
		ctype = "uint64_t"
		break
	default:
		ctype = "void"
	}

	return ctype + ptrRef
}

func separatePtr(gotype string) (newgotype string, ptr string) {
	re, _ := regexp.Compile(`(^\**)+`)
	matches := re.FindStringSubmatch(gotype)
	if len(matches) > 1 {
		return gotype[len(matches[1]):], matches[1]
	}

	return gotype, ""
}

func fromMapType(gotype string) string {
	return gotype
}

func fromSliceType(gotype string) string {
	gotype = gotype[2:]
	ctype := fromGoType(gotype)
	return ctype + "*"
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
		ctype := fromGoType(typeNameBuf.String())
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
