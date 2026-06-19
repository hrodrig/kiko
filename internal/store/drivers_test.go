package store

import (
	"context"
	"testing"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
)

func TestPostgresOpenFailsWithoutServer(t *testing.T) {
	_, err := Open(config.DatabaseCfg{
		Driver: "postgres",
		Host:   "127.0.0.1",
		Port:   1,
		User:   "kiko",
		DBName: "kiko",
	})
	if err == nil {
		t.Fatal("expected postgres open error")
	}
}

func TestMySQLOpenFailsWithoutServer(t *testing.T) {
	_, err := Open(config.DatabaseCfg{
		Driver: "mysql",
		Host:   "127.0.0.1",
		Port:   1,
		User:   "kiko",
		DBName: "kiko",
	})
	if err == nil {
		t.Fatal("expected mysql open error")
	}
}

func TestSQLiteSaveHitsEmpty(t *testing.T) {
	st, err := Open(config.DatabaseCfg{Driver: "sqlite", Path: ":memory:"})
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	defer st.Close()
	if err := st.SaveHits(nil); err != nil {
		t.Fatalf("SaveHits(nil) = %v", err)
	}
}

func TestSQLiteSaveHitsBatch(t *testing.T) {
	st, err := Open(config.DatabaseCfg{Driver: "sqlite", Path: ":memory:"})
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	defer st.Close()

	hits := []hit.Hit{
		{Host: "site.test", Path: "/p"},
		{Host: "site.test", Path: "/q", Referrer: "", Title: ""},
	}
	if err := st.SaveHits(hits); err != nil {
		t.Fatalf("SaveHits() = %v", err)
	}
	if err := st.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() = %v", err)
	}
}

func TestSQLiteSaveHitsAfterClose(t *testing.T) {
	st, err := Open(config.DatabaseCfg{Driver: "sqlite", Path: ":memory:"})
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	st.Close()
	if err := st.SaveHits([]hit.Hit{{Host: "x", Path: "/"}}); err == nil {
		t.Fatal("expected error saving after close")
	}
}
