package store

import (
	"context"
	"os"
	"testing"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
)

func TestPostgresIntegration(t *testing.T) {
	cfg := postgresCfgFromEnv(t)
	st := openIntegrationStore(t, cfg)
	defer st.Close()

	runIntegrationSaveHits(t, st)
}

func TestMySQLIntegration(t *testing.T) {
	cfg := mysqlCfgFromEnv(t)
	st := openIntegrationStore(t, cfg)
	defer st.Close()

	runIntegrationSaveHits(t, st)
}

func postgresCfgFromEnv(t *testing.T) config.DatabaseCfg {
	t.Helper()
	if dsn := os.Getenv("KIKO_TEST_POSTGRES_DSN"); dsn != "" {
		return config.DatabaseCfg{Driver: "postgres", DSN: dsn}
	}
	host := os.Getenv("KIKO_TEST_POSTGRES_HOST")
	if host == "" {
		t.Skip("set KIKO_TEST_POSTGRES_DSN or KIKO_TEST_POSTGRES_HOST for postgres integration")
	}
	return config.DatabaseCfg{
		Driver:   "postgres",
		Host:     host,
		Port:     envInt("KIKO_TEST_POSTGRES_PORT", 5432),
		User:     envOr("KIKO_TEST_POSTGRES_USER", "kiko"),
		Password: os.Getenv("KIKO_TEST_POSTGRES_PASSWORD"),
		DBName:   envOr("KIKO_TEST_POSTGRES_DB", "kiko_test"),
		SSLMode:  envOr("KIKO_TEST_POSTGRES_SSLMODE", "disable"),
	}
}

func mysqlCfgFromEnv(t *testing.T) config.DatabaseCfg {
	t.Helper()
	if dsn := os.Getenv("KIKO_TEST_MYSQL_DSN"); dsn != "" {
		return config.DatabaseCfg{Driver: "mysql", DSN: dsn}
	}
	host := os.Getenv("KIKO_TEST_MYSQL_HOST")
	if host == "" {
		t.Skip("set KIKO_TEST_MYSQL_DSN or KIKO_TEST_MYSQL_HOST for mysql integration")
	}
	return config.DatabaseCfg{
		Driver:   "mysql",
		Host:     host,
		Port:     envInt("KIKO_TEST_MYSQL_PORT", 3306),
		User:     envOr("KIKO_TEST_MYSQL_USER", "kiko"),
		Password: os.Getenv("KIKO_TEST_MYSQL_PASSWORD"),
		DBName:   envOr("KIKO_TEST_MYSQL_DB", "kiko_test"),
	}
}

func openIntegrationStore(t *testing.T, cfg config.DatabaseCfg) Store {
	t.Helper()
	st, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	return st
}

func runIntegrationSaveHits(t *testing.T, st Store) {
	t.Helper()
	hits := []hit.Hit{
		{
			Host:        "integration.test",
			Path:        "/a",
			Referrer:    "https://google.com/search",
			VisitorHash: "abc",
			Browser:     "Chrome",
			OS:          "Linux",
			Channel:     "organic",
		},
	}
	if err := st.SaveHits(hits); err != nil {
		t.Fatalf("SaveHits() = %v", err)
	}
	if err := st.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() = %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n := 0
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if n == 0 && v != "0" {
		return def
	}
	return n
}
