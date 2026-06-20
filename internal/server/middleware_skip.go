package server

import "strings"

// MiddlewareSkip lists request paths exempt from rate limiting.
type MiddlewareSkip struct {
	Exact    []string
	Prefixes []string
}

// PublicMiddlewareSkip returns paths that must stay reachable without rate limits
// (probes, static tracking script).
func PublicMiddlewareSkip() MiddlewareSkip {
	return MiddlewareSkip{
		Exact: []string{
			HealthzPath,
			ReadyzPath,
			VersionPath,
			"/kiko.js",
		},
	}
}

func (s MiddlewareSkip) matches(path string) bool {
	for _, p := range s.Exact {
		if path == p {
			return true
		}
	}
	for _, p := range s.Prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
