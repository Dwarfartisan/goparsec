package gparsec

import (
	"strings"
)

func Try(parser Parser) Parser {
	return func(st ParseState) (interface{}, error) {
		pos := st.Pos()
		result, err := parser(st)
		if err == nil {
			return result, nil
		} else {
			st.SeekTo(pos)
			return nil, err
		}
	}
}
func Bind(parser Parser, fun func(interface{}) Parser) Parser {
	return func(st ParseState) (interface{}, error) {
		result, err := parser(st)
		if err != nil {
			return nil, err
		}
		return fun(result)(st)
	}
}

func Bind_(parserx, parsery Parser) Parser {
	return func(st ParseState) (interface{}, error) {
		_, err := parserx(st)
		if err != nil {
			return nil, err
		}
		return parsery(st)
	}
}

// try one parser, if it fails (without consuming input) try the next
func Either(parserx, parsery Parser) Parser {
	return func(st ParseState) (interface{}, error) {
		pos := st.Pos()
		x, err := parserx(st)
		if err == nil {
			return x, nil
		} else {
			if st.Pos() == pos {
				return parsery(st)
			}
		}
		return nil, err
	}
}
func Return(v interface{}) Parser {
	return func(st ParseState) (interface{}, error) {
		return v, nil
	}
}
func Option(v interface{}, parser Parser) Parser {
	return func(st ParseState) (interface{}, error) {
		return Either(parser, Return(v))(st)
	}
}
func Many1(parser Parser) Parser {
	head := func(value interface{}) Parser {
		tail := func(values interface{}) Parser {
			return Return(append([]interface{}{value}, values.([]interface{})...))
		}
		return Bind(Many(parser), tail)
	}
	return Bind(parser, head)
}
func Many(parser Parser) Parser {
	return func(st ParseState) (interface{}, error) {
		return Option([]interface{}{}, Many1(parser))(st)
	}
}
func Fail(message string) Parser {
	return func(st ParseState) (interface{}, error) {
		return nil, st.Trap(message)
	}
}
func OneOf(runes string) Parser {
	return func(st ParseState) (interface{}, error) {
		r, ok, err := st.Next(func(ru rune) bool { return strings.IndexRune(runes, ru) >= 0 })
		if err != nil {
			return nil, err
		}

		if ok {
			return r, nil
		} else {
			return nil, st.Trap("Excepted one of '%s' but got '%c'", string(runes), r)
		}
	}
}
func NoneOf(runes string) Parser {
	return func(st ParseState) (interface{}, error) {
		r, ok, err := st.Next(func(ru rune) bool { return strings.IndexRune(runes, ru) < 0 })
		if err != nil {
			return nil, err
		}

		if ok {
			return r, nil
		} else {
			return nil, st.Trap("Excepted one of '%s' but got '%c'", string(runes), r)
		}
	}
}
func Between(start, end, p Parser) Parser {
	keep := func(x interface{}) Parser {
		return Bind_(end, Return(x))
	}
	return Bind_(start, Bind(p, keep))
}
func SepBy1(p, sep Parser) Parser {
	head := func(x interface{}) Parser {
		tail := func(xs interface{}) Parser {
			return Return(append([]interface{}{x}, xs.([]interface{})...))
		}
		return Bind(Many(Bind_(sep, p)), tail)
	}
	return Bind(p, head)
}
func SepBy(p, sep Parser) Parser {
	return Option([]interface{}{}, SepBy1(p, sep))
}
func ManyTil(p, end Parser) Parser {
	head := func(x interface{}) Parser {
		tail := func(xs interface{}) Parser {
			return Return(append([]interface{}{x}, xs.([]interface{})...))
		}

		return Bind(ManyTil(p, end), tail)
	}
	term := Bind_(Try(end), Return([]interface{}{}))
	return Either(term, Bind(p, head))
}
func Maybe(p Parser) Parser {
	return Option(nil, Bind_(p, Return(nil)))
}
func Skip(p Parser) Parser {
	return Maybe(Many(p))
}

func Choice(parsers ...Parser) Parser {
	return func(st ParseState) (interface{}, error) {
		var err error
		var result interface{}
		for _, parser := range parsers {
			result, err = parser(st)
			if err == nil {
				return result, nil
			}
		}
		return nil, err
	}
}

// 其实我比较希望把上面那个东西实现成下面这个样子，就是好像在golang里不太经济……
// func Choice(parsers ...Parser) Parser {
// 	switch len(parsers) {
// 	case 0:
// 		panic(errors.New("empty choice chain"))
// 	case 1:
// 		return parsers[0]
// 	case 2:
// 		return Either(Try(parsers[0]), Try(parsers[1]))
// 	default:
// 		return Either(Try(parsers[0]), Choice(parsers[1:]))
// 	}
// }
