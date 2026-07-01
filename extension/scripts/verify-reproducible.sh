#!/usr/bin/env bash
# extension/scripts/verify-reproducible.sh
# Dùng: verify-reproducible.sh <commit> <shipped-artifact.zip>
set -euo pipefail
COMMIT="$1"
SHIPPED="$2"
WORK="$(mktemp -d)"

# Ensure we have git configured to find the commit
git worktree add --detach "$WORK" "$COMMIT"
(
  cd "$WORK"
  # Use npm ci if lockfile is present
  npm ci
  node scripts/build-deterministic.mjs
  cd dist
  zip -rX -D -X ../rebuilt.zip .
)

# shasum is available on macOS, sha256sum on Linux
if command -v shasum >/dev/null 2>&1; then
  REBUILT_HASH="$(shasum -a 256 "$WORK/rebuilt.zip" | awk '{print $1}')"
  SHIPPED_HASH="$(shasum -a 256 "$SHIPPED" | awk '{print $1}')"
else
  REBUILT_HASH="$(sha256sum "$WORK/rebuilt.zip" | awk '{print $1}')"
  SHIPPED_HASH="$(sha256sum "$SHIPPED" | awk '{print $1}')"
fi

git worktree remove --force "$WORK"

echo "rebuilt:  $REBUILT_HASH"
echo "shipped:  $SHIPPED_HASH"
if [ "$REBUILT_HASH" != "$SHIPPED_HASH" ]; then
  echo "HASH KHÔNG KHỚP - bản ship khác source công khai" >&2
  exit 1
fi
echo "OK - reproducible: bản ship khớp source @ $COMMIT"
