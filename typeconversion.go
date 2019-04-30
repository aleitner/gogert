package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

func fromGoType(gotype string) string {
	gotypeWithoutPtr, ptrRef := separatePtr(gotype)

	if strings.HasPrefix(gotypeWithoutPtr, "[") {
		return fromSliceType(gotypeWithoutPtr) + ptrRef
	}

	if strings.HasPrefix(gotypeWithoutPtr, "map") {
		return fromMapType(gotypeWithoutPtr) + ptrRef
	}

	switch gotypeWithoutPtr {
	case "string":
		return "uint8_t*" + ptrRef
	case "bool":
		return "bool" + ptrRef
	case "int8":
		return "int8_t" + ptrRef
	case "byte":
	case "uint8":
		return "uint8_t" + ptrRef
	case "int16":
		return "int16_t" + ptrRef
	case "uint16":
		return "uint16_t" + ptrRef
	case "int32":
		return "int32_t" + ptrRef
	case "rune":
	case "uint32":
		return "uint32_t" + ptrRef
	case "int64":
		return "int64_t" + ptrRef
	case "uint64":
		return "uint64_t" + ptrRef
	case "float32":
		return "float" + ptrRef
	case "float64":
		return "double" + ptrRef
	case "int":
		return "int" + ptrRef
	case "uint":
		return "unsigned int" + ptrRef
	case "uintptr": // TODO: This is platform dependent
		return "uint64_t" + ptrRef
	case "complex64": // TODO: unsupported
	case "complex128": // TODO: unsupported
	default:
		return "void" + ptrRef
	}

	return ""
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

	key, value := keyValueFromMap(gotype)

	ckey := fromGoType(key)
	cvalue := fromGoType(value)

	reg, err := regexp.Compile("[^a-zA-Z0-9_]+")
	if err != nil {
		log.Fatal(err)
	}
	mapName := reg.ReplaceAllString(strings.ReplaceAll(fmt.Sprintf("MAP_%s_%s", ckey, cvalue), "*", ""), "-")

	if strings.HasPrefix(value, "map") {
		fmt.Printf("struct %s {\n\t%s key; // gotype: %s\n\tstruct Something* value; // gotype: %s\n};\n\n", mapName, ckey, key, value)
		return fmt.Sprintf("struct %s", mapName)
	}

	fmt.Printf("struct %s {\n\t%s key; // gotype: %s\n\t%s value; // gotype: %s\n};\n\n", mapName, ckey, key, cvalue, value)

	return fmt.Sprintf("struct %s", mapName)
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

func fromSliceType(gotype string) string {
	re, _ := regexp.Compile(`^([[0-9]*])`)
	matches := re.FindStringSubmatch(gotype)

	if len(matches) > 1 {
		prefix := matches[1]

		if len(prefix) > 2 {
			return fromGoType(gotype[len(prefix):]) + prefix
		} else {
			return fromGoType(gotype[len(prefix):]) + "*"
		}
	}

	return "void *"
}
