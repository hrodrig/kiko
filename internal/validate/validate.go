package validate

import (
	"net"
	"strings"
)

// Allowlist checks if host is in the allowed list.
// Each entry can be:
//   - a hostname: "gghstats.com"
//   - an IP: "192.168.1.1"
//   - a CIDR range: "10.0.0.0/8"
//
// Empty allowlist = accept all.
func Allowlist(host string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	// strip port if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	for _, a := range allowed {
		if a == host {
			return true
		}
		// CIDR match
		if strings.Contains(a, "/") {
			_, cidr, err := net.ParseCIDR(a)
			if err != nil {
				continue
			}
			ip := net.ParseIP(host)
			if ip != nil && cidr.Contains(ip) {
				return true
			}
		}
	}
	return false
}

// IsBot returns true if the request looks like a bot.
func IsBot(ua string) bool {
	if ua == "" {
		return true
	}
	if len(ua) < 10 {
		return true
	}
	low := strings.ToLower(ua)
	bots := []string{
		"bot", "crawler", "spider", "scraper",
		"curl", "wget", "go-http-client",
		"python-requests", "python-urllib",
		"mastodon", "twitterbot", "slack",
		"whatsapp", "telegrambot", "discordbot",
		"semrush", "ahrefsbot", "rogerbot",
		"googlebot", "bingbot", "yandex",
		"baiduspider", "duckduckbot",
		"headless", "chrome-lighthouse",
	}
	for _, b := range bots {
		if strings.Contains(low, b) {
			return true
		}
	}
	return false
}

// Prefetch returns true if the request is a prefetch/prerender.
func Prefetch(ua, purpose, secPurpose string) bool {
	if purpose == "prefetch" || secPurpose == "prefetch" {
		return true
	}
	if strings.Contains(strings.ToLower(ua), "prefetch") {
		return true
	}
	return false
}
