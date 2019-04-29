package main

//CExport A
type A struct {
	B      B
	I      int
	Hi     memory.Size
	events map[string]*Event
	Bptr   *B
	Bptr2  **B
	Iptr   *int
}

//Butts CExport BString
type BString string

/* hello */

//garbo
//CExport B
// function does this
type B struct {
	Str     BString
	numbers []int64
}

func main() {}
