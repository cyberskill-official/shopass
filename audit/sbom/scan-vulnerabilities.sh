#!/usr/bin/env bash
# Fail-closed vulnerability scan. Do not claim "0 high CVEs" without scanning.
set -euo pipefail

echo "Vulnerability scan is not wired to a real tool (e.g. npm audit / trivy / osv-scanner)."
echo "Refusing to report a clean CVE result. Wire a real scanner, then re-run."
echo "CVE scan status: NOT_RUN (fail-closed)"
exit 1
