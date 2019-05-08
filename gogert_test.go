package gogert

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var conversions = []struct {
	gotype string // input
	ctype  string // expected result
}{
	{"B", "struct B"},     // struct
	{"*B", "struct B*"},   // pointer struct
	{"**B", "struct B**"}, // pointer to a pointer of a struct

	{"int64", "long long"},   // basic type
	{"string", "char*"},      // basic type
	{"float32", "float"},     // basic type
	{"float64", "double"},    // basic type
	{"int", "int"},           // basic type
	{"uint", "unsigned int"}, // basic type
	{"*int64", "long long*"}, // pointer to basic type
	{"[]int64", "long long*"},
	{"[3]int64", "long long[3]"},
	{"[][]int64", "long long**"},
	{"*[]int64", "long long**"},
	{"[]B", "struct B*"},
	{"[]*B", "struct B**"},
	{"[][]*int64", "long long***"},
	{"[]*[]int64", "long long***"},
	{"[3][4]int64", "long long[4][3]"},
	{"map[string]*Event", "struct MAP_char_struct_Event*"},
	{"map[string]map[string]*Event", "struct MAP_char_struct_MAP_char_struct_Event*"},
	{"struct {\n\ta bool\n\tb int\n}", "struct {\n\tbool a; // gotype: bool\n\tint b; // gotype: int\n}"},
}

func TestFromGoType(t *testing.T) {
	converter, err := NewConverter()
	if err != nil {
		require.NoError(t, err)
	}

	for _, tt := range conversions {
		actual, _ := converter.FromGoType(tt.gotype)
		if actual != tt.ctype {
			t.Errorf("fromGoType(%s): expected %s, actual %s", tt.gotype, tt.ctype, actual)
		}
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
