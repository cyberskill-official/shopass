#!/usr/bin/env bash
set -euo pipefail

# This script assumes it is run from the root of the monorepo or services/comply.
# If run from root, we need to adjust the path to cmd/auditscan.
if [ -d "services/comply" ]; then
    cd services/comply
    ROOT_DIR="../.."
    CMD_PATH="./cmd/auditscan"
elif [ -d "cmd/auditscan" ]; then
    ROOT_DIR="../.."
    CMD_PATH="./cmd/auditscan"
else
    echo "Please run from repo root or services/comply"
    exit 1
fi

n=$(go run "$CMD_PATH" --root "$ROOT_DIR" --count)
if [ "$n" -ne 0 ]; then
  echo "no-cleartext gate FAILED: $n vi pham (xem log duoi day)"
  go run "$CMD_PATH" --root "$ROOT_DIR"
  exit 1
fi
echo "no-cleartext gate PASSED"
