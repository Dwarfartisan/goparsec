package main

import (
	"fmt"
	. "github.com/Dwarfartisan/goparsec"
	"io/ioutil"
	"os"
)

var Brackets = Between(Rune('['), Rune(']'), Many1(NoneOf("]")))
var Parentheses = Between(Rune('('), Rune(')'), Many1(NoneOf(")")))

var Entry = Between(String("[entry:"), Rune(']'), Many1(NoneOf("]")))

func HTTP(st ParseState) (interface{}, error) {
	parser := Between(String("[http://"), Rune(']'), Many1(NoneOf("]")))
	data, err := parser(st)
	if err == nil {
		return map[string]interface{}{
			"lnk": "http://" + string(data.([]rune)),
		}, nil
	} else {
		return nil, err
	}
}

var Link = Bind(Brackets,
	func(cap interface{}) Parser {
		return func(st ParseState) (interface{}, error) {
			lnk, err := Parentheses(st)
			if err == nil {
				return map[string]interface{}{"cap": cap, "lnk": lnk}, nil
			} else {
				return nil, err
			}
		}
	},
)
var Code = Choice(Try(Entry), Try(Link), Try(HTTP))

func MissMatch(st ParseState) (interface{}, error) {
	p := Bind(Rune('['), func(r interface{}) Parser {
		return func(st ParseState) (interface{}, error) {
			data, err := plain(st)
			buf := data.([]interface{})
			buffer := make([]rune, 0, len(buf))
			for _, item := range buf {
				buffer = append(buffer, item.(rune))
			}

			if err == nil {
				return string(r.(rune)) + string(buffer), nil
			} else {
				return nil, err
			}
		}
	})
	return p(st)
}

var plain = Many1(NoneOf("["))
var content = Choice(Try(plain), Try(Code), MissMatch)
var Paragraph = ManyTil(content, Eof)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(err)
		return
	}
	st := MemoryParseState(string(data))
	para, err := Paragraph(st)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(para)
}
