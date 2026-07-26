# Shopass web data path (R18)

**Decision (2026-07-26):** option **(b)** — keep the GraphQL BFF experimental; do not put it on the default compose path.

## Production / default local path

```text
Browser
  → Next.js (web) same-origin routes
       ├─ /api/auth/*  → GATEWAY_INTERNAL_BASE_URL → gateway → authsvc
       └─ apiFetch("/v1/…") → NEXT_PUBLIC_API_BASE_URL ("" = same origin
            proxied/rewritten to gateway in deploy) → gateway → track|price|deal|…
```

Evidence: `web/lib/api.ts`, `web/lib/server-auth.ts`, and callers under `web/lib/{track,chart,alerts,wishlist,billing}`. No `web/` import of `/graphql`.

## Experimental BFF

- Compose profile: `bff` (`deploy/docker-compose.yml`, `docker-compose.production.yml`).
- Gateway still has optional `BFF_UPSTREAM_URL` for `/graphql` when the profile is up.
- Bring up: `docker compose … --profile bff up -d` and set `BFF_UPSTREAM_URL=http://bff:8085`.

## Why not (a) right now

Migrating chart/wishlist reads to GraphQL would duplicate a working REST surface without a current product consumer. Revisit when a client needs GraphQL aggregation/DataLoader benefits.
