package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/log"
	"github.com/hrodrig/kiko/internal/store"
	"github.com/hrodrig/kiko/internal/visitor"
)

func TestStatsSummary(t *testing.T) {
	st, handler := testStatsHandler(t, "secret")
	defer st.Close()

	hits := []hit.Hit{
		{Host: "gghstats.com", Path: "/?utm_source=launch", VisitorHash: "a"},
		{Host: "gghstats.com", Path: "/", VisitorHash: "b"},
	}
	for i := range hits {
		hits[i].Normalize()
	}
	if err := st.SaveHits(hits); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/summary?host=gghstats.com&since=2020-01-01", nil)
	req.Header.Set("X-API-Key", "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["hits"].(float64) != 2 {
		t.Errorf("hits = %v", out["hits"])
	}
	if cc := rec.Header().Get("Cache-Control"); cc == "" {
		t.Error("expected cache header")
	}
}

func TestStatsPathsOpenKey(t *testing.T) {
	st, handler := testStatsHandler(t, "")
	defer st.Close()
	if err := st.SaveHits([]hit.Hit{{Host: "x.com", Path: "/", VisitorHash: "a"}}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/paths?host=x.com&since=2020-01-01", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestStatsUnauthorized(t *testing.T) {
	_, handler := testStatsHandler(t, "secret")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/summary?host=x.com", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func testStatsHandler(t *testing.T, apiKey string) (store.Store, http.Handler) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kiko.db")
	st, err := store.Open(config.DatabaseCfg{Driver: "sqlite", Path: path})
	if err != nil {
		t.Fatal(err)
	}
	buf := hit.NewBuffer(100)
	l := log.New(nil, log.Off)
	sv := New(st, buf, l, nil, visitor.NewHasher("test"), nil,
		WithStats(StatsConfig{APIKey: apiKey}, NewAPIRateLimiter(100, 100)))
	return st, sv.Handler()
}

func TestExtractAPIKey(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer tok")
	if extractAPIKey(r) != "tok" {
		t.Fatal("bearer")
	}
	r = httptest.NewRequest(http.MethodGet, "/?key=q", nil)
	if extractAPIKey(r) != "q" {
		t.Fatal("query")
	}
}
