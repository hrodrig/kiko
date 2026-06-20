package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/log"
	"github.com/hrodrig/kiko/internal/store"
	"github.com/hrodrig/kiko/internal/version"
	"github.com/hrodrig/kiko/internal/visitor"
)

func TestVersionEndpoint(t *testing.T) {
	version.Version = "0.4.1"
	version.Commit = "deadbeef"
	version.BuildDate = "2026-06-20"
	version.Branch = "main"
	s := testServer(nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, VersionPath, nil)
	s.mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	var body version.Info
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Version != "0.4.1" || body.Commit != "deadbeef" || body.BuildDate != "2026-06-20" || body.Branch != "main" {
		t.Errorf("body = %+v", body)
	}
	if body.String() != "kiko 0.4.1 (commit deadbeef, built 2026-06-20, branch main)" {
		t.Errorf("String() = %q", body.String())
	}
}

func TestVersionEndpointNoAuth(t *testing.T) {
	buf := hit.NewBuffer(16)
	s := New(store.NewNop(), buf, log.New(nil, log.Off), visitor.NewHasher("test"), nil,
		WithStats(StatsConfig{APIKey: "secret"}, NewAPIRateLimiter(100, 100)),
		WithIngest(mustFilter(t, nil), nil, false))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, VersionPath, nil)
	s.Handler().ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200 without API key", w.Code)
	}
}
