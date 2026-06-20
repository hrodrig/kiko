package netutil

import (
	"net"
	"net/http"
	"testing"
)

func TestClientIPTrustProxy(t *testing.T) {
	r := httptestNew("10.0.0.5, 203.0.113.10, 198.51.100.2", "")
	got := ClientIP(r, true)
	if got != "203.0.113.10" {
		t.Errorf("ClientIP = %q", got)
	}
}

func TestClientIPNoTrustUsesFirstHop(t *testing.T) {
	r := httptestNew("10.0.0.5, 203.0.113.10", "")
	got := ClientIP(r, false)
	if got != "10.0.0.5" {
		t.Errorf("ClientIP = %q", got)
	}
}

func TestClientIPRemoteAddr(t *testing.T) {
	r := httptestNew("", "192.0.2.1:1234")
	got := ClientIP(r, false)
	if got != "192.0.2.1" {
		t.Errorf("ClientIP = %q", got)
	}
}

func TestIsPrivateOrLoopback(t *testing.T) {
	if !IsPrivateOrLoopback(net.ParseIP("127.0.0.1")) {
		t.Fatal("loopback")
	}
	if IsPrivateOrLoopback(net.ParseIP("203.0.113.1")) {
		t.Fatal("public")
	}
}

func TestClientIPTrustProxyRealIP(t *testing.T) {
	r := httptestNew("", "192.0.2.1:1234")
	r.Header.Set("X-Real-IP", "203.0.113.8")
	got := ClientIP(r, true)
	if got != "203.0.113.8" {
		t.Errorf("got %q", got)
	}
}

func TestClientIPTrustProxyAllPrivate(t *testing.T) {
	r := httptestNew("10.0.0.1, 192.168.1.1", "192.0.2.1:1234")
	if got := ClientIP(r, true); got != "" {
		t.Errorf("got %q want empty", got)
	}
}

func TestClientIPRealIPNoTrust(t *testing.T) {
	r := httptestNew("", "192.0.2.1:1234")
	r.Header.Set("X-Real-IP", "203.0.113.9")
	if got := ClientIP(r, false); got != "203.0.113.9" {
		t.Errorf("got %q", got)
	}
}

func TestClientIPRemoteHostNoPort(t *testing.T) {
	r := httptestNew("", "192.0.2.5")
	if got := ClientIP(r, false); got != "192.0.2.5" {
		t.Errorf("got %q", got)
	}
}

func TestClientIPEmptyRemote(t *testing.T) {
	r := httptestNew("", "")
	if got := ClientIP(r, false); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestIsPrivateOrLoopbackNil(t *testing.T) {
	if !IsPrivateOrLoopback(nil) {
		t.Fatal("nil should be private")
	}
}

func TestIsPrivateOrLoopbackRFC1918(t *testing.T) {
	for _, ip := range []string{"172.16.0.1", "192.168.0.1", "10.1.2.3"} {
		if !IsPrivateOrLoopback(net.ParseIP(ip)) {
			t.Fatalf("%s should be private", ip)
		}
	}
}

func TestClientIPStripZone(t *testing.T) {
	r := httptestNew("fe80::1%eth0", "192.0.2.1:1234")
	if got := ClientIP(r, false); got != "fe80::1" {
		t.Errorf("got %q", got)
	}
}

func httptestNew(xff, remote string) *http.Request {
	r := &http.Request{Header: make(http.Header)}
	if xff != "" {
		r.Header.Set("X-Forwarded-For", xff)
	}
	r.RemoteAddr = remote
	return r
}
