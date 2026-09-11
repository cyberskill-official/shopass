#!/usr/bin/env bash
# Dry-run: verify backup upload tooling without writing remote objects (R12 prep).
# Does not invent credentials. Exits non-zero if BACKUP_S3_URI is set but the
# chosen tool is missing or misconfigured.
set -euo pipefail

ENV_FILE="${ENV_FILE:-/etc/shopass/runtime.env}"
if [[ -f "${ENV_FILE}" ]]; then
  # shellcheck disable=SC1090
  set -a
  # Only load known BACKUP_* keys to avoid executing arbitrary env.
  while IFS= read -r line; do
    case "$line" in
      BACKUP_*=*) export "$line" ;;
    esac
  done < <(grep -E '^BACKUP_[A-Z0-9_]+=' "${ENV_FILE}" || true)
  set +a
fi

echo "backup-upload-dry-run"
echo "  BACKUP_S3_URI=${BACKUP_S3_URI:-<unset>}"
echo "  BACKUP_UPLOAD_TOOL=${BACKUP_UPLOAD_TOOL:-<auto>}"
echo "  BACKUP_LOCAL_DIR=${BACKUP_LOCAL_DIR:-/var/backups/shopass}"

if [[ -z "${BACKUP_S3_URI:-}" ]]; then
  echo "OK: no BACKUP_S3_URI — local-only mode (R12 still needs_stephen for off-host)."
  echo "Stephen ask: provision bucket + keys; set BACKUP_S3_URI and BACKUP_UPLOAD_TOOL=rclone|aws in ${ENV_FILE} (mode 0600)."
  exit 0
fi

TOOL="${BACKUP_UPLOAD_TOOL:-}"
if [[ -z "${TOOL}" ]]; then
  if command -v rclone >/dev/null 2>&1; then
    TOOL=rclone
  elif command -v aws >/dev/null 2>&1; then
    TOOL=aws
  else
    echo "FAIL: BACKUP_S3_URI set but neither rclone nor aws in PATH" >&2
    exit 1
  fi
fi

case "${TOOL}" in
  rclone)
    command -v rclone >/dev/null 2>&1 || { echo "FAIL: rclone missing" >&2; exit 1; }
    echo "checking rclone remote (lsd / dry)…"
    # List parent without mutating. rclone returns non-zero if remote missing.
    rclone lsd "${BACKUP_S3_URI}" --max-depth 1 >/dev/null
    echo "OK: rclone can list ${BACKUP_S3_URI}"
    ;;
  aws)
    command -v aws >/dev/null 2>&1 || { echo "FAIL: aws cli missing" >&2; exit 1; }
    echo "checking aws s3 ls (no write)…"
    aws s3 ls "${BACKUP_S3_URI%/}/" >/dev/null
    echo "OK: aws can list ${BACKUP_S3_URI}"
    ;;
  none)
    echo "OK: tool=none (upload disabled despite URI — unexpected for DR)"
    ;;
  *)
    echo "FAIL: unknown BACKUP_UPLOAD_TOOL=${TOOL}" >&2
    exit 1
    ;;
esac

echo "dry-run complete — no objects uploaded"
