package server

import (
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// APIRateLimiter limits stats API requests per API key (or client IP when no key).
type APIRateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
}

// NewAPIRateLimiter creates a per-key token bucket for the stats API.
func NewAPIRateLimiter(requestsPerSec, burst int) *APIRateLimiter {
	r := rate.Limit(requestsPerSec)
	if r <= 0 {
		r = rate.Inf
	}
	if burst <= 0 {
		burst = 1
	}
	return &APIRateLimiter{rate: r, burst: burst}
}

// Middleware wraps stats routes with per-key rate limiting.
func (rl *APIRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := extractAPIKey(r)
		if key == "" {
			key = "ip:" + clientIP(r)
		}
		lim := rl.get(key)
		if !lim.Allow() {
			w.Header().Set("Retry-After", "60")
			writeJSONError(w, http.StatusTooManyRequests, "rate_limit_exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *APIRateLimiter) get(key string) *rate.Limiter {
	if v, ok := rl.limiters.Load(key); ok {
		return v.(*rate.Limiter)
	}
	lim := rate.NewLimiter(rl.rate, rl.burst)
	v, _ := rl.limiters.LoadOrStore(key, lim)
	return v.(*rate.Limiter)
}

// StatsMiddlewareSkip returns paths exempt from ingest rate limits but not stats auth.
func StatsMiddlewareSkip() MiddlewareSkip {
	skip := PublicMiddlewareSkip()
	skip.Prefixes = append(skip.Prefixes, "/api/v1/stats/")
	return skip
}

// WrapStats applies API rate limiting to /api/v1/stats/* only.
func WrapStats(base http.Handler, rl *APIRateLimiter) http.Handler {
	if rl == nil {
		return base
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/stats/") {
			rl.Middleware(base).ServeHTTP(w, r)
			return
		}
		base.ServeHTTP(w, r)
	})
}
