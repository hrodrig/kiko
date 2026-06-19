package store_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/hrodrig/kiko/internal/analyzer"
	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/store"
)

func TestAnalyzerSummary(t *testing.T) {
	az, st := openAnalyzer(t)
	defer st.Close()
	seedAnalyzerHits(t, st)

	q := analyzerQuery()
	sum, err := az.Summary(context.Background(), q)
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if sum.Hits != 3 || sum.Uniques != 2 {
		t.Errorf("summary hits/uniques = %d/%d, want 3/2", sum.Hits, sum.Uniques)
	}
}

func TestAnalyzerBreakdowns(t *testing.T) {
	az, st := openAnalyzer(t)
	defer st.Close()
	seedAnalyzerHits(t, st)
	q := analyzerQuery()
	ctx := context.Background()

	if _, err := az.UTMSources(ctx, q); err != nil {
		t.Fatalf("UTMSources: %v", err)
	}
	for _, fn := range []func(context.Context, analyzer.Query) error{
		func(c context.Context, q analyzer.Query) error { _, err := az.Paths(c, q); return err },
		func(c context.Context, q analyzer.Query) error { _, err := az.Refs(c, q); return err },
		func(c context.Context, q analyzer.Query) error { _, err := az.Timeline(c, q); return err },
		func(c context.Context, q analyzer.Query) error { _, err := az.Channels(c, q); return err },
		func(c context.Context, q analyzer.Query) error { _, err := az.Browsers(c, q); return err },
		func(c context.Context, q analyzer.Query) error { _, err := az.OS(c, q); return err },
	} {
		if err := fn(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	q.Interval = "hour"
	if _, err := az.Timeline(ctx, q); err != nil {
		t.Fatal(err)
	}
}

func openAnalyzer(t *testing.T) (*analyzer.Analyzer, store.Store) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kiko.db")
	st, err := store.Open(config.DatabaseCfg{Driver: "sqlite", Path: path})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	acc := st.(store.DBAccessor)
	db, driver := acc.StatsDB()
	return analyzer.New(db, driver), st
}

func seedAnalyzerHits(t *testing.T, st store.Store) {
	t.Helper()
	hits := []hit.Hit{
		{Host: "example.com", Path: "/a?utm_source=newsletter&utm_medium=email", VisitorHash: "u1", Browser: "Chrome", OS: "Linux"},
		{Host: "example.com", Path: "/a", VisitorHash: "u2", Browser: "Firefox", OS: "macOS", Channel: "organic", Source: "Google"},
		{Host: "example.com", Path: "/b", VisitorHash: "u1", Referrer: "https://google.com", Channel: "organic", Source: "Google"},
	}
	for i := range hits {
		hits[i].Normalize()
	}
	if err := st.SaveHits(hits); err != nil {
		t.Fatalf("SaveHits: %v", err)
	}
}

func analyzerQuery() analyzer.Query {
	return analyzer.Query{
		Host:     "example.com",
		Since:    time.Now().UTC().Add(-time.Hour),
		Until:    time.Now().UTC().Add(time.Minute),
		Limit:    5,
		Interval: "day",
	}
}
