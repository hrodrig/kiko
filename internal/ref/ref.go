// Package ref normalizes referrer URLs and assigns a coarse traffic channel.
package ref

import (
	"net/url"
	"strings"
)

// Channel is a coarse acquisition label for analytics rollups.
type Channel string

const (
	ChannelDirect   Channel = "direct"
	ChannelOrganic  Channel = "organic"
	ChannelSocial   Channel = "social"
	ChannelEmail    Channel = "email"
	ChannelReferral Channel = "referral"
)

// Info is the parsed referrer used when storing hits.
type Info struct {
	Referrer string
	Channel  Channel
	Source   string // display label, e.g. Google, Facebook
}

// Parse normalizes referrer for the given site host and classifies the channel.
func Parse(referrer, siteHost string) Info {
	if strings.TrimSpace(referrer) == "" {
		return Info{Channel: ChannelDirect}
	}

	raw := strings.TrimSpace(referrer)
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return Info{Referrer: raw, Channel: ChannelReferral}
	}

	if channelFromQuery(u.Query()) == ChannelEmail {
		return Info{Referrer: cleanURL(u), Channel: ChannelEmail}
	}

	refHost := hostKey(u.Hostname())
	site := hostKey(siteHost)
	if refHost == site || refHost == "" {
		return Info{Referrer: cleanURL(u), Channel: ChannelDirect}
	}

	ch := classifyHost(refHost)
	return Info{Referrer: cleanURL(u), Channel: ch, Source: sourceLabel(refHost, ch)}
}

func sourceLabel(host string, ch Channel) string {
	if ch == ChannelDirect {
		return ""
	}
	if ch == ChannelOrganic {
		return searchSourceName(host)
	}
	if ch == ChannelSocial {
		return socialSourceName(host)
	}
	if ch == ChannelEmail {
		return "Email"
	}
	return host
}

func searchSourceName(host string) string {
	switch {
	case strings.Contains(host, "google."):
		return "Google"
	case strings.Contains(host, "bing.com"):
		return "Bing"
	case strings.Contains(host, "duckduckgo.com"):
		return "DuckDuckGo"
	case strings.Contains(host, "yahoo.com"):
		return "Yahoo"
	case strings.Contains(host, "yandex."):
		return "Yandex"
	case strings.Contains(host, "baidu.com"):
		return "Baidu"
	case strings.Contains(host, "ecosia.org"):
		return "Ecosia"
	case strings.Contains(host, "kagi.com"):
		return "Kagi"
	default:
		return host
	}
}

func socialSourceName(host string) string {
	switch {
	case strings.Contains(host, "twitter.com"), host == "t.co", strings.Contains(host, "x.com"):
		return "Twitter/X"
	case strings.Contains(host, "facebook.com"), strings.Contains(host, "fb.com"):
		return "Facebook"
	case strings.Contains(host, "instagram.com"):
		return "Instagram"
	case strings.Contains(host, "linkedin.com"):
		return "LinkedIn"
	case strings.Contains(host, "reddit.com"):
		return "Reddit"
	case strings.Contains(host, "youtube.com"):
		return "YouTube"
	case strings.Contains(host, "tiktok.com"):
		return "TikTok"
	case strings.Contains(host, "mastodon."):
		return "Mastodon"
	case strings.Contains(host, "bsky.app"):
		return "Bluesky"
	case strings.Contains(host, "news.ycombinator.com"), strings.Contains(host, "hn.algolia.com"):
		return "Hacker News"
	default:
		return host
	}
}

func channelFromQuery(q url.Values) Channel {
	medium := strings.ToLower(q.Get("utm_medium"))
	if medium == "email" || medium == "e-mail" {
		return ChannelEmail
	}
	return ""
}

func classifyHost(host string) Channel {
	if isSearchEngine(host) {
		return ChannelOrganic
	}
	if isSocial(host) {
		return ChannelSocial
	}
	return ChannelReferral
}

func isSearchEngine(host string) bool {
	for _, s := range []string{
		"google.", "bing.com", "duckduckgo.com", "yahoo.com", "yandex.",
		"baidu.com", "ecosia.org", "kagi.com", "startpage.com", "qwant.com",
	} {
		if strings.Contains(host, s) {
			return true
		}
	}
	return false
}

func isSocial(host string) bool {
	for _, s := range []string{
		"twitter.com", "t.co", "x.com", "facebook.com", "fb.com", "instagram.com",
		"linkedin.com", "reddit.com", "threads.net", "mastodon.", "bsky.app",
		"dev.to", "news.ycombinator.com", "hn.algolia.com", "youtube.com",
		"tiktok.com", "pinterest.com", "whatsapp.com", "telegram.",
	} {
		if strings.Contains(host, s) {
			return true
		}
	}
	return false
}

func hostKey(host string) string {
	h := strings.ToLower(strings.TrimSpace(host))
	h = strings.TrimPrefix(h, "www.")
	if i := strings.IndexByte(h, ':'); i >= 0 {
		h = h[:i]
	}
	return h
}

func cleanURL(u *url.URL) string {
	u.Fragment = ""
	u.RawQuery = ""
	u.User = nil
	out := u.String()
	if strings.HasSuffix(out, "/") && u.Path == "" {
		out = strings.TrimSuffix(out, "/")
	}
	return out
}
