#!/usr/bin/env bash
# seed-hits.sh — Send sample tracking hits to a local kiko server.
#
# Usage:
#   ./scripts/seed-hits.sh [base_url]
#
# Default base_url is http://localhost:8080 — override via first argument.
set -euo pipefail

BASE="${1:-http://localhost:8080}"
HIT_URL="${BASE}/hit"
GIF_URL="${BASE}/hit.gif"

red=$(printf '\033[31m')
grn=$(printf '\033[32m')
cya=$(printf '\033[36m')
rst=$(printf '\033[0m')

ok()   { printf "  %s✓%s %s\n" "$grn" "$rst" "$1"; }
fail() { printf "  %s✗%s %s\n" "$red" "$rst" "$1"; }
info() { printf "%s·%s %s\n" "$cya" "$rst" "$1"; }

die() { echo "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "missing: $1"; }
need curl

# ------------------------------------------------------------------
# POST /hit — JSON body
# ------------------------------------------------------------------
info "POST /hit — JSON"

send_json() {
  local host="$1" path="$2" referrer="$3" title="$4" width="$5"
  local desc="${host}${path}"
  local code
  code=$(curl -s -o /dev/null -w '%{http_code}' \
    -H 'Content-Type: application/json' \
    -H 'User-Agent: Mozilla/5.0 (seed-hits; X11; Linux x86_64)' \
    -d "$(cat <<BODY
{
  "host": "${host}",
  "path": "${path}",
  "referrer": "${referrer}",
  "title": "${title}",
  "width": ${width}
}
BODY
)" \
    "${HIT_URL}")
  if [ "$code" = "200" ] || [ "$code" = "204" ]; then
    ok "$desc → $code"
  else
    fail "$desc → $code"
  fi
}

send_json "example.com"       "/"               ""                     "Homepage"           1920
send_json "example.com"       "/blog"           "https://example.com/" "Blog — Example"     1440
send_json "example.com"       "/about"          "https://example.com/" "About Us"           1920
send_json "example.com"       "/blog/post-1"    "https://example.com/blog" "Post 1 — Example" 1366
send_json "example.com"       "/contact"        "https://example.com/about" "Contact"          1024

send_json "myshop.example"    "/products"       ""                     "Our Products"       1920
send_json "myshop.example"    "/products/abc"   "https://myshop.example/products" "Product ABC" 1536
send_json "myshop.example"    "/cart"           "https://myshop.example/products/abc" "Cart" 1920
send_json "myshop.example"    "/checkout"       "https://myshop.example/cart" "Checkout"        1920
send_json "myshop.example"    "/"               "https://google.com/search?q=myshop" "Home — MyShop" 1440

# ------------------------------------------------------------------
# GET /hit.gif — pixel fallback
# ------------------------------------------------------------------
info "GET /hit.gif — pixel"

send_gif() {
  local host="$1" path="$2" referrer="$3" title="$4" width="$5"
  local desc="${host}${path} (gif)"
  local code
  code=$(curl -s -o /dev/null -w '%{http_code}' \
    -H 'User-Agent: Mozilla/5.0 (seed-hits; iPhone; CPU iPhone OS 17_0)' \
    "${GIF_URL}?h=$(printf '%s' "$host" | jq -sRr @uri)&p=$(printf '%s' "$path" | jq -sRr @uri)&r=$(printf '%s' "$referrer" | jq -sRr @uri)&t=$(printf '%s' "$title" | jq -sRr @uri)&w=${width}")
  if [ "$code" = "200" ] || [ "$code" = "204" ]; then
    ok "$desc → $code"
  else
    fail "$desc → $code"
  fi
}

send_gif "docs.example.com"   "/getting-started" ""                     "Quick Start"        1920
send_gif "docs.example.com"   "/api"             "https://docs.example.com/getting-started" "API Reference" 1920
send_gif "docs.example.com"   "/faq"             "https://docs.example.com/api" "FAQ" 1440

# ------------------------------------------------------------------
# Summary
# ------------------------------------------------------------------
echo
info "Done. Hits sent to ${BASE}"
info "Check stats at ${BASE}/api/v1/stats or the kiko logs."
