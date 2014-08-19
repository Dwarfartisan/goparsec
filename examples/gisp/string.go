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
		// FIXME:引号的解析偷懒了，单双引号的应该分开。
		case '\'':
			return '\'', nil
		case '"':
			return '"', nil
		case '\\':
			return '\\', nil
		case 't':
			return '\t', nil
		default:
			return nil, st.Trap("Unknown escape sequence \\%c", r)
		}
	} else {
		return nil, err
	}
})

var RuneParser = Bind(
	Between(Rune('\''), Rune('\''),
		Either(Try(EscapeChar), NoneOf("'"))),
	ReturnString)

var StringParser = Bind(
	Between(Rune('"'), Rune('"'),
		Many(Either(Try(EscapeChar), NoneOf("\"")))),
	ReturnString)
