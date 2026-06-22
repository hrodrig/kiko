#!/usr/bin/env sh
set -e

ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$ROOT/testing"

docker compose -f docker-compose.yml down -v 2>/dev/null || true
docker compose -f docker-compose.yml up -d --build --wait

BASE=http://127.0.0.1:18080
UA="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/120.0.0.0"

curl -fsS "$BASE/api/v1/healthz" | grep -q '"status":"ok"'
curl -fsS "$BASE/api/v1/readyz" | grep -q '"status":"ok"'

  curl -fsS -X POST "$BASE/api" \
  -H "Content-Type: application/json" \
  -H "User-Agent: $UA" \
  -d '{"host":"e2e.test","path":"/hello","referrer":"https://example.com"}' \
  -o /dev/null

sleep 12

curl -fsS "$BASE/api/v1/stats/summary?host=e2e.test&since=2020-01-01" \
  -H "X-API-Key: e2e-secret" | grep -q '"hits"'

echo "E2E OK"
docker compose -f docker-compose.yml down -v
