package store

import (
	"testing"

	"github.com/hrodrig/kiko/internal/hit"
)

func TestNopStore(t *testing.T) {
	n := NewNop()
	if err := n.SaveHits(nil); err != nil {
		t.Errorf("NopStore.SaveHits(nil) = %v", err)
	}
	if err := n.SaveHits([]hit.Hit{}); err != nil {
		t.Errorf("NopStore.SaveHits([]) = %v", err)
	}
	if err := n.SaveHits([]hit.Hit{{Host: "test", Path: "/"}}); err != nil {
		t.Errorf("NopStore.SaveHits([hit]) = %v", err)
	}
}
