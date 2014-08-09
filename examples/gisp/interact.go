package gisp

import (
	parsec "github.com/Dwarfartisan/goparsec"
)

var ParseValue = parsec.Choice(StringVal, Number)
