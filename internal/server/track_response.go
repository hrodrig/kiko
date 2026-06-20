package server

import (
	"encoding/json"
	"net/http"
)

const (
	headerKikoDropped = "X-Kiko-Dropped"
	headerDebugReq    = "X-Debug-Request"
)

type trackOutcome struct {
	accepted bool
	reason   RejectReason
	clientIP string
}

func debugRequest(r *http.Request) bool {
	return r.Header.Get(headerDebugReq) == "true"
}

func respondTrack(w http.ResponseWriter, r *http.Request, out trackOutcome) {
	if debugRequest(r) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"client_ip": out.clientIP,
			"accepted":  boolStr(out.accepted),
			"reason":    string(out.reason),
		})
		return
	}
	if !out.accepted {
		w.Header().Set(headerKikoDropped, "1")
	}
	servePixel(w)
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
