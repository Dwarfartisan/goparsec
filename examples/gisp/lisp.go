package gisp

import (
	"fmt"
	. "github.com/Dwarfartisan/goparsec"
	"strings"
)

type Env interface {
	Define(name string, value interface{}) error
	SetVar(name string, value interface{}) error
	Lookup(name string) (interface{}, bool)
	Local(name string) (interface{}, bool)
	Global(name string) (interface{}, bool)
}

type Environment struct {
	Meta    map[string]interface{}
	Content map[string]function
}

func (this Environment) Lookup(name string) (interface{}, bool) {
	if v, ok := this.Local(name); ok {
		return v, true
	} else {
		return this.Global(name)
	}
}

func (this Environment) Local(name string) (interface{}, bool) {
	if v, ok := this.Content[name]; ok {
		return v, true
	} else {
		return nil, false
	}
}

func (this Environment) Global(name string) (interface{}, bool) {
	if o, ok := this.Meta["global"]; ok {
		outer := o.(Env)
		return outer.Lookup(name)
	} else {
		return nil, false
	}
}

func eval(env Env, lisp interface{}) (interface{}, error) {
	// a lisp data or go value
	if l, ok := lisp.(Lisp); ok {
		value, err := l.Eval(env)
		return value, err
	} else {
		return lisp, nil
	}
}

type element func(args ...interface{}) (interface{}, error)
type function func(Env) element

type Lisp interface {
	Eval(env Env) (interface{}, error)
}

type List []interface{}

func (this List) String() string {
	frags := []string{}
	for _, item := range this {
		frags = append(frags, fmt.Sprintf("%v", item))
	}
	body := strings.Join(frags, " ")
	return fmt.Sprintf("(%s)", body)
}

func (this List) Eval(env Env) (interface{}, error) {
	l := len(this)
	if l == 0 {
		return nil, nil
	} else {
		var lisp interface{}
		switch fun := this[0].(type) {
		case Atom:
			var ok bool
			if lisp, ok = env.Lookup(fun.Name); !ok {
				return nil, fmt.Errorf("any callable named %s not found", fun.Name)
			}
		case List:
			var err error
			lisp, err = fun.Eval(env)
			if err != nil {
				return nil, err
			}
		}
		switch item := lisp.(type) {
		case function:
			return item(env)(this[1:]...)
		case Function:
			return item.Eval(env)
		case Lambda:
			fun := item.Call(this[1:]...)
			return fun.Eval(env)
		case Let:
			return item.Eval(env)
		default:
			return nil, fmt.Errorf("%v:%t is't callable", this[0], this[0])
		}
	}
}

func bodyParser(st ParseState) (interface{}, error) {
	value, err := SepBy(ValueParser, Many1(Space))(st)
	return value, err
}

func ListParser(st ParseState) (interface{}, error) {
	one := Bind(AtomParser, func(atom interface{}) Parser {
		return Bind_(Rune(')'), Return(List{atom}))
	})
	list, err := Either(Try(Bind_(Rune('('), one)),
		Between(Rune('('), Rune(')'), bodyParser))(st)
	if err == nil {
		return List(list.([]interface{})), nil
	} else {
		return nil, err
	}
}

type Quote struct {
	Lisp interface{}
}

func (this Quote) Eval(env Env) (interface{}, error) {
	return this.Lisp, nil
}

func QuoteParser(st ParseState) (interface{}, error) {
	lisp, err := Bind_(Rune('\''), ValueParser)(st)
	if err == nil {
		return Quote{lisp}, nil
	} else {
		return nil, err
	}
}

func ValueParser(st ParseState) (interface{}, error) {
	value, err := Choice(StringParser,
		NumberParser,
		QuoteParser,
		RuneParser,
		StringParser,
		BoolParser,
		NilParser,
		AtomParser,
		ListParser)(st)
	return value, err
}

type GispParser struct {
	Meta    map[string]interface{}
	Content map[string]interface{}
}

// 给定若干可以组合的基准环境
func NewGisp(buildins map[string]Environment) (*GispParser, error) {
	ret := GispParser{
		Meta: map[string]interface{}{
			"category": "gisp",
			"buildins": buildins,
		},
		Content: map[string]interface{}{},
	}
	return &ret, nil
}

func (this GispParser) Define(name string, value interface{}) error {
	if _, ok := this.Content[name]; ok {
		return fmt.Errorf("var %s exists", name)
	} else {
		this.Content[name] = value
	}
	return nil
}

func (this GispParser) SetVar(name string, value interface{}) error {
	if _, ok := this.Content[name]; ok {
		this.Content[name] = value
		return nil
	} else {
		return fmt.Errorf("Setable var %s not found", name)
	}
}

func (this GispParser) Local(name string) (interface{}, bool) {
	if value, ok := this.Content[name]; ok {
		return value, true
	} else {
		return nil, false
	}
}

func (this GispParser) Lookup(name string) (interface{}, bool) {
	if value, ok := this.Global(name); ok {
		return value, true
	} else {
		return this.Local(name)
	}
}

// look up in buildins
func (this GispParser) Global(name string) (interface{}, bool) {
	buildins := this.Meta["buildins"].(map[string]Environment)
	for _, env := range buildins {
		if v, ok := env.Lookup(name); ok {
			return v, true
		}
	}
	return nil, false
}

func (this GispParser) Parse(code string) (interface{}, error) {
	st := MemoryParseState(code)

	value, err := ValueParser(st)
	if err != nil {
		return nil, err
	}
	switch lisp := value.(type) {
	case Lisp:
		return lisp.Eval(this)
	default:
		return lisp, nil
	}
}
