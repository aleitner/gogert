package gogert

import (
	"bytes"
	"fmt"
	"log"
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
	ctype  string
	name   string
	gotype string
}

// CStructMeta contains all necessary information about a c struct converted from a go struct
type CStructMeta struct {
	name                  string
	fields                []*Field
	dependencyStructNames []string
	hasPointer            bool
}

func (meta *CStructMeta) String() (cstruct string) {
	var cstructBytes bytes.Buffer
	_, err := fmt.Fprintf(&cstructBytes, structBegin, meta.name, meta.name)
	if err != nil {
		return ""
	}

	for _, field := range meta.fields {
		_, err = fmt.Fprintf(&cstructBytes, fieldFormat, field.ctype, field.name, field.gotype)
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

func (c *TypeConverter) fromGoType(gotype string) (ctype string, dependentTypes []*CStructMeta) {
	gotypeWithoutPtr, ptrRef := separatePtr(gotype)

	if strings.HasPrefix(gotypeWithoutPtr, "[") {
		ctype, dependentTypes = c.fromSliceType(gotypeWithoutPtr)
	} else if strings.HasPrefix(gotypeWithoutPtr, "map") {
		ctype, dependentTypes = c.fromMapType(gotypeWithoutPtr)
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

func (c *TypeConverter) fromComplexType(gotype string) (ctype string, dependentTypes []*CStructMeta) {
	fmt.Println("Complex type conversion ", gotype)
	// convert?
	return gotype, dependentTypes
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
	ckey, keydependencies := c.fromGoType(key)
	cvalue, valuedependencies := c.fromGoType(value)

	// add dependencies created from converting the keys and values
	dependencies = append(dependencies, keydependencies...)
	dependencies = append(dependencies, valuedependencies...)

	// Determine map stuct name
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	mapName := reg.ReplaceAllString(strings.ReplaceAll(fmt.Sprintf("MAP_%s_%s", ckey, cvalue), "*", ""), "_")

	mapStruct := &CStructMeta{name: mapName}
	dependencies = append(dependencies, mapStruct)

	// if the cvalue is a struct then we need to recognize that as a dependency
	if strings.Contains(cvalue, "struct") {
		mapStruct.dependencyStructNames = append(mapStruct.dependencyStructNames, cvalue)
	}

	keyField := &Field{
		name:   "key",
		ctype:  ckey,
		gotype: key,
	}

	valueField := &Field{
		name:   "value",
		ctype:  cvalue,
		gotype: value,
	}

	mapStruct.fields = append(mapStruct.fields, keyField, valueField)

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

		ctype, dependencies := c.fromGoType(gotype[len(prefix):])

		// Check if slice has a specified size or not
		if len(prefix) > 2 {
			return ctype + prefix, dependencies
		}

		return ctype + "*", dependencies
	}

	return "void *", dependencies
}
