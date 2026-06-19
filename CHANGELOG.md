# Changelog

## [Unreleased]

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
