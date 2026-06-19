package store

import (
	"strings"
	"testing"
)

func TestInsertHitSQL(t *testing.T) {
	if !strings.Contains(insertHitSQL("postgres"), "$1") {
		t.Fatal("postgres insert should use $1 placeholders")
	}
	if !strings.Contains(insertHitSQL("sqlite"), "?") {
		t.Fatal("sqlite insert should use ? placeholders")
	}
}
