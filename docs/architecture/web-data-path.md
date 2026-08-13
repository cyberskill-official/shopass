# Shopass web data path (R18)

**Decision (2026-07-26):** option **(b)** — keep the GraphQL BFF experimental; do not put it on the default compose path.

**Decision (2026-08-13):** production API is **same-origin `/v1`** on `https://shopass.cyberskill.world`. Do not depend on `api.shopass.cyberskill.world` (no DNS).

## Production / default local path

```text
Browser
  → Caddy {$APP_DOMAIN}   (production; only public port)
       ├─ /v1/auth/*     → 404 (token-pair JSON must not hit the browser)
       ├─ /v1/* /graphql → private gateway (JWT, WAF, allowlist)
       └─ else           → Next.js web
            ├─ /api/auth/*  → GATEWAY_INTERNAL_BASE_URL → gateway → authsvc
            ├─ /api/healthz → cheap liveness JSON (Caddy health_uri)
            └─ /v1/*        → unused in prod (Caddy already intercepted)

Local Docker without Caddy (make up):
Browser → Next.js :3000
       ├─ /api/auth/* → GATEWAY_INTERNAL_BASE_URL → gateway → authsvc
       └─ /v1/* (except /v1/auth/*) → app/v1/[...path] → gateway only
            (never rewrite to pricesvc/tracksvc; JWT still verified at gateway)
```

`NEXT_PUBLIC_API_BASE_URL` stays empty so `apiFetch("/v1/…")` is same-origin.

Evidence: `web/lib/api.ts`, `web/lib/server-auth.ts`, `web/app/v1/[...path]/route.ts`, and callers under `web/lib/{track,chart,alerts,wishlist,billing}`. No `web/` import of `/graphql`.

## Experimental BFF

- Compose profile: `bff` (`deploy/docker-compose.yml`, `docker-compose.production.yml`).
- Gateway still has optional `BFF_UPSTREAM_URL` for `/graphql` when the profile is up.
- Bring up: `docker compose … --profile bff up -d` and set `BFF_UPSTREAM_URL=http://bff:8085`.

## Why not (a) right now

Migrating chart/wishlist reads to GraphQL would duplicate a working REST surface without a current product consumer. Revisit when a client needs GraphQL aggregation/DataLoader benefits.
