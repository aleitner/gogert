package gogert

import (
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
	case "uint8":
		return "uint8_t" + ptrRef
	case "byte":
		return "uint8_t" + ptrRef
	case "int16":
		return "int16_t" + ptrRef
	case "uint16":
		return "uint16_t" + ptrRef
	case "int32":
		return "int32_t" + ptrRef
	case "rune":
		return "int32_t" + ptrRef
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
	case "uintptr": //TODO: This is platform dependent
		return "uint64_t" + ptrRef
	default:
		return "void" + ptrRef
	}
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
