# Kiko — Roadmap

> `v0.1.0` → `v1.0.0`
> Minimal privacy-first web analytics collector in Go.

---

## Vision

**kiko** is a Google Analytics replacement that fits in a <15MB Go binary.
No cookies, no Node, no external dependencies. Runs in a MicroK8s Pod,
writes to PostgreSQL, passes all audits.

---

## Phases

### 🟢 Phase 0: Foundation (`v0.1.0`) — 1-2 sprints

**Goal:** Project skeleton compiles, quality gates pass, structure ready.

- [ ] `go mod init github.com/hrodrig/kiko`
- [ ] `cmd/kiko/main.go` with Cobra CLI (`serve` subcommand)
- [ ] `internal/version/` with ldflags injection
- [ ] `internal/config/` with Viper (YAML + env)
- [ ] `VERSION` file (`0.1.0`), `LICENSE` (MIT)
- [ ] `Makefile` with targets: `build`, `test`, `lint`, `cover`, `security`, `release-check`
- [ ] `.goreleaser.yaml` v2: linux/darwin/windows/freebsd/openbsd × amd64/arm64
- [ ] GitHub Actions: `ci.yml`, `security.yml`, `codeql.yml`, `release.yml`
- [ ] `Dockerfile` + `Dockerfile.release` (distroless/static)
- [ ] `contrib/freebsd/` + `contrib/openbsd/port/` (port skeletons)
- [ ] `contrib/man/man1/kiko.1` (man page skeleton)
- [ ] `codecov.yml`
- [ ] `README.md` + `CHANGELOG.md` + `CONTRIBUTING.md` + `SECURITY.md`
- [ ] `.gitignore` (deny-all + allowlist)
- [ ] Quality gates passing: gocyclo ≤14, coverage ≥80%, govulncheck clean, grype clean

**Success criteria:** `make release-check` passes cleanly locally and in CI.

---

### 🟡 Phase 1: Core Engine (`v0.2.0`) — 2-3 sprints

**Goal:** kiko receives hits, buffers in memory, persists to PostgreSQL.

- [ ] `internal/hit/` — `Hit` type, `VisitorHash`
- [ ] `internal/buffer/` — `MemBuffer` with channel, flush every 10s
- [ ] `internal/store/` — `Store` interface + PostgreSQL implementation via `pgx`
- [ ] `internal/server/` — HTTP handlers:
  - `POST /hit` — JSON tracking endpoint
  - `GET /hit.gif` — pixel fallback
  - `GET /kiko.js` — tracking script
  - `GET /health` — health check
- [ ] `internal/log/` — leveled logger (Microsoft LogLevel semantics)
- [ ] `internal/validate/` — host allowlist, bot detection, prefetch filtering
- [ ] `internal/visitor/` — `generateVisitorHash(ip, ua) string` with `crypto/sha256`
- [ ] `internal/ua/` — minimal parser (browser name + OS, no regex)
- [ ] `internal/ref/` — referrer parser + basic channel classifier
- [ ] SQL migration: `CREATE TABLE kiko_hits`, `kiko_paths`, `kiko_refs`
- [ ] `internal/stats/` — aggregation pipeline:
  - `updateHitCounts()` — hourly upsert
  - `updateRefCounts()` — hourly upsert
- [ ] `kiko.js` — tracking script (~500B, vanilla JS)
- [ ] Tests: unit + integration with PostgreSQL

**Success criteria:** `make run` starts kiko, `curl -X POST localhost:8080/hit` returns GIF and hit appears in PostgreSQL.

---

### 🟡 Phase 2: Query API (`v0.3.0`) — 2-3 sprints

**Goal:** kiko exposes REST API for aggregated stats. Dashboard is a separate repo.

- [ ] `internal/analyzer/` — aggregation queries to PostgreSQL:
  - `GET /api/v1/stats/summary?host=&since=&until=` — hits, uniques, top path
  - `GET /api/v1/stats/paths?host=&since=&until=&limit=` — top paths with counts
  - `GET /api/v1/stats/refs?host=&since=&until=&limit=` — top referrers
  - `GET /api/v1/stats/timeline?host=&since=&until=&interval=` — time series by day/hour
  - `GET /api/v1/stats/visitors?host=&since=&until=` — unique visitors
- [ ] JSON output with cache headers (CDN-friendly)
- [ ] Rate limiting by API key
- [ ] Tests: unit + integration with PostgreSQL

**Success criteria:** `curl localhost:8090/api/v1/stats/summary?host=gghstats.com` returns JSON with real data.

---

### 🟠 Phase 3: Hardening & Distribution (`v0.4.0`) — 1-2 sprints

**Goal:** Ready for production in MicroK8s.

- [ ] Multi-level rate limiting (by IP, by host)
- [ ] Bot filtering: prefetch headers, known bots, UA validation
- [ ] IP ignore list (configurable, exclude own IPs)
- [ ] `configs/kiko.yml.sample` — documented config
- [ ] `scripts/install.sh` — `curl | sh` installer
- [ ] First real release to GitHub (tag v0.4.0)
- [ ] Homebrew cask published
- [ ] .deb + .rpm packages via nfpm
- [ ] Docker multi-arch published to GHCR
- [ ] Man page complete
- [ ] E2E test with docker-compose (kiko + PostgreSQL + curl assertions)
- [ ] MicroK8s deployment docs:
  - Deployment + Service + ConfigMap
  - Ingress with Traefik + auth middleware
- [ ] CHANGELOG v0.4.0

**Success criteria:** `brew install hrodrig/kiko/kiko && kiko serve` works,
Docker image on GHCR with grype 0 vulnerabilities.

---

### 🔵 Phase 4: Maturity (`v0.5.0` → `v1.0.0`)

**Goal:** Feature-complete for real production use.

- [ ] CSV export (raw hits by date range)
- [ ] Data retention policy (auto-purge old hits, keep aggregated stats)
- [ ] Custom events (optional: `POST /event` with metadata)
- [ ] Geography: optional GeoIP via MaxMind GeoLite2 (country-level)
- [ ] Channel classification (organic, social, direct, referral, email, ai)
- [ ] Prometheus metrics endpoint (`/metrics`) for kiko self-monitoring
- [ ] Load testing: ensure 10k hits/s throughput on a single Pod
- [ ] Complete docs: SPECIFICATIONS.md + ROADMAP.md + API docs

**Success criteria:** kiko replaces GoatCounter in production for gghstats.com and kzero.dev.

---

### 🟣 Post-v1.0 (Future)

- [ ] **Dashboard** — separate repo. Consumes kiko API. Design TBD (Go templates, SPA, whatever)
- [ ] SQLite backend (for deployments without PostgreSQL)
- [ ] ClickHouse backend (for high throughput)
- [ ] Real-time webhook (stream hits to external systems)
- [ ] Multi-tenant (single kiko for N sites)
- [ ] Plugins: custom filters, notifications, S3 export

---

## Prioritization

| | P0 | P1 | P2 | P3 | P4 |
|---|---|---|---|---|---|
| **Impact** | Foundation | Functional | API | Production | Maturity |
| **Effort** | Low | Medium | Medium | Medium | High |
| **Risk** | Low | Low | Medium | Medium | Low |
| **Dependencies** | Go | PostgreSQL | P1 | P1+P2 | P1+P2+P3 |

**Next:** Phase 0. Done.
