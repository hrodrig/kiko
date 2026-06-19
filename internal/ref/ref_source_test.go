package ref

import "testing"

func TestSourceLabelGoogle(t *testing.T) {
	info := Parse("https://www.google.com/search?q=kiko", "example.com")
	if info.Source != "Google" || info.Channel != ChannelOrganic {
		t.Errorf("Parse google = %+v", info)
	}
}

func TestSourceLabelTwitter(t *testing.T) {
	info := Parse("https://t.co/abc", "example.com")
	if info.Source != "Twitter/X" || info.Channel != ChannelSocial {
		t.Errorf("Parse twitter = %+v", info)
	}
}
