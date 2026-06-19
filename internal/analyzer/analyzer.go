package analyzer

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"
)

// Analyzer runs read-only stats queries against kiko_hits and aggregate tables.
type Analyzer struct {
	db     *sql.DB
	driver string
}

// New attaches an Analyzer to an open SQL store.
func New(db *sql.DB, driver string) *Analyzer {
	return &Analyzer{db: db, driver: driver}
}

// Query holds common filter parameters for stats endpoints.
type Query struct {
	Host     string
	Since    time.Time
	Until    time.Time
	Limit    int
	Interval string // hour or day
}

// ParseQuery builds Query from URL values.
func ParseQuery(v url.Values) (Query, error) {
	host := v.Get("host")
	if host == "" {
		return Query{}, fmt.Errorf("host is required")
	}
	until := time.Now().UTC()
	if s := v.Get("until"); s != "" {
		t, err := parseTime(s)
		if err != nil {
			return Query{}, fmt.Errorf("until: %w", err)
		}
		until = t
	}
	since := until.Add(-30 * 24 * time.Hour)
	if s := v.Get("since"); s != "" {
		t, err := parseTime(s)
		if err != nil {
			return Query{}, fmt.Errorf("since: %w", err)
		}
		since = t
	}
	if !since.Before(until) {
		return Query{}, fmt.Errorf("since must be before until")
	}
	limit := 10
	if s := v.Get("limit"); s != "" {
		var n int
		if _, err := fmt.Sscanf(s, "%d", &n); err != nil || n <= 0 {
			return Query{}, fmt.Errorf("invalid limit")
		}
		limit = n
	}
	if limit > 100 {
		limit = 100
	}
	interval := v.Get("interval")
	if interval == "" {
		interval = "day"
	}
	if interval != "hour" && interval != "day" {
		return Query{}, fmt.Errorf("interval must be hour or day")
	}
	return Query{
		Host:     host,
		Since:    since,
		Until:    until,
		Limit:    limit,
		Interval: interval,
	}, nil
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// Summary holds headline metrics for a site.
type Summary struct {
	Host     string `json:"host"`
	Since    string `json:"since"`
	Until    string `json:"until"`
	Hits     int64  `json:"hits"`
	Uniques  int64  `json:"uniques"`
	TopPath  string `json:"top_path,omitempty"`
	TopPathN int64  `json:"top_path_hits,omitempty"`
}

// Row is a labeled count pair used in breakdown tables.
type Row struct {
	Label   string `json:"label"`
	Hits    int64  `json:"hits"`
	Uniques int64  `json:"uniques"`
}

// PathRow is a path with hit counts.
type PathRow struct {
	Path    string `json:"path"`
	Title   string `json:"title,omitempty"`
	Hits    int64  `json:"hits"`
	Uniques int64  `json:"uniques"`
}

// RefRow is a referrer with hit counts.
type RefRow struct {
	Referrer string `json:"referrer"`
	Source   string `json:"source,omitempty"`
	Hits     int64  `json:"hits"`
	Uniques  int64  `json:"uniques"`
}

// TimelinePoint is one bucket in a time series.
type TimelinePoint struct {
	Period  string `json:"period"`
	Hits    int64  `json:"hits"`
	Uniques int64  `json:"uniques"`
}

// Visitors holds unique visitor count for a range.
type Visitors struct {
	Host    string `json:"host"`
	Since   string `json:"since"`
	Until   string `json:"until"`
	Uniques int64  `json:"uniques"`
}

func (a *Analyzer) Summary(ctx context.Context, q Query) (Summary, error) {
	hits, uniques, err := a.countHits(ctx, q)
	if err != nil {
		return Summary{}, err
	}
	out := Summary{
		Host:    q.Host,
		Since:   q.Since.Format(time.RFC3339),
		Until:   q.Until.Format(time.RFC3339),
		Hits:    hits,
		Uniques: uniques,
	}
	path, n, err := a.topPath(ctx, q)
	if err != nil {
		return Summary{}, err
	}
	out.TopPath = path
	out.TopPathN = n
	return out, nil
}

func (a *Analyzer) Paths(ctx context.Context, q Query) ([]PathRow, error) {
	sqlQ, args := pathsSQL(a.driver, q)
	return scanPaths(ctx, a.db, sqlQ, args)
}

func (a *Analyzer) Refs(ctx context.Context, q Query) ([]RefRow, error) {
	sqlQ, args := refsSQL(a.driver, q)
	return scanRefs(ctx, a.db, sqlQ, args)
}

func (a *Analyzer) Channels(ctx context.Context, q Query) ([]Row, error) {
	return a.groupBy(ctx, q, "channel", "direct")
}

func (a *Analyzer) Browsers(ctx context.Context, q Query) ([]Row, error) {
	return a.groupBy(ctx, q, "browser", "unknown")
}

func (a *Analyzer) OS(ctx context.Context, q Query) ([]Row, error) {
	return a.groupBy(ctx, q, "os", "unknown")
}

func (a *Analyzer) UTMSources(ctx context.Context, q Query) ([]Row, error) {
	return a.groupByNonEmpty(ctx, q, "utm_source")
}

func (a *Analyzer) Timeline(ctx context.Context, q Query) ([]TimelinePoint, error) {
	sqlQ, args := timelineSQL(a.driver, q)
	rows, err := a.db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TimelinePoint
	for rows.Next() {
		var p TimelinePoint
		if err := rows.Scan(&p.Period, &p.Hits, &p.Uniques); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (a *Analyzer) Visitors(ctx context.Context, q Query) (Visitors, error) {
	_, uniques, err := a.countHits(ctx, q)
	if err != nil {
		return Visitors{}, err
	}
	return Visitors{
		Host:    q.Host,
		Since:   q.Since.Format(time.RFC3339),
		Until:   q.Until.Format(time.RFC3339),
		Uniques: uniques,
	}, nil
}

func (a *Analyzer) countHits(ctx context.Context, q Query) (hits, uniques int64, err error) {
	sqlQ, args := countSQL(a.driver, q)
	err = a.db.QueryRowContext(ctx, sqlQ, args...).Scan(&hits, &uniques)
	return hits, uniques, err
}

func (a *Analyzer) topPath(ctx context.Context, q Query) (path string, n int64, err error) {
	sqlQ, args := topPathSQL(a.driver, q)
	err = a.db.QueryRowContext(ctx, sqlQ, args...).Scan(&path, &n)
	if err == sql.ErrNoRows {
		return "", 0, nil
	}
	return path, n, err
}

func (a *Analyzer) groupBy(ctx context.Context, q Query, col, emptyLabel string) ([]Row, error) {
	sqlQ, args := groupSQL(a.driver, q, col, emptyLabel)
	rows, err := a.db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.Label, &r.Hits, &r.Uniques); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (a *Analyzer) groupByNonEmpty(ctx context.Context, q Query, col string) ([]Row, error) {
	sqlQ, args := groupNonEmptySQL(a.driver, q, col)
	rows, err := a.db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.Label, &r.Hits, &r.Uniques); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanPaths(ctx context.Context, db *sql.DB, sqlQ string, args []any) ([]PathRow, error) {
	rows, err := db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PathRow
	for rows.Next() {
		var r PathRow
		var title sql.NullString
		if err := rows.Scan(&r.Path, &title, &r.Hits, &r.Uniques); err != nil {
			return nil, err
		}
		if title.Valid {
			r.Title = title.String
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanRefs(ctx context.Context, db *sql.DB, sqlQ string, args []any) ([]RefRow, error) {
	rows, err := db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RefRow
	for rows.Next() {
		var r RefRow
		var ref, src sql.NullString
		if err := rows.Scan(&ref, &src, &r.Hits, &r.Uniques); err != nil {
			return nil, err
		}
		if ref.Valid {
			r.Referrer = ref.String
		}
		if src.Valid {
			r.Source = src.String
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
