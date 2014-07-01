// parsec 的部分代码实现参考了 https://github.com/sanyaade-buildtools/goparsec
// 和 https://github.com/prataprc/goparsec
// 但是我需要一个面向 unicode 的简洁实现，所以只好重写了自己的版本。
package gparsec

type Parser func(ParseState) (interface{}, error)
