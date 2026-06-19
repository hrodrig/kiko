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
	return Info{Referrer: cleanURL(u), Channel: ch}
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
