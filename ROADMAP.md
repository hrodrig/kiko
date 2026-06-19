# Kiko — Roadmap

> `v0.1.0` → `v1.0.0`
> Recolector de analítica web minimalista en Go.

---

## Visión

**kiko** es un reemplazo de Google Analytics que cabe en un binario Go de <15MB.
Sin cookies, sin Node, sin dependencias externas. Corre en un Pod de MicroK8s,
escribe a PostgreSQL, y pasa todas las auditorías.

---

## Fases

### 🟢 Fase 0: Foundation (`v0.1.0`) — 1-2 sprints

**Objetivo:** Esqueleto del proyecto compila, pasa quality gates, estructura lista.

- [ ] `go mod init github.com/hrodrig/kiko`
- [ ] `cmd/kiko/main.go` con Cobra CLI (`serve` subcommand)
- [ ] `internal/version/` con ldflags injection
- [ ] `internal/config/` con Viper (YAML + env)
- [ ] `VERSION` file (`0.1.0`), `LICENSE` (MIT)
- [ ] `Makefile` con targets: `build`, `test`, `lint`, `cover`, `security`, `release-check`
- [ ] `.goreleaser.yaml` v2: linux/darwin/windows/freebsd/openbsd × amd64/arm64
- [ ] GitHub Actions: `ci.yml`, `security.yml`, `codeql.yml`, `release.yml`
- [ ] `Dockerfile` + `Dockerfile.release` (distroless/static)
- [ ] `contrib/freebsd/` + `contrib/openbsd/port/` (skeleton ports)
- [ ] `contrib/man/man1/kiko.1` (man page skeleton)
- [ ] `codecov.yml`
- [ ] `README.md` + `CHANGELOG.md` + `CONTRIBUTING.md` + `SECURITY.md`
- [ ] `.cursor/rules/` (release-tests)
- [ ] Quality gates pasando: gocyclo ≤14, coverage ≥80%, govulncheck clean, grype clean

**Criterio de éxito:** `make release-check` pasa limpio en local y CI.

---

### 🟡 Fase 1: Core Engine (`v0.2.0`) — 2-3 sprints

**Objetivo:** kiko recibe hits, los bufferiza en memoria, y los persiste a PostgreSQL.

- [ ] `internal/hit/` — tipos `Hit`, `VisitorHash`
- [ ] `internal/buffer/` — `MemBuffer` con `RWMutex`, flush cada 10s
- [ ] `internal/store/` — interfaz `Store` + implementation PostgreSQL con `pgx`
- [ ] `internal/server/` — HTTP handlers:
  - `POST /hit` — endpoint JSON principal
  - `GET /hit.gif` — fallback pixel
  - `GET /health` — health check
- [ ] `internal/server/middleware/` — rate limiter (token bucket, simple)
- [ ] `internal/visitor/` — `generateVisitorHash(ip, ua) string` con `crypto/sha256`
- [ ] `internal/ua/` — parser mínimo (solo browser name + OS, sin regex)
- [ ] `internal/ref/` — referrer parser + channel classifier básico
- [ ] Migración SQL: `CREATE TABLE kiko_hits`, `kiko_paths`, `kiko_refs`
- [ ] `internal/stats/` — pipeline de agregación:
  - `updateHitCounts()` — upsert por hora
  - `updateRefCounts()` — upsert por hora
- [ ] `kiko.js` — script de tracking (~3.5KB, vanilla JS)
- [ ] Tests: unitarios para buffer, store, visitor hash, server handlers
- [ ] Integración: `docker-compose.yml` con PostgreSQL + kiko

**Criterio de éxito:** `make run` levanta kiko, `curl -X POST localhost:8080/hit -d '{"host":"test.dev","path":"/"}'` devuelve GIF y el hit aparece en PostgreSQL.

---

### 🟡 Fase 2: API de Consultas (`v0.3.0`) — 2-3 sprints

**Objetivo:** kiko expone API REST para consultar stats agregadas. El dashboard será otro repo.

