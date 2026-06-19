package visitor

import (
	"testing"
	"time"
)

func TestHashDeterministic(t *testing.T) {
	fixed := time.Date(2026, 6, 19, 15, 0, 0, 0, time.UTC)
	h := Hasher{salt: "test-salt", now: func() time.Time { return fixed }}

	a := h.Hash("203.0.113.5", "Mozilla/5.0")
	b := h.Hash("203.0.113.5", "Mozilla/5.0")
	if a != b {
		t.Fatalf("hash not deterministic: %q vs %q", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("hash len = %d; want 64", len(a))
	}
}

func TestHashChangesWithDay(t *testing.T) {
	h := Hasher{salt: "test-salt", now: func() time.Time {
		return time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)
	}}
	a := h.Hash("1.2.3.4", "Mozilla/5.0")

	h.now = func() time.Time {
		return time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	}
	b := h.Hash("1.2.3.4", "Mozilla/5.0")
	if a == b {
		t.Fatal("hash should change when UTC day changes")
	}
}

func TestHashChangesWithIPOrUA(t *testing.T) {
	fixed := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)
	h := Hasher{salt: "test-salt", now: func() time.Time { return fixed }}

	base := h.Hash("1.2.3.4", "Mozilla/5.0")
	if base == h.Hash("1.2.3.5", "Mozilla/5.0") {
		t.Error("hash should change with IP")
	}
	if base == h.Hash("1.2.3.4", "Other/1.0") {
		t.Error("hash should change with UA")
	}
}

func TestNewHasherDefaultSalt(t *testing.T) {
	h := NewHasher("")
	if !h.DevSalt() {
		t.Error("empty salt should use dev default")
	}
	if h.Hash("a", "b") == "" {
		t.Fatal("hash empty")
	}
}
