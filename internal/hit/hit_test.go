package hit

import "testing"

func TestNewBufferDefaultCapacity(t *testing.T) {
	b := NewBuffer(0)
	for i := 0; i < 4096; i++ {
		b.Append(Hit{Host: "x", Path: "/"})
	}
	b.Append(Hit{Host: "x", Path: "/"})
	if b.Drops() != 1 {
		t.Errorf("Drops() = %d; want 1", b.Drops())
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		in   Hit
		want string
	}{
		{Hit{Path: ""}, "/"},
		{Hit{Path: "/"}, "/"},
		{Hit{Path: "/blog"}, "/blog"},
		{Hit{Path: "/blog?page=2"}, "/blog?page=2"},
		{Hit{Path: "/blog?utm_source=x&page=2"}, "/blog?page=2"},
	}
	for _, tt := range tests {
		h := tt.in
		h.Normalize()
		if h.Path != tt.want {
			t.Errorf("Normalize({Path:%q}) = %q; want %q", tt.in.Path, h.Path, tt.want)
		}
	}
}

func TestNormalizeUTM(t *testing.T) {
	h := Hit{Path: "/?utm_source=newsletter&utm_medium=email"}
	h.Normalize()
	if h.UTMSource != "newsletter" || h.UTMMedium != "email" {
		t.Errorf("utm = %+v", h)
	}
}
