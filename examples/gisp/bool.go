package gisp

import (
	"fmt"
	. "github.com/Dwarfartisan/goparsec"
)

var BoolParser = Bind(Choice(String("true"), String("false")), func(input interface{}) Parser {
	return func(st ParseState) (interface{}, error) {
		switch input.(string) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("Unexcept bool token %v", input)
		}
	}
})

var NilParser = Bind_(String("nil"), Return(nil))
