package store

import "database/sql"

// DBAccessor exposes the underlying SQL connection for read-only analytics.
type DBAccessor interface {
	StatsDB() (*sql.DB, string)
}

func (s *sqlStore) StatsDB() (*sql.DB, string) {
	return s.db, s.driver
}
