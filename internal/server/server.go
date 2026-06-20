package server

import (
	_ "embed"
	"encoding/json"
	"io"
	stdlog "log"
	"net/http"
	"strings"

	"github.com/hrodrig/kiko/internal/analyzer"
	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/log"
	"github.com/hrodrig/kiko/internal/netutil"
	"github.com/hrodrig/kiko/internal/store"
	"github.com/hrodrig/kiko/internal/visitor"
)

// HealthzPath is the Kubernetes liveness probe (process up; no dependency checks).
const HealthzPath = "/api/v1/healthz"

// ReadyzPath is the Kubernetes readiness probe (DB reachable, buffer healthy).
const ReadyzPath = "/api/v1/readyz"

//go:embed kiko.js
var trackingJS string

//go:embed pixel.gif
var pixelGIF []byte

type Server struct {
	store       store.Store
	buf         hit.Buffer
	mux         *http.ServeMux
	log         *log.Logger
	visitor     visitor.Hasher
	rateLimiter *RateLimiter
	apiLimiter  *APIRateLimiter
	hostLimiter *HostRateLimiter
	filter      *HitFilter
	trustProxy  bool
	stats       StatsConfig
}

func New(s store.Store, buf hit.Buffer, l *log.Logger, v visitor.Hasher, rl *RateLimiter, opts ...ServerOption) *Server {
	sv := &Server{
		store:       s,
		buf:         buf,
		mux:         http.NewServeMux(),
		log:         l,
		visitor:     v,
		rateLimiter: rl,
	}
	for _, o := range opts {
		o(sv)
	}
	sv.mux.HandleFunc("GET /kiko.js", sv.serveJS)
	sv.mux.HandleFunc("POST /hit", sv.trackHit)
	sv.mux.HandleFunc("GET /hit.gif", sv.trackGIF)
	sv.mux.HandleFunc("GET "+HealthzPath, sv.healthz)
	sv.mux.HandleFunc("GET "+ReadyzPath, sv.readyz)
	if acc, ok := s.(store.DBAccessor); ok {
		db, driver := acc.StatsDB()
		registerStats(sv.mux, analyzer.New(db, driver), sv.stats, l)
	}
	return sv
}

// ServerOption configures optional Server behavior.
type ServerOption func(*Server)

// WithStats enables the read-only stats API.
func WithStats(cfg StatsConfig, apiRL *APIRateLimiter) ServerOption {
	return func(s *Server) {
		s.stats = cfg
		s.apiLimiter = apiRL
	}
}

// WithIngest configures hit filtering and per-host rate limits.
func WithIngest(filter *HitFilter, hostRL *HostRateLimiter, trustProxy bool) ServerOption {
	return func(s *Server) {
		s.filter = filter
		s.hostLimiter = hostRL
		s.trustProxy = trustProxy
	}
}

func (s *Server) Handler() http.Handler {
	h := http.Handler(s.mux)
	if s.rateLimiter != nil {
		h = s.rateLimiter.Middleware(h, PublicMiddlewareSkip())
	}
	h = WrapStats(h, s.apiLimiter)
	return h
}

func (s *Server) serveJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, trackingJS)
}

func (s *Server) trackHit(w http.ResponseWriter, r *http.Request) {
	var h hit.Hit
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		s.log.Debug("hit decode error: %v", err)
		respondTrack(w, r, trackOutcome{accepted: false, reason: RejectNotAllowed})
		return
	}
	h.Normalize()
	s.ingest(w, r, h)
}

func (s *Server) trackGIF(w http.ResponseWriter, r *http.Request) {
	h := hit.Hit{
		Host:     r.URL.Query().Get("h"),
		Path:     r.URL.Query().Get("p"),
		Referrer: r.URL.Query().Get("r"),
		Title:    r.URL.Query().Get("t"),
		Width:    queryInt(r.URL.Query().Get("w")),
	}
	h.Normalize()
	s.ingest(w, r, h)
}

func (s *Server) ingest(w http.ResponseWriter, r *http.Request, h hit.Hit) {
	out := trackOutcome{accepted: true}
	if s.filter != nil {
		out.reason, out.clientIP = s.filter.Check(h, r)
		if out.reason != RejectNone {
			out.accepted = false
			s.log.Debug("reject: %s host=%s ip=%s", out.reason, h.Host, out.clientIP)
			respondTrack(w, r, out)
			return
		}
	} else {
		out.clientIP = s.clientIP(r)
	}
	if s.hostLimiter != nil && !s.hostLimiter.Allow(h.Host) {
		out.accepted = false
		out.reason = RejectHostRateLimit
		s.log.Debug("reject: host rate limit %s", h.Host)
		respondTrack(w, r, out)
		return
	}
	h.VisitorHash = s.visitor.Hash(out.clientIP, r.Header.Get("User-Agent"))
	enrichHit(&h, r.Header.Get("User-Agent"))
	s.buf.Append(h)
	s.log.Debug("hit: %s %s (ref: %s)", h.Host, h.Path, h.Referrer)
	respondTrack(w, r, out)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	code := http.StatusOK

	if err := s.store.Ping(r.Context()); err != nil {
		status = "degraded"
		code = http.StatusServiceUnavailable
		s.log.Warn("readyz: database ping failed: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"status":       status,
		"buffer_len":   s.buf.Len(),
		"buffer_drops": s.buf.Drops(),
	})
}

func (s *Server) clientIP(r *http.Request) string {
	return clientIP(r, s.trustProxy)
}

func servePixel(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(pixelGIF)
}

func clientIP(r *http.Request, trustProxy bool) string {
	return netutil.ClientIP(r, trustProxy)
}

func shorten(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func queryInt(s string) int {
	if s == "" {
		return 0
	}
	var n int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func init() {
	if len(pixelGIF) == 0 {
		stdlog.Fatal("pixel.gif not embedded")
	}
	if !strings.HasPrefix(trackingJS, "(function") {
		stdlog.Fatal("kiko.js looks wrong")
	}
}
