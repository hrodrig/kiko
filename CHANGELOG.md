# Changelog

## [Unreleased]

## [0.5.0] - 2026-06-23

### Added

- **kiko.js `data-endpoint` attribute** — set a custom endpoint URL via
  `script[data-endpoint]` for first-party tracking and CSP compliance.
- **`window.kiko.track` API** — expose manual `send()` function for modern
  SPAs to trigger pageviews programmatically.
- **`height` field** — capture `screen.height` alongside existing `screen.width`.
- **`language` field** — capture `navigator.language` on each hit.
- **`URLSearchParams` GIF fallback** — replace hand-rolled `encodeURIComponent`
  with `URLSearchParams` for proper encoding of GIF query parameters.

### Changed

- **kiko.js** — reformatted with `'use strict'` and expanded variable names
  for readability.

## [0.4.4] - 2026-06-22

### Changed

- **Ingest endpoints** — `POST /hit` → `POST /api`, `GET /hit.gif` → `GET /api.gif`
  to avoid ad-blocker detection. Updated `kiko.js`, `seed-hits.sh`, docs, and tests.

## [0.4.3] - 2026-06-21

### Added

- **Log format** — each line now includes `- kiko -` before level for multi-service log
  federation (`2026-06-22T01:40:25Z  - kiko - INFO  ...`).
- **Config search path** — `./kiko.yml` → `~/.kiko/kiko.yml` → `/etc/kiko/kiko.yml`;
  startup logs `Using config file: <path>` or `Using config file: none, default settings`.
- **Log level banner** — `Debug level set to: DEBUG` on startup.
- **Database path logged** — shows `database path: ./data/kiko.db` on startup
  (SQLite only).
- **`scripts/seed-hits.sh`** — send sample tracking hits to a local kiko server
  (POST /api and GET /api.gif).

### Changed

- Config search path extended to include `~/.kiko/kiko.yml`.

## [0.4.2] - 2026-06-21

### Fixed

- **kiko.js** — sendBeacon and api.gif now use absolute URL derived from script source.
  Works when kiko.js loaded from different origin than page (cross-domain tracking).

## [0.4.1] - 2026-06-20

### Added

- **`GET /api/v1/version`** — public JSON build metadata (`version`, `commit`, `build_date`, `branch`); same fields as `kiko version` via shared `version.BuildInfo()`.

## [0.4.0] - 2026-06-19

### Added

- **Phase 3 hardening:** referrer spam blocklist, optional datacenter IP blocklist, IP ignore list.
- **Proxy-aware client IP** (`filter.trust_proxy`) via `internal/netutil/` — first public IP from `X-Forwarded-For`.
- **Debug ingest:** `X-Kiko-Dropped: 1` on silent rejects; `X-Debug-Request: true` returns JSON with client IP.
- **Per-host rate limiting** (`rate_limit.host_requests_per_sec`).
- **`kiko.js` SPA support** — auto pageviews on `pushState` / `popstate`; `?hash=1` enables hash routing.
- **`scripts/install.sh`** — release installer (used from [kiko-selfhosted](https://github.com/hrodrig/kiko-selfhosted) docs).
- **E2E:** `testing/docker-compose.yml` + `make e2e` (developer CI only).
- Man page updated for v0.4.0 endpoints and config.

## [0.3.0] - 2026-06-19

### Added

- **`internal/analyzer/`** — read-only stats queries over `kiko_hits` (summary, paths, refs, timeline, visitors, channels, browsers, OS, UTM sources).
- **Stats API** — `GET /api/v1/stats/*` with JSON responses and `Cache-Control: public, max-age=60`.
- **API auth** — optional `api.key` / `KIKO_API_KEY`; `X-API-Key` or `Authorization: Bearer`.
- **API rate limiting** — per-key token bucket (`api.rate_limit`, default 30 req/s).
- **`internal/utm/`** — parse `utm_*` from hit path; strip from stored path; persist on `kiko_hits`.
- **Referrer source labels** — `source` column (Google, Twitter/X, …) from `internal/ref/`.
- **`GET /api/v1/stats/utm`** — breakdown by `utm_source`.

## [0.2.0] - 2026-06-19

### Added

- **`internal/ua/`** — User-Agent parser (browser + OS, no regex).
- **`internal/ref/`** — referrer normalization and channel classifier (direct, organic, social, email, referral).
- Hit enrichment on ingest: `browser`, `os`, `channel` persisted in `kiko_hits`.
- PostgreSQL/MySQL integration tests (skip unless `KIKO_TEST_POSTGRES_*` / `KIKO_TEST_MYSQL_*` env vars set).

## [0.1.0] - TBD

### Added

- Project skeleton: Cobra CLI, Viper config, version injection
- Makefile with quality gates: lint, test, cover-check, security, release-check
- GoReleaser v2 config: 5 OS × 2 arch, Docker, Homebrew, nfpm
- GitHub Actions: CI, Security (govulncheck + grype), CodeQL, Release
- Docker multi-stage build (distroless/static)
- FreeBSD and OpenBSD port skeletons
- SPECIFICATIONS.md and ROADMAP.md
