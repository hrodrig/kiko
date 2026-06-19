package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/log"
	"github.com/hrodrig/kiko/internal/store"
)

func testServer(allowed []string) *Server {
	buf := hit.NewBuffer(4096)
	st := store.NewNop()
	l := log.New(nil, log.Trace)
	return New(st, buf, l, allowed)
}

func TestServeJS(t *testing.T) {
	s := testServer(nil)
	if s.Handler() == nil {
		t.Fatal("Handler() nil")
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/kiko.js", nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/javascript") {
		t.Errorf("Content-Type = %q; want application/javascript", ct)
	}
	if !strings.Contains(w.Body.String(), "(function") {
		t.Error("kiko.js body missing (function")
	}
}

func TestHealth(t *testing.T) {
	s := testServer(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", ReadyzPath, nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf(`status = %v; want "ok"`, body["status"])
	}
	if _, ok := body["buffer_len"]; !ok {
		t.Error("health missing buffer_len")
	}
}

func TestTrackHit(t *testing.T) {
	s := testServer(nil)
	body := `{"host":"test.dev","path":"/blog","referrer":"https://google.com"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/hit", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "image/gif") {
		t.Errorf("Content-Type = %q; want image/gif", ct)
	}
	if w.Body.Len() != 45 {
		t.Errorf("body size = %d; want 45 (GIF pixel)", w.Body.Len())
	}
}

func TestTrackHit_RejectBot(t *testing.T) {
	s := testServer(nil)
	body := `{"host":"test.dev","path":"/"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/hit", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "Googlebot/2.1")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	// still returns GIF, but hit was discarded silently
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "image/gif") {
		t.Errorf("Content-Type = %q; want image/gif", ct)
	}
}

func TestTrackHit_RejectHost(t *testing.T) {
	s := testServer([]string{"gghstats.com"})
	body := `{"host":"evil.com","path":"/"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/hit", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
}

func TestTrackHit_AllowedHost(t *testing.T) {
	s := testServer([]string{"gghstats.com"})
	body := `{"host":"gghstats.com","path":"/"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/hit", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
}

func TestTrackHit_RejectPrefetch(t *testing.T) {
	s := testServer(nil)
	body := `{"host":"test.dev","path":"/"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/hit", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120")
	r.Header.Set("Purpose", "prefetch")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
}

func TestTrackGIF(t *testing.T) {
	s := testServer(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/hit.gif?p=/test&h=test.dev&r=https://google.com", nil)
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	if w.Body.Len() != 45 {
		t.Errorf("body size = %d; want 45", w.Body.Len())
	}
}

func TestServePixel(t *testing.T) {
	w := httptest.NewRecorder()
	servePixel(w)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	if w.Body.Len() != 45 {
		t.Errorf("body size = %d; want 45", w.Body.Len())
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "image/gif") {
		t.Errorf("Content-Type = %q; want image/gif", ct)
	}
}

func TestShorten(t *testing.T) {
	if got := shorten("hello", 10); got != "hello" {
		t.Errorf("shorten = %q", got)
	}
	if got := shorten("hello world this is long", 5); got != "hello..." {
		t.Errorf("shorten = %q", got)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		fwd  string
		addr string
		want string
	}{
		{"", "192.168.1.1:8080", "192.168.1.1"},
		{"10.0.0.1", "192.168.1.1:8080", "10.0.0.1"},
		{"203.0.113.5, 10.0.0.1", "192.168.1.1:8080", "203.0.113.5"},
		{"", "", ""},
	}
	for _, tt := range tests {
		r := httptest.NewRequest("GET", "/", nil)
		if tt.fwd != "" {
			r.Header.Set("X-Forwarded-For", tt.fwd)
		}
		r.RemoteAddr = tt.addr
		got := clientIP(r)
		if got != tt.want {
			t.Errorf("clientIP(fwd=%q, addr=%q) = %q; want %q", tt.fwd, tt.addr, got, tt.want)
		}
	}
}
