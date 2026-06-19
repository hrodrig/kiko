package analyzer

import "time"

func timeArgs(driver string, q Query) (since, until any) {
	switch driver {
	case "postgres":
		return q.Since, q.Until
	case "mysql":
		return q.Since, q.Until
	default:
		return q.Since.UTC().Format(time.RFC3339), q.Until.UTC().Format(time.RFC3339)
	}
}

func countSQL(driver string, q Query) (string, []any) {
	since, until := timeArgs(driver, q)
	switch driver {
	case "postgres":
		return `SELECT COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = $1 AND created_at >= $2 AND created_at < $3`,
			[]any{q.Host, since, until}
	case "mysql":
		return `SELECT COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?`,
			[]any{q.Host, since, until}
	default:
		return `SELECT COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?`,
			[]any{q.Host, since, until}
	}
}

func topPathSQL(driver string, q Query) (string, []any) {
	since, until := timeArgs(driver, q)
	switch driver {
	case "postgres":
		return `SELECT path, COUNT(*) AS hits FROM kiko_hits
			WHERE host = $1 AND created_at >= $2 AND created_at < $3
			GROUP BY path ORDER BY hits DESC LIMIT 1`,
			[]any{q.Host, since, until}
	case "mysql":
		return `SELECT path, COUNT(*) AS hits FROM kiko_hits
			WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY path ORDER BY hits DESC LIMIT 1`,
			[]any{q.Host, since, until}
	default:
		return `SELECT path, COUNT(*) AS hits FROM kiko_hits
			WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY path ORDER BY hits DESC LIMIT 1`,
			[]any{q.Host, since, until}
	}
}

func pathsSQL(driver string, q Query) (string, []any) {
	since, until := timeArgs(driver, q)
	switch driver {
	case "postgres":
		return `SELECT path, MAX(title), COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = $1 AND created_at >= $2 AND created_at < $3
			GROUP BY path ORDER BY COUNT(*) DESC LIMIT $4`,
			[]any{q.Host, since, until, q.Limit}
	case "mysql":
		return `SELECT path, MAX(title), COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY path ORDER BY COUNT(*) DESC LIMIT ?`,
			[]any{q.Host, since, until, q.Limit}
	default:
		return `SELECT path, MAX(title), COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY path ORDER BY COUNT(*) DESC LIMIT ?`,
			[]any{q.Host, since, until, q.Limit}
	}
}

func refsSQL(driver string, q Query) (string, []any) {
	since, until := timeArgs(driver, q)
	switch driver {
	case "postgres":
		return `SELECT COALESCE(referrer, ''), MAX(source), COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = $1 AND created_at >= $2 AND created_at < $3
			GROUP BY referrer ORDER BY COUNT(*) DESC LIMIT $4`,
			[]any{q.Host, since, until, q.Limit}
	case "mysql":
		return `SELECT COALESCE(referrer, ''), MAX(source), COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY referrer ORDER BY COUNT(*) DESC LIMIT ?`,
			[]any{q.Host, since, until, q.Limit}
	default:
		return `SELECT COALESCE(referrer, ''), MAX(source), COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY referrer ORDER BY COUNT(*) DESC LIMIT ?`,
			[]any{q.Host, since, until, q.Limit}
	}
}

func groupSQL(driver string, q Query, col, emptyLabel string) (string, []any) {
	since, until := timeArgs(driver, q)
	labelExpr := coalesceCol(driver, col, emptyLabel)
	switch driver {
	case "postgres":
		return `SELECT ` + labelExpr + `, COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = $1 AND created_at >= $2 AND created_at < $3
			GROUP BY 1 ORDER BY COUNT(*) DESC LIMIT $4`,
			[]any{q.Host, since, until, q.Limit}
	case "mysql":
		return `SELECT ` + labelExpr + `, COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY 1 ORDER BY COUNT(*) DESC LIMIT ?`,
			[]any{q.Host, since, until, q.Limit}
	default:
		return `SELECT ` + labelExpr + `, COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY 1 ORDER BY COUNT(*) DESC LIMIT ?`,
			[]any{q.Host, since, until, q.Limit}
	}
}

func groupNonEmptySQL(driver string, q Query, col string) (string, []any) {
	since, until := timeArgs(driver, q)
	switch driver {
	case "postgres":
		return `SELECT ` + col + `, COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = $1 AND created_at >= $2 AND created_at < $3
			AND ` + col + ` IS NOT NULL AND ` + col + ` <> ''
			GROUP BY ` + col + ` ORDER BY COUNT(*) DESC LIMIT $4`,
			[]any{q.Host, since, until, q.Limit}
	case "mysql":
		return `SELECT ` + col + `, COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			AND ` + col + ` IS NOT NULL AND ` + col + ` <> ''
			GROUP BY ` + col + ` ORDER BY COUNT(*) DESC LIMIT ?`,
			[]any{q.Host, since, until, q.Limit}
	default:
		return `SELECT ` + col + `, COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			AND ` + col + ` IS NOT NULL AND ` + col + ` <> ''
			GROUP BY ` + col + ` ORDER BY COUNT(*) DESC LIMIT ?`,
			[]any{q.Host, since, until, q.Limit}
	}
}

func coalesceCol(driver, col, emptyLabel string) string {
	_ = driver
	return `COALESCE(NULLIF(` + col + `, ''), '` + emptyLabel + `')`
}

func timelineSQL(driver string, q Query) (string, []any) {
	since, until := timeArgs(driver, q)
	bucket := timelineBucket(driver, q.Interval)
	switch driver {
	case "postgres":
		return `SELECT ` + bucket + ` AS period, COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = $1 AND created_at >= $2 AND created_at < $3
			GROUP BY period ORDER BY period`,
			[]any{q.Host, since, until}
	case "mysql":
		return `SELECT ` + bucket + ` AS period, COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY period ORDER BY period`,
			[]any{q.Host, since, until}
	default:
		return `SELECT ` + bucket + ` AS period, COUNT(*), COUNT(DISTINCT visitor_hash)
			FROM kiko_hits WHERE host = ? AND created_at >= ? AND created_at < ?
			GROUP BY period ORDER BY period`,
			[]any{q.Host, since, until}
	}
}

func timelineBucket(driver, interval string) string {
	if interval == "hour" {
		switch driver {
		case "postgres":
			return `date_trunc('hour', created_at)`
		case "mysql":
			return `DATE_FORMAT(created_at, '%Y-%m-%dT%H:00:00Z')`
		default:
			return `strftime('%Y-%m-%dT%H:00:00Z', created_at)`
		}
	}
	switch driver {
	case "postgres":
		return `date_trunc('day', created_at)`
	case "mysql":
		return `DATE_FORMAT(created_at, '%Y-%m-%d')`
	default:
		return `strftime('%Y-%m-%d', created_at)`
	}
}
