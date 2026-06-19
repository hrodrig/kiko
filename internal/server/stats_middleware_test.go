package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIRateLimiter(t *testing.T) {
	rl := NewAPIRateLimiter(1, 1)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	h := rl.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/summary", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("first = %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second = %d", rec2.Code)
	}
}

func TestParseQueryErrors(t *testing.T) {
	_, handler := testStatsHandler(t, "")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/summary", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}
