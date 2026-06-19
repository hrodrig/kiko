package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if cfg.Listen != ":8080" {
		t.Errorf("Listen = %q; want :8080", cfg.Listen)
	}
	if cfg.Buffer.FlushInterval != 10 {
		t.Errorf("FlushInterval = %d; want 10", cfg.Buffer.FlushInterval)
	}
	if cfg.RateLimit.RequestsPerSec != 100 {
		t.Errorf("RequestsPerSec = %d; want 100", cfg.RateLimit.RequestsPerSec)
	}
}

func TestLoadFromFile(t *testing.T) {
	content := []byte(`
listen: ":9090"
log_level: debug
database:
  host: pg.example.com
  port: 5432
  user: kiko
  password: secret
  dbname: kiko_prod
allowed_hosts:
  - gghstats.com
  - 10.0.0.0/8
`)
	dir := t.TempDir()
	path := filepath.Join(dir, "kiko.yml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%q) = %v", path, err)
	}
	if cfg.Listen != ":9090" {
		t.Errorf("Listen = %q; want :9090", cfg.Listen)
	}
	if cfg.Database.Host != "pg.example.com" {
		t.Errorf("DB host = %q; want pg.example.com", cfg.Database.Host)
	}
	if len(cfg.AllowedHosts) != 2 || cfg.AllowedHosts[0] != "gghstats.com" {
		t.Errorf("AllowedHosts = %v; want [gghstats.com 10.0.0.0/8]", cfg.AllowedHosts)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	t.Setenv("KIKO_LISTEN", ":9999")
	t.Setenv("KIKO_LOG_LEVEL", "error")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if cfg.Listen != ":9999" {
		t.Errorf("Listen = %q; want :9999", cfg.Listen)
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	t.Setenv("KIKO_LOG_LEVEL", "bogus")
	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for bogus log level")
	}
}
