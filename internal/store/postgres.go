package store

import (
	"github.com/hrodrig/kiko/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openPostgres(cfg config.DatabaseCfg) (*sqlStore, error) {
	db, err := sqlOpen("pgx", cfg.DSNString())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)

	s := &sqlStore{db: db, driver: "postgres"}
	if err := migrate(s.db, "postgres"); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
