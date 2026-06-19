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
	hits := hitsTableSQL(driver)
	agg := aggregateTableSQL(driver)
	out := make([]string, 0, len(hits)+len(agg))
	out = append(out, hits...)
	out = append(out, agg...)
	return out
}

func hitsTableSQL(driver string) []string {
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

func aggregateTableSQL(driver string) []string {
	switch driver {
	case "postgres":
		return []string{
			`CREATE TABLE IF NOT EXISTS kiko_paths (
				id SERIAL PRIMARY KEY,
				host VARCHAR(255) NOT NULL,
				path TEXT NOT NULL,
				title TEXT,
				UNIQUE (host, path)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_refs (
				id SERIAL PRIMARY KEY,
				host VARCHAR(255) NOT NULL,
				referrer TEXT NOT NULL,
				UNIQUE (host, referrer)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_hit_counts (
				host VARCHAR(255) NOT NULL,
				path_id INTEGER NOT NULL REFERENCES kiko_paths(id),
				hour TIMESTAMPTZ NOT NULL,
				total INTEGER NOT NULL DEFAULT 0,
				uniques INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (host, path_id, hour)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_ref_counts (
				host VARCHAR(255) NOT NULL,
				ref_id INTEGER NOT NULL REFERENCES kiko_refs(id),
				hour TIMESTAMPTZ NOT NULL,
				total INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (host, ref_id, hour)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_hit_uniques (
				host VARCHAR(255) NOT NULL,
				path_id INTEGER NOT NULL REFERENCES kiko_paths(id),
				hour TIMESTAMPTZ NOT NULL,
				visitor_hash CHAR(64) NOT NULL,
				PRIMARY KEY (host, path_id, hour, visitor_hash)
			)`,
		}
	case "mysql":
		return []string{
			`CREATE TABLE IF NOT EXISTS kiko_paths (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				host VARCHAR(255) NOT NULL,
				path TEXT NOT NULL,
				title TEXT,
				UNIQUE KEY uq_kiko_paths_host_path (host, path(255))
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_refs (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				host VARCHAR(255) NOT NULL,
				referrer TEXT NOT NULL,
				UNIQUE KEY uq_kiko_refs_host_ref (host, referrer(255))
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_hit_counts (
				host VARCHAR(255) NOT NULL,
				path_id BIGINT NOT NULL,
				hour DATETIME(6) NOT NULL,
				total INT NOT NULL DEFAULT 0,
				uniques INT NOT NULL DEFAULT 0,
				PRIMARY KEY (host, path_id, hour),
				FOREIGN KEY (path_id) REFERENCES kiko_paths(id)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_ref_counts (
				host VARCHAR(255) NOT NULL,
				ref_id BIGINT NOT NULL,
				hour DATETIME(6) NOT NULL,
				total INT NOT NULL DEFAULT 0,
				PRIMARY KEY (host, ref_id, hour),
				FOREIGN KEY (ref_id) REFERENCES kiko_refs(id)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_hit_uniques (
				host VARCHAR(255) NOT NULL,
				path_id BIGINT NOT NULL,
				hour DATETIME(6) NOT NULL,
				visitor_hash CHAR(64) NOT NULL,
				PRIMARY KEY (host, path_id, hour, visitor_hash),
				FOREIGN KEY (path_id) REFERENCES kiko_paths(id)
			)`,
		}
	default:
		return []string{
			`CREATE TABLE IF NOT EXISTS kiko_paths (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				host TEXT NOT NULL,
				path TEXT NOT NULL,
				title TEXT,
				UNIQUE(host, path)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_refs (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				host TEXT NOT NULL,
				referrer TEXT NOT NULL,
				UNIQUE(host, referrer)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_hit_counts (
				host TEXT NOT NULL,
				path_id INTEGER NOT NULL,
				hour TEXT NOT NULL,
				total INTEGER NOT NULL DEFAULT 0,
				uniques INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (host, path_id, hour),
				FOREIGN KEY (path_id) REFERENCES kiko_paths(id)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_ref_counts (
				host TEXT NOT NULL,
				ref_id INTEGER NOT NULL,
				hour TEXT NOT NULL,
				total INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (host, ref_id, hour),
				FOREIGN KEY (ref_id) REFERENCES kiko_refs(id)
			)`,
			`CREATE TABLE IF NOT EXISTS kiko_hit_uniques (
				host TEXT NOT NULL,
				path_id INTEGER NOT NULL,
				hour TEXT NOT NULL,
				visitor_hash TEXT NOT NULL,
				PRIMARY KEY (host, path_id, hour, visitor_hash),
				FOREIGN KEY (path_id) REFERENCES kiko_paths(id)
			)`,
		}
	}
}
