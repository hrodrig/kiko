package cli

import (
	"testing"

	"github.com/hrodrig/kiko/internal/version"
)

func TestVersionOutput(t *testing.T) {
	// just verify package doesn't panic on basic use
	version.Version = "0.1.0"
	version.Commit = "abc123"
	code := Execute()
	if code != 0 {
		t.Errorf("Execute() = %d; want 0", code)
	}
}
