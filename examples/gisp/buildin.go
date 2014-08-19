package gisp

import (
	"fmt"
	"reflect"
)

var Axiom = Environment{
	Meta: map[string]interface{}{
		"name":     "axiom",
		"category": "environment",
	},
	Content: map[string]function{
		"quote": func(env Env) element {
			return func(args ...interface{}) (interface{}, error) {
				return Quote{args[0]}, nil
			}
		},
		"var": func(env Env) element {
			return func(args ...interface{}) (interface{}, error) {
				value, err := eval(env, args[1])
				if err != nil {
					return nil, err
				}
				err = env.Define((args[0].(Atom)).Name, value)
				return nil, err
			}
		},
		"set": func(env Env) element {
			return func(args ...interface{}) (interface{}, error) {
				value, err := eval(env, args[1])
				if err != nil {
					return nil, err
				}
				err = env.SetVar((args[0].(Atom)).Name, value)
				if err == nil {
					return nil, err
				} else {
					return value, nil
				}
			}
		},
		"equal": func(env Env) element {
			return func(args ...interface{}) (interface{}, error) {
				x, err := eval(env, args[0])
				if err != nil {
					return nil, err
				}
				y, err := eval(env, args[1])
				if err != nil {
					return nil, err
				}
				return reflect.DeepEqual(x, y), nil
			}
		},
		"cond": func(env Env) element {
			return func(args ...interface{}) (interface{}, error) {
				cases := args[0].([]interface{})
				l := len(args)
				var els interface{}
				if l > 1 {
					els = args[1]
				} else {
					els = nil
				}

				for _, b := range cases { // FIXME: need a else
					branch := b.([]interface{})
					cond := branch[0].(List)
					result, err := eval(env, cond)
					if err != nil {
						return nil, err
					}
					if ok := result.(bool); ok {
						return eval(env, branch[1])
					}
				}
				// else branch
				if els != nil {
					return eval(env, els)
				} else {
					return nil, nil
				}
			}
		},
		"car": func(env Env) element {
			return func(args ...interface{}) (interface{}, error) {
				// FIXME: out range error
				lisp, err := eval(env, args[0])
				if err != nil {
					return nil, err
				}
				return (lisp.(List))[0], nil
			}
		},
		"cdr": func(env Env) element {
			return func(args ...interface{}) (interface{}, error) {
				// FIXME: out range error
				lisp, err := eval(env, args[0])
				if err != nil {
					return nil, err
				}
				return (lisp.(List))[1:], nil
			}
		},
		// atom while true both lisp atom or go value
		"atom": func(env Env) element {
			return func(args ...interface{}) (interface{}, error) {
				arg := args[0]
				if l, ok := arg.(List); ok {
					return len(l) == 0, nil
				} else {
					return true, nil
				}
			}
		},
		// 照搬 cons 运算符对于 golang 嵌入没有足够的收益，这里的 concat 是一个 cons 的变形，
		// 它总是返回包含所有参数的 List 。
		"concat": func(env Env) element {
			return func(args ...interface{}) (interface{}, error) {
				return List(args), nil
			}
		},
	},
}

var Propositions Environment = Environment{
	Meta: map[string]interface{}{
		"name":     "propositions",
		"category": "environment",
	},
	Content: map[string]function{
		"lambda": LambdaExpr,
		"let":    LetExpr,
		"+":      addExpr,
		"add":    addExpr,
		"-":      subExpr,
		"sub":    subExpr,
		"*":      mulExpr,
		"mul":    mulExpr,
		"/":      divExpr,
		"div":    divExpr,
	},
}

type Lambda struct {
	Meta    map[string]interface{}
	Content List
}

// (lambda (args...) body)
func DeclareLambda(env Env, args List, lisps ...interface{}) (*Lambda, error) {
	ret := Lambda{map[string]interface{}{
		"category":          "lambda",
		"formal parameters": args,
		"local":             map[string]interface{}{},
	}, List{}}
	prepare := map[string]bool{}
	for _, lisp := range lisps {
		err := ret.prepare(env, prepare, lisp)
		if err != nil {
			return nil, err
		}
	}
	return &ret, nil
}

func (this *Lambda) prepare(env Env, prepare map[string]bool, content interface{}) error {
	next := map[string]bool{}
	for key := range prepare {
		next[key] = true
	}
	var err error = nil
	switch lisp := content.(type) {
	case Atom:
		err = this.prepareAtom(env, next, lisp)
		return err
	case List:
		err = this.prepareList(env, next, lisp)
	}
	if err == nil {
		this.Content = append(this.Content, content)
	}
	return err
}

