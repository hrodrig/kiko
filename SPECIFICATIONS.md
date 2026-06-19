# Kiko — Especificaciones Técnicas

> Recolector de analítica web minimalista, escrito en Go.
> Privacidad primero. Sin cookies. Sin Node/NPM en producción. Sin bloat.

---

## 1. Filosofía

**kiko** nace de la frustración con la sobreingeniería de Google Analytics, Matomo, y similares.
Sigue los mismos principios de gghstats, kzero, groot, vision:

- **"Boring Hardware"** — herramientas predecibles, mantenibles, sin magia
- **Un solo binario estático** — Go puro, CGO disabled, distroless
- **Zero Node en producción** — sin runtime de JavaScript en los servidores
- **Privacidad por diseño** — sin cookies, sin datos personales almacenados
- **Pasa todas las auditorías** — govulncheck, grype, gocyclo, cover, go vet, gofmt

---

## 2. Arquitectura

```
┌─────────────┐     POST /hit        ┌──────────────────────┐
│  Astro/Web  │ ──────────────────►  │   kiko (Go binary)   │
│  (script    │     GET /hit.gif     │                      │
│   3.5KB)    │ ◄──── 43px GIF ──── │  ┌────────────────┐  │
└─────────────┘                      │  │  MemBuffer      │  │
                                     │  │  ([]Hit, mutex) │  │
                                     │  └───────┬────────┘  │
                                     │          │ flush cada │
                                     │          │ 10s        │
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

### 2.1 Componentes

| Componente | Descripción | Lenguaje | Estado |
|-----------|-------------|----------|--------|
| **kiko** | Backend recolector: recibe hits, buffer en memoria, batch insert a PostgreSQL | Go | MVP |
| **kiko.js** | Script de tracking (~3.5KB) que envía hits vía sendBeacon o <img> | JS | MVP |
| **dashboard** | Repositorio separado. Consume API de kiko. Go nativo, SPA, o lo que se decida luego | — | Futuro |

### 2.2 Flujo de un hit

1. Browser carga `kiko.js` → detecta `path`, `referrer`, `title`, `screen.width`
2. Envía `POST /hit` con JSON body vía `navigator.sendBeacon()`, fallback a `GET /hit.gif?p=...`
3. **kiko** recibe, calcula `visitor_hash = SHA-256(ip + ua + salt_diario)`, lo pone en buffer de memoria
4. Cada 10s, batch flush: normaliza paths/referrers, upsertea stats
5. Siempre responde con GIF transparente 43-byte (éxito o error — indistinguible)

### 2.3 Privacy por diseño

- **Sin cookies** — tracking vía `visitor_hash` efímero
- **Salt diario** — el hash cambia cada día, el visitante es "nuevo" al día siguiente
- **IP solo en memoria** — nunca se persiste a disco, solo para el hash
- **Sin datos personales** — no se almacena email, nombre, ni identificador persistente
- **GDPR-ready** — no necesita banner de cookies, no almacena PII

---

## 3. Esquema de Base de Datos (PostgreSQL)

```sql
-- Tabla de hits crudos (append-only)
CREATE TABLE kiko_hits (
    id          BIGSERIAL PRIMARY KEY,
    host        VARCHAR(255) NOT NULL,       -- gghstats.com, kzero.dev...
    path        TEXT NOT NULL,               -- /blog, /docs/install...
    referrer    TEXT,                        -- Fuente de tráfico
    visitor_hash CHAR(64) NOT NULL,          -- SHA-256(ip+ua+salt)
    screen_width SMALLINT,                   -- Para stats de resoluciones
    title       TEXT,                        -- Page title del hit
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kiko_hits_host_date ON kiko_hits (host, created_at DESC);

-- Tabla de paths normalizados
CREATE TABLE kiko_paths (
    id      SERIAL PRIMARY KEY,
    host    VARCHAR(255) NOT NULL,
    path    TEXT NOT NULL,
    title   TEXT,
    UNIQUE(host, path)
);

-- Tabla de referrers normalizados
CREATE TABLE kiko_refs (
    id      SERIAL PRIMARY KEY,
    host    VARCHAR(255) NOT NULL,
    referrer TEXT NOT NULL,
    UNIQUE(host, referrer)
);

-- Conteos agregados por hora (para dashboards rápidos)
CREATE TABLE kiko_hit_counts (
    host        VARCHAR(255) NOT NULL,
    path_id     INTEGER NOT NULL REFERENCES kiko_paths(id),
    hour        TIMESTAMPTZ NOT NULL,
    total       INTEGER NOT NULL DEFAULT 0,
    uniques     INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (host, path_id, hour)
);

-- Conteos por referrer por hora
CREATE TABLE kiko_ref_counts (
    host        VARCHAR(255) NOT NULL,
    ref_id      INTEGER NOT NULL REFERENCES kiko_refs(id),
    hour        TIMESTAMPTZ NOT NULL,
    total       INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (host, ref_id, hour)
);
```

**Strategy de agregación:** Batch upsert con `ON CONFLICT DO UPDATE SET total = kiko_hit_counts.total + EXCLUDED.total`.

---

## 4. API

### `POST /hit`
Endpoint principal de tracking.

**Headers:** `Content-Type: application/json`

**Body:**
```json
{
  "host": "gghstats.com",
  "path": "/blog/mi-post",
  "referrer": "https://dev.to/alguien",
  "title": "Mi Post | GGHStats",
  "width": 1920
}
```

**Response:** `200 OK` — `Content-Type: image/gif` — 43 bytes de GIF transparente.

### `GET /hit.gif`
Fallback para navegadores sin sendBeacon.

**Query params:** `p`, `r`, `t`, `w`

**Response:** Mismo GIF 43-byte.

### `GET /health`
Health check.

**Response:** `{"status": "ok", "version": "0.1.0", "uptime": 12345}`

---

## 5. Script de Tracking (`kiko.js`)

```javascript
// ~3.5KB, sin dependencias externas
(function() {
  var d = {
    host: location.hostname,
    path: location.pathname,
    referrer: document.referrer || '',
    title: document.title,
    width: screen.width
  };
  var b = new Blob([JSON.stringify(d)], {type: 'application/json'});
  navigator.sendBeacon && navigator.sendBeacon('/hit', b) ||
    new Image().src = '/hit.gif?p=' + encodeURIComponent(d.path) +
      '&r=' + encodeURIComponent(d.referrer) +
      '&t=' + encodeURIComponent(d.title) +
      '&w=' + d.width;
})();
```

---

## 6. Stack Tecnológico

| Capa | Tecnología | Razón |
|------|-----------|-------|
| Lenguaje | **Go 1.26+** | Binario estático, eco con el ecosistema |
| CLI | **Cobra** | Estándar, mismo pattern que gghstats/kzero/groot |
| Config | **Viper** | YAML + env vars, mismo pattern |
| DB | **PostgreSQL** | vía `pgx` (driver puro Go, sin CGO) |
| Buffer | **Memoria + RWMutex** | Sin dependencias externas para buffer |
| HTTP | **net/http** estándar | Sin frameworks, chi/mux es overkill para 3 endpoints |
| Templates | — | Dashboard será otro repo |
| CI | **GitHub Actions** | Mismo pattern: ci.yml + security.yml + release.yml |
| Release | **GoReleaser** | v2, multi-OS/arch, Homebrew, dockers_v2 |
| Contenedor | **distroless/static** | gcr.io base, nonroot user |
| SBOM | **syft + cosign** | SPDX + CycloneDX, keyless signing |
| Homebrew | **hrodrig/homebrew-kiko** | Tap separado |

---

## 7. Quality Gates

| Gate | Threshold | Dónde | Bloquea |
|------|-----------|-------|---------|
| `gofmt -s` | Sin diff | `make lint`, CI | ✅ |
| `go vet ./...` | 0 warnings | `make lint`, CI | ✅ |
| `gocyclo -over 14` | ≤14 por función | `make lint`, CI | ✅ |
| `go test -race ./...` | All pass | `make test`, CI | ✅ |
| `go test -coverprofile` | ≥80% | `make cover-check`, CI | ✅ |
| `govulncheck ./...` | "No vulnerabilities" | `make security`, CI | ✅ |
| `grype --fail-on high` | 0 high/critical | `make docker-scan`, CI | ✅ |
| CodeQL | Sin security alerts | codeql.yml | ✅ |
| VERSION semver | `MAJOR.MINOR.PATCH` | release-check | ✅ |
| Docker running | `docker info` | release-check, docker-* | ✅ |
| HOMEBREW_TAP_TOKEN | set | release.yml | ✅ |

---

## 8. Makefile Targets

```makefile
build         # go build con ldflags (version, commit, date)
install       # go install a $GOBIN
test          # go test -race ./...
cover         # test + coverage.out + report
cover-check   # test + coverage gate ≥80%
lint          # gofmt -s + go vet + gocyclo -over 14
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

## 9. Proyectos de Referencia

### Inspiración directa

| Proyecto | Qué tomar | Qué evitar |
|----------|-----------|------------|
| **GoatCounter** | Buffer en memoria + batch flush, GIF tracking pixel, sesiones sin cookies, upsert de stats | SQLite bottleneck, jQuery frontend, sin escalado horizontal |
| **Pirsch** | Siphash fingerprint (UA+IP+salt+date), worker pipeline channel+batch, zero-allocation UA parser, channel classification local | ClickHouse-only (pesado), sin frontend open-source, denormalización excesiva |

### Proyectos propios (patrón a replicar)

| Proyecto | Patrón clave |
|----------|-------------|
| **gghstats** | Makefile con quality gates, GoReleaser v2, distroless Docker, BSD ports sync, release-check gate |
| **kzero** | gocyclo ≤14, coverage 80%, security metatarget, cross-compile 5 OS × 2 arch |
| **groot** | Misma estructura: cmd/ + internal/, VERSION file, codecov.yml, man page en contrib/ |

---

## 10. Plataformas Soportadas

| OS | Arch | Formato |
|----|------|---------|
| Linux | amd64, arm64 | tar.gz, .deb, .rpm, Docker |
| macOS | amd64, arm64 | tar.gz, Homebrew |
| Windows | amd64, arm64 | zip |
| FreeBSD | amd64, arm64 | tar.gz, port |
| OpenBSD | amd64, arm64 | tar.gz, port |
