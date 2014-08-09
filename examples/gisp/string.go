package gisp

import (
	. "github.com/Dwarfartisan/goparsec"
)

var EscapeChar = Bind_(Rune('\\'), func(st ParseState) (interface{}, error) {
	r, err := OneOf("nrt\"\\")(st)
	if err == nil {
		ru := r.(rune)
		switch ru {
		case 'r':
			return '\r', nil
		case 'n':
			return '\n', nil
		case '"':
			return '"', nil
		case '\\':
			return '\\', nil
		case 't':
			return '\t', nil
		default:
			return nil, st.Trap("Can't escape \\%c", r)
		}
	} else {
		return nil, err
	}
})

var StringVal = Bind(
	Between(Rune('"'), Rune('"'),
		Many(Either(Try(EscapeChar), NoneOf("\"")))),
	ReturnString)
