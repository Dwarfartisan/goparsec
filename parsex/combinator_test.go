package parsex

import (
	"fmt"
	"testing"
	"time"
)

var fromParser = Bind_(String("from"), TimeVar)

func TestFrom(t *testing.T) {
	theTime := time.Now()
	var fromData0 = []interface{}{"from", theTime}
	state := &StateInMemory{fromData0, 0}
	val, err := fromParser(state)
	if err == nil {
		fmt.Sprintf("success:%v\n", val)
	} else {
		t.Fatalf("except extract the time value %v", theTime)
	}
}
