package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
)

func TestSQLiteAggregation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kiko.db")
	st, err := Open(config.DatabaseCfg{Driver: "sqlite", Path: path})
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	defer st.Close()

	hour := time.Now().UTC().Truncate(time.Hour).Format(time.RFC3339)
	hits := []hit.Hit{
		{Host: "test.dev", Path: "/blog", Referrer: "https://google.com", VisitorHash: "hash-a"},
		{Host: "test.dev", Path: "/blog", Referrer: "https://google.com", VisitorHash: "hash-b"},
		{Host: "test.dev", Path: "/", VisitorHash: "hash-a"},
	}
	if err := st.SaveHits(hits); err != nil {
		t.Fatalf("SaveHits() = %v", err)
	}

	db, err := sqlOpen("sqlite", path+"?_journal_mode=WAL")
	if err != nil {
		t.Fatalf("sqlOpen() = %v", err)
	}
	defer db.Close()

	var pathCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM kiko_paths WHERE host = 'test.dev'`).Scan(&pathCount); err != nil {
		t.Fatalf("count paths: %v", err)
	}
	if pathCount != 2 {
		t.Errorf("paths = %d, want 2", pathCount)
	}

	var refCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM kiko_refs WHERE host = 'test.dev'`).Scan(&refCount); err != nil {
		t.Fatalf("count refs: %v", err)
	}
	if refCount != 1 {
		t.Errorf("refs = %d, want 1", refCount)
	}

	var blogTotal, blogUniques int
	err = db.QueryRow(`
		SELECT hc.total, hc.uniques
		FROM kiko_hit_counts hc
		JOIN kiko_paths p ON p.id = hc.path_id
		WHERE p.path = '/blog' AND hc.hour = ?`, hour).Scan(&blogTotal, &blogUniques)
	if err != nil {
		t.Fatalf("query blog counts: %v", err)
	}
	if blogTotal != 2 {
		t.Errorf("blog total = %d, want 2", blogTotal)
	}
	if blogUniques != 2 {
		t.Errorf("blog uniques = %d, want 2", blogUniques)
	}

	var refTotal int
	err = db.QueryRow(`
		SELECT rc.total
		FROM kiko_ref_counts rc
		JOIN kiko_refs r ON r.id = rc.ref_id
		WHERE r.referrer = 'https://google.com' AND rc.hour = ?`, hour).Scan(&refTotal)
	if err != nil {
		t.Fatalf("query ref counts: %v", err)
	}
	if refTotal != 2 {
		t.Errorf("ref total = %d, want 2", refTotal)
	}

	if err := st.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() = %v", err)
	}
}
