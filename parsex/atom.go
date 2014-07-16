package parsex

import (
	"io"
	"reflect"
	"unicode"
)

func always(r rune) bool {
	return true
}
func deep_equals(x interface{}) func(rune) bool {
	return func(data interface{}) bool {
		return reflect.DeepEqual(x, data)
	}
}

func Rune(r rune) Parser {
	return func(st ParseState) (interface{}, error) {
		ru, ok, err := st.Next(equals(r))
		if err != nil {
			return nil, err
		}
		if ok {
			return r, nil
		} else {
			return nil, st.Trap("rune '%c' nomatch rune pattern '%c'", ru, r)
		}
	}
}

func Eof(st ParseState) (interface{}, error) {
	r, _, err := st.Next(always)
	if err == nil {
		return nil, st.Trap("Except EOF but got %c", r)
	} else {
		if err == io.EOF {
			return nil, nil
		} else {
			return nil, err
		}
	}
}

// parsex 的 String 尝试匹配 State 的下一个 Token，这与 parsec 不同
func String(s string) Parser {
	return func(st ParseState) (interface{}, error) {
		pos := st.Pos()

		// try and match string
		_, ok, err := st.Next(equals(s))
		if err != nil {
			st.SeekTo(pos)
			return nil, err
		}

		if !ok {
			st.SeekTo(pos)
			// the string failed to match
			return nil, st.Trap("Expected '%s'", s)
		}

		return s, nil
	}
}

func AnyRune(st ParseState) (interface{}, error) {
	c, _, err := st.Next(always)

	if err == nil {
		return c, nil
	} else {
		if err == io.EOF {
			return nil, st.Trap("Unexpected end of file")
		} else {
			return nil, err
		}
	}
}

func RuneChecker(checker func(rune) bool, expected string) Parser {
	return func(st ParseState) (interface{}, error) {
		r, ok, err := st.Next(checker)

		if err == nil {
			if ok {
				return r, nil
			} else {
				return nil, st.Trap("Expected %s but '%c'", expected, r)
			}
		} else {
			if err == io.EOF {
				return nil, st.Trap("Unexpected end of file")
			} else {
				return nil, err
			}
		}
	}
}

var Space = RuneChecker(unicode.IsSpace, "space")
var Spaces = Skip(Space)
var NewLineRunes = "\r\n"
var NewLine = OneOf(NewLineRunes)
var Eol = Either(Eof, NewLine)