func (this Lambda) prepareAtom(env Env, prepare map[string]bool, one Atom) error {
	if _, ok := prepare[one.Name]; ok {
		return nil
	}
	next := map[string]bool{}
	for key := range prepare {
		next[key] = true
	}

	for _, arg := range this.Meta["formal parameters"].(List) {
		if (arg.(Atom)).Name == one.Name {
			return nil
		}
	}
	if _, ok := prepare[one.Name]; !ok {
		if v, ok := env.Lookup(one.Name); ok {
			local := (this.Meta["local"]).(map[string]interface{})
			local[one.Name] = v
		} else {
			return fmt.Errorf("%s not found", one.Name)
		}
	}
	return nil
}

func (this Lambda) prepareList(env Env, prepare map[string]bool, content List) error {
	next := map[string]bool{}
	for key := range prepare {
		next[key] = true
	}
	var err error = nil
	fun := content[0].(Atom)
	switch fun.Name {
	case "define":
		name := content[1].(string)
		if err != nil {
			return err
		} else {
			next[name] = true
		}
	case "lambda":
		args := content[1].(List)
		for _, a := range args {
			arg := a.(Atom)
			next[arg.Name] = true
		}
	case "let":
		for _, def := range content[1].(List) {
			define := def.(List)
			name := define[0].(string)
			next[name] = true
		}
	}
	for _, l := range content {
		switch lisp := l.(type) {
		case List:
			err = this.prepareList(env, next, lisp)
		case Atom:
			err = this.prepareAtom(env, next, lisp)
		}
	}
	return err
}

// create a lambda s-expr can be eval
func (this Lambda) Call(args ...interface{}) Function {
	meta := map[string]interface{}{}
	for k, v := range this.Meta {
		meta[k] = v
	}
	meta["actual parameters"] = List(args)
	meta["my"] = map[string]interface{}{}
	l := len(this.Content)
	content := make([]interface{}, l)
	for idx, data := range this.Content {
		content[idx] = data
	}
	return Function{meta, content}
}

func LambdaExpr(env Env) element {
	return func(args ...interface{}) (interface{}, error) {
		_args := args[0].(List)
		ret, err := DeclareLambda(env, _args, args[1:]...)
		if err == nil {
			return *ret, nil
		} else {
			return nil, err
		}
	}
}

type Function struct {
	Meta    map[string]interface{}
	Content []interface{}
}

func (this Function) Local(name string) (interface{}, bool) {
	my := this.Meta["my"].(map[string]interface{})
	if value, ok := my[name]; ok {
		return value, true
	}
	if value, ok := this.Parameter(name); ok {
		return value, true
	}

	local := this.Meta["local"].(map[string]interface{})
	value, ok := local[name]
	return value, ok
}

func (this Function) Parameter(name string) (interface{}, bool) {
	formals := this.Meta["formal parameters"].(List)
	actuals := this.Meta["actual parameters"].(List)
	for idx := range formals {
		formal := formals[idx].(Atom)
		if formal.Name == name {
			return actuals[idx], true
		}
	}
	return nil, false
}

func (this Function) Global(name string) (interface{}, bool) {
	global := this.Meta["global"].(Env)
	return global.Lookup(name)
}

func (this Function) Lookup(name string) (interface{}, bool) {
	if value, ok := this.Local(name); ok {
		return value, true
	} else {
		return this.Global(name)
	}
}

func (this Function) SetVar(name string, value interface{}) error {
	mine := this.Meta["my"].(map[string]interface{})
	if _, ok := mine[name]; ok {
		mine[name] = value
		return nil
	} else {
		local := this.Meta["local"].(map[string]interface{})
		if _, ok := local[name]; ok {
			local[name] = value
			return nil
		} else {
			global := this.Meta["global"].(Env)
			return global.SetVar(name, value)
		}
	}
}

func (this Function) Define(name string, value interface{}) error {
	mine := this.Meta["my"].(map[string]interface{})
	if _, ok := mine[name]; ok {
		return fmt.Errorf("%s was exists.", name)
	} else {
		mine[name] = value
		return nil
	}
}

