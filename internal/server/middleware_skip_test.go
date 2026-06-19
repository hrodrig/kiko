package server

import "testing"

func TestMiddlewareSkipMatches(t *testing.T) {
	skip := MiddlewareSkip{
		Exact:    []string{"/exact"},
		Prefixes: []string{"/prefix/"},
	}
	tests := []struct {
		path string
		want bool
	}{
		{"/exact", true},
		{"/other", false},
		{"/prefix/foo", true},
		{"/prefixed", false},
	}
	for _, tt := range tests {
		if got := skip.matches(tt.path); got != tt.want {
			t.Errorf("matches(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
