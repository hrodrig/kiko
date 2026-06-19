package store

import (
	"database/sql"
	"fmt"
	"strings"
)

func migrate(db *sql.DB, driver string) error {
	for _, stmt := range schemaSQL(driver) {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate (%s): %w", driver, err)
		}
	}
	if driver == "mysql" {
		if _, err := db.Exec(`CREATE INDEX idx_kiko_hits_host_date ON kiko_hits (host, created_at)`); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				return fmt.Errorf("migrate (mysql index): %w", err)
			}
		}
	}
	return nil
}

func schemaSQL(driver string) []string {
	switch driver {
	case "postgres":
		return []string{
			`CREATE TABLE IF NOT EXISTS kiko_hits (
				id BIGSERIAL PRIMARY KEY,
				host VARCHAR(255) NOT NULL,
				path TEXT NOT NULL,
				referrer TEXT,
				visitor_hash CHAR(64) NOT NULL DEFAULT '',
				screen_width SMALLINT,
				title TEXT,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
			`CREATE INDEX IF NOT EXISTS idx_kiko_hits_host_date ON kiko_hits (host, created_at DESC)`,
		}
	case "mysql":
		return []string{
			`CREATE TABLE IF NOT EXISTS kiko_hits (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				host VARCHAR(255) NOT NULL,
				path TEXT NOT NULL,
				referrer TEXT,
				visitor_hash CHAR(64) NOT NULL DEFAULT '',
				screen_width SMALLINT,
				title TEXT,
				created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
			)`,
		}
	default:
		return []string{
			`CREATE TABLE IF NOT EXISTS kiko_hits (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				host TEXT NOT NULL,
				path TEXT NOT NULL,
				referrer TEXT,
				visitor_hash TEXT NOT NULL DEFAULT '',
				screen_width INTEGER,
				title TEXT,
				created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
			)`,
			`CREATE INDEX IF NOT EXISTS idx_kiko_hits_host_date ON kiko_hits (host, created_at)`,
		}
	}
}
