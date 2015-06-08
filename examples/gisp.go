package main

//gisp is a go parsec example, create a shceme like parser. you can run the interpreter though
//    go run gisp

import (
	"bufio"
	"fmt"
	"github.com/Dwarfartisan/goparsec/examples/gisp"
	"os"
)

var prompt = ">>> "

func main() {
	interactive()
}

func interactive() {
	parser, err := gisp.NewGisp(map[string]gisp.Environment{
		"axiom": gisp.Axiom,
		"prop":  gisp.Propositions,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		buf, _, _ := reader.ReadLine()
		re, err := parser.Parse(string(buf))
		if err == nil {
			parseAndPrint(re)
		} else {
			fmt.Println(err)
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
	case gisp.List:
		fmt.Printf("List: %v\n", v)
	case gisp.Function:
		fmt.Printf("Lambda: %v\n", v)
	case nil:
		fmt.Println("Nil")
	default:
		fmt.Printf("Unexpect: %v\n", v)
	}
}
