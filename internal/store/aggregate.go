package store

import (
	"database/sql"
	"time"

	"github.com/hrodrig/kiko/internal/hit"
)

type pathAgg struct {
	host  string
	path  string
	title string
	total int
	hash  map[string]struct{}
}

type refAgg struct {
	host     string
	referrer string
	total    int
}

func (s *sqlStore) aggregateHits(tx *sql.Tx, hits []hit.Hit, hour time.Time) error {
	paths := map[string]*pathAgg{}
	refs := map[string]*refAgg{}

	for _, h := range hits {
		pk := h.Host + "\x00" + h.Path
		p, ok := paths[pk]
		if !ok {
			p = &pathAgg{host: h.Host, path: h.Path, title: h.Title, hash: map[string]struct{}{}}
			paths[pk] = p
		}
		p.total++
		if h.Title != "" && p.title == "" {
			p.title = h.Title
		}
		if h.VisitorHash != "" {
			p.hash[h.VisitorHash] = struct{}{}
		}

		if h.Referrer == "" {
			continue
		}
		rk := h.Host + "\x00" + h.Referrer
		r, ok := refs[rk]
		if !ok {
			r = &refAgg{host: h.Host, referrer: h.Referrer}
			refs[rk] = r
		}
		r.total++
	}

	for _, p := range paths {
		pathID, err := s.upsertPath(tx, p.host, p.path, p.title)
		if err != nil {
			return err
		}
		if err := s.addHitCount(tx, p.host, pathID, hour, p.total); err != nil {
			return err
		}
		for vh := range p.hash {
			if err := s.addHitUnique(tx, p.host, pathID, hour, vh); err != nil {
				return err
			}
		}
	}

	for _, r := range refs {
		refID, err := s.upsertRef(tx, r.host, r.referrer)
		if err != nil {
			return err
		}
		if err := s.addRefCount(tx, r.host, refID, hour, r.total); err != nil {
			return err
		}
	}
	return nil
}

func (s *sqlStore) upsertPath(tx *sql.Tx, host, path, title string) (int64, error) {
	switch s.driver {
	case "postgres":
		var id int64
		err := tx.QueryRow(
			`INSERT INTO kiko_paths (host, path, title) VALUES ($1, $2, $3)
			 ON CONFLICT (host, path) DO UPDATE SET title = COALESCE(NULLIF(excluded.title, ''), kiko_paths.title)
			 RETURNING id`,
			host, path, nullString(title),
		).Scan(&id)
		return id, err
	case "mysql":
		res, err := tx.Exec(
			`INSERT INTO kiko_paths (host, path, title) VALUES (?, ?, ?)
			 ON DUPLICATE KEY UPDATE title = COALESCE(NULLIF(VALUES(title), ''), title), id = LAST_INSERT_ID(id)`,
			host, path, nullString(title),
		)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	default:
		if _, err := tx.Exec(
			`INSERT INTO kiko_paths (host, path, title) VALUES (?, ?, ?)
			 ON CONFLICT(host, path) DO UPDATE SET title = COALESCE(NULLIF(excluded.title, ''), kiko_paths.title)`,
			host, path, nullString(title),
		); err != nil {
			return 0, err
		}
		var id int64
		err := tx.QueryRow(`SELECT id FROM kiko_paths WHERE host = ? AND path = ?`, host, path).Scan(&id)
		return id, err
	}
}

func (s *sqlStore) upsertRef(tx *sql.Tx, host, referrer string) (int64, error) {
	switch s.driver {
	case "postgres":
		var id int64
		err := tx.QueryRow(
			`INSERT INTO kiko_refs (host, referrer) VALUES ($1, $2)
			 ON CONFLICT (host, referrer) DO UPDATE SET referrer = excluded.referrer
			 RETURNING id`,
			host, referrer,
		).Scan(&id)
		return id, err
	case "mysql":
		res, err := tx.Exec(
			`INSERT INTO kiko_refs (host, referrer) VALUES (?, ?)
			 ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)`,
			host, referrer,
		)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	default:
		if _, err := tx.Exec(
			`INSERT INTO kiko_refs (host, referrer) VALUES (?, ?)
			 ON CONFLICT(host, referrer) DO NOTHING`,
			host, referrer,
		); err != nil {
			return 0, err
		}
		var id int64
		err := tx.QueryRow(`SELECT id FROM kiko_refs WHERE host = ? AND referrer = ?`, host, referrer).Scan(&id)
		return id, err
	}
}

func (s *sqlStore) addHitCount(tx *sql.Tx, host string, pathID int64, hour time.Time, delta int) error {
	q, args := hitCountUpsertSQL(s.driver, host, pathID, hour, delta)
	_, err := tx.Exec(q, args...)
	return err
}

