package server

import (
	"net/http"

	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/netutil"
	"github.com/hrodrig/kiko/internal/validate"
)

// FilterConfig controls ingest-side rejection rules.
type FilterConfig struct {
	AllowedHosts       []string
	IgnoreIPs          []string
	TrustProxy         bool
	BlockDatacenterIPs bool
	DatacenterExtra    []string
}

// HitFilter applies allowlist, bot, spam, and IP block rules.
type HitFilter struct {
	cfg        FilterConfig
	datacenter *validate.DatacenterBlocklist
}

// NewHitFilter builds a filter from config. Datacenter blocklist is optional.
func NewHitFilter(cfg FilterConfig) (*HitFilter, error) {
	f := &HitFilter{cfg: cfg}
	if cfg.BlockDatacenterIPs {
		bl, err := validate.NewDatacenterBlocklist(cfg.DatacenterExtra)
		if err != nil {
			return nil, err
		}
		f.datacenter = bl
	}
	return f, nil
}

// RejectReason explains why a hit was dropped.
type RejectReason string

const (
	RejectNone          RejectReason = ""
	RejectNotAllowed    RejectReason = "not_allowed"
	RejectPrefetch      RejectReason = "prefetch"
	RejectBot           RejectReason = "bot"
	RejectReferrerSpam  RejectReason = "referrer_spam"
	RejectIgnoredIP     RejectReason = "ignored_ip"
	RejectDatacenterIP  RejectReason = "datacenter_ip"
	RejectNoClientIP    RejectReason = "no_client_ip"
	RejectHostRateLimit RejectReason = "host_rate_limit"
)

// Check returns a reject reason or empty when the hit should be accepted.
func (f *HitFilter) Check(h hit.Hit, r *http.Request) (RejectReason, string) {
	ip := netutil.ClientIP(r, f.cfg.TrustProxy)
	if f.cfg.TrustProxy && ip == "" && hasProxyHeader(r) {
		return RejectNoClientIP, ip
	}
	if !f.hostAllowed(h.Host) && !validate.Allowlist(ip, f.cfg.AllowedHosts) {
		return RejectNotAllowed, ip
	}
	ua := r.Header.Get("User-Agent")
	if validate.Prefetch(ua, r.Header.Get("Purpose"), r.Header.Get("Sec-Purpose")) {
		return RejectPrefetch, ip
	}
	if validate.IsBot(ua) {
		return RejectBot, ip
	}
	if validate.ReferrerSpam(h.Referrer) {
		return RejectReferrerSpam, ip
	}
	if validate.IgnoreIP(ip, f.cfg.IgnoreIPs) {
		return RejectIgnoredIP, ip
	}
	if f.datacenter != nil && f.datacenter.Contains(ip) {
		return RejectDatacenterIP, ip
	}
	return RejectNone, ip
}

func (f *HitFilter) hostAllowed(host string) bool {
	return validate.Allowlist(host, f.cfg.AllowedHosts)
}

func hasProxyHeader(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Real-IP") != ""
}
