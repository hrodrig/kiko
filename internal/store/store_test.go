package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
)

func TestNopStore(t *testing.T) {
	n := NewNop()
	if err := n.SaveHits(nil); err != nil {
		t.Errorf("NopStore.SaveHits(nil) = %v", err)
	}
	if err := n.Ping(context.Background()); err != nil {
		t.Errorf("NopStore.Ping() = %v", err)
	}
	if err := n.Close(); err != nil {
		t.Errorf("NopStore.Close() = %v", err)
	}
}

func TestSQLiteOpenAndSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kiko.db")
	st, err := Open(config.DatabaseCfg{Driver: "sqlite", Path: path})
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	defer st.Close()

	hits := []hit.Hit{
		{Host: "test.dev", Path: "/blog", Referrer: "https://google.com", Title: "Post", Width: 1920},
		{Host: "test.dev", Path: "/"},
	}
	if err := st.SaveHits(hits); err != nil {
		t.Fatalf("SaveHits() = %v", err)
	}
	if err := st.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() = %v", err)
	}

	// second open should migrate idempotently
	st2, err := Open(config.DatabaseCfg{Driver: "sqlite", Path: path})
	if err != nil {
		t.Fatalf("re-Open() = %v", err)
	}
	st2.Close()
}

func TestOpenUnsupportedDriver(t *testing.T) {
	_, err := Open(config.DatabaseCfg{Driver: "oracle"})
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}
