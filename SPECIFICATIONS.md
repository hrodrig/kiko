# Kiko — Technical Specifications

> Minimal, privacy-first web analytics collector written in Go.
> No cookies. No Node in production. No bloat.

---

## 1. Philosophy

**kiko** is a privacy-first, lightweight web analytics collector. No cookies, no JavaScript runtime on servers, one static binary.
It follows the same principles as gghstats, kzero, groot, vision:

- **"Boring Hardware"** — predictable, maintainable tools, no magic
- **Single static binary** — pure Go, CGO disabled, distroless
- **Zero Node in production** — no JavaScript runtime on servers
- **Privacy by design** — no cookies, no personal data stored
- **Passes all audits** — govulncheck, grype, gocyclo, cover, go vet, gofmt

---

## 2. Architecture

```
┌─────────────┐     POST /hit        ┌──────────────────────┐
│  Astro/Web  │ ──────────────────►  │   kiko (Go binary)   │
│  (script    │     GET /hit.gif     │                      │
│   500B)     │ ◄──── 43px GIF ──── │  ┌────────────────┐  │
└─────────────┘                      │  │  MemBuffer      │  │
                                     │  │  (chan Hit,     │  │
                                     │  │   cap 4096)     │  │
                                     │  └───────┬────────┘  │
                                     │          │ flush every│
                                     │          │ 10s       │
                                     │          ▼            │
                                     │  ┌────────────────┐  │
                                     │  │  BatchInserter  │  │
                                     │  └───────┬────────┘  │
                                     └──────────┼───────────┘
                                                │
                                                ▼
                                     ┌──────────────────────┐
                                     │    PostgreSQL         │
                                     │                      │
                                     │  hits table          │
                                     │  hit_counts (upsert) │
                                     │  ref_counts (upsert) │
                                     └──────────────────────┘
```

### 2.1 Components

| Component | Description | Language | Status |
|-----------|-------------|----------|--------|
| **kiko** | Collector backend: receives hits, in-memory buffer, batch insert to PostgreSQL | Go | MVP |
| **kiko.js** | Tracking script (~500B) sending hits via sendBeacon or `<img>` fallback | JS | MVP |
| **dashboard** | Separate repo. Consumes kiko API. Go native, SPA, TBD | — | Future |

### 2.2 Hit flow

1. Browser loads `kiko.js` → detects `path`, `referrer`, `title`, `screen.width`
2. Sends `POST /hit` with JSON body via `navigator.sendBeacon()`, fallback to `GET /hit.gif?p=...`
3. **kiko** receives, calculates `visitor_hash = SHA-256(ip + ua + daily_salt)`, appends to memory buffer
4. Every 10s, batch flush: normalizes paths/referrers, upserts stats
5. Always responds with 43-byte transparent GIF (success or error — indistinguishable)

### 2.3 Privacy by design

- **No cookies** — tracking via ephemeral `visitor_hash`
- **Daily salt** — hash changes every day, visitor is "new" the next day
- **IP in memory only** — never persisted to disk, only used for the hash
- **No personal data** — no email, name, or persistent identifier stored
- **GDPR-ready** — no cookie banner needed, no PII stored

---

## 3. Database Schema (PostgreSQL)

```sql
-- Raw hits table (append-only)
CREATE TABLE kiko_hits (
    id           BIGSERIAL PRIMARY KEY,
    host         VARCHAR(255) NOT NULL,       -- gghstats.com, kzero.dev...
    path         TEXT NOT NULL,               -- /blog, /docs/install...
    referrer     TEXT,                        -- Traffic source
    visitor_hash CHAR(64) NOT NULL,           -- SHA-256(ip+ua+salt)
    screen_width SMALLINT,                    -- Screen resolution stats
    title        TEXT,                        -- Page title
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kiko_hits_host_date ON kiko_hits (host, created_at DESC);

-- Normalized paths
CREATE TABLE kiko_paths (
    id      SERIAL PRIMARY KEY,
    host    VARCHAR(255) NOT NULL,
    path    TEXT NOT NULL,
    title   TEXT,
    UNIQUE(host, path)
);

-- Normalized referrers
CREATE TABLE kiko_refs (
    id       SERIAL PRIMARY KEY,
    host     VARCHAR(255) NOT NULL,
    referrer TEXT NOT NULL,
    UNIQUE(host, referrer)
);

-- Hourly aggregated counts (for fast dashboards)
CREATE TABLE kiko_hit_counts (
    host        VARCHAR(255) NOT NULL,
    path_id     INTEGER NOT NULL REFERENCES kiko_paths(id),
    hour        TIMESTAMPTZ NOT NULL,
    total       INTEGER NOT NULL DEFAULT 0,
    uniques     INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (host, path_id, hour)
);

-- Hourly referrer counts
CREATE TABLE kiko_ref_counts (
    host        VARCHAR(255) NOT NULL,
    ref_id      INTEGER NOT NULL REFERENCES kiko_refs(id),
    hour        TIMESTAMPTZ NOT NULL,
    total       INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (host, ref_id, hour)
);
```

**Aggregation strategy:** Batch upsert with `ON CONFLICT DO UPDATE SET total = kiko_hit_counts.total + EXCLUDED.total`.

---

## 4. API

### `POST /hit`
Main tracking endpoint.

**Headers:** `Content-Type: application/json`

**Body:**
```json
{
  "host": "gghstats.com",
  "path": "/blog/my-post",
  "referrer": "https://dev.to/someone",
  "title": "My Post | GGHStats",
  "width": 1920
}
```

