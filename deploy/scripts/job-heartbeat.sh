#!/usr/bin/env sh
# Push a job success heartbeat to Pushgateway (R17).
# Usage: job-heartbeat.sh <job_name>
# Optional env: PUSHGATEWAY_URL (default http://127.0.0.1:9091)
set -eu
JOB_NAME="${1:?job_name required (scrape|forecast)}"
PUSHGATEWAY_URL="${PUSHGATEWAY_URL:-http://127.0.0.1:9091}"
NOW="$(date +%s)"
BODY="shopass_job_last_success_unixtime{job_name=\"${JOB_NAME}\"} ${NOW}
"
printf '%s' "$BODY" | curl -fsS --data-binary @- \
  "${PUSHGATEWAY_URL}/metrics/job/shopass_jobs/instance/${JOB_NAME}"
echo "heartbeat ok job=${JOB_NAME} ts=${NOW}"
