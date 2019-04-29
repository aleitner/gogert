package gogert

import (
	"os"
	"testing"
)

var conversions = []struct {
	gotype string // input
	ctype  string // expected result
}{
	{"B", "void"},     // struct
	{"*B", "void*"},   // pointer struct
	{"**B", "void**"}, // pointer to a pointer of a struct

	{"int64", "int64_t"},     // basic type
	{"string", "uint8_t*"},   // basic type
	{"float32", "float"},     // basic type
	{"float64", "double"},    // basic type
	{"int", "int"},           // basic type
	{"uint", "unsigned int"}, // basic type
	{"*int64", "int64_t*"},   // pointer to basic type

	{"memory.Size", "void"}, // custom basic type

	{"[]int64", "int64_t*"},
	{"[][]int64", "int64_t**"},
	{"*[]int64", "int64_t**"},
	{"[][]*int64", "int64_t***"},
	{"[]*[]int64", "int64_t***"},
	{"map[string]*Event", ""},
	{"map[string]map[string]*Event", ""},
}

func TestFromGoType(t *testing.T) {

	for _, tt := range conversions {
		actual := fromGoType(tt.gotype)
		if actual != tt.ctype {
			t.Errorf("fromGoType(%s): expected %s, actual %s", tt.gotype, tt.ctype, actual)
		}
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
