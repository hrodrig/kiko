package server

import (
	"testing"

	"github.com/hrodrig/kiko/internal/hit"
)

func TestEnrichHit(t *testing.T) {
	h := hit.Hit{
		Host:     "example.com",
		Path:     "/",
		Referrer: "https://www.google.com/search?q=test",
	}
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	enrichHit(&h, ua)

	if h.Browser != "Chrome" || h.OS != "Windows" {
		t.Errorf("browser/os = %q / %q; want Chrome / Windows", h.Browser, h.OS)
	}
	if h.Channel != "organic" {
		t.Errorf("channel = %q; want organic", h.Channel)
	}
	if h.Source != "Google" {
		t.Errorf("source = %q; want Google", h.Source)
	}
	if h.Referrer != "https://www.google.com/search" {
		t.Errorf("referrer = %q", h.Referrer)
	}
}
