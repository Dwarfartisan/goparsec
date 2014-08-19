package gisp

import (
	"fmt"
	. "github.com/Dwarfartisan/goparsec"
	"reflect"
	"unicode"
)

const LISP reflect.Kind = reflect.Invalid + 256

type Atom struct {
	Name string
}

func (this Atom) String() string {
	return this.Name
}

func (this Atom) Eval(env Env) (interface{}, error) {
	if value, ok := env.Lookup(this.Name); ok {
		return value, nil
	} else {
		return nil, fmt.Errorf("value of atom %s not found", this.Name)
	}
}

func nameChecker(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

func AtomParser(st ParseState) (interface{}, error) {
	a, err := Bind(Many1(NoneOf("'() \t\r\n.")),
		ReturnString)(st)
	if err == nil {
		return Atom{a.(string)}, nil
	} else {
		return nil, err
	}
}
