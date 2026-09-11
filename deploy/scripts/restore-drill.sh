#!/usr/bin/env bash
# Restore drill into an isolated scratch Postgres (never overwrites production).
#
# Usage:
#   deploy/scripts/restore-drill.sh /var/backups/shopass/shopass_YYYY-MM-DD_SHA.sql.gz
#
# Optional env:
#   DRILL_CONTAINER   default shopass-restore-drill
#   DRILL_IMAGE       default timescale/timescaledb:2.28.2-pg16
#   DRILL_KEEP=1      leave container running for inspection
set -euo pipefail

DUMP="${1:?path to .sql.gz dump required}"
if [[ ! -f "${DUMP}" ]]; then
  echo "dump not found: ${DUMP}" >&2
  exit 1
fi

DRILL_CONTAINER="${DRILL_CONTAINER:-shopass-restore-drill}"
DRILL_IMAGE="${DRILL_IMAGE:-timescale/timescaledb:2.28.2-pg16}"
DRILL_KEEP="${DRILL_KEEP:-0}"
DRILL_USER="${DRILL_USER:-shopass}"
DRILL_DB="${DRILL_DB:-shopass}"
DRILL_PASS="${DRILL_PASS:-drill-only-not-prod}"

START_TS="$(date +%s)"
echo "restore-drill start dump=${DUMP} container=${DRILL_CONTAINER}"

docker rm -f "${DRILL_CONTAINER}" >/dev/null 2>&1 || true

# Cap memory so a second Timescale on a small VPS does not OOM the host.
docker run -d --name "${DRILL_CONTAINER}" \
  --memory=1024m \
  -e POSTGRES_USER="${DRILL_USER}" \
  -e POSTGRES_PASSWORD="${DRILL_PASS}" \
  -e POSTGRES_DB="${DRILL_DB}" \
  "${DRILL_IMAGE}" \
  postgres -c shared_buffers=64MB -c max_connections=20 >/dev/null

cleanup() {
  if [[ "${DRILL_KEEP}" != "1" ]]; then
    docker rm -f "${DRILL_CONTAINER}" >/dev/null 2>&1 || true
  else
    echo "DRILL_KEEP=1 — left ${DRILL_CONTAINER} running (password=${DRILL_PASS})"
  fi
}
trap cleanup EXIT

# Timescale image runs timescaledb-tune and restarts once during first boot.
# Wait for a stable ready window so restore is not interrupted mid-init.
echo "waiting for scratch postgres (stable ready after Timescale tune restart)..."
ready_streak=0
for _ in $(seq 1 180); do
  if docker exec "${DRILL_CONTAINER}" pg_isready -U "${DRILL_USER}" -d "${DRILL_DB}" >/dev/null 2>&1 \
    && docker exec "${DRILL_CONTAINER}" \
      psql -U "${DRILL_USER}" -d "${DRILL_DB}" -c "SELECT 1" >/dev/null 2>&1; then
    ready_streak=$((ready_streak + 1))
    if [[ "${ready_streak}" -ge 8 ]]; then
      break
    fi
  else
    ready_streak=0
  fi
  sleep 1
done
if [[ "${ready_streak}" -lt 8 ]]; then
  echo "scratch postgres did not become stably ready" >&2
  docker logs "${DRILL_CONTAINER}" 2>&1 | tail -n 50 || true
  exit 1
fi

LOG="/tmp/shopass-restore-drill.$$.log"

docker exec -i "${DRILL_CONTAINER}" \
  psql -U "${DRILL_USER}" -d "${DRILL_DB}" -v ON_ERROR_STOP=0 \
  -c "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;" \
  -c "SELECT timescaledb_pre_restore();" >/dev/null 2>&1 || true

echo "restoring dump (this may take minutes)..."
set +e
gunzip -c "${DUMP}" | docker exec -i "${DRILL_CONTAINER}" \
  psql -U "${DRILL_USER}" -d "${DRILL_DB}" -v ON_ERROR_STOP=0 >"${LOG}" 2>&1
RESTORE_RC=$?
set -e

docker exec -i "${DRILL_CONTAINER}" \
  psql -U "${DRILL_USER}" -d "${DRILL_DB}" -v ON_ERROR_STOP=0 \
  -c "SELECT timescaledb_post_restore();" >>"${LOG}" 2>&1 || true

ROW_COUNTS="$(docker exec "${DRILL_CONTAINER}" psql -U "${DRILL_USER}" -d "${DRILL_DB}" -Atc \
  "SELECT 'app_user_tbl=' || (SELECT count(*)::text FROM information_schema.tables WHERE table_schema='public' AND table_name='app_user') ||
          ' app_user_rows=' || COALESCE((SELECT count(*)::text FROM app_user), '0') ||
          ' tracked_product_rows=' || COALESCE((SELECT count(*)::text FROM tracked_product), '0') ||
          ' price_snapshot_rel=' || COALESCE((SELECT to_regclass('public.price_snapshot')::text), 'null');" 2>>"${LOG}" \
  || echo "row_count_query_failed")"

END_TS="$(date +%s)"
ELAPSED="$((END_TS - START_TS))"
DUMP_BYTES="$(wc -c < "${DUMP}" | tr -d ' ')"

echo "restore-drill complete elapsed_sec=${ELAPSED} dump_bytes=${DUMP_BYTES} restore_psql_rc=${RESTORE_RC} ${ROW_COUNTS}"
echo "transcript: ${LOG}"
echo "--- restore log tail ---"
tail -n 40 "${LOG}" || true

if [[ "${ROW_COUNTS}" == *row_count_query_failed* ]] || [[ "${ROW_COUNTS}" != *app_user_tbl=1* ]]; then
  echo "restore-drill FAILED: application tables missing" >&2
  exit 1
fi
