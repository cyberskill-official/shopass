#!/usr/bin/env bash
# Local/staging helper: post a dropping price series into pricesvc ingest
# (no HTTPS_PROXY / live Shopee). Requires compose stack with pricesvc on :8081.
#
# Usage (from repo root, after make up):
#   ./scripts/simulate_price_series.sh
#   PRODUCT_ID=100 ./scripts/simulate_price_series.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE=(docker compose -f deploy/docker-compose.yml)
if [ -f deploy/.env ]; then
  COMPOSE+=(--env-file deploy/.env)
fi

PRODUCT_ID="${PRODUCT_ID:-100}"
PRICE_URL="${PRICE_URL:-http://127.0.0.1:8081}"
LIST_PRICE="${LIST_PRICE:-250000}"

if [ ! -f deploy/.env ]; then
  echo "REFUSING: deploy/.env missing (cp deploy/.env.example deploy/.env)" >&2
  exit 1
fi

TOKEN="$(python3 - <<'PY'
from pathlib import Path
for line in Path("deploy/.env").read_text().splitlines():
    s = line.strip()
    if not s or s.startswith("#") or "=" not in s:
        continue
    k, v = s.split("=", 1)
    if k == "PRICE_INTERNAL_SERVICE_TOKEN":
        print(v.strip())
        break
else:
    raise SystemExit("PRICE_INTERNAL_SERVICE_TOKEN not set in deploy/.env")
PY
)"
if [ -z "$TOKEN" ]; then
  echo "REFUSING: PRICE_INTERNAL_SERVICE_TOKEN empty" >&2
  exit 1
fi

code="$(curl -s -o /dev/null -w '%{http_code}' --connect-timeout 2 "$PRICE_URL/" || true)"
if [ -z "$code" ] || [ "$code" = "000" ]; then
  echo "REFUSING: pricesvc not reachable at $PRICE_URL (make up first)" >&2
  exit 1
fi

echo "== ensure tracked_product($PRODUCT_ID) =="
"${COMPOSE[@]}" exec -T db psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-shopass}" -v ON_ERROR_STOP=1 -c \
  "INSERT INTO tracked_product(id, platform_id, platform_item_id, first_seen)
   VALUES ($PRODUCT_ID, 1, 'sim-555:777', now() - INTERVAL '100 days')
   ON CONFLICT (id) DO NOTHING;"

post_snap() {
  local ts="$1" price="$2"
  local body resp http
  body="$(python3 -c "import json; print(json.dumps({
    'product_id': int('$PRODUCT_ID'),
    'ts': '$ts',
    'price': int('$price'),
    'list_price': int('$LIST_PRICE'),
    'stock': 10,
    'flash_sale': False,
  }))")"
  resp="$(curl -sS -w '\nHTTP:%{http_code}' -X POST "$PRICE_URL/v1/price/snapshots" \
    -H "Content-Type: application/json" \
    -H "X-Service-Token: $TOKEN" \
    -d "$body")"
  http="${resp##*$'\n'HTTP:}"
  body_out="${resp%$'\n'HTTP:*}"
  echo "  ts=$ts price=$price -> HTTP $http $body_out"
  case "$http" in
    200|201) ;;
    *) echo "FAIL: ingest returned HTTP $http" >&2; exit 1 ;;
  esac
}

echo "== ingest dropping series for product $PRODUCT_ID =="
# Distinct ts + changing price so delta-only ingest writes each point.
post_snap "2026-07-01T00:00:00Z" 199000
post_snap "2026-07-02T00:00:00Z" 179000
post_snap "2026-07-03T00:00:00Z" 149000

echo "== price_snapshot($PRODUCT_ID) =="
"${COMPOSE[@]}" exec -T db psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-shopass}" -c \
  "SELECT product_id, ts, price FROM price_snapshot WHERE product_id=$PRODUCT_ID ORDER BY ts;"

echo "OK: simulated price series (chart: /products/$PRODUCT_ID/chart)"
