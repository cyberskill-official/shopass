#!/usr/bin/env bash
# Fail when any DPIA row in docs/compliance/DPIA.md has review_due < today
# without a waiver: reviewed:... cell.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
FILE="$ROOT/docs/compliance/DPIA.md"
TODAY="${DPIA_TODAY:-$(date -u +%Y-%m-%d)}"

if [[ ! -f "$FILE" ]]; then
  echo "DPIA gate FAILED: missing $FILE"
  exit 1
fi

python3 - "$FILE" "$TODAY" <<'PY'
import re, sys
from datetime import date

path, today_s = sys.argv[1], sys.argv[2]
today = date.fromisoformat(today_s)
text = open(path, encoding="utf-8").read()
# table rows: | DPIA-00N | ... | filing | review_due | owner | waiver |
row_re = re.compile(
    r"^\|\s*(DPIA-\d+)\s*\|\s*([^|]+)\|\s*(\d{4}-\d{2}-\d{2})\s*\|\s*(\d{4}-\d{2}-\d{2})\s*\|\s*([^|]*)\|\s*([^|]*)\|",
    re.M,
)
failures = []
found = 0
for m in row_re.finditer(text):
    found += 1
    did, activity, _filed, due_s, _owner, waiver = (x.strip() for x in m.groups())
    due = date.fromisoformat(due_s)
    waived = waiver.lower().startswith("reviewed:")
    if due < today and not waived:
        failures.append(f"{did} ({activity.strip()}) review_due={due_s} < today={today_s}")

if found == 0:
    print("DPIA gate FAILED: no DPIA rows parsed")
    sys.exit(1)
if failures:
    print("DPIA gate FAILED: overdue reviews without waiver:")
    for f in failures:
        print(" -", f)
    sys.exit(1)
print(f"DPIA gate PASSED ({found} rows; today={today_s})")
PY
