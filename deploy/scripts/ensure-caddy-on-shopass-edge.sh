#!/usr/bin/env bash
# Ensure CyberOS Caddy is attached to shopass-edge (shared-host topology).
# Safe to run repeatedly; no downtime when already connected.
set -euo pipefail

EDGE_NET="${SHOPASS_EDGE_NETWORK:-shopass-edge}"
CADDY_CONTAINER="${CYBEROS_CADDY_CONTAINER:-cyberos-p0-caddy-1}"

if ! docker network inspect "${EDGE_NET}" >/dev/null 2>&1; then
  echo "missing network ${EDGE_NET}; create via shopass compose cyberos-shared overlay first" >&2
  exit 1
fi

if ! docker inspect "${CADDY_CONTAINER}" >/dev/null 2>&1; then
  echo "missing container ${CADDY_CONTAINER}" >&2
  exit 1
fi

if docker network inspect "${EDGE_NET}" --format '{{range .Containers}}{{.Name}}{{"\n"}}{{end}}' \
  | grep -qx "${CADDY_CONTAINER}"; then
  echo "ok: ${CADDY_CONTAINER} already on ${EDGE_NET}"
  exit 0
fi

docker network connect "${EDGE_NET}" "${CADDY_CONTAINER}"
echo "connected ${CADDY_CONTAINER} -> ${EDGE_NET}"
