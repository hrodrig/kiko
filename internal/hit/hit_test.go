package hit

import "testing"

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
