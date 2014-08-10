package gisp

import (
	. "github.com/Dwarfartisan/goparsec"
	"unicode"
)

type Atom struct {
	Name string
}

func (this Atom) String() string {
	return this.Name
}

func nameChecker(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

func AtomParser(st ParseState) (interface{}, error) {
	a, err := Bind(Many(RuneChecker(nameChecker, "letter or number")),
                ReturnString)(st)
	if err == nil {
		return Atom{a.(string)}, nil
	} else {
		return nil, err
	}
}
