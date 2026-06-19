package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
)

// Store persists hits and reports backend health.
type Store interface {
	SaveHits(hits []hit.Hit) error
	Ping(ctx context.Context) error
	Close() error
}

type sqlStore struct {
	db     *sql.DB
	driver string
}

// Open connects to the configured database backend and runs migrations.
func Open(cfg config.DatabaseCfg) (Store, error) {
	driver := cfg.NormalizedDriver()
	switch driver {
	case "sqlite":
		return openSQLite(cfg)
	case "postgres":
		return openPostgres(cfg)
	case "mysql":
		return openMySQL(cfg)
	default:
		return nil, fmt.Errorf("store: unsupported driver %q (want sqlite, postgres, mysql)", cfg.Driver)
	}
}

func (s *sqlStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *sqlStore) Close() error {
	return s.db.Close()
}

func (s *sqlStore) SaveHits(hits []hit.Hit) error {
	if len(hits) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertHitSQL(s.driver))
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, h := range hits {
		if _, err := stmt.Exec(h.Host, h.Path, nullString(h.Referrer), h.Width, nullString(h.Title)); err != nil {
			return fmt.Errorf("insert hit: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func insertHitSQL(driver string) string {
	switch driver {
	case "postgres":
		return `INSERT INTO kiko_hits (host, path, referrer, screen_width, title)
			VALUES ($1, $2, $3, $4, $5)`
	default:
		return `INSERT INTO kiko_hits (host, path, referrer, screen_width, title)
			VALUES (?, ?, ?, ?, ?)`
	}
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
