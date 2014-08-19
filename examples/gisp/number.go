package gisp

import (
	. "github.com/Dwarfartisan/goparsec"
	"strconv"
)

func NumberParser(st ParseState) (interface{}, error) {
	f, err := Try(Float)(st)
	if err == nil {
		return strconv.ParseFloat(f.(string), 64)
	}
	i, err := Int(st)
	if err == nil {
		return strconv.Atoi(i.(string))
	}
	return nil, err
}