func (s *sqlStore) addHitUnique(tx *sql.Tx, host string, pathID int64, hour time.Time, visitorHash string) error {
	inserted, err := s.insertHitUnique(tx, host, pathID, hour, visitorHash)
	if err != nil || !inserted {
		return err
	}
	q, args := hitUniqueBumpSQL(s.driver, host, pathID, hour)
	_, err = tx.Exec(q, args...)
	return err
}

func (s *sqlStore) insertHitUnique(tx *sql.Tx, host string, pathID int64, hour time.Time, visitorHash string) (bool, error) {
	q, args := hitUniqueInsertSQL(s.driver, host, pathID, hour, visitorHash)
	res, err := tx.Exec(q, args...)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *sqlStore) addRefCount(tx *sql.Tx, host string, refID int64, hour time.Time, delta int) error {
	q, args := refCountUpsertSQL(s.driver, host, refID, hour, delta)
	_, err := tx.Exec(q, args...)
	return err
}

func hourValue(driver string, hour time.Time) any {
	switch driver {
	case "postgres":
		return hour
	case "mysql":
		return hour.UTC()
	default:
		return hour.UTC().Format(time.RFC3339)
	}
}

func hitCountUpsertSQL(driver string, host string, pathID int64, hour time.Time, delta int) (string, []any) {
	hv := hourValue(driver, hour)
	switch driver {
	case "postgres":
		return `INSERT INTO kiko_hit_counts (host, path_id, hour, total, uniques)
			VALUES ($1, $2, $3, $4, 0)
			ON CONFLICT (host, path_id, hour) DO UPDATE SET total = kiko_hit_counts.total + EXCLUDED.total`,
			[]any{host, pathID, hv, delta}
	case "mysql":
		return `INSERT INTO kiko_hit_counts (host, path_id, hour, total, uniques)
			VALUES (?, ?, ?, ?, 0)
			ON DUPLICATE KEY UPDATE total = total + VALUES(total)`,
			[]any{host, pathID, hv, delta}
	default:
		return `INSERT INTO kiko_hit_counts (host, path_id, hour, total, uniques)
			VALUES (?, ?, ?, ?, 0)
			ON CONFLICT(host, path_id, hour) DO UPDATE SET total = total + excluded.total`,
			[]any{host, pathID, hv, delta}
	}
}

func hitUniqueInsertSQL(driver string, host string, pathID int64, hour time.Time, visitorHash string) (string, []any) {
	hv := hourValue(driver, hour)
	switch driver {
	case "postgres":
		return `INSERT INTO kiko_hit_uniques (host, path_id, hour, visitor_hash)
			VALUES ($1, $2, $3, $4) ON CONFLICT (host, path_id, hour, visitor_hash) DO NOTHING`,
			[]any{host, pathID, hv, visitorHash}
	case "mysql":
		return `INSERT IGNORE INTO kiko_hit_uniques (host, path_id, hour, visitor_hash)
			VALUES (?, ?, ?, ?)`,
			[]any{host, pathID, hv, visitorHash}
	default:
		return `INSERT OR IGNORE INTO kiko_hit_uniques (host, path_id, hour, visitor_hash)
			VALUES (?, ?, ?, ?)`,
			[]any{host, pathID, hv, visitorHash}
	}
}

func hitUniqueBumpSQL(driver string, host string, pathID int64, hour time.Time) (string, []any) {
	hv := hourValue(driver, hour)
	switch driver {
	case "postgres":
		return `UPDATE kiko_hit_counts SET uniques = uniques + 1
			WHERE host = $1 AND path_id = $2 AND hour = $3`,
			[]any{host, pathID, hv}
	case "mysql":
		return `UPDATE kiko_hit_counts SET uniques = uniques + 1
			WHERE host = ? AND path_id = ? AND hour = ?`,
			[]any{host, pathID, hv}
	default:
		return `UPDATE kiko_hit_counts SET uniques = uniques + 1
			WHERE host = ? AND path_id = ? AND hour = ?`,
			[]any{host, pathID, hv}
	}
}

func refCountUpsertSQL(driver string, host string, refID int64, hour time.Time, delta int) (string, []any) {
	hv := hourValue(driver, hour)
	switch driver {
	case "postgres":
		return `INSERT INTO kiko_ref_counts (host, ref_id, hour, total)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (host, ref_id, hour) DO UPDATE SET total = kiko_ref_counts.total + EXCLUDED.total`,
			[]any{host, refID, hv, delta}
	case "mysql":
		return `INSERT INTO kiko_ref_counts (host, ref_id, hour, total)
			VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE total = total + VALUES(total)`,
			[]any{host, refID, hv, delta}
	default:
		return `INSERT INTO kiko_ref_counts (host, ref_id, hour, total)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(host, ref_id, hour) DO UPDATE SET total = total + excluded.total`,
			[]any{host, refID, hv, delta}
	}
}
