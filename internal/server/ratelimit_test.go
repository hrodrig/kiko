package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiterBlocksAfterBurst(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{RequestsPerSec: 1, Burst: 2})
	defer rl.Shutdown()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), MiddlewareSkip{})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/hit", nil)
		req.RemoteAddr = "10.0.0.2:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("burst request %d: got %d, want 200", i, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/hit", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("got %d, want 429", rec.Code)
	}
}

func TestRateLimiterSkipsHealthAndScript(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{RequestsPerSec: 1, Burst: 1})
	defer rl.Shutdown()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), PublicMiddlewareSkip())

	req := httptest.NewRequest(http.MethodGet, "/hit", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatal("expected first /hit to pass")
	}
	req = httptest.NewRequest(http.MethodGet, "/hit", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected /hit rate limited, got %d", rec.Code)
	}

	for _, path := range []string{HealthzPath, ReadyzPath, "/health", "/kiko.js"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("exempt %s: got %d, want 200", path, rec.Code)
		}
	}
}
