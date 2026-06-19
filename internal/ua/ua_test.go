package ua

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		ua      string
		browser string
		os      string
	}{
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Chrome", "Windows",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			"Safari", "macOS",
		},
		{
			"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
			"Firefox", "Linux",
		},
		{
			"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			"Safari", "iOS",
		},
		{
			"Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
			"Chrome", "Android",
		},
		{"", "", ""},
	}
	for _, tc := range tests {
		got := Parse(tc.ua)
		if got.Browser != tc.browser || got.OS != tc.os {
			t.Errorf("Parse(%q) = %+v; want browser=%q os=%q", tc.ua, got, tc.browser, tc.os)
		}
	}
}
