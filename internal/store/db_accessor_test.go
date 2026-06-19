package store_test

import (
	"path/filepath"
	"testing"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/store"
)

func TestDBAccessor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kiko.db")
	st, err := store.Open(config.DatabaseCfg{Driver: "sqlite", Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	acc, ok := st.(store.DBAccessor)
	if !ok {
		t.Fatal("missing DBAccessor")
	}
	db, driver := acc.StatsDB()
	if db == nil || driver != "sqlite" {
		t.Fatalf("db=%v driver=%q", db, driver)
	}
}
