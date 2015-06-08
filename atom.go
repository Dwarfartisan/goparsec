package goparsec

import (
	"io"
	"unicode"
)

func always(r rune) bool {
	return true
}
func equals(r rune) func(rune) bool {
	return func(ru rune) bool {
		return ru == r
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
		st.SeekTo(st.Pos() - 1)
		return nil, st.Trap("expect EOF but got %c", r)
	} else {
		if err == io.EOF {
			return nil, nil
		} else {
			return nil, err
		}
	}
}

func String(s string) Parser {
	return func(st ParseState) (interface{}, error) {
		pos := st.Pos()

		// try and match each character
		for _, r := range []rune(s) {
			_, ok, err := st.Next(equals(r))
			if err != nil {
				st.SeekTo(pos)
				return nil, err
			}

			if !ok {
				st.SeekTo(pos)
				// the string failed to match
				return nil, st.Trap("Expected '%s'", s)
			}
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
var Letter = RuneChecker(unicode.IsLetter, "letter")
var Number = RuneChecker(unicode.IsNumber, "number")
var Spaces = Skip(Space)
var NewLineRunes = "\r\n"
var NewLine = OneOf(NewLineRunes)
var Eol = Either(Eof, NewLine)
var Digit = RuneChecker(unicode.IsDigit, "digit")

func Int(st ParseState) (interface{}, error) {
	pos := st.Pos()
	values := []interface{}{}
	_, err := Try(Rune('-'))(st)
	if err == nil {
		values = append(values, '-')
	}
	v, err := Many1(Digit)(st)
	if err == nil {
		values = append(values, v.([]interface{})...)
		return ExtractString(values), nil
	} else {
		st.SeekTo(pos)
		return nil, err
	}
}

var UnsignedFloat = Bind(Many(Digit), func(input interface{}) Parser {
	return func(st ParseState) (interface{}, error) {
		value, err := Bind_(Rune('.'), Many1(Digit))(st)
		if err != nil {
			return nil, err
		}
		ret := []interface{}{}
		ret = append(ret, input.([]interface{})...)
		ret = append(ret, '.')
		ret = append(ret, value.([]interface{})...)
		return ExtractString(ret), nil
	}
})

var Float = Either(UnsignedFloat,
	Bind_(Rune('-'), func(st ParseState) (interface{}, error) {
		value, err := UnsignedFloat(st)
		if err == nil {
			return ExtractString(append([]interface{}{'-'},
				value.([]interface{})...)), nil
		} else {
			return nil, err
		}
	}))
