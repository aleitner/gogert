package main

import "testing"

var conversions = []struct {
	gotype string // input
	ctype  string // expected result
}{
	{"B", "void"},     // struct
	{"*B", "void*"},   // pointer struct
	{"**B", "void**"}, // pointer to a pointer of a struct

	{"int64", "uint64_t"},    // basic type
	{"string", "uint8_t*"},   // basic type
	{"float32", "float"},     // basic type
	{"float64", "double"},    // basic type
	{"int", "int"},           // basic type
	{"uint", "unsigned int"}, // basic type
	{"*int64", "uint64_t*"},  // pointer to basic type

	{"memory.Size", "void"}, // custom basic type

	{"[]int64", "uint64_t*"},
	{"[][]int64", "uint64_t**"},
	{"*[]int64", "uint64_t***"},
	{"[][]*int64", "uint64_t***"},
	{"[]*[]int64", "uint64_t***"},
	{"map[string]*Event", ""},
	{"map[string]map[string]*Event", ""},
}

func TestfromGoType(t *testing.T) {

	for _, tt := range conversions {
		actual := fromGoType(tt.gotype)
		if actual != tt.ctype {
			t.Errorf("fromGoType(%s): expected %s, actual %s", tt.gotype, tt.ctype, actual)
		}
	}
}