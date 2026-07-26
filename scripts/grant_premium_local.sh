#!/usr/bin/env bash
# Local/staging review helper: grant active premium_basic without checkout/IPN.
# Keeps tracksvc→billsvc gating real. NOT for production.
#
# Usage (from repo root, after make up):
#   USER_ID=4242 ./scripts/grant_premium_local.sh
#   make grant-premium USER_ID=4242
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE=(docker compose -f deploy/docker-compose.yml)
if [ -f deploy/.env ]; then
  COMPOSE+=(--env-file deploy/.env)
fi

USER_ID="${USER_ID:-}"
if [ -z "$USER_ID" ] || ! [[ "$USER_ID" =~ ^[1-9][0-9]*$ ]]; then
  echo "REFUSING: set USER_ID to a positive integer (e.g. USER_ID=4242)" >&2
  exit 1
fi

# Guard: compose db must be up and bound to loopback (dev stack only).
if ! "${COMPOSE[@]}" ps --status running db 2>/dev/null | grep -q db; then
  echo "REFUSING: compose service 'db' is not running (make up first)" >&2
  exit 1
fi
PORT_LINE="$("${COMPOSE[@]}" port db 5432 2>/dev/null || true)"
case "$PORT_LINE" in
  127.0.0.1:*|\[::1\]:*) ;;
  *)
    echo "REFUSING: db is not published on loopback ($PORT_LINE); local/staging review only" >&2
    exit 1
    ;;
esac

echo "== grant premium_basic to user_id=$USER_ID (local/staging) =="
"${COMPOSE[@]}" exec -T db psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-shopass}" -v ON_ERROR_STOP=1 <<SQL
INSERT INTO app_user(id) VALUES ($USER_ID) ON CONFLICT DO NOTHING;
UPDATE subscription SET status = 'canceled'
  WHERE user_id = $USER_ID AND status = 'active';
INSERT INTO subscription (user_id, plan_id, renews_at, status)
SELECT $USER_ID, id, now() + interval '30 days', 'active'
FROM plan_catalog
WHERE tier = 'premium_basic' AND active = true
LIMIT 1;
SQL

SUB="$("${COMPOSE[@]}" exec -T db psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-shopass}" -tA -c \
  "SELECT status || ' plan_id=' || plan_id FROM subscription WHERE user_id=$USER_ID AND status='active' ORDER BY id DESC LIMIT 1;")"
if [ -z "$SUB" ]; then
  echo "FAIL: no active subscription after grant (is plan_catalog seeded?)" >&2
  exit 1
fi
echo "subscription: $SUB"
echo "OK: next POST /v1/alerts bottom_predicted for X-User-Id: $USER_ID should return 201"
