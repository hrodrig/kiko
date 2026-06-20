package server

import "testing"

func TestHostRateLimiter(t *testing.T) {
	h := NewHostRateLimiter(1, 1)
	if !h.Allow("example.com") {
		t.Fatal("first allow")
	}
	if h.Allow("example.com") {
		t.Fatal("second should block")
	}
	if !h.Allow("other.com") {
		t.Fatal("different host should allow")
	}
}

func TestHostRateLimiterDisabled(t *testing.T) {
	if NewHostRateLimiter(0, 0) != nil {
		t.Fatal("zero rps should disable")
	}
}
