package main

import (
	"bufio"
	"fmt"
	parsec "github.com/Dwarfartisan/goparsec"
	"github.com/Dwarfartisan/goparsec/examples/gisp"
	"os"
)

var prompt = ">>> "

func main() {
	interactive()
}

func interactive() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		buf, _, _ := reader.ReadLine()
		st := parsec.MemoryParseState(string(buf))
		value, err := gisp.ParseValue(st)
		if err == nil {
			parseAndPrint(value)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func parseAndPrint(value interface{}) {
	switch v := value.(type) {
	case int:
		fmt.Printf("Int: %d\n", v)
	case float64:
		fmt.Printf("Float64: %f\n", v)
	case string:
		fmt.Printf("String: %s\n", v)
	case gisp.Atom:
		fmt.Printf("Atom: %v\n", v)
	case nil:
		fmt.Println("Nil")
	default:
		fmt.Printf("Unexcept: %v\n", v)
	}
}
