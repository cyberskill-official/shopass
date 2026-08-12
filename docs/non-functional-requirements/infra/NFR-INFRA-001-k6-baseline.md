# NFR-INFRA-001 — k6 baseline log (R20)

Regression gate script: `loadtests/k6/nfr-infra-001.js`  
Workflow: `.github/workflows/k6-weekly.yml` (Mon 01:00 UTC + dispatch)

## Traffic model

- Mixed arrival-rate ramp to **50 RPS** (~60% chart / ~40% tracked-products).
- Targets: chart p95 &lt; 500ms; cache-class p95 &lt; 300ms; error rate &lt; 5%.
- Auth: Bearer JWT via `K6_ACCESS_TOKEN` or pooled `K6_ACCESS_TOKENS`.
- **Not hit:** `/v1/compare`, `/v1/wishlists` (gateway 404 in closed beta).

## Rate-limit caveat

Gateway default ~100 req/min per user. Sustained 50 RPS needs a token pool or staging limit raise — document the chosen mitigation in each weekly row.

## Results

| Date (UTC) | Commit | Base URL | Product | Tokens | Chart p95 | Cache p95 | Errors | CI |
|------------|--------|----------|---------|--------|-----------|-----------|--------|-----|
| _pending first green weekly run_ | — | `shopass.cyberskill.world` | secret | — | — | — | — | — |

Append a row after each weekly (or dispatch) run. Paste artifact / Actions URL in the CI column.
