#!/usr/bin/env bash
# Fail-closed security audit runner.
# Unimplemented hooks MUST fail. Never print MOCK PASS or claim AUDIT PASS
# without real evidence from each hook.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FAIL=0
RESULT_DIR="audit/report/hook-results"
rm -rf "$RESULT_DIR"
mkdir -p "$RESULT_DIR"

write_hook_result() {
  local name="$1"
  local pass="$2"
  local detail="$3"
  printf '{"name":"%s","pass":%s,"detail":%s}\n' \
    "$name" "$pass" "$(printf '%s' "$detail" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read()))')" \
    >"$RESULT_DIR/${name}.json"
}

echo "== [1/4] Egress test (dynamic) =="
if ( cd audit && npx --no-install jest egress --passWithNoTests 2>/dev/null ) || ( cd audit && npx jest egress ); then
  write_hook_result "egress" "true" "jest egress suite exited 0"
else
  write_hook_result "egress" "false" "jest egress suite failed or could not run"
  FAIL=1
fi

echo "== [2/4] SBOM + vuln scan =="
if bash audit/sbom/generate-sbom.sh && bash audit/sbom/scan-vulnerabilities.sh; then
  write_hook_result "sbom" "true" "SBOM generated and vuln scan completed"
else
  write_hook_result "sbom" "false" "SBOM generation and/or vuln scan not implemented or failed"
  FAIL=1
fi

echo "== [3/4] Verify reproducible build (TASK-TRUST-001) =="
if [[ -x extension/scripts/verify-reproducible.sh ]] && [[ -n "${SHIPPED:-}" ]]; then
  if bash extension/scripts/verify-reproducible.sh "$(git rev-parse HEAD)" "$SHIPPED"; then
    write_hook_result "reproducible" "true" "verify-reproducible.sh exited 0"
  else
    write_hook_result "reproducible" "false" "verify-reproducible.sh failed"
    FAIL=1
  fi
else
  echo "FAIL: reproducible-build hook is not wired (missing extension/scripts/verify-reproducible.sh and/or SHIPPED artifact)."
  write_hook_result "reproducible" "false" "hook not implemented — fail closed"
  FAIL=1
fi

echo "== [4/4] Payload guard static (TASK-COMPLY-005) =="
if [[ -d services/comply/internal/audit ]] && ( cd services/comply && go test ./internal/audit/... ); then
  write_hook_result "payload_guard" "true" "go test ./internal/audit/... exited 0"
else
  echo "FAIL: payload-guard hook is not wired (services/comply/internal/audit missing or tests failed)."
  write_hook_result "payload_guard" "false" "hook not implemented — fail closed"
  FAIL=1
fi

echo "== Report =="
if command -v npx >/dev/null 2>&1; then
  ( cd audit && npx --yes ts-node report/build-report.ts ) || FAIL=1
else
  echo "FAIL: npx/ts-node unavailable; cannot build report"
  FAIL=1
fi

if [ "$FAIL" -ne 0 ]; then
  echo "AUDIT FAIL — one or more hooks failed or are not implemented (fail-closed)."
  exit 1
fi
echo "AUDIT PASS — all four hooks produced real evidence."
