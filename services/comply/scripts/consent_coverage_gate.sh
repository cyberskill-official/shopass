#!/usr/bin/env bash
# Validate docs/compliance/CONSENT-SURFACES.md against purpose allowlist and
# enforcement status rules (api/client/deferred+waiver).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
FILE="$ROOT/docs/compliance/CONSENT-SURFACES.md"
TYPES="$ROOT/services/comply/internal/consent/types.go"

python3 - "$FILE" "$TYPES" "$ROOT" <<'PY'
import re, sys
from pathlib import Path

surfaces_path, types_path, root = map(Path, sys.argv[1:4])
text = surfaces_path.read_text(encoding="utf-8")
types = types_path.read_text(encoding="utf-8")

purposes = set(re.findall(r'Purpose\w+\s+Purpose\s*=\s*"([^"]+)"', types))
if not purposes:
    print("consent gate FAILED: no Purpose constants in types.go")
    sys.exit(1)

row_re = re.compile(
    r"^\|\s*(CS-\d+)\s*\|\s*([^|]+)\|\s*`([^`]+)`\s*\|\s*(api|client|deferred)\s*\|\s*`([^`]+)`\s*\|\s*([^|]*)\|",
    re.M,
)
failures = []
found = 0
for m in row_re.finditer(text):
    found += 1
    sid, surface, purpose, status, anchor, waiver = (x.strip() for x in m.groups())
    if purpose not in purposes:
        failures.append(f"{sid}: unknown purpose `{purpose}`")
    anchor_path = root / anchor
    if not anchor_path.exists():
        failures.append(f"{sid}: missing code_anchor {anchor}")
    if status == "deferred" and not waiver.lower().startswith("reviewed:"):
        failures.append(f"{sid}: deferred without waiver reviewed:<id>")
    if status == "api":
        # consent API itself records grants; other api surfaces must mention IsAllowed nearby later.
        blob = ""
        if anchor_path.is_file():
            blob = anchor_path.read_text(encoding="utf-8", errors="ignore")
        elif anchor_path.is_dir():
            for p in anchor_path.rglob("*.go"):
                blob += p.read_text(encoding="utf-8", errors="ignore")
        if "IsAllowed" not in blob and "Grant(" not in blob and "Withdraw(" not in blob:
            failures.append(f"{sid}: api surface lacks IsAllowed/Grant/Withdraw in {anchor}")

if found == 0:
    print("consent gate FAILED: no surface rows parsed")
    sys.exit(1)
if failures:
    print("consent gate FAILED:")
    for f in failures:
        print(" -", f)
    sys.exit(1)
print(f"consent gate PASSED ({found} surfaces; purposes={sorted(purposes)})")
PY
