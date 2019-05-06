package main

import "github.com/aleitner/gogert/samples/pkg"

//CExport Sample
type Sample struct {
	Pkg pkg.Pkg
}

//CExport A
type A struct {
	B              B
	I              int
	Hi             BString
	events         map[string]*Sample
	eventsSuperMap map[*B]map[int]*Sample
	Bptr           *B
	Bptr2          **B
	Iptr           *int
	str            string
}

//This shouldn't CExport BString
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
