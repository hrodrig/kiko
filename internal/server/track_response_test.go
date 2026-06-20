package server

import "testing"

func TestBoolStr(t *testing.T) {
	if boolStr(true) != "true" || boolStr(false) != "false" {
		t.Fatal("boolStr")
	}
}
