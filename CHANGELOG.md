# Changelog

## [Unreleased]

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
