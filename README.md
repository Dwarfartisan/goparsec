# GoParsec

Haskell Parsec Libraray's golang version

PLEASE USE [goparsec2](http://github.com/Dwarfartisan/goparsec2) BUT THIS. GOPARSEC2 LIKE HASKELL PRASEC LIBRARY THAN GOPARSEC. IT SUPPORT MONAD BIND/THEN/OVER.

## Document

### Paper Style

http://bms.tratao.com/desktop/doc/0c3802e4ee404a71407f34996eff98ef

### Slide Style

http://bms.tratao.com/slide/simple/0c3802e4ee404a71407f34996eff98ef##1

## 概述

还是那句话，公司自用的东西肯定不会坑，但是可能达到一个可用阶段后会停一段时间。

第一步先支持几个常用组合子，让公司的项目可以做下去，然后……

我的项目大量的参考了 https://github.com/sanyaade-buildtools/goparsec ，但是这个项目面向
byte流，而我需要一个面向 unicode 或更通用的规则解析的工具。所以只好重新实现了一遍。

这里 https://github.com/vito/go-parse 还有一个实现，但是五年没有更新了。

## 测试

由于用在团队内部一个 markdown like 的文档内容解析功能上， 所以目前的测试用例都是基于 markdown 转
json 这个场景展开的。

对 parsec 和 parsex 的深度运用，可以参见 http://github.com/Dwarfartisan/gisp ，这是
一个类 lisp 语言的解释器，里面大量运用了 parsec 和 parsex 。

## Parsex

Parsex 是 go parsec 的一个包。parsec 的主干是以一个字符流 state parser 接口为基础的。在工作
中，我遇到了一些更通用的规则构造和解析的需求。所以在 parsec/parsex 包里，我写了一个基于
[]interface{} 序列的组合子功能。其实这本质上是因为 golang 没有泛型，否则直接把 state parer
泛型化就可以了。这也是 Haskell 原版的 parsec 没有这种区分的原因。

parsex/combinator_test.go 文件中包含了一个测试用例，可以看作是对一个经过词法解析的简单 token
流的语法解析。

## 最后

如果有同行参与进来一起把它通用化当然最好啦，口嫌体正直什么的我是不会做的 ε-(´∀｀; ) 。

对了我还准备再用 swift 做一遍，这个主意怎么看都酷毙了，目前唯一就缺个会写代码的了……
