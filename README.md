# kiko

[![Version](https://img.shields.io/badge/version-0.1.0-blue)](https://github.com/hrodrig/kiko/releases)
[![CI](https://github.com/hrodrig/kiko/actions/workflows/ci.yml/badge.svg)](https://github.com/hrodrig/kiko/actions)
[![codecov](https://codecov.io/gh/hrodrig/kiko/graph/badge.svg)](https://codecov.io/gh/hrodrig/kiko)
[![Go 1.26.4](https://img.shields.io/badge/go-1.26.4-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/hrodrig/kiko)](https://goreportcard.com/report/github.com/hrodrig/kiko)

**Repo:** [github.com/hrodrig/kiko](https://github.com/hrodrig/kiko) · **Releases:** [Releases](https://github.com/hrodrig/kiko/releases)

Privacy-first web analytics collector. No cookies. No Node in production. One static binary.

```bash
git clone https://github.com/hrodrig/kiko
cd kiko
make build
./kiko serve
```

## Table of Contents

- [Why kiko](#why-kiko)
- [How it works](#how-it-works)
- [Install](#install)
- [Usage](#usage)
- [Quality gates](#quality-gates)
- [Related](#related)
- [License](#license)

## Why kiko

kiko is a privacy-first, lightweight web analytics collector:

- **No cookies** — SHA-256 visitor hash with daily salt. No GDPR banner needed.
- **No Node in production** — tracking script is 500 bytes JS. Server is a static Go binary.
- **Passes audits** — govulncheck, grype, gocyclo, cover. Same standard as sibling projects.
- **Single binary** — Go, CGO disabled, distroless. 2.5MB compiled.

## How it works

```
                  kiko.js (500B)
                       │
                  POST /hit
                       ▼
┌─────────────────────────────┐
│        kiko (Go binary)     │
│                             │
│  ┌─────────────────────┐   │
│  │    MemBuffer         │   │
│  │  (chan Hit, cap 4k) │   │
│  └─────────┬───────────┘   │
│            │ flush cada 10s │
│            ▼                │
│  ┌─────────────────────┐   │
│  │   BatchInserter     │   │
│  │   → PostgreSQL      │   │
│  └─────────────────────┘   │
└─────────────────────────────┘
```

Each hit is buffered in memory (buffered channel, non-blocking). Every 10s it flushes to PostgreSQL in batch. Dashboard is a separate repo.

## Install

```bash
# Homebrew (coming soon)
# brew install hrodrig/kiko/kiko

# From source
git clone https://github.com/hrodrig/kiko
cd kiko
make build
sudo cp kiko /usr/local/bin/

# Docker
docker pull ghcr.io/hrodrig/kiko:latest
```

### Platform packages

| OS | Arch | Format |
|----|------|--------|
| Linux | amd64, arm64 | tar.gz, .deb, .rpm, Docker |
| macOS | amd64, arm64 | tar.gz, Homebrew |
| Windows | amd64, arm64 | zip |
| FreeBSD | amd64, arm64 | tar.gz, port |
| OpenBSD | amd64, arm64 | tar.gz, port |

## Usage

```bash
# Start server
kiko serve

# With custom config
kiko serve -c /etc/kiko/kiko.yml

# Tracking: add to your HTML
<script defer src="https://analytics.tudominio.com/kiko.js"></script>
```

### Config (`kiko.yml`)

```yaml
# HTTP listen address
listen: ":8080"

# Public URL for tracking script
public_url: "https://analytics.yourdomain.com"

# Log level: debug, info, warn, error
log_level: info

# PostgreSQL connection
database:
  host: localhost
  port: 5432
  user: kiko
  password: ""
  dbname: kiko
  sslmode: disable

# In-memory hit buffer
buffer:
  flush_interval: 10   # seconds between batch flushes
  capacity: 4096        # max hits in channel

# Rate limiting (per-IP)
rate_limit:
  requests_per_sec: 100
  burst: 200

# Only accept hits from these hosts (empty = accept all)
allowed_hosts:
```

All fields overridable via env vars with `KIKO_` prefix:

| Env | Maps to |
|-----|---------|
| `KIKO_LISTEN` | `listen` |
| `KIKO_PUBLIC_URL` | `public_url` |
| `KIKO_LOG_LEVEL` | `log_level` |
| `KIKO_DATABASE_HOST` | `database.host` |
| `KIKO_DATABASE_PORT` | `database.port` |
| `KIKO_DATABASE_USER` | `database.user` |
| `KIKO_DATABASE_PASSWORD` | `database.password` |
| `KIKO_DATABASE_DBNAME` | `database.dbname` |
| `KIKO_DATABASE_SSLMODE` | `database.sslmode` |
| `KIKO_DATABASE_DSN` | `database.dsn` (overrides all) |
| `KIKO_BUFFER_FLUSH_INTERVAL` | `buffer.flush_interval` |
| `KIKO_BUFFER_CAPACITY` | `buffer.capacity` |
| `KIKO_RATE_LIMIT_REQUESTS_PER_SEC` | `rate_limit.requests_per_sec` |
| `KIKO_RATE_LIMIT_BURST` | `rate_limit.burst` |
| `KIKO_ALLOWED_HOSTS` | `allowed_hosts` (comma-separated) |

### Log levels

| Level | Value | Description |
|-------|-------|-------------|
| trace | 0 | Diagnostic detail, most verbose |
| debug | 1 | Debugging information |
| info | 2 | General operational messages (default) |
| warn | 3 | Non-critical issues |
| error | 4 | Runtime errors |
| fatal | 5 | Critical failure, process exits |
| off | 6 | Nothing logged |

Set via `log_level` in config or `KIKO_LOG_LEVEL` env var.

### API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/kiko.js` | GET | Script de tracking |
| `/hit` | POST | Tracking endpoint (JSON) |
| `/hit.gif` | GET | Fallback pixel tracking |
| `/health` | GET | Health check |

## Quality gates

| Gate | Threshold | Enforced |
|------|-----------|----------|
| gofmt -s | No diff | CI + release |
| go vet | 0 warnings | CI + release |
| gocyclo | ≤ 14 | CI + release |
| govulncheck | 0 vulnerabilities | CI + release |
| grype | 0 high/critical | CI + release |
| go test -cover | ≥ 80% | CI + release |
| CodeQL | Clean | CI |

## Related

- [SPECIFICATIONS.md](SPECIFICATIONS.md) — architecture, schema, API
- [ROADMAP.md](ROADMAP.md) — development phases
- [gghstats](https://github.com/hrodrig/gghstats) — GitHub traffic stats
- [kzero](https://github.com/hrodrig/kzero) — Kubernetes pipeline CLI
- [groot](https://github.com/hrodrig/groot) — Kubernetes log collector
- [pgwd](https://github.com/hrodrig/pgwd) — PostgreSQL connection watchdog

## License

MIT — [LICENSE](LICENSE)
