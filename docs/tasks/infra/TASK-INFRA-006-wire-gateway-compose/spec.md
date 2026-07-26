---
id: TASK-INFRA-006
title: "Wire gateway into local compose; unpublish service ports; kill client X-User-Id trust"
module: INFRA
priority: MUST
status: implementing
class: improvement
verify: T
phase: harden
milestone: Wave 1 - unblock
slice: R1
owner: Stephen Cheng (Founder)
created: 2026-07-26
related_frs: [TASK-INFRA-001, TASK-AUTH-002]
depends_on: [TASK-INFRA-001, TASK-AUTH-002]
blocks: []
source_decisions:
  - "Converted from improvement R1 (docs/tasks/improvement/TASKS-A-blockers.md)"
  - "Local compose must match production trust model: only gateway (+ web + optional db) publishes host ports"
  - "Gateway strips client X-User-* then sets verified identity from JWT"
language: "Docker Compose + Go gateway (existing)"
service: shopass/deploy/
new_files:
  - deploy/docker-compose.yml
  - deploy/.env.example
  - scripts/simulate_price_series.sh
---

# TASK-INFRA-006 — Wire gateway into local compose

## §1 — Normative

1. **MUST** add `redis` + `gateway` to [`deploy/docker-compose.yml`](../../../deploy/docker-compose.yml) (dev stack), mirroring production upstream env wiring from [`deploy/docker-compose.production.yml`](../../../deploy/docker-compose.production.yml).
2. **MUST** publish host ports only for `gateway`, `web`, and optionally `db` (loopback-bound). **MUST NOT** publish `pricesvc`, `dealsvc`, `authsvc`, `tracksvc`, `billsvc`, `bff`, or `notifsvc` to the host.
3. **MUST** set `GATEWAY_INTERNAL_BASE_URL=http://gateway:8080` on `web` so Next.js auth handlers call the gateway.
4. **MUST** keep internal service-to-service URLs on the compose network (unchanged).
5. **MUST** update local helpers that previously hit host service ports (e.g. `simulate-prices`) to reach `pricesvc` on the compose network.
6. Forged `X-User-Id` on gateway without JWT **MUST** yield **401** (existing gateway unit coverage + compose smoke).

## §2 — Acceptance

- `docker compose -f deploy/docker-compose.yml config` shows no host ports on private services.
- `curl -H 'X-User-Id: 1' http://127.0.0.1:${GATEWAY_PORT}/v1/alerts` → 401 without Bearer.
- `make up` brings gateway healthy; `make simulate-prices` still writes snapshots.

## §3 — Out of scope

Production Caddy/TLS (R11), CSRF (R6), Vault `_FILE` migration, Google OAuth.
