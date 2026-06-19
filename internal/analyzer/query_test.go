package analyzer_test

import (
	"net/url"
	"testing"

	"github.com/hrodrig/kiko/internal/analyzer"
)

func TestParseQuery(t *testing.T) {
	q, err := analyzer.ParseQuery(url.Values{
		"host":  {"x.com"},
		"since": {"2026-01-01"},
		"until": {"2026-01-31"},
		"limit": {"20"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if q.Host != "x.com" || q.Limit != 20 {
		t.Errorf("query = %+v", q)
	}
}

func TestParseQueryValidation(t *testing.T) {
	if _, err := analyzer.ParseQuery(nil); err == nil {
		t.Fatal("expected error without host")
	}
	_, err := analyzer.ParseQuery(url.Values{
		"host":     {"x.com"},
		"interval": {"week"},
	})
	if err == nil {
		t.Fatal("expected interval error")
	}
	_, err = analyzer.ParseQuery(url.Values{
		"host":  {"x.com"},
		"since": {"2026-06-01T00:00:00Z"},
		"until": {"2026-06-02T00:00:00Z"},
	})
	if err != nil {
		t.Fatal(err)
	}
}
