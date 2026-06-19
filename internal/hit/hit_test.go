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
	}
	for _, tt := range tests {
		tt.in.Normalize()
		if tt.in.Path != tt.want {
			t.Errorf("Normalize({Path:%q}) = %q; want %q", tt.in.Path, tt.in.Path, tt.want)
		}
	}
}
