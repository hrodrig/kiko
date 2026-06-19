package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hrodrig/kiko/internal/config"
	_ "modernc.org/sqlite"
)

func openSQLite(cfg config.DatabaseCfg) (*sqlStore, error) {
	path := cfg.Path
	if path == "" {
		path = "./data/kiko.db"
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sqlOpen("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	s := &sqlStore{db: db, driver: "sqlite"}
	if err := migrate(s.db, "sqlite"); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
