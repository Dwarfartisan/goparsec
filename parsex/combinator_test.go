package parsex

import (
	"testing"
	"time"
)

// 该测试用例模拟一个词法分析后的 token 流，定义一个时间区间规则，生成结果是一个匹配函数。
// 定义规则为 [from xxx] [to xxx] 。
// 匹配函数检查输入的时间是否在匹配区间内，并给出左边界到输入时间的间隔值。对于 to Time 校验，
// 因为没有左边界，返回是否在区间的校验结果，和 nil 。
//
// 这个验证证明 parsex 可以接在词法分析后面作为语法分析器使用，或者直接构造一个一般性的规则系统。

var (
	now       = time.Now()
	dur, _    = time.ParseDuration("-24h")
	yesterday = now.Add(dur)
)

var fromParser = Bind_(String("from"), TimeVal)

func from(st ParsexState) (interface{}, error) {
	if ts, err := fromParser(st); err == nil {
		return func(t time.Time) (bool, interface{}) {
			return !t.Before(ts.(time.Time)), t.Sub(ts.(time.Time))
		}, nil
	} else {
		return nil, err
	}
}

var toParser = Bind_(String("to"), TimeVal)

func to(st ParsexState) (interface{}, error) {
	if ts, err := toParser(st); err == nil {
		return func(t time.Time) (bool, interface{}) {
			return !t.After(ts.(time.Time)), nil
		}, nil
	} else {
		return nil, err
	}
}

var fromToParser = Union(fromParser, toParser)

func fromto(st ParsexState) (interface{}, error) {
	if ts, err := fromToParser(st); err == nil {
		return func(ti time.Time) (bool, interface{}) {
			ft := ts.([]interface{})
			f := ft[0].(time.Time)
			t := ft[1].(time.Time)
			return !ti.Before(f) && !ti.After(t), f.Sub(ti)
		}, nil
	} else {
		return nil, err
	}
}

func TestFromP0(t *testing.T) {
	var fromData0 = []interface{}{"from", now}
	state := &StateInMemory{fromData0, 0}
	val, err := fromParser(state)
	if err == nil {
		t.Logf("success:%v\n", val)
	} else {
		t.Fatalf("except extract the time value %v but %v", now, err)
	}
}

func TestFrom0(t *testing.T) {
	fromData0 := []interface{}{"from", yesterday}
	state := &StateInMemory{fromData0, 0}
	chk, err := from(state)
	checker := chk.(func(time.Time) (bool, interface{}))
	if err == nil {
		if ok, dur := checker(now); ok {
			t.Logf("now after yesterday %v", dur)
		} else {
			t.Logf("except now after yesterday 24hours but %v", dur)
		}
	} else {
		t.Fatalf("except extract the time checker from %v but %v", now, err)
	}
}

func TestFrom1(t *testing.T) {
	fromData0 := []interface{}{"from", now}
	state := &StateInMemory{fromData0, 0}
	chk, err := from(state)
	checker := chk.(func(time.Time) (bool, interface{}))
	if err == nil {
		if ok, dur := checker(now); ok {
			t.Logf("now after yesterday %v", dur)
		} else {
			t.Logf("except now after yesterday 0hours but %v", dur)
		}
	} else {
		t.Fatalf("except extract the time checker from %v but %v", now, err)
	}
}

func TestFrom2(t *testing.T) {
	fromData0 := []interface{}{"from", now}
	state := &StateInMemory{fromData0, 0}
	chk, err := from(state)
	checker := chk.(func(time.Time) (bool, interface{}))
	if err == nil {
		if ok, dur := checker(yesterday); !ok {
			t.Logf("now after yesterday %v", dur)
		} else {
			t.Logf("except now after yesterday 0hours but %v", dur)
		}
	} else {
		t.Fatalf("except extract the time checker from %v but %v", now, err)
	}
}

func TestToP0(t *testing.T) {
	var toData0 = []interface{}{"to", now}
	state := &StateInMemory{toData0, 0}
	val, err := toParser(state)
	if err == nil {
		t.Logf("success:%v\n", val)
	} else {
		t.Fatalf("except extract the time value %v but %v", now, err)
	}
}

func TestTo0(t *testing.T) {
	toData0 := []interface{}{"to", now}
	state := &StateInMemory{toData0, 0}
	chk, err := to(state)
	if err == nil {
		checker := chk.(func(time.Time) (bool, interface{}))
		if ok, dur := checker(yesterday); !ok {
			t.Fatalf("except now after yesterday 24hours but %v", dur)
		}
	} else {
		t.Fatalf("except extract the time checker from %v but %v", now, err)
	}
}

func TestTo1(t *testing.T) {
	toData0 := []interface{}{"to", now}
	state := &StateInMemory{toData0, 0}
	chk, err := to(state)
	if err == nil {
		checker := chk.(func(time.Time) (bool, interface{}))
		if ok, dur := checker(now); !ok {
			t.Fatalf("except now after yesterday 0hours but %v", dur)
		}
	} else {
		t.Fatalf("except extract the time checker from %v but %v", now, err)
	}
}

func TestTo2(t *testing.T) {
	toData0 := []interface{}{"to", yesterday}
	state := &StateInMemory{toData0, 0}
	chk, err := to(state)
	if err == nil {
		checker := chk.(func(time.Time) (bool, interface{}))
		if ok, dur := checker(now); ok {
			t.Fatalf("except now after yesterday 0hours but %v", dur)
		}
	} else {
		t.Fatalf("except extract the time checker from %v but %v", now, err)
	}
}

func TestFromToP0(t *testing.T) {
	var toData0 = []interface{}{"from", yesterday, "to", now}
	state := &StateInMemory{toData0, 0}
	val, err := fromToParser(state)
	if err == nil {
		t.Logf("success:%v\n", val)
	} else {
		t.Fatalf("except extract the time values %v but %v", []interface{}{yesterday, now}, err)
	}
}

func TestFromTo0(t *testing.T) {
	data := []interface{}{"from", yesterday, "to", now}
	state := &StateInMemory{data, 0}
	checker, err := fromto(state)
	if err == nil {
		d, _ := time.ParseDuration("12h")
		date := yesterday.Add(d)
		if ok, dur := checker.(func(time.Time) (bool, interface{}))(date); ok {
			t.Logf("success:%v\n", dur)
		} else {
			t.Fatalf("except find %v in duration from %v to %v but failed", date, yesterday, now)
		}
	} else {
		t.Fatalf("except create a duration checker from %v to %v but failed: %v", yesterday, now, err)
	}
}
