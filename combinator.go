package goparsec

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
			return nil, st.Trap("Excepted one of \"%s\" but got '%c'", runes, r)
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
			return nil, st.Trap("Excepted none of \"%s\" but got '%c'", string(runes), r)
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

func Union(parsers ...Parser) Parser {
	return func(st ParseState) (interface{}, error) {
		var ret = make([]interface{}, 0, len(parsers))
		for _, parser := range parsers {
			val, err := parser(st)
			if err == nil {
				if val != nil {
					ret = append(ret, val)
				}
			} else {
				return nil, err
			}
		}
		return ret, nil
	}
}

func UnionAll(parsers ...Parser) Parser {
	return func(st ParseState) (interface{}, error) {
		var ret = make([]interface{}, 0, len(parsers))
		for _, parser := range parsers {
			val, err := parser(st)
			if err == nil {
				ret = append(ret, val)
			} else {
				return nil, err
			}
		}
		return ret, nil
	}
}

// Choice 实现为以下逻辑的迭代版本：
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
// 其实我比较希望把下面那个东西实现成上面这个样子，就是好像在golang里不太经济……
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

// Binds 相当于用 Bind 对一个 func(interface{})Parser 链做左折叠，起始参数为 first
func Binds(first Parser, then ...func(interface{}) Parser) Parser {
	if len(then) == 0 {
		return first
	}
	return func(st ParseState) (interface{}, error) {
		ret, err := first(st)
		if err != nil {
			return nil, err
		}
		next := then[0](ret)
		return Binds(next, then[1:]...)(st)
	}
}

// Binds_ 逐个尝试每一个 Parser，直至发生错误或者到达最后，如果到达最后一个 Parser，
// 返回其结果
func Binds_(parsers ...Parser) Parser {
	if len(parsers) < 2 {
		return Fail("combinator Binds_ need parsers more than 2 as args")
	}
	if len(parsers) == 2 {
		return Bind_(parsers[0], parsers[1])
	}
	return Bind_(parsers[0], Binds_(parsers[1:]...))
}
