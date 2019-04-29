package main

//CExport A
type A struct {
	B              B
	I              int
	Hi             memory.Size
	events         map[string]*Event
	eventsSuperMap map[string]map[string]*Event
	Bptr           *B
	Bptr2          **B
	Iptr           *int
}

//Butts CExport BString
type BString string

/* hello */

//garbo
//CExport B
// function does this
type B struct {
	Str               BString
	numbers           []int64
	numbersInNumbers  []*[]int64
	numbersInNumbers2 [][]int64
	numbersInNumbers3 [][]*int64
	arrayPtr          *[]int64
}

func main() {}
