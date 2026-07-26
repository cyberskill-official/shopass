# k6 load tests (R20)

Mixed gateway load against **NFR-INFRA-001**:

| Class | Endpoint | p95 target |
|-------|----------|------------|
| `price_chart` | `GET /v1/products/{id}/chart?range=90d` | &lt; 500ms |
| `read_cached` | `GET /v1/tracked-products` | &lt; 300ms |

`/v1/compare` and `/v1/wishlists` are **not** published on the gateway in closed beta (404). Revisit when those routes ship.

## Local

```bash
export K6_BASE_URL=https://api.shopass.cyberskill.world
export K6_ACCESS_TOKEN='…'          # or K6_ACCESS_TOKENS=t1,t2,t3
export K6_PRODUCT_ID=1
k6 run loadtests/k6/nfr-infra-001.js
```

Install: https://grafana.com/docs/k6/latest/set-up/install-k6/

## Rate limits

Gateway default is ~100 req/min per user. A single token at 50 RPS will 429.
For the weekly gate, either:

1. Set `K6_ACCESS_TOKENS` to a pool of staging load-test JWTs, or
2. Raise / bypass the limit on staging (ops change).

## CI

`.github/workflows/k6-weekly.yml` — Mondays 01:00 UTC + `workflow_dispatch`.
Requires repo secrets: `K6_BASE_URL`, `K6_ACCESS_TOKEN` (or `K6_ACCESS_TOKENS`), `K6_PRODUCT_ID`.

Without secrets the job skips (does not fail the schedule).
