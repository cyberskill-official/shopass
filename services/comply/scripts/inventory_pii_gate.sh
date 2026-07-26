#!/usr/bin/env bash
# Drift check: personal-data fields listed in DATA-INVENTORY.md must appear in
# SQL migrations (or adapters). New migration lines creating email/phone/password
# columns without an inventory mention (or -- reviewed:) fail the gate.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
INV="$ROOT/docs/compliance/DATA-INVENTORY.md"

python3 - "$ROOT" "$INV" <<'PY'
import re, sys
from pathlib import Path

root, inv_path = map(Path, sys.argv[1:3])
inv = inv_path.read_text(encoding="utf-8")
# fields like `app_user.email` or `payment.order_ref`
fields = set(re.findall(r"`([a-z_][a-z0-9_]*\.[a-z_][a-z0-9_]*)`", inv, re.I))
if len(fields) < 5:
    print("inventory gate FAILED: too few fields parsed from DATA-INVENTORY.md")
    sys.exit(1)

sql_files = list((root / "db" / "migrations").glob("*.sql"))
sql_files += list((root / "services").glob("**/migrations/*.sql"))
corpus = "\n".join(p.read_text(encoding="utf-8", errors="ignore") for p in sql_files)
# also include adapters that anonymize
for p in (root / "services" / "comply" / "internal" / "adapters").glob("*.go"):
    corpus += "\n" + p.read_text(encoding="utf-8", errors="ignore")

missing = sorted(f for f in fields if f.split(".", 1)[1] not in corpus and f not in corpus)
# require column name presence at least
missing = [f for f in fields if f.split(".", 1)[1] not in corpus]
# soften: ignore wildcard rows like consent_policy.*
missing = [f for f in missing if not f.endswith(".*") and "*" not in f]

# Column definitions that look like personal/credential storage.
# Table names alone (e.g. CREATE TABLE refresh_token) are not findings —
# platform session tokens on server are covered by auditscan separately.
col_re = re.compile(
    r"(?i)\b(email|phone|password_hash|password|passwd|secret_plain|access_token|refresh_token)\s+(TEXT|VARCHAR|CITEXT|BYTEA)",
)
bad_new = []
for p in sql_files:
    text = p.read_text(encoding="utf-8", errors="ignore")
    for i, line in enumerate(text.splitlines(), 1):
        m = col_re.search(line)
        if not m:
            continue
        if "-- reviewed:" in line or "# reviewed:" in line or "-- audit:allow" in line:
            continue
        col = m.group(1).lower()
        # password_hash is expected (argon2id); inventory tracks account email/phone.
        if col == "password_hash":
            continue
        if any(f.endswith("." + col) or f.endswith("." + col.replace("_hash", "")) for f in fields):
            continue
        bad_new.append(f"{p.relative_to(root)}:{i}: {line.strip()[:120]}")

failures = []
# Only fail missing inventory fields that look personal (heuristic)
personalish = [f for f in missing if any(x in f for x in ("email", "phone", "name", "ip", "user_agent", "order_ref", "transaction"))]
if personalish:
    failures.append("inventory fields not found in SQL/adapters: " + ", ".join(personalish[:20]))
if bad_new:
    failures.append("sensitive migration lines without inventory/-- reviewed:")
    failures.extend("  " + x for x in bad_new[:20])

if failures:
    print("inventory PII gate FAILED:")
    for f in failures:
        print(" -", f)
    sys.exit(1)
print(f"inventory PII gate PASSED ({len(fields)} inventory fields scanned)")
PY
