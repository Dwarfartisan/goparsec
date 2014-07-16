// parsec 的部分代码实现参考了 https://github.com/sanyaade-buildtools/goparsec
// 和 https://github.com/prataprc/goparsec
// 但是我需要一个面向 unicode 的简洁实现，所以只好重写了自己的版本。
package goparsec

import (
	"fmt"
	"strings"
)

type Parser func(ParseState) (interface{}, error)

// 因为几个基础的 parser 获取到的是 []interface{} ，内部保存 string 或 rune 。所以经常遇到传递出来的
// inteface{} 要转为 []string 再转 string 的情况，所以这里提供两个
// 工具函数。

// func ExtraString 将 interface{} 转成 string，如果输入数据与前面提到的规范不符，会 panic
func ExtractString(input interface{}) string {
	data := input.([]interface{})
	l := len(data)
	buffer := make([]string, l)
	for index, item := range data {
		switch it := item.(type) {
		case string:
			buffer[index] = it
		case rune:
			buffer[index] = string([]rune{it})
		default:
			buffer[index] = fmt.Sprintf("%v", it)
		}
	}
	return strings.Join(buffer, "")
}

// func ReturnString 用 Return 包装 ExtraString，使其适用于 Bind 这样的组合子。
func ReturnString(input interface{}) Parser {
	return Return(ExtractString(input))
}
