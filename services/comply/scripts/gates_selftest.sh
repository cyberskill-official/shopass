#!/usr/bin/env bash
# Red/green self-test for R33 gates (not run in CI by default).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT"

bash services/comply/scripts/no_cleartext_gate.sh
bash services/comply/scripts/consent_coverage_gate.sh
bash services/comply/scripts/inventory_pii_gate.sh
bash services/comply/scripts/dpia_review_gate.sh

# Seed overdue DPIA → must fail without waiver.
if DPIA_TODAY=2099-01-01 bash services/comply/scripts/dpia_review_gate.sh; then
  echo "selftest FAILED: expected overdue DPIA failure"
  exit 1
fi
echo "selftest: overdue DPIA correctly failed"

echo "R33 gates_selftest PASSED"
