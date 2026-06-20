package server

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter is a per-IP token-bucket rate limiter.
type RateLimiter struct {
	limiters   sync.Map // string → *rateLimiterEntry
	rate       rate.Limit
	burst      int
	maxIdle    time.Duration
	done       chan struct{}
	trustProxy bool
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitConfig holds rate limiter parameters (requests per second).
type RateLimitConfig struct {
	RequestsPerSec int
	Burst          int
	TrustProxy     bool
}

// NewRateLimiter creates a per-IP token bucket rate limiter. Call Shutdown on cleanup.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	r := rate.Limit(cfg.RequestsPerSec)
	if r <= 0 {
		r = rate.Inf
	}
	burst := cfg.Burst
	if burst <= 0 {
		burst = 1
	}
	rl := &RateLimiter{
		rate:       r,
		burst:      burst,
		maxIdle:    5 * time.Minute,
		done:       make(chan struct{}),
		trustProxy: cfg.TrustProxy,
	}
	go rl.cleanupLoop()
	return rl
}

// Shutdown stops the background cleanup goroutine.
func (rl *RateLimiter) Shutdown() {
	select {
	case <-rl.done:
	default:
		close(rl.done)
	}
}

// Middleware wraps next with per-IP rate limiting. Paths in skip are exempt.
func (rl *RateLimiter) Middleware(next http.Handler, skip MiddlewareSkip) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skip.matches(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		ip := clientIP(r, rl.trustProxy)
		entry := rl.getOrCreate(ip)
		if !entry.limiter.Allow() {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate_limit_exceeded"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) getOrCreate(ip string) *rateLimiterEntry {
	if v, ok := rl.limiters.Load(ip); ok {
		e := v.(*rateLimiterEntry)
		e.lastSeen = time.Now()
		return e
	}
	entry := &rateLimiterEntry{
		limiter:  rate.NewLimiter(rl.rate, rl.burst),
		lastSeen: time.Now(),
	}
	v, _ := rl.limiters.LoadOrStore(ip, entry)
	return v.(*rateLimiterEntry)
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-rl.done:
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

func (rl *RateLimiter) cleanup() {
	now := time.Now()
	rl.limiters.Range(func(key, value any) bool {
		e := value.(*rateLimiterEntry)
		if now.Sub(e.lastSeen) > rl.maxIdle {
			rl.limiters.Delete(key)
		}
		return true
	})
}
