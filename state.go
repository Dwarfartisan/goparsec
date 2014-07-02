// parser state 包参考了 golang 的内置包定义，部分代码模仿或来自 text/scanner ，其中有向
// https://github.com/sanyaade-buildtools/goparsec 学习一部分设计思路
package goparsec

import (
	"errors"
	"fmt"
	"io"
)

type ParseError struct {
	Line    int
	Column  int
	Pos     int
	Message string
}

func (err ParseError) Error() string {
	return fmt.Sprintf("pos %d line %d column %d:\n%s",
		err.Pos, err.Line, err.Column, err.Message)
}

type ParseState interface {
	Next(pred func(rune) bool) (r rune, ok bool, err error)
	Line() int
	Column() int
	Pos() int
	SeekTo(int)
	Trap(message string, args ...interface{}) error
}

type StateInMemory struct {
	buffer   []rune
	newLines []int
	line     int
	column   int
	pos      int
}

func MemoryParseState(data string) ParseState {
	buffer := ([]rune)(data)
	newLines := []int{}
	last := len(buffer) - 1
	for idx, r := range buffer {
		if r == '\n' {
			newLines = append(newLines, idx)
		}
	}
	if buffer[last] != '\n' {
		newLines = append(newLines, last)
	}
	return &StateInMemory{buffer, newLines, 1, 1, 0}
}

func (this *StateInMemory) Next(pred func(rune) bool) (r rune, match bool, err error) {
	buffer := (*this).buffer
	if (*this).pos < len(buffer) {
		ru := buffer[(*this).pos]
		if pred(ru) {
			(*this).pos++
			if ru == '\r' {
				(*this).line++
				(*this).column = 0
			} else {
				(*this).column++
			}
			return ru, true, nil
		} else {
			return ru, false, nil
		}
	} else {
		return '\000', false, io.EOF
	}
}

func (this *StateInMemory) Line() int {
	return (*this).line
}

func (this *StateInMemory) Column() int {
	return (*this).column
}

func (this *StateInMemory) Pos() int {
	return (*this).pos
}

func (this *StateInMemory) SeekTo(pos int) {
	end := len((*this).buffer) - 1
	if pos < 0 || pos > end {
		message := fmt.Sprintf("%d out range [0, %d]", pos, end)
		panic(errors.New(message))
	}
	(*this).pos = pos
	top := len((*this).newLines) - 1
	for idx, _ := range (*this).newLines {
		line := top - idx
		start := (*this).newLines[line]
		if start < pos {
			(*this).line = line + 2
			(*this).column = pos - start
			return
		}
	}
}

func (this *StateInMemory) Trap(message string, args ...interface{}) error {
	return ParseError{(*this).line, (*this).column, (*this).pos,
		fmt.Sprintf(message, args...)}
}
