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

func Always(pos int, x interface{}) (interface{}, error) {
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

// match anyone else a eof or panic
func AnyOne(st ParsexState) (interface{}, error) {
	return st.Next(Always)
}
func TheOne(one interface{}) Parser {
	return func(st ParsexState) (interface{}, error) {
		_, err := st.Next(equals(one))
		if err == nil {
			return one, nil
		} else {
			return nil, err
		}
	}
}

func Eof(st ParsexState) (interface{}, error) {
	r, err := st.Next(Always)
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
	c, err := st.Next(Always)

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
	case complex128:
		if imag(val) == 0 {
			return int(real(val)), nil
		} else {
			return nil, TypeError{"int", val, pos}
		}
	default:
		return nil, TypeError{"int", val, pos}
	}
}
func canbeFloat64(pos int, x interface{}) (interface{}, error) {
	switch val := x.(type) {
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	case complex64:
		if imag(val) == 0 {
			return float64(real(val)), nil
		} else {
			return nil, TypeError{"float64", val, pos}
		}
	case complex128:
		if imag(val) == 0 {
			return real(val), nil
		} else {
			return nil, TypeError{"float64", val, pos}
		}
	default:
		return nil, TypeError{"float64", val, pos}
	}
}

func intType(pos int, x interface{}) (interface{}, error) {
	if _, ok := x.(int); ok {
		return x, nil
	} else {
		return nil, TypeError{"int", x, pos}
	}
}

func float64Type(pos int, x interface{}) (interface{}, error) {
	if _, ok := x.(float64); ok {
		return x, nil
	} else {
		return nil, TypeError{"float64", x, pos}
	}
}

func IntVal(st ParsexState) (interface{}, error) {
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
func AnyFloat64(st ParsexState) (interface{}, error) {
	i, err := st.Next(canbeFloat64)
	if err == nil {
		return i, nil
	} else {
		return nil, err
	}
}
func Float64Val(st ParsexState) (interface{}, error) {
	i, err := st.Next(float64Type)
	if err == nil {
		return i, nil
	} else {
		return nil, err
	}
}

//FIXME: string 类型是一个坑，好大的坑，目前没有更好的办法？
//       不同 format 类型的坑
func AnyTime(st ParsexState) (interface{}, error) {
	t, err := st.Next(func(pos int, x interface{}) (interface{}, error) {
		switch x.(type) {
		case string:
			//只针对这一种情况，其他情况，如2006/01/02 15:04 ， 02/01/2006 ...还有其他格式的字符串
			result, e := time.Parse("2006-01-02 15:04", x.(string))
			if e == nil {
				return result, nil
			} else {
				return nil, TypeError{"time", x, pos}
			}
		case time.Time:
			if _, ok := x.(time.Time); ok {
				return x, nil
			} else {
				return nil, TypeError{"time", x, pos}
			}
		default:
			return nil, TypeError{"time", x, pos}
		}

	})
	if err == nil {
		return t, nil
	} else {
		return nil, err
	}
}

func TimeVal(st ParsexState) (interface{}, error) {
	t, err := st.Next(func(pos int, x interface{}) (interface{}, error) {
		if _, ok := x.(time.Time); ok {
			return x, nil
		} else {
			return nil, TypeError{"time", x, pos}
		}

	})
	if err == nil {
		return t, nil
	} else {
		return nil, err
	}
}

func StringVal(st ParsexState) (interface{}, error) {
	t, err := st.Next(func(pos int, x interface{}) (interface{}, error) {
		if _, ok := x.(string); ok {
			return x, nil
		} else {
			return nil, TypeError{"string", x, pos}
		}
	})
	if err == nil {
		return t, nil
	} else {
		return nil, err
	}
}

func Nil(st ParsexState) (interface{}, error) {
	t, err := st.Next(func(pos int, x interface{}) (interface{}, error) {
		if x == nil {
			return x, nil
		} else {
			return nil, TypeError{"nil", x, pos}
		}
	})
	if err == nil {
		return t, nil
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
