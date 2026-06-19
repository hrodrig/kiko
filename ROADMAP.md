# Kiko — Roadmap

> `v0.1.0` → `v1.0.0`
> Minimal privacy-first web analytics collector in Go.

---

## Vision

**kiko** is a privacy-first web analytics **collector**. No cookies, no Node, one static binary.
**kui** (*kiko* + *ui*) is the planned analytics UI — separate repo, not in scope here.

Writes to PostgreSQL, passes all audits.

---

## Phases

### ✅ Phase 0: Foundation (`v0.1.0`) — Done

**Goal:** Project skeleton compiles, quality gates pass, structure ready.

- [x] `go mod init github.com/hrodrig/kiko`
- [x] `cmd/kiko/main.go` with Cobra CLI (`serve` subcommand)
- [x] `internal/version/` with ldflags injection
- [x] `internal/config/` with Viper (YAML + env)
- [x] `internal/cli/` with Cobra wiring
- [x] `VERSION` file (`0.1.0`), `LICENSE` (MIT)
- [x] `Makefile` with targets: `build`, `test`, `lint`, `cover`, `security`, `release-check`
- [x] `.goreleaser.yaml` v2: linux/darwin/windows/freebsd/openbsd × amd64/arm64
- [x] GitHub Actions: `ci.yml`, `security.yml`, `codeql.yml`, `release.yml`
- [x] `Dockerfile` + `Dockerfile.release` (distroless/static)
- [x] `contrib/freebsd/` + `contrib/openbsd/port/` (port skeletons)
- [x] `contrib/man/man1/kiko.1` (man page skeleton)
- [x] `contrib/systemd/kiko.service`
- [x] `codecov.yml`
- [x] `README.md` + `CHANGELOG.md` + `CONTRIBUTING.md` + `SECURITY.md`
- [x] `SPECIFICATIONS.md` + `ROADMAP.md`
- [x] `.gitignore` (deny-all + allowlist)
- [x] `homebrew-kiko` tap repo created
- [x] Quality gates passing: gocyclo ≤14, coverage 84.1% ≥80%, govulncheck clean, grype clean, lint clean

**Bonus (started Phase 1 items):**
- [x] `internal/hit/` with Hit type + Buffer (channel-based)
- [x] `internal/server/` with HTTP handlers (kiko.js, hit, hit.gif, health)
- [x] `internal/log/` with leveled logger (Trace → Off)
- [x] `internal/validate/` with host allowlist, bot detection, prefetch filtering
- [x] `internal/store/` with Store interface + NopStore
- [x] Tests: 84.1% total coverage (validate 96.7%, store 100%, server 85.9%)
- [x] Pushed to `github.com/hrodrig/kiko`

**Success criteria:** `make release-check` passes cleanly locally and in CI.

---

### 🟡 Phase 1: Core Engine (`v0.2.0`) — Done

**Goal:** kiko receives hits, buffers in memory, persists and aggregates to SQLite/PostgreSQL/MySQL.

- [x] `internal/hit/` — `Hit` type, `VisitorHash`, mutex buffer
- [x] `internal/store/` — `Store` interface + SQLite/PostgreSQL/MySQL via `Open()`
- [x] `internal/server/` — HTTP handlers + healthz/readyz
- [x] `internal/log/` — leveled logger
- [x] `internal/validate/` — host allowlist, bot detection, prefetch filtering
- [x] `internal/visitor/` — daily SHA-256 visitor hash
- [x] SQL migration: `kiko_hits`, `kiko_paths`, `kiko_refs`, `kiko_hit_counts`, `kiko_ref_counts`
- [x] Aggregation pipeline in `SaveHits()` — hourly upserts + unique dedup
- [x] Per-IP rate limiting (`golang.org/x/time/rate`, pattern from gghstats)
- [x] `kiko.js` — tracking script (~500B)
- [x] `internal/ua/` — minimal parser (browser name + OS, no regex)
- [x] `internal/ref/` — referrer parser + basic channel classifier
- [x] Tests: integration with PostgreSQL/MySQL (skip unless `KIKO_TEST_*` env set)

**Success criteria:** `make run` starts kiko, `curl -X POST localhost:8080/hit` returns GIF and hit appears in DB with hourly aggregates.

---

### 🟡 Phase 2: Query API (`v0.3.0`) — 2-3 sprints

**Goal:** kiko exposes REST API for aggregated stats. **kui** (UI, separate repo) consumes that API.

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

- [ ] Multi-level rate limiting (by IP, by host) — per-IP done in Phase 1
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

- [ ] **[kui](https://github.com/hrodrig/kui)** — analytics UI (*kiko* + *ui*). Separate repo. Consumes kiko query API. Design TBD (Go templates, SPA, whatever)
- [ ] SQLite backend (for deployments without PostgreSQL) — done in Phase 1 (default)
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
