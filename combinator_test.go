package goparsec

import (
	"testing"
)

func TestBinds_(t *testing.T) {
	data := "int opt int"
	st := MemoryParseState(data)
	checker := Binds_(String("int"), Spaces, String("opt"), Spaces, String("int"), Eof)
	_, err := checker(st)
	if err != nil {
		t.Fatalf("expect the Binds_ checker success but %v", err)
	}
}

func TestBindsFail_(t *testing.T) {
	data := "int opt float"
	st := MemoryParseState(data)
	checker := Binds_(String("int"), Spaces, String("opt"), Spaces, String("int"), Eof)
	_, err := checker(st)
	if err == nil {
		t.Fatalf("expect the Binds_ checker failed at \"%s\"", data)
	}
}
