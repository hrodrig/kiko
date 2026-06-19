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
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Database.Driver = %q; want sqlite", cfg.Database.Driver)
	}
	if cfg.Database.Path != "./data/kiko.db" {
		t.Errorf("Database.Path = %q; want ./data/kiko.db", cfg.Database.Path)
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

func TestLoadInvalidDriver(t *testing.T) {
	t.Setenv("KIKO_DATABASE_DRIVER", "oracle")
	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}

func TestLoadInvalidBufferCapacity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kiko.yml")
	content := []byte("buffer:\n  capacity: 0\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for zero buffer capacity")
	}
}

func TestLoadMySQLDefaultPort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kiko.yml")
	content := []byte("database:\n  driver: mysql\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if cfg.Database.Port != 3306 {
		t.Errorf("mysql port = %d; want 3306", cfg.Database.Port)
	}
}

func TestLoadInvalidRateLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kiko.yml")
	content := []byte("rate_limit:\n  enabled: true\n  requests_per_sec: 0\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for zero requests_per_sec")
	}
}

func TestLoadRateLimitDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kiko.yml")
	content := []byte("rate_limit:\n  enabled: false\n  requests_per_sec: 0\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if cfg.RateLimit.Enabled {
		t.Error("expected rate limit disabled")
	}
}

func TestDatabaseDSNHelpers(t *testing.T) {
	cfg := DatabaseCfg{
		User:     "u",
		Password: "p",
		Host:     "db.local",
		Port:     5432,
		DBName:   "kiko",
		SSLMode:  "require",
	}
	if got := cfg.DSNString(); got == "" {
		t.Fatal("DSNString empty")
	}
	cfg.DSN = "postgres://override"
	if got := cfg.DSNString(); got != "postgres://override" {
		t.Errorf("DSNString override = %q", got)
	}
	cfg.DSN = ""
	cfg.Port = 3306
	if got := cfg.MySQLDSN(); got == "" {
		t.Fatal("MySQLDSN empty")
	}
}

func TestLoadInvalidAPIRateLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kiko.yml")
	content := []byte("api:\n  rate_limit:\n    enabled: true\n    requests_per_sec: 0\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected api rate limit error")
	}
}

func TestLoadAPIKeyEnv(t *testing.T) {
	t.Setenv("KIKO_API_KEY", "secret")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.API.Key != "secret" {
		t.Errorf("API.Key = %q", cfg.API.Key)
	}
}

func TestNormalizedDriver(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "sqlite"},
		{"sqlite3", "sqlite"},
		{"postgresql", "postgres"},
		{"mysql", "mysql"},
	}
	for _, tt := range tests {
		cfg := DatabaseCfg{Driver: tt.in}
		if got := cfg.NormalizedDriver(); got != tt.want {
			t.Errorf("NormalizedDriver(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
