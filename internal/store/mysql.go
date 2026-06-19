package store

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/hrodrig/kiko/internal/config"
)

func openMySQL(cfg config.DatabaseCfg) (*sqlStore, error) {
	db, err := sqlOpen("mysql", cfg.MySQLDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)

	s := &sqlStore{db: db, driver: "mysql"}
	if err := migrate(s.db, "mysql"); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
