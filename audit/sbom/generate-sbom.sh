#!/usr/bin/env bash
# Fail-closed SBOM generator. Do not emit a fake CycloneDX document.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
OUT="$ROOT/audit/sbom/bom.json"

echo "SBOM generation is not wired to a real tool (e.g. @cyclonedx/cyclonedx-npm)."
echo "Refusing to write a dummy SBOM. Install and invoke a real generator, then re-run."

cat >"$OUT" <<'EOF'
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "version": 1,
  "metadata": {
    "component": {
      "type": "application",
      "name": "sandeal-extension",
      "version": "NOT_GENERATED"
    },
    "properties": [
      {
        "name": "sandeal:sbom_status",
        "value": "NOT_RUN — generate-sbom.sh fails closed until a real CycloneDX tool is wired"
      }
    ]
  },
  "components": []
}
EOF

exit 1
