package main

import (
	"encoding/json"
	"fmt"
	. "github.com/Dwarfartisan/goparsec"
	"io/ioutil"
	"os"
)

// return a text skip newline and exclude the runes
func TextWithout(runes string) Parser {
	var newline = Bind_(Many1(OneOf(NewLineRunes)), Return(""))
	var others = Bind(Many1(NoneOf(runes+NewLineRunes)), ReturnString)
	var content = Many1(Choice(Try(newline), Try(others)))

	return func(st ParseState) (interface{}, error) {
		data, err := content(st)
		if err == nil {
			var text = ""
			for _, item := range data.([]interface{}) {
				text += item.(string)
			}
			return text, nil
		} else {
			return nil, err
		}
	}
}

var Brackets = Between(Rune('['), Rune(']'), TextWithout("]"))
var Parentheses = Between(Rune('('), Rune(')'), TextWithout(")"))

var Entry = Between(String("[entry:"), Rune(']'), TextWithout("]"))

func HTTP(st ParseState) (interface{}, error) {
	parser := Between(String("[http://"), Rune(']'), TextWithout("]"))
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

var plain = TextWithout("[")
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
	out, _ := json.Marshal(para)
	fmt.Println(string(out))
}
