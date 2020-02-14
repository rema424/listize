package main

import (
	"fmt"
	"listize/listize"
)

func main() {
	fmt.Println(listize.MakeFuncSource(listize.Struct{Name: "Struct_1", Fields: []listize.Field{{Name: "Field_1", Type: "string"}}}))
	panic("must implement")
}
