package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/log"
	"github.com/hrodrig/kiko/internal/visitor"
)

type pingFailStore struct{}

func (p *pingFailStore) SaveHits([]hit.Hit) error { return nil }
func (p *pingFailStore) Ping(context.Context) error {
	return context.DeadlineExceeded
}
func (p *pingFailStore) Close() error { return nil }

func TestHealthz(t *testing.T) {
	s := testServer(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", HealthzPath, nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf(`status = %q; want "ok"`, body["status"])
	}
}

func TestHealthzIgnoresDBFailure(t *testing.T) {
	buf := hit.NewBuffer(16)
	s := New(&pingFailStore{}, buf, log.New(nil, log.Off), nil, visitor.NewHasher("test"), nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", HealthzPath, nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz status = %d; want 200 even when DB down", w.Code)
	}
}

func TestReadyzDegraded(t *testing.T) {
	buf := hit.NewBuffer(16)
	s := New(&pingFailStore{}, buf, log.New(nil, log.Off), nil, visitor.NewHasher("test"), nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", ReadyzPath, nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d; want 503", w.Code)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "degraded" {
		t.Errorf("status = %v; want degraded", body["status"])
	}
}
