package main

//CExport A
type A struct {
	B  B
	I  int
	Hi memory.Size
}

//Butts CExport BString
type BString string

/* hello */

//garbo
//CExport B
// function does this
type B struct {
	Str BString
}

func main() {}
