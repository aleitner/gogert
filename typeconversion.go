package main

import (
	"fmt"
	"log"
	"plugin"
	"regexp"
	"strings"
)

type typeConverter struct {
	plugin *plugin.Plugin
}

// NewConverter creates typeConverter for converting gotype to ctype
func NewConverter(object string) (*typeConverter, error) {
	c := &typeConverter{}

	if object == "" {
		return c, nil
	}

	p, err := plugin.Open(object)
	if err != nil {
		return c, err
	}

	c.plugin = p

	return c, nil
}

func (c *typeConverter) fromGoType(gotype string) (ctype string, dependentTypes []*CStructMeta) {
	gotypeWithoutPtr, ptrRef := separatePtr(gotype)

	if strings.HasPrefix(gotypeWithoutPtr, "[") {
		ctype, dependentTypes = c.fromSliceType(gotypeWithoutPtr)
	} else if strings.HasPrefix(gotypeWithoutPtr, "map") {
		ctype, dependentTypes = c.fromMapType(gotypeWithoutPtr)
	} else {
		ctype = c.fromBasicType(gotypeWithoutPtr)
	}

	// todo: check if type is a custom type or a struct
	if c.plugin != nil && ctype == "void*" {
		fmt.Println("We need to do something about ", gotype)
		// convert?
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

func (c *typeConverter) fromBasicType(gotype string) string {
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

func (c *typeConverter) fromMapType(gotype string) (ctype string, dependencies []*CStructMeta) {

	key, value := keyValueFromMap(gotype)

	ckey, keydependencies := c.fromGoType(key)
	cvalue, valuedependencies := c.fromGoType(value)

	dependencies = append(dependencies, keydependencies...)
	dependencies = append(dependencies, valuedependencies...)

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	mapName := reg.ReplaceAllString(strings.ReplaceAll(fmt.Sprintf("MAP_%s_%s", ckey, cvalue), "*", ""), "_")

	mapStruct := &CStructMeta{name: mapName}
	dependencies = append(dependencies, mapStruct)

	if strings.Contains(cvalue, "struct") {
		mapStruct.dependencies = append(mapStruct.dependencies, cvalue)
	}

	if strings.HasPrefix(value, "map") {
		_, err = fmt.Fprintf(&mapStruct.structDeclaration, "struct %s {\n\t%s key; // gotype: %s\n\t%s value; // gotype: %s\n};\n\n", mapName, ckey, key, cvalue, value)
		if err != nil {
			log.Fatal(err)
		}
		return fmt.Sprintf("struct %s*", mapName), dependencies
	}

	_, err = fmt.Fprintf(&mapStruct.structDeclaration, "struct %s {\n\t%s key; // gotype: %s\n\t%s value; // gotype: %s\n};\n\n", mapName, ckey, key, cvalue, value)
	if err != nil {
		log.Fatal(err)
	}

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

func (c *typeConverter) fromSliceType(gotype string) (ctype string, dependencies []*CStructMeta) {
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