- [ ] `internal/analyzer/` — queries de agregación a PostgreSQL:
  - `GET /api/v1/stats/summary?host=&since=&until=` — hits, uniques, top path
  - `GET /api/v1/stats/paths?host=&since=&until=&limit=` — top paths con counts
  - `GET /api/v1/stats/refs?host=&since=&until=&limit=` — top referrers
  - `GET /api/v1/stats/timeline?host=&since=&until=&interval=` — serie temporal por día/hora
  - `GET /api/v1/stats/visitors?host=&since=&until=` — unique visitors
- [ ] Output JSON, con cabeceras de caché (CDN-friendly)
- [ ] Rate limiting por API key
- [ ] Tests: unitarios + integración con PostgreSQL de test

**Criterio de éxito:** `curl localhost:8090/api/v1/stats/summary?host=gghstats.com` devuelve JSON con datos reales.

---

### 🟠 Fase 3: Blindaje y Distribución (`v0.4.0`) — 1-2 sprints

**Objetivo:** Ready for production en MicroK8s.

- [ ] Rate limiting multi-nivel (por IP, por host)
- [ ] Bot filtering: prefetch headers, known bots, UA validation
- [ ] IP ignore list (configurable, exclude自家 IPs)
- [ ] `configs/kiko.yml.sample` — config completa documentada
- [ ] `scripts/install.sh` — `curl | sh` installer
- [ ] Release real a GitHub (tag v0.4.0)
- [ ] Homebrew cask publicado
- [ ] .deb + .rpm packages via nfpm
- [ ] Docker multi-arch publicado a GHCR
- [ ] Man page completa
- [ ] E2E test con docker-compose (kiko + PostgreSQL + curl assertions)
- [ ] Documentación de despliegue en MicroK8s:
  - Deployment + Service + ConfigMap
  - Ingress con Traefik + auth middleware
- [ ] CHANGELOG v0.4.0

**Criterio de éxito:** `brew install hrodrig/kiko/kiko && kiko serve` funciona,
Docker image en GHCR con grype 0 vulnerabilities.

---

### 🔵 Fase 4: Madurez (`v0.5.0` → `v1.0.0`)

**Objetivo:** Feature-complete para usar en producción real.

- [ ] Export CSV (raw hits por rango de fecha)
- [ ] Data retention policy (auto-purge hits viejos, mantener stats agregadas)
- [ ] Eventos personalizados (opcional: `POST /event` con metadata)
- [ ] Geografía: GeoIP opcional via MaxMind GeoLite2 (country-level)
- [ ] Channel classification (organic, social, direct, referral, email, ai)
- [ ] Dashboard: real-time updates vía polling ligero
- [ ] Dashboard: filtros por rango de fecha, host, path
- [ ] Dashboard: modo embed (iframe, sin auth, read-only)
- [ ] Prometheus metrics endpoint (`/metrics`) para monitoreo del propio kiko
- [ ] Load testing: asegurar throughput de 10k hits/s en un solo Pod
- [ ] Documentación completa: SPECIFICATIONS.md + ROADMAP.md + API docs

**Criterio de éxito:** kiko reemplaza GoatCounter en producción para gghstats.com y kzero.dev.

---

### 🟣 Post-v1.0 (Futuro lejano)

- [ ] **Dashboard** — repo separado. Consume API de kiko. Diseño TBD (Go templates, SPA, lo que sea)
- [ ] SQLite backend (para despliegues sin PostgreSQL)
- [ ] ClickHouse backend (para alto throughput)
- [ ] Webhook de hits en tiempo real (streaming a sistemas externos)
- [ ] Multi-tenant (un solo kiko para N sitios)
- [ ] Plugins: filtros personalizados, notificaciones, export a S3

---

## Priorización

| | F0 | F1 | F2 | F3 | F4 |
|---|---|---|---|---|---|
| **Impacto** | Fundación | Funcional | Visual | Producción | Madurez |
| **Esfuerzo** | Bajo | Medio | Medio | Medio | Alto |
| **Riesgo** | Bajo | Bajo | Medio | Medio | Bajo |
| **Dependencias** | Go | PostgreSQL | F1 | F1+F2 | F1+F2+F3 |

**Next:** Fase 0. Arrancar repo con estructura y quality gates.
