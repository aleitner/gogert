package gogert

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

var (
	structBegin = "#ifndef CSTRUCTS_%s\n\tstruct %s {\n"
	fieldFormat = "\t\t%s %s; // gotype: %s\n"
	end         = "\t};\n#endif\n\n"
)

// TypeConverter contains necessary info for converting a gotype to a c type
type TypeConverter struct {
}

// Field contains all information about a particular field that was converted from go to C
type Field struct {
	CType  string
	Name   string
	GoType string
}

func (f *Field) String() string {
	return fmt.Sprintf(fieldFormat, f.CType, f.Name, f.GoType)
}

// CStructMeta contains all necessary information about a c struct converted from a go struct
type CStructMeta struct {
	Name                  string
	Fields                []*Field
	DependencyStructNames []string
	hasPointer            bool
}

// NewCStructMeta returns a new CStructMeta struct
func NewCStructMeta(name string, hasPointer bool) (*CStructMeta, error) {
	return &CStructMeta{Name: name, hasPointer: hasPointer}, nil
}

func (meta *CStructMeta) String() (cstruct string) {
	var cstructBytes bytes.Buffer
	_, err := fmt.Fprintf(&cstructBytes, structBegin, meta.Name, meta.Name)
	if err != nil {
		return ""
	}

	for _, field := range meta.Fields {
		_, err = fmt.Fprintf(&cstructBytes, field.String())
		if err != nil {
			return ""
		}
	}

	if meta.hasPointer {
		_, err = fmt.Fprint(&cstructBytes, "\t\t__SIZE_TYPE__ ptrRef; // gotype: uintptr\n")
		if err != nil {
			return ""
		}
	}

	_, err = fmt.Fprintf(&cstructBytes, end)
	if err != nil {
		return ""
	}

	return cstructBytes.String()
}

// NewConverter creates TypeConverter for converting gotype to ctype
func NewConverter() (*TypeConverter, error) {
	return &TypeConverter{}, nil
}

// FromGoType will convert a go type string to a ctypestring
func (c *TypeConverter) FromGoType(gotype string) (ctype string, dependentTypes []*CStructMeta) {
	gotypeWithoutPtr, ptrRef := separatePtr(gotype)

	if strings.HasPrefix(gotypeWithoutPtr, "[") {
		ctype, dependentTypes = c.fromSliceType(gotypeWithoutPtr)
	} else if strings.HasPrefix(gotypeWithoutPtr, "map") {
		ctype, dependentTypes = c.fromMapType(gotypeWithoutPtr)
	} else if isAnonymousStruct(gotypeWithoutPtr) {
		ctype, dependentTypes = c.fromAnonymousStruct(gotypeWithoutPtr)
	} else {
		ctype = c.fromBasicType(gotypeWithoutPtr)
	}

	// todo: check if type is a custom type or a struct
	if ctype == "void*" {
		// ctype, dependentTypes = c.fromComplexType(gotypeWithoutPtr)
	}

	// Don't add ptr because void automatically adds one
	if ctype == "void*" {
		if len(ptrRef) > 0 {
			ptrRef = ptrRef[1:]
		}
	}

	return ctype + ptrRef, dependentTypes
}

func separatePtr(gotype string) (newgotype string, ptr string) {
	re, _ := regexp.Compile(`(^\**)+`)
	matches := re.FindStringSubmatch(gotype)
	if len(matches) > 1 {
		return gotype[len(matches[1]):], matches[1]
	}

	return gotype, ""
}

func isAnonymousStruct(gotype string) bool {
	re, _ := regexp.Compile(`^struct`)
	return re.MatchString(gotype)
}

func (c *TypeConverter) fromAnonymousStruct(gotype string) (ctype string, dependentTypes []*CStructMeta) {
	anonymousBegin := "struct {\n"
	anonymousField := "\t%s %s; // gotype: %s\n"
	anonymousEnd := "}"

	var ctypeBytes bytes.Buffer

	_, err := fmt.Fprintf(&ctypeBytes, anonymousBegin)
	if err != nil {
		return "void*", dependentTypes
	}

	// Add package main for parsing
	src := fmt.Sprintf("package main;\ntype anonymous %s", gotype)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return "void*", dependentTypes
	}

	// hard coding looking these up
	typeDecl := f.Decls[0].(*ast.GenDecl)
	t := typeDecl.Specs[0].(*ast.TypeSpec)
	structDecl := t.Type.(*ast.StructType)
	fields := structDecl.Fields.List

	for _, field := range fields {
		typeExpr := field.Type

		start := typeExpr.Pos() - 1
		end := typeExpr.End() - 1

		// grab it in source
		typeInSource := src[start:end]

		fieldName := field.Names[0].Name

		fieldCType, fieldDependencies := c.FromGoType(typeInSource)

		_, err := fmt.Fprintf(&ctypeBytes, anonymousField, fieldCType, fieldName, typeInSource)
		if err != nil {
			return "void*", dependentTypes
		}

		dependentTypes = append(dependentTypes, fieldDependencies...)
	}

