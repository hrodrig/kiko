package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/hrodrig/kiko/internal/analyzer"
)

const statsCacheMaxAge = 60

// StatsConfig configures the read-only stats API.
type StatsConfig struct {
	APIKey string
}

type statsHandler struct {
	az  *analyzer.Analyzer
	cfg StatsConfig
	log debugLogger
}

type debugLogger interface {
	Debug(format string, args ...any)
	Warn(format string, args ...any)
}

func registerStats(mux *http.ServeMux, az *analyzer.Analyzer, cfg StatsConfig, log debugLogger) {
	if az == nil {
		return
	}
	h := &statsHandler{az: az, cfg: cfg, log: log}
	mux.HandleFunc("GET /api/v1/stats/summary", h.summary)
	mux.HandleFunc("GET /api/v1/stats/paths", h.paths)
	mux.HandleFunc("GET /api/v1/stats/refs", h.refs)
	mux.HandleFunc("GET /api/v1/stats/timeline", h.timeline)
	mux.HandleFunc("GET /api/v1/stats/visitors", h.visitors)
	mux.HandleFunc("GET /api/v1/stats/channels", h.channels)
	mux.HandleFunc("GET /api/v1/stats/browsers", h.browsers)
	mux.HandleFunc("GET /api/v1/stats/os", h.os)
	mux.HandleFunc("GET /api/v1/stats/utm", h.utm)
}

func (h *statsHandler) summary(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, func(ctx context.Context, q analyzer.Query) (any, error) {
		return h.az.Summary(ctx, q)
	})
}

func (h *statsHandler) paths(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, func(ctx context.Context, q analyzer.Query) (any, error) {
		return h.az.Paths(ctx, q)
	})
}

func (h *statsHandler) refs(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, func(ctx context.Context, q analyzer.Query) (any, error) {
		return h.az.Refs(ctx, q)
	})
}

func (h *statsHandler) timeline(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, func(ctx context.Context, q analyzer.Query) (any, error) {
		return h.az.Timeline(ctx, q)
	})
}

func (h *statsHandler) visitors(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, func(ctx context.Context, q analyzer.Query) (any, error) {
		return h.az.Visitors(ctx, q)
	})
}

func (h *statsHandler) channels(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, func(ctx context.Context, q analyzer.Query) (any, error) {
		return h.az.Channels(ctx, q)
	})
}

func (h *statsHandler) browsers(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, func(ctx context.Context, q analyzer.Query) (any, error) {
		return h.az.Browsers(ctx, q)
	})
}

func (h *statsHandler) os(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, func(ctx context.Context, q analyzer.Query) (any, error) {
		return h.az.OS(ctx, q)
	})
}

func (h *statsHandler) utm(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, func(ctx context.Context, q analyzer.Query) (any, error) {
		return h.az.UTMSources(ctx, q)
	})
}

type statsFn func(context.Context, analyzer.Query) (any, error)

func (h *statsHandler) serve(w http.ResponseWriter, r *http.Request, fn statsFn) {
	if !h.authorize(w, r) {
		return
	}
	q, err := analyzer.ParseQuery(r.URL.Query())
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	out, err := fn(ctx, q)
	if err != nil {
		h.log.Debug("stats error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "query failed")
		return
	}
	writeJSONStats(w, out)
}

func (h *statsHandler) authorize(w http.ResponseWriter, r *http.Request) bool {
	key := h.cfg.APIKey
	if key == "" {
		return true
	}
	got := extractAPIKey(r)
	if got == key {
		return true
	}
	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
	return false
}

func extractAPIKey(r *http.Request) string {
	if k := r.Header.Get("X-API-Key"); k != "" {
		return k
	}
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	return r.URL.Query().Get("key")
}

func writeJSONStats(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age="+itoa(statsCacheMaxAge))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
