package validate

import "testing"

func TestAllowlistHostname(t *testing.T) {
	allowed := []string{"gghstats.com", "kzero.dev"}
	cases := []struct {
		host string
		want bool
	}{
		{"gghstats.com", true},
		{"kzero.dev", true},
		{"evil.com", false},
		{"gghstats.com.evil.com", false},
	}
	for _, c := range cases {
		got := Allowlist(c.host, allowed)
		if got != c.want {
			t.Errorf("Allowlist(%q, %v) = %v; want %v", c.host, allowed, got, c.want)
		}
	}
}

func TestAllowlistEmpty(t *testing.T) {
	if !Allowlist("anything", nil) {
		t.Error("Allowlist with nil should accept all")
	}
	if !Allowlist("anything", []string{}) {
		t.Error("Allowlist with empty should accept all")
	}
}

func TestAllowlistIP(t *testing.T) {
	allowed := []string{"127.0.0.1", "10.0.0.0/8"}
	cases := []struct {
		host string
		want bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"192.168.1.1", false},
	}
	for _, c := range cases {
		got := Allowlist(c.host, allowed)
		if got != c.want {
			t.Errorf("Allowlist(%q, %v) = %v; want %v", c.host, allowed, got, c.want)
		}
	}
}

func TestAllowlistWithPort(t *testing.T) {
	allowed := []string{"127.0.0.1", "10.0.0.0/8"}
	cases := []struct {
		host string
		want bool
	}{
		{"127.0.0.1:8080", true},
		{"10.0.0.1:3000", true},
		{"192.168.1.1:9000", false},
	}
	for _, c := range cases {
		got := Allowlist(c.host, allowed)
		if got != c.want {
			t.Errorf("Allowlist(%q, %v) = %v; want %v", c.host, allowed, got, c.want)
		}
	}
}

func TestIsBot(t *testing.T) {
	cases := []struct {
		ua   string
		want bool
	}{
		{"", true},
		{"sh", true}, // too short
		{"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", true},
		{"curl/8.0", true},
		{"python-requests/2.31", true},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36", false},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) Mobile/15E148", false},
	}
	for _, c := range cases {
		got := IsBot(c.ua)
		if got != c.want {
			t.Errorf("IsBot(%q) = %v; want %v", c.ua, got, c.want)
		}
	}
}

func TestPrefetch(t *testing.T) {
	cases := []struct {
		ua      string
		purpose string
		want    bool
	}{
		{"", "prefetch", true},
		{"Mozilla/5.0 ...", "", false},
		{"Mozilla/5.0 (compatible; Prefetch)", "", true},
		{"Mozilla/5.0 ...", "other", false},
	}
	for _, c := range cases {
		got := Prefetch(c.ua, c.purpose)
		if got != c.want {
			t.Errorf("Prefetch(%q, %q) = %v; want %v", c.ua, c.purpose, got, c.want)
		}
	}
}