_:
	fmt.Fprintf(&ctypeBytes, anonymousEnd)
	if err != nil {
		return "void*", dependentTypes
	}

	return ctypeBytes.String(), dependentTypes
}

func (c *TypeConverter) fromComplexType(gotype string) (ctype string, dependentTypes []*CStructMeta) {
	// fmt.Println(gotype)

	return ctype, dependentTypes
}

func (c *TypeConverter) fromBasicType(gotype string) string {
	switch gotype {
	case "string":
		return "char*"
	case "bool":
		return "bool"
	case "int8":
		return "signed char"
	case "byte":
	case "uint8":
		return "unsigned char"
	case "int16":
		return "short"
	case "uint16":
		return "unsigned short"
	case "int32":
		return "int"
	case "rune":
	case "uint32":
		return "unsigned int"
	case "int64":
		return "long long"
	case "uint64":
		return "unsigned long long"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "int":
		return "int"
	case "uint":
		return "unsigned int"
	case "uintptr": // TODO: This is platform dependent
		return "__SIZE_TYPE__"
	case "complex64":
		return "float _Complex"
	case "complex128":
		return "double _Complex"
	default:
		return "void*"
	}

	return ""
}

func (c *TypeConverter) fromMapType(gotype string) (ctype string, dependencies []*CStructMeta) {
	// split go map into a key and value
	key, value := keyValueFromMap(gotype)

	// Convert the key and value go types to ctypes
	ckey, keydependencies := c.FromGoType(key)
	cvalue, valuedependencies := c.FromGoType(value)

	// add dependencies created from converting the keys and values
	dependencies = append(dependencies, keydependencies...)
	dependencies = append(dependencies, valuedependencies...)

	// Determine map stuct name
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return "void*", dependencies
	}
	mapName := reg.ReplaceAllString(strings.ReplaceAll(fmt.Sprintf("MAP_%s_%s", ckey, cvalue), "*", ""), "_")

	mapStruct, err := NewCStructMeta(mapName, false)
	if err != nil {
		return "void*", dependencies
	}
	dependencies = append(dependencies, mapStruct)

	// if the cvalue is a struct then we need to recognize that as a dependency
	if strings.Contains(cvalue, "struct") {
		mapStruct.DependencyStructNames = append(mapStruct.DependencyStructNames, cvalue)
	}

	keyField := &Field{
		Name:   "key",
		CType:  ckey,
		GoType: key,
	}

	valueField := &Field{

		Name:   "value",
		CType:  cvalue,
		GoType: value,
	}

	mapStruct.Fields = append(mapStruct.Fields, keyField, valueField)

	return fmt.Sprintf("struct %s*", mapName), dependencies
}

func keyValueFromMap(mapstr string) (key string, value string) {
	key = ""
	value = ""

	re, _ := regexp.Compile(`^map\[([^\]]+)\]`)
	matches := re.FindStringSubmatch(mapstr)
	if len(matches) > 1 {
		key = matches[1]
	}

	re, _ = regexp.Compile(`^map\[[^\]]+\](.+)`)
	matches = re.FindStringSubmatch(mapstr)
	if len(matches) > 1 {
		value = matches[1]
	}

	return key, value
}

func (c *TypeConverter) fromSliceType(gotype string) (ctype string, dependencies []*CStructMeta) {
	re, _ := regexp.Compile(`^([[0-9]*])`)
	matches := re.FindStringSubmatch(gotype)

	if len(matches) > 1 {
		prefix := matches[1]

		ctype, dependencies := c.FromGoType(gotype[len(prefix):])

		// Check if slice has a specified size or not
		if len(prefix) > 2 {
			return ctype + prefix, dependencies
		}

		return ctype + "*", dependencies
	}

	return "void *", dependencies
}