**Response:** `200 OK` — `Content-Type: image/gif` — 43 bytes transparent GIF.

### `GET /hit.gif`
Fallback for browsers without sendBeacon.

**Query params:** `p`, `r`, `t`, `w`, `h`

**Response:** Same 43-byte GIF.

### `GET /kiko.js`
Serves the tracking script (immutable, cached 24h).

### `GET /health`
Health check.

**Response:** `{"status": "ok", "version": "0.1.0", "uptime": 12345}`

---

## 5. Tracking script (`kiko.js`)

```javascript
// ~500B, zero dependencies
(function(){
  var d = {
    host: location.hostname,
    path: location.pathname + location.search,
    referrer: document.referrer || '',
    title: document.title,
    width: screen.width
  };
  var b = new Blob([JSON.stringify(d)], {type:'application/json'});
  try {
    if (!navigator.sendBeacon('/hit', b)) throw 0;
  } catch(e) {
    (new Image()).src = '/hit.gif?p=' + encodeURIComponent(d.path) +
      '&r=' + encodeURIComponent(d.referrer) +
      '&t=' + encodeURIComponent(d.title) +
      '&w=' + d.width +
      '&h=' + encodeURIComponent(d.host);
  }
})();
```

---

## 6. Tech Stack

| Layer | Technology | Reason |
|-------|-----------|--------|
| Language | **Go 1.26+** | Static binary, ecosystem alignment |
| CLI | **Cobra** | Standard, same pattern as gghstats/kzero/groot |
| Config | **Viper** | YAML + env vars, same pattern |
| DB | **PostgreSQL** | via `pgx` (pure Go driver, no CGO) |
| Buffer | **Memory + chan** | Zero external deps |
| HTTP | **net/http** stdlib | 3 endpoints, chi is overkill |
| CI | **GitHub Actions** | Same pattern: ci.yml + security.yml + release.yml |
| Release | **GoReleaser** | v2, multi-OS/arch, Homebrew, dockers_v2 |
| Container | **distroless/static** | gcr.io base, nonroot user |
| SBOM | **syft + cosign** | SPDX + CycloneDX, keyless signing |
| Homebrew | **hrodrig/homebrew-kiko** | Separate tap |

---

## 7. Quality Gates

| Gate | Threshold | Where | Blocks |
|------|-----------|-------|--------|
| `gofmt -s` | No diff | `make lint`, CI | ✅ |
| `go vet ./...` | 0 warnings | `make lint`, CI | ✅ |
| `gocyclo -over 14` | ≤14 per function | `make lint`, CI | ✅ |
| `go test -race ./...` | All pass | `make test`, CI | ✅ |
| `go test -coverprofile` | ≥80% | `make cover-check`, CI | ✅ |
| `govulncheck ./...` | "No vulnerabilities" | `make security`, CI | ✅ |
| `grype --fail-on high` | 0 high/critical | `make docker-scan`, CI | ✅ |
| CodeQL | No security alerts | codeql.yml | ✅ |
| VERSION semver | `MAJOR.MINOR.PATCH` | release-check | ✅ |
| Docker running | `docker info` | release-check, docker-* | ✅ |
| HOMEBREW_TAP_TOKEN | set | release.yml | ✅ |

---

## 8. Makefile Targets

```makefile
build         # go build with ldflags (version, commit, date)
install       # go install to $GOBIN
test          # go test -race ./...
cover         # test + coverage.out + report
cover-check   # test + coverage gate ≥80%
lint          # mapstructure pin + gofmt -s + go vet + gocyclo -over 14
lint-fix      # gofmt -s -w
security      # govulncheck + gocyclo + grype
docker-build  # Docker build multi-stage
docker-scan   # docker-build + grype image scan
release-check # lint → test → cover-check → security → docker-scan
release       # release-check + goreleaser (main branch only)
snapshot      # goreleaser snapshot local
port-freebsd-sync   # VERSION → contrib/freebsd/Makefile
port-openbsd-sync   # VERSION → contrib/openbsd/port/Makefile
dist-freebsd  # cross-compile freebsd tarball
dist-openbsd  # cross-compile openbsd tarball
```

---

## 9. Reference Projects

### Direct inspiration

| Project | What to take | What to avoid |
|---------|-------------|---------------|
| **GoatCounter** | In-memory buffer + batch flush, GIF tracking pixel, cookie-less sessions, upsert stats | SQLite bottleneck, jQuery frontend, no horizontal scaling |
| **Pirsch** | Siphash fingerprint (UA+IP+salt+date), worker pipeline channel+batch, zero-allocation UA parser, local channel classification | ClickHouse-only (heavy), no open-source frontend, excessive denormalization |

### Sibling projects (pattern to replicate)

| Project | Key pattern |
|---------|-------------|
| **gghstats** | Makefile with quality gates, GoReleaser v2, distroless Docker, BSD ports sync, release-check gate |
| **kzero** | gocyclo ≤14, coverage 80%, security meta-target, cross-compile 5 OS × 2 arch |
| **groot** | Same structure: cmd/ + internal/, VERSION file, codecov.yml, man page in contrib/ |

---

## 10. Supported Platforms

| OS | Arch | Format |
|----|------|--------|
| Linux | amd64, arm64 | tar.gz, .deb, .rpm, Docker |
| macOS | amd64, arm64 | tar.gz, Homebrew |
| Windows | amd64, arm64 | zip |
| FreeBSD | amd64, arm64 | tar.gz, port |
| OpenBSD | amd64, arm64 | tar.gz, port |
