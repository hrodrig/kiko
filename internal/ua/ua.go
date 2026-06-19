// Package ua parses User-Agent strings into browser and OS names without regex.
package ua

import "strings"

// Info holds coarse browser and OS labels for analytics (not a full UA database).
type Info struct {
	Browser string
	OS      string
}

// Parse returns browser and OS names from a User-Agent header value.
func Parse(ua string) Info {
	if ua == "" {
		return Info{}
	}
	low := strings.ToLower(ua)
	return Info{
		Browser: browser(low),
		OS:      osName(low),
	}
}

func browser(low string) string {
	switch {
	case strings.Contains(low, "edg/") || strings.Contains(low, "edge/"):
		return "Edge"
	case strings.Contains(low, "opr/") || strings.Contains(low, "opera"):
		return "Opera"
	case strings.Contains(low, "chrome/") || strings.Contains(low, "crios/"):
		return "Chrome"
	case strings.Contains(low, "firefox/") || strings.Contains(low, "fxios/"):
		return "Firefox"
	case strings.Contains(low, "safari/") && !strings.Contains(low, "chrome"):
		return "Safari"
	case strings.Contains(low, "msie") || strings.Contains(low, "trident/"):
		return "IE"
	default:
		return "Other"
	}
}

func osName(low string) string {
	switch {
	case strings.Contains(low, "iphone") || strings.Contains(low, "ipad"):
		return "iOS"
	case strings.Contains(low, "android"):
		return "Android"
	case strings.Contains(low, "windows nt"):
		return "Windows"
	case strings.Contains(low, "mac os x") || strings.Contains(low, "macintosh"):
		return "macOS"
	case strings.Contains(low, "cros"):
		return "Chrome OS"
	case strings.Contains(low, "linux"):
		return "Linux"
	default:
		return "Other"
	}
}
