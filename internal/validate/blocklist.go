package validate

import (
	"net"
	"net/url"
	"strings"
)

// ReferrerSpam reports known spam referrer hostnames.
func ReferrerSpam(referrer string) bool {
	if strings.TrimSpace(referrer) == "" {
		return false
	}
	u, err := url.Parse(referrer)
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))
	for _, s := range referrerSpamDomains {
		if host == s || strings.HasSuffix(host, "."+s) {
			return true
		}
	}
	return false
}

var referrerSpamDomains = []string{
	"buttons-for-website.com",
	"semalt.com",
	"darodar.com",
	"ilovevitaly.com",
	"priceg.com",
	"blackhatworth.com",
	"best-seo-offer.com",
	"best-seo-solution.com",
	"googlsucks.com",
	"humanorbot.com",
	"simple-share-buttons.com",
	"social-buttons.com",
	"sharebutton.net",
}

// IgnoreIP returns true when ip should be excluded from stats (operator traffic).
func IgnoreIP(ip string, ignore []string) bool {
	if ip == "" || len(ignore) == 0 {
		return false
	}
	return Allowlist(ip, ignore)
}

// DatacenterBlocklist matches cloud/hosting CIDR ranges.
type DatacenterBlocklist struct {
	nets []*net.IPNet
}

// NewDatacenterBlocklist builds a blocklist from embedded defaults plus optional extra CIDR strings.
func NewDatacenterBlocklist(extra []string) (*DatacenterBlocklist, error) {
	cidrs := append([]string(nil), defaultDatacenterCIDRs...)
	cidrs = append(cidrs, extra...)
	var nets []*net.IPNet
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return nil, err
		}
		nets = append(nets, n)
	}
	return &DatacenterBlocklist{nets: nets}, nil
}

// Contains reports whether ip falls in a blocked datacenter range.
func (b *DatacenterBlocklist) Contains(ip string) bool {
	if b == nil || len(b.nets) == 0 {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, n := range b.nets {
		if n.Contains(parsed) {
			return true
		}
	}
	return false
}

// Lightweight sample of well-known cloud/hosting ranges (CE-style filter).
var defaultDatacenterCIDRs = []string{
	"35.192.0.0/12",
	"34.128.0.0/10",
	"52.0.0.0/11",
	"54.0.0.0/10",
	"13.52.0.0/14",
	"20.0.0.0/11",
	"40.64.0.0/10",
	"104.16.0.0/12",
	"172.64.0.0/13",
}