func (this Function) Eval(env Env) (interface{}, error) {
	args := this.Meta["actual parameters"].(List)
	actual := make(List, len(args))
	for idx, arg := range args {
		a, err := eval(env, arg)
		if err == nil {
			actual[idx] = a
		} else {
			return nil, err
		}
	}
	this.Meta["actual parameters"] = args
	this.Meta["global"] = env
	l := len(this.Content)
	switch l {
	case 0:
		return nil, nil
	case 1:
		return eval(this, this.Content[0])
	default:
		for _, expr := range this.Content[:l-2] {
			_, err := eval(this, expr)
			if err != nil {
				return nil, err
			}
		}
		return eval(this, this.Content[l-1])
	}
}

type Let struct {
	Meta    map[string]interface{}
	Content List
}

// let => (let ((a, value), (b, value)...) ...)
func LetExpr(env Env) element {
	return func(args ...interface{}) (interface{}, error) {
		local := map[string]interface{}{}
		vars := args[0].(List)
		for _, v := range vars {
			declares := v.(List)
			name := (declares[0].(Atom)).Name
			value, err := eval(env, (declares[1]))
			if err != nil {
				return nil, err
			}
			local[name] = value
		}
		meta := map[string]interface{}{
			"local": map[string]interface{}{},
		}
		let := Let{meta, args}
		return let.Eval(env)
	}
}

func (this Let) Define(name string, value interface{}) error {
	if _, ok := this.Local(name); ok {
		return fmt.Errorf("local name %s is exists", name)
	}
	local := this.Meta["local"].(map[string]interface{})
	local[name] = value
	return nil
}

func (this Let) SetVar(name string, value interface{}) error {
	if _, ok := this.Local(name); ok {
		local := this.Meta["local"].(map[string]interface{})
		local[name] = value
		return nil
	} else {
		global := this.Meta["global"].(Env)
		return global.SetVar(name, value)
	}
}

func (this Let) Local(name string) (interface{}, bool) {
	local := this.Meta["local"].(map[string]interface{})
	value, ok := local[name]
	return value, ok
}

func (this Let) Lookup(name string) (interface{}, bool) {
	if value, ok := this.Local(name); ok {
		return value, true
	} else {
		return this.Global(name)
	}
}

func (this Let) Global(name string) (interface{}, bool) {
	global := this.Meta["global"].(Env)
	return global.Lookup(name)
}

func (this Let) Eval(env Env) (interface{}, error) {
	this.Meta["global"] = env
	l := len(this.Content)
	switch l {
	case 0:
		return nil, nil
	case 1:
		return eval(this, this.Content[0])
	default:
		for _, expr := range this.Content[:l-2] {
			eval(this, expr)
		}
		expr := this.Content[l-1]
		return eval(this, expr)
	}
}

func tofloat64(env Env, arg interface{}) (float64, error) {
	switch value := arg.(type) {
	case float64:
		return value, nil
	case float32:
		return float64(value), nil
	case int:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case Lisp:
		v, err := value.Eval(env)
		if err != nil {
			return 0.0, err
		}
		return tofloat64(env, v)
	default:
		return 0.0, fmt.Errorf("%v isn't avalid number", arg)
	}
}

func addExpr(env Env) element {
	return func(args ...interface{}) (interface{}, error) {
		x := 0.0
		for _, arg := range args {
			v, err := tofloat64(env, arg)
			if err == nil {
				x += v
			} else {
				return nil, err
			}
		}
		return x, nil
	}
}

func subExpr(env Env) element {
	return func(args ...interface{}) (interface{}, error) {
		x, err := tofloat64(env, args[0])
		if err != nil {
			return nil, err
		}
		for _, arg := range args[1:] {
			value, err := tofloat64(env, arg)
			if err != nil {
				return nil, err
			}
			x -= value
		}
		return x, nil
	}
}

func mulExpr(env Env) element {
	return func(args ...interface{}) (interface{}, error) {
		x, err := tofloat64(env, args[0])
		if err != nil {
			return nil, err
		}
		for _, arg := range args[1:] {
			value, err := tofloat64(env, arg)
			if err != nil {
				return nil, err
			}
			x *= value
		}
		return x, nil
	}
}

func divExpr(env Env) element {
	return func(args ...interface{}) (interface{}, error) {
		x, err := tofloat64(env, args[0])
		if err != nil {
			return nil, err
		}
		for _, arg := range args[1:] {
			value, err := tofloat64(env, arg)
			if err != nil {
				return nil, err
			}
			x /= value
		}
		return x, nil
	}
}
