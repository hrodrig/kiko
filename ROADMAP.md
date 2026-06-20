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

### ✅ Phase 2: Query API (`v0.3.0`) — Done

**Goal:** kiko exposes REST API for aggregated stats. **kui** (UI, separate repo) consumes that API.

- [x] `internal/analyzer/` — aggregation queries to PostgreSQL:
  - `GET /api/v1/stats/summary?host=&since=&until=` — hits, uniques, top path
  - `GET /api/v1/stats/paths?host=&since=&until=&limit=` — top paths with counts
  - `GET /api/v1/stats/refs?host=&since=&until=&limit=` — top referrers
  - `GET /api/v1/stats/timeline?host=&since=&until=&interval=` — time series by day/hour
  - `GET /api/v1/stats/visitors?host=&since=&until=` — unique visitors
  - `GET /api/v1/stats/channels?host=&since=&until=` — breakdown by channel (direct, organic, social, …)
  - `GET /api/v1/stats/browsers?host=&since=&until=&limit=` — breakdown by browser
  - `GET /api/v1/stats/os?host=&since=&until=&limit=` — breakdown by OS
  - `GET /api/v1/stats/utm?host=&since=&until=&limit=` — breakdown by utm_source
- [x] JSON output with cache headers (CDN-friendly)
- [x] Rate limiting by API key
- [x] Tests: unit + integration with SQLite (Postgres/MySQL via existing `KIKO_TEST_*` hooks)

**Ingest enrichment (Plausible-inspired, collector side):**

- [x] **UTM capture** — parse `utm_source`, `utm_medium`, `utm_campaign`, `utm_term`, `utm_content` from hit path/query; persist on `kiko_hits`; expose in stats API breakdowns
- [x] **Referrer source labels** — display names for major search/social referrers; store `source` alongside normalized `referrer` URL

**Success criteria:** `curl localhost:8080/api/v1/stats/summary?host=gghstats.com` returns JSON with real data; UTM params on `?utm_source=newsletter&utm_medium=email` appear in DB and API.

---

### ✅ Phase 3: Hardening & Distribution (`v0.4.0`) — Done

**Goal:** Ready for production in MicroK8s.

- [x] Multi-level rate limiting (by IP, by host) — per-IP done in Phase 1
- [x] Bot filtering: prefetch headers, known bots, UA validation
- [x] **Referrer spam blocklist** — static domain list; drop hits with known spam referrers
- [x] **Optional datacenter IP blocklist** — static CIDR file; configurable on/off (lightweight CE-style filter)
- [x] IP ignore list (configurable, exclude own IPs)
- [x] **Proxy-aware client IP** — first valid IP from `X-Forwarded-For`; reject private/loopback when proxy misconfigured
- [x] **Debug ingest headers** — `X-Kiko-Dropped: 1` on silent reject (bot/prefetch/spam); optional `X-Debug-Request: true` returns IP used for visitor hash (Traefik/Ingress troubleshooting)
- [x] **`kiko.js` SPA support** — auto pageviews on `history.pushState` / `popstate`; optional `hashchange` for hash-based routing; keep script ~500B–1KB
- [x] `configs/kiko.yml.sample` — documented config
- [x] `scripts/install.sh` — release installer; documented in **kiko-selfhosted**
- [x] Production-hardened release (tag v0.4.0)
- [x] Homebrew cask published
- [x] .deb + .rpm packages via nfpm
- [x] Docker multi-arch published to GHCR
- [x] Man page complete
- [x] E2E test with docker-compose (kiko + PostgreSQL + curl assertions) — dev CI in this repo
- [x] MicroK8s / Compose / Helm deployment docs — **[kiko-selfhosted](https://github.com/hrodrig/kiko-selfhosted)**
- [x] CHANGELOG v0.4.0

**Success criteria:** `brew install hrodrig/kiko/kiko && kiko serve` works;
SPA navigation fires multiple pageviews; Docker image on GHCR with grype 0 vulnerabilities.

---

### 🔵 Phase 4: Maturity (`v0.5.0` → `v1.0.0`)

**Goal:** Feature-complete for real production use.

- [ ] CSV export (raw hits by date range)
- [ ] Data retention policy (auto-purge old hits, keep aggregated stats)
- [ ] **Custom events** — extend `POST /hit` (or `POST /event`) with `name` + optional `props` (JSON, capped); aggregate by event name for **kui**
- [ ] **Auto-capture (optional, `kiko.js` flags)** — outbound link clicks, file downloads, form submissions (Plausible-style, off by default)
- [ ] Geography: optional GeoIP via MaxMind GeoLite2 (country-level)
- [ ] Channel classification: add **ai** referrer bucket (ChatGPT, Perplexity, …) — direct/organic/social/email/referral done in Phase 1
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

**Next:** Phase 4 — Maturity (CSV export, custom events, GeoIP).

---

## Reference: [Plausible Analytics](https://github.com/plausible/analytics)

Ideas adopted into this roadmap (patterns and contracts only — no AGPL code):

| Idea | Phase | Notes |
|------|-------|-------|
| Stats API breakdowns (channels, browsers, OS) | 2 | Shape for **kui** |
| UTM params from page URL | 2 | Campaign tracking |
| referer-parser source labels | 2 | Better than hardcoded host lists |
| SPA `pushState` / hash routing in tracker | 3 | ~500B–1KB budget |
| Proxy IP + debug/dropped headers | 3 | Ingress/Treafik ops |
| Referrer spam + optional DC IP blocklist | 3 | CE-style bot hygiene |
| Custom events + auto-capture | 4 | Optional, off by default |

Tracker reference (MIT): [@plausible-analytics/tracker](https://www.npmjs.com/package/@plausible-analytics/tracker) — read-only for SPA/event patterns; **kiko.js** stays self-hosted and minimal.
