package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hrodrig/kiko/internal/hit"
)

func TestFilterReferrerSpam(t *testing.T) {
	f, err := NewHitFilter(FilterConfig{})
	if err != nil {
		t.Fatal(err)
	}
	reason, _ := f.Check(hit.Hit{Host: "example.com", Path: "/", Referrer: "https://semalt.com/x"}, uaRequest())
	if reason != RejectReferrerSpam {
		t.Fatalf("reason = %q", reason)
	}
}

func TestFilterIgnoreIP(t *testing.T) {
	f, err := NewHitFilter(FilterConfig{IgnoreIPs: []string{"203.0.113.0/24"}})
	if err != nil {
		t.Fatal(err)
	}
	r := uaRequest()
	r.RemoteAddr = "203.0.113.5:8080"
	reason, _ := f.Check(hit.Hit{Host: "example.com", Path: "/"}, r)
	if reason != RejectIgnoredIP {
		t.Fatalf("reason = %q", reason)
	}
}

func TestFilterDatacenter(t *testing.T) {
	f, err := NewHitFilter(FilterConfig{BlockDatacenterIPs: true})
	if err != nil {
		t.Fatal(err)
	}
	r := uaRequest()
	r.RemoteAddr = "35.200.1.1:8080"
	reason, _ := f.Check(hit.Hit{Host: "example.com", Path: "/"}, r)
	if reason != RejectDatacenterIP {
		t.Fatalf("reason = %q", reason)
	}
}

func TestFilterNoClientIPBehindProxy(t *testing.T) {
	f, err := NewHitFilter(FilterConfig{TrustProxy: true})
	if err != nil {
		t.Fatal(err)
	}
	r := uaRequest()
	r.Header.Set("X-Forwarded-For", "10.0.0.1")
	reason, ip := f.Check(hit.Hit{Host: "example.com", Path: "/"}, r)
	if reason != RejectNoClientIP || ip != "" {
		t.Fatalf("reason=%q ip=%q", reason, ip)
	}
}

func TestDebugRequestJSON(t *testing.T) {
	s := testServer(nil)
	body := `{"host":"test.dev","path":"/"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120")
	r.Header.Set(headerDebugReq, "true")
	r.RemoteAddr = "203.0.113.1:8080"
	s.mux.ServeHTTP(w, r)
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("content-type = %q", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), "203.0.113.1") {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func uaRequest() *http.Request {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120")
	return r
}
