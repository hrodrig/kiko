package server

import (
	"sync"

	"golang.org/x/time/rate"
)

// HostRateLimiter applies a per-site-host token bucket on ingest.
type HostRateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
}

// NewHostRateLimiter creates a per-host limiter. requestsPerSec <= 0 disables limits.
func NewHostRateLimiter(requestsPerSec, burst int) *HostRateLimiter {
	if requestsPerSec <= 0 {
		return nil
	}
	if burst <= 0 {
		burst = 1
	}
	return &HostRateLimiter{
		rate:  rate.Limit(requestsPerSec),
		burst: burst,
	}
}

// Allow reports whether another hit for host may proceed.
func (h *HostRateLimiter) Allow(host string) bool {
	if h == nil || host == "" {
		return true
	}
	v, _ := h.limiters.LoadOrStore(host, rate.NewLimiter(h.rate, h.burst))
	return v.(*rate.Limiter).Allow()
}
