package ref

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		ref, host string
		wantRef   string
		wantCh    Channel
	}{
		{"", "example.com", "", ChannelDirect},
		{"https://example.com/about", "example.com", "https://example.com/about", ChannelDirect},
		{"https://www.example.com/blog?utm=x", "example.com", "https://www.example.com/blog", ChannelDirect},
		{
			"https://www.google.com/search?q=kiko",
			"example.com",
			"https://www.google.com/search",
			ChannelOrganic,
		},
		{"https://t.co/abc", "example.com", "https://t.co/abc", ChannelSocial},
		{"https://dev.to/post", "example.com", "https://dev.to/post", ChannelSocial},
		{"https://other.site/page", "example.com", "https://other.site/page", ChannelReferral},
		{
			"https://example.com/landing?utm_medium=email&utm_source=news",
			"example.com",
			"https://example.com/landing",
			ChannelEmail,
		},
	}
	for _, tc := range tests {
		got := Parse(tc.ref, tc.host)
		if got.Referrer != tc.wantRef || got.Channel != tc.wantCh {
			t.Errorf("Parse(%q, %q) = %+v; want ref=%q channel=%q",
				tc.ref, tc.host, got, tc.wantRef, tc.wantCh)
		}
	}
}
