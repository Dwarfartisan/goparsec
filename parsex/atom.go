package parsex

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"
	"unicode"
)

type NotEqual struct {
	Except interface{}
	Value  interface{}
	Pos    int
}

func (this NotEqual) Error() string {
	return fmt.Sprintf("Postion: %d,  Except %v but got %v", this.Pos, this.Except, this.Value)
}

type TypeError struct {
	Type  string
	Value interface{}
	Pos   int
}

func (this TypeError) Error() string {
	return fmt.Sprintf("Postion: %d,  Except %v as a %s", this.Pos, this.Value, this.Type)
}

func equals(x interface{}) func(int, interface{}) (interface{}, error) {
	return func(pos int, data interface{}) (interface{}, error) {
		if reflect.DeepEqual(x, data) {
			return x, nil
		} else {
			return data, NotEqual{x, data, pos}
		}
	}
}

func always(pos int, x interface{}) (interface{}, error) {
	return x, nil
}

func Rune(r rune) Parser {
	return func(st ParsexState) (interface{}, error) {
		ru, err := st.Next(equals(r))
		if err != nil {
			return nil, err
		} else {
			return ru, nil
		}
	}
}

func Eof(st ParsexState) (interface{}, error) {
	r, err := st.Next(always)
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
	return func(st ParsexState) (interface{}, error) {
		pos := st.Pos()

		// try and match string
		_, err := st.Next(equals(s))
		if err != nil {
			st.SeekTo(pos)
			return nil, err
		}
		return s, nil
	}
}

func AnyRune(st ParsexState) (interface{}, error) {
	c, err := st.Next(always)

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
func canbeInt(pos int, x interface{}) (interface{}, error) {
	switch val := x.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case int32:
		return int(val), nil
	case float64:
		return int(val), nil
	case float32:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	case complex64:
		if imag(val) == 0 {
			return int(real(val)), nil
		} else {
			return nil, TypeError{"int", val, pos}
		}
	default:
		return nil, TypeError{"int", val, pos}
	}
}
func intType(pos int, x interface{}) (interface{}, error) {
	if _, ok := x.(int); ok {
		return x, nil
	} else {
		return nil, TypeError{"int", x, pos}
	}
}
func IntVar(st ParsexState) (interface{}, error) {
	i, err := st.Next(intType)
	if err == nil {
		return i, nil
	} else {
		return nil, err
	}
}
func AnyInt(st ParsexState) (interface{}, error) {
	i, err := st.Next(canbeInt)
	if err == nil {
		return i, nil
	} else {
		return nil, err
	}
}
func TimeVar(st ParsexState) (interface{}, error) {
	i, err := st.Next(func(pos int, x interface{}) (interface{}, error) {
		if _, ok := x.(time.Time); ok {
			return x, nil
		} else {
			return nil, TypeError{"time", x, pos}
		}
	})
	if err == nil {
		return i, nil
	} else {
		return nil, err
	}
}

func RuneChecker(checker func(int, interface{}) (interface{}, error), expected string) Parser {
	return func(st ParsexState) (interface{}, error) {
		r, err := st.Next(checker)

		if err == nil {
			return r, nil
		} else {
			if err == io.EOF {
				return nil, st.Trap("Unexpected end of file")
			} else {
				return nil, err
			}
		}
	}
}

var Space = RuneChecker(func(pos int, x interface{}) (interface{}, error) {
	if unicode.IsSpace(x.(rune)) {
		return x, nil
	} else {
		message := fmt.Sprintf("Except space but got %v", x)
		return x, errors.New(message)
	}
}, "space")
var Spaces = Skip(Space)
var NewLineRunes = []interface{}{"\r", "\n"}
var NewLine = OneOf(NewLineRunes)
var Eol = Either(Eof, NewLine)
