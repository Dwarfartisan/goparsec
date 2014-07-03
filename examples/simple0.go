package main

import (
	"fmt"
	. "github.com/Dwarfartisan/goparsec"
)

func main() {
	input := "a,b c,d,"
	fmt.Println(input)
	data, err := SepBy(Bind(Many1(NoneOf(", ")), ReturnString),
		Many1(OneOf(", ")))(MemoryParseState(input))
	if err == nil {
		for _, r := range data.([]interface{}) {
			fmt.Printf("%s\n", r.(string))
		}
	} else {
		fmt.Println(err)
	}
}
