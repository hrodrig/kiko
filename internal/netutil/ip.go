package netutil

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP returns the client IP for visitor hashing and filtering.
// When trustProxy is true, X-Forwarded-For / X-Real-IP are scanned for the first
// global-unicast address; private and loopback hops are skipped.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if ip := firstPublicFromHeader(xff); ip != "" {
				return ip
			}
			return ""
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			if ip := firstPublicFromHeader(xri); ip != "" {
				return ip
			}
			return ""
		}
	} else if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := strings.TrimSpace(strings.Split(xff, ",")[0]); ip != "" {
			return stripZone(ip)
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return stripZone(xri)
	}
	return remoteHost(r.RemoteAddr)
}

func firstPublicFromHeader(raw string) string {
	for _, part := range strings.Split(raw, ",") {
		ip := net.ParseIP(stripZone(strings.TrimSpace(part)))
		if ip == nil || IsPrivateOrLoopback(ip) {
			continue
		}
		return ip.String()
	}
	return ""
}

// IsPrivateOrLoopback reports RFC1918, loopback, link-local, and ULA addresses.
func IsPrivateOrLoopback(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10,
			ip4[0] == 127,
			ip4[0] == 0:
			return true
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return true
		case ip4[0] == 192 && ip4[1] == 168:
			return true
		case ip4[0] == 169 && ip4[1] == 254:
			return true
		}
		return false
	}
	return ip.IsPrivate()
}

func remoteHost(addr string) string {
	if addr == "" {
		return ""
	}
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

func stripZone(s string) string {
	if i := strings.IndexByte(s, '%'); i >= 0 {
		return s[:i]
	}
	return s
}
