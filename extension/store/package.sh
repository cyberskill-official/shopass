#!/usr/bin/env bash
# Build a submission zip from a clean production dist/.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
rm -rf dist shopass-extension-store.zip
npm ci --no-audit --no-fund
npm test
npm run typecheck
npm run build
# Optional: strip localhost host permission for store zip
python3 - <<'PY'
import json
from pathlib import Path
p = Path("dist/manifest.json")
m = json.loads(p.read_text())
m["host_permissions"] = [h for h in m["host_permissions"] if "127.0.0.1" not in h]
p.write_text(json.dumps(m, indent=2, ensure_ascii=False) + "\n")
print("stripped localhost host_permission for store zip")
PY
(cd dist && zip -r -X ../shopass-extension-store.zip .)
shasum -a 256 shopass-extension-store.zip
echo "Wrote $ROOT/shopass-extension-store.zip"
