package analyzer_test

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

func TestAnalyzerSQLiteIntegration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kiko.db")
	st, err := store.Open(config.DatabaseCfg{Driver: "sqlite", Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	acc := st.(store.DBAccessor)
	db, driver := acc.StatsDB()
	az := analyzer.New(db, driver)

	hits := []hit.Hit{
		{Host: "site.test", Path: "/x?utm_source=a", VisitorHash: "1", Browser: "Chrome", OS: "Linux", Channel: "direct"},
		{Host: "site.test", Path: "/y", VisitorHash: "2", Browser: "Safari", OS: "iOS", Channel: "social", Source: "Twitter/X", Referrer: "https://t.co/x"},
	}
	for i := range hits {
		hits[i].Normalize()
	}
	if err := st.SaveHits(hits); err != nil {
		t.Fatal(err)
	}

	q := analyzer.Query{
		Host:     "site.test",
		Since:    time.Now().UTC().Add(-time.Hour),
		Until:    time.Now().UTC().Add(time.Minute),
		Limit:    10,
		Interval: "day",
	}
	ctx := context.Background()

	run := []struct {
		name string
		fn   func() error
	}{
		{"Summary", func() error { _, err := az.Summary(ctx, q); return err }},
		{"Paths", func() error { _, err := az.Paths(ctx, q); return err }},
		{"Refs", func() error { _, err := az.Refs(ctx, q); return err }},
		{"Timeline", func() error { _, err := az.Timeline(ctx, q); return err }},
		{"Visitors", func() error { _, err := az.Visitors(ctx, q); return err }},
		{"Channels", func() error { _, err := az.Channels(ctx, q); return err }},
		{"Browsers", func() error { _, err := az.Browsers(ctx, q); return err }},
		{"OS", func() error { _, err := az.OS(ctx, q); return err }},
		{"UTM", func() error { _, err := az.UTMSources(ctx, q); return err }},
	}
	for _, tc := range run {
		if err := tc.fn(); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
	}

	q.Interval = "hour"
	if _, err := az.Timeline(ctx, q); err != nil {
		t.Fatal(err)
	}
}
