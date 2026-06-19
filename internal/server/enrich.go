package server

import (
	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/ref"
	"github.com/hrodrig/kiko/internal/ua"
)

func enrichHit(h *hit.Hit, userAgent string) {
	ui := ua.Parse(userAgent)
	h.Browser = ui.Browser
	h.OS = ui.OS

	ri := ref.Parse(h.Referrer, h.Host)
	h.Referrer = ri.Referrer
	h.Channel = string(ri.Channel)
	h.Source = ri.Source
}
