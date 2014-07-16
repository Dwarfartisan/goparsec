// parsex state 包参考了 golang 的内置包定义，部分代码模仿或来自 text/scanner ，其中有向
// https://github.com/sanyaade-buildtools/goparsec 学习一部分设计思路
package parsex

import (
	"errors"
	"fmt"
	"io"
)

type ParsexError struct {
	Pos     int
	Message string
}

func (err ParsexError) Error() string {
	return fmt.Sprintf("pos %d :\n%s",
		err.Pos, err.Message)
}

type ParsexState interface {
	Next(pred func(interface{}) bool) (x interface{}, ok bool, err error)
	Pos() int
	SeekTo(int)
	Trap(message string, args ...interface{}) error
}

type StateInMemory struct {
	buffer []interface{}
	pos    int
}

func (this *StateInMemory) Next(pred func(interface{}) bool) (x interface{}, match bool, err error) {
	buffer := (*this).buffer
	if (*this).pos < len(buffer) {
		x := buffer[(*this).pos]
		if pred(x) {
			(*this).pos++
			return x, true, nil
		} else {
			return x, false, nil
		}
	} else {
		return nil, false, io.EOF
	}
}

func (this *StateInMemory) Pos() int {
	return (*this).pos
}

func (this *StateInMemory) SeekTo(pos int) {
	end := len((*this).buffer)
	if pos < 0 || pos > end {
		message := fmt.Sprintf("%d out range [0, %d]", pos, end)
		panic(errors.New(message))
	}
	(*this).pos = pos
}

func (this *StateInMemory) Trap(message string, args ...interface{}) error {
	return ParsexError{(*this).pos,
		fmt.Sprintf(message, args...)}
}