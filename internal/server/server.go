package server

import (
	_ "embed"
	"encoding/json"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"strings"

	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/log"
	"github.com/hrodrig/kiko/internal/store"
	"github.com/hrodrig/kiko/internal/validate"
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
	store        store.Store
	buf          hit.Buffer
	mux          *http.ServeMux
	log          *log.Logger
	allowedHosts []string
}

func New(s store.Store, buf hit.Buffer, l *log.Logger, allowedHosts []string) *Server {
	sv := &Server{
		store:        s,
		buf:          buf,
		mux:          http.NewServeMux(),
		log:          l,
		allowedHosts: allowedHosts,
	}
	sv.mux.HandleFunc("GET /kiko.js", sv.serveJS)
	sv.mux.HandleFunc("POST /hit", sv.trackHit)
	sv.mux.HandleFunc("GET /hit.gif", sv.trackGIF)
	sv.mux.HandleFunc("GET "+HealthzPath, sv.healthz)
	sv.mux.HandleFunc("GET "+ReadyzPath, sv.readyz)
	sv.mux.HandleFunc("GET /health", sv.readyz) // deprecated: use ReadyzPath
	return sv
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) serveJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, trackingJS)
}

func (s *Server) accept(h hit.Hit, r *http.Request) bool {
	// allowlist: hostname OR IP/CIDR match (empty = accept all)
	if validate.Allowlist(h.Host, s.allowedHosts) {
		goto checkBot
	}
	if validate.Allowlist(clientIP(r), s.allowedHosts) {
		goto checkBot
	}
	s.log.Debug("reject: not allowed: host=%s ip=%s", h.Host, clientIP(r))
	return false

checkBot:
	// bot check
	ua := r.Header.Get("User-Agent")
	if validate.Prefetch(ua, r.Header.Get("Purpose")) {
		s.log.Debug("reject: prefetch from %s", h.Host)
		return false
	}
	if validate.IsBot(ua) {
		s.log.Debug("reject: bot from %s (ua: %s)", h.Host, shorten(ua, 40))
		return false
	}
	return true
}

func (s *Server) trackHit(w http.ResponseWriter, r *http.Request) {
	var h hit.Hit
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		s.log.Debug("hit decode error: %v", err)
		servePixel(w)
		return
	}
	h.Normalize()
	if s.accept(h, r) {
		s.buf.Append(h)
		s.log.Debug("hit: %s %s (ref: %s)", h.Host, h.Path, h.Referrer)
	}
	servePixel(w)
}

func (s *Server) trackGIF(w http.ResponseWriter, r *http.Request) {
	h := hit.Hit{
		Host:     r.URL.Query().Get("h"),
		Path:     r.URL.Query().Get("p"),
		Referrer: r.URL.Query().Get("r"),
		Title:    r.URL.Query().Get("t"),
		Width:    0,
	}
	h.Normalize()
	if s.accept(h, r) {
		s.buf.Append(h)
		s.log.Debug("hit: %s %s (ref: %s)", h.Host, h.Path, h.Referrer)
	}
	servePixel(w)
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

func servePixel(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(pixelGIF)
}

// clientIP extracts the client IP from X-Forwarded-For or RemoteAddr.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if rip := r.RemoteAddr; rip != "" {
		if h, _, err := net.SplitHostPort(rip); err == nil {
			return h
		}
		return rip
	}
	return ""
}

func shorten(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func init() {
	if len(pixelGIF) == 0 {
		stdlog.Fatal("pixel.gif not embedded")
	}
	if !strings.HasPrefix(trackingJS, "(function") {
		stdlog.Fatal("kiko.js looks wrong")
	}
}
