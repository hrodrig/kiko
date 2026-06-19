package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/log"
)

func TestInitRuntimeSQLite(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseCfg{Driver: "sqlite", Path: ":memory:"},
		Buffer:   config.BufferCfg{FlushInterval: 10, Capacity: 64},
		Log:      log.New(nil, log.Off),
	}

	st, buf, handler, cleanup, err := initRuntime(cfg)
	if err != nil {
		t.Fatalf("initRuntime() = %v", err)
	}
	defer cleanup()

	ts := httptest.NewServer(handler)
	defer ts.Close()

	body := `{"host":"test.dev","path":"/ok"}`
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/hit", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /hit = %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want 200", resp.StatusCode)
	}

	flushHits(st, buf, cfg.Log)
	if err := st.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() = %v", err)
	}
}

func TestInitRuntimeBadDatabase(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseCfg{Driver: "postgres", Host: "127.0.0.1", Port: 1, DBName: "kiko"},
		Buffer:   config.BufferCfg{FlushInterval: 10, Capacity: 64},
		Log:      log.New(nil, log.Off),
	}
	_, _, _, _, err := initRuntime(cfg)
	if err == nil {
		t.Fatal("expected database open error")
	}
}

func TestFlushHitsWarnsOnDrops(t *testing.T) {
	st := &fakeStore{}
	buf := hit.NewBuffer(1)
	buf.Append(hit.Hit{Host: "a", Path: "/"})
	buf.Append(hit.Hit{Host: "b", Path: "/"})
	flushHits(st, buf, log.New(nil, log.Warn))
	if buf.Drops() != 1 {
		t.Errorf("drops = %d; want 1", buf.Drops())
	}
}
