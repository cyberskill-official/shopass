#!/usr/bin/env bash
# Nightly Postgres logical dump for Shopass (R12).
#
# Writes: ${BACKUP_LOCAL_DIR}/shopass_YYYY-MM-DD_SCHEMA.gitsha.sql.gz
# Optional off-host copy when BACKUP_S3_URI is set (rclone or aws cli).
# On success, pushes shopass_job_last_success_unixtime{job_name="backup"} to Pushgateway.
#
# Required: Docker Compose production stack with healthy `db`.
# Env (typically /etc/shopass/runtime.env):
#   BACKUP_LOCAL_DIR     default /var/backups/shopass
#   BACKUP_RETENTION_DAYS default 30
#   BACKUP_S3_URI        e.g. s3:shopass-backups/pg  (rclone remote:path) or s3://bucket/prefix
#   BACKUP_UPLOAD_TOOL   rclone | aws | none  (default: rclone if BACKUP_S3_URI set, else none)
#   BACKUP_SCHEMA_GIT_SHA optional override; else git -C /srv/shopass rev-parse --short HEAD
#   COMPOSE_PROJECT / paths via same conventions as systemd scrape units
set -euo pipefail

ROOT="${SHOPASS_ROOT:-/srv/shopass}"
COMPOSE_FILE="${COMPOSE_FILE:-${ROOT}/deploy/docker-compose.production.yml}"
ENV_FILE="${ENV_FILE:-/etc/shopass/runtime.env}"
BACKUP_LOCAL_DIR="${BACKUP_LOCAL_DIR:-/var/backups/shopass}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
POSTGRES_USER="${POSTGRES_USER:-shopass}"
POSTGRES_DB="${POSTGRES_DB:-shopass}"
PUSHGATEWAY_URL="${PUSHGATEWAY_URL:-http://127.0.0.1:9091}"

if [[ ! -f "${COMPOSE_FILE}" ]]; then
  echo "missing compose file: ${COMPOSE_FILE}" >&2
  exit 1
fi

mkdir -p "${BACKUP_LOCAL_DIR}"
chmod 0700 "${BACKUP_LOCAL_DIR}" 2>/dev/null || true

DATE_UTC="$(date -u +%F)"
if [[ -n "${BACKUP_SCHEMA_GIT_SHA:-}" ]]; then
  SCHEMA_SHA="${BACKUP_SCHEMA_GIT_SHA}"
elif [[ -d "${ROOT}/.git" ]]; then
  SCHEMA_SHA="$(git -C "${ROOT}" rev-parse --short HEAD 2>/dev/null || echo unknown)"
else
  SCHEMA_SHA="unknown"
fi

OUT_NAME="shopass_${DATE_UTC}_${SCHEMA_SHA}.sql.gz"
OUT_PATH="${BACKUP_LOCAL_DIR}/${OUT_NAME}"
TMP_PATH="${OUT_PATH}.partial"

compose() {
  if [[ -f "${ENV_FILE}" ]]; then
    docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" "$@"
  else
    docker compose -f "${COMPOSE_FILE}" "$@"
  fi
}

echo "backup start date=${DATE_UTC} schema_sha=${SCHEMA_SHA} out=${OUT_PATH}"

compose exec -T db \
  pg_dump -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" --no-owner --no-acl \
  | gzip -c > "${TMP_PATH}"

mv -f "${TMP_PATH}" "${OUT_PATH}"
BYTES="$(wc -c < "${OUT_PATH}" | tr -d ' ')"
echo "backup local ok bytes=${BYTES} path=${OUT_PATH}"

# Retention (local)
find "${BACKUP_LOCAL_DIR}" -maxdepth 1 -type f -name 'shopass_*.sql.gz' -mtime "+${BACKUP_RETENTION_DAYS}" -print -delete \
  || true

UPLOAD_TOOL="${BACKUP_UPLOAD_TOOL:-}"
if [[ -z "${UPLOAD_TOOL}" ]]; then
  if [[ -n "${BACKUP_S3_URI:-}" ]]; then
    if command -v rclone >/dev/null 2>&1; then
      UPLOAD_TOOL=rclone
    elif command -v aws >/dev/null 2>&1; then
      UPLOAD_TOOL=aws
    else
      UPLOAD_TOOL=none
      echo "BACKUP_S3_URI set but neither rclone nor aws found; local-only" >&2
    fi
  else
    UPLOAD_TOOL=none
  fi
fi

case "${UPLOAD_TOOL}" in
  none|"")
    echo "upload skipped (no BACKUP_S3_URI or tool=none)"
    ;;
  rclone)
    if [[ -z "${BACKUP_S3_URI:-}" ]]; then
      echo "rclone upload requires BACKUP_S3_URI" >&2
      exit 1
    fi
    rclone copyto "${OUT_PATH}" "${BACKUP_S3_URI%/}/${OUT_NAME}"
    echo "upload rclone ok dest=${BACKUP_S3_URI%/}/${OUT_NAME}"
    # Remote retention best-effort (rclone delete older than N days)
    rclone delete --min-age "${BACKUP_RETENTION_DAYS}d" "${BACKUP_S3_URI}" || true
    ;;
  aws)
    if [[ -z "${BACKUP_S3_URI:-}" ]]; then
      echo "aws upload requires BACKUP_S3_URI=s3://bucket/prefix" >&2
      exit 1
    fi
    aws s3 cp "${OUT_PATH}" "${BACKUP_S3_URI%/}/${OUT_NAME}"
    echo "upload aws ok dest=${BACKUP_S3_URI%/}/${OUT_NAME}"
    ;;
  *)
    echo "unknown BACKUP_UPLOAD_TOOL=${UPLOAD_TOOL}" >&2
    exit 1
    ;;
esac

# Heartbeat (same metric family as R17 jobs)
if command -v curl >/dev/null 2>&1; then
  NOW="$(date +%s)"
  BODY="shopass_job_last_success_unixtime{job_name=\"backup\"} ${NOW}
"
  if printf '%s' "${BODY}" | curl -fsS --data-binary @- \
    "${PUSHGATEWAY_URL}/metrics/job/shopass_jobs/instance/backup"; then
    echo "heartbeat ok job=backup ts=${NOW}"
  else
    echo "heartbeat failed (Pushgateway unreachable); dump still succeeded" >&2
  fi
fi

echo "backup complete ${OUT_NAME}"
