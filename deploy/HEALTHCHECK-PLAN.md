# Health and readiness plan

The production Compose manifest intentionally declares health checks only for
Postgres and Redis. It does **not** add misleading shell-based probes to the
application services: the Go runtime images are distroless and do not contain
`sh`, `curl`, or `wget`, and several services do not expose a health route.

Current endpoint coverage:

| Service | Current endpoint | Safe Compose health check now? |
| --- | --- | --- |
| `db` | `pg_isready` | Yes |
| `redis` | `redis-cli ping` | Yes |
| `authsvc` | `/healthz` | No - liveness only; it does not verify DB connectivity |
| `tracksvc` | `/healthz` | No - liveness only; it does not verify DB connectivity |
| `dealsvc` | `/healthz` | No - liveness only; it does not verify DB connectivity |
| `bff` | `/healthz` | No - image/probe support still needed |
| `gateway` | `/healthz` | No - liveness only; it does not verify Redis/JWKS reachability. Caddy now probes this URI. |
| `pricesvc` | `/healthz` | No - liveness only; distroless images still cannot shell-probe |
| `notifsvc` | `/healthz` | No - liveness only; distroless images still cannot shell-probe |
| `web` | `/api/healthz` | No - cheap JSON liveness; Caddy probes it. Not a dependency-aware `/readyz`. |
| `complysvc` | `/healthz` | No - liveness only |

## Partial edge step (2026-08-13)

Caddy `health_uri` on the gateway (`/healthz`) and web (`/api/healthz`)
upstreams reduces sending traffic to a process that is not listening yet.
It does **not** replace `/livez`+`/readyz`, distroless probe binaries, or
Compose `service_healthy` for application containers. A 502 on the public
host can still mean the stack is down, DNS/TLS is wrong, or Caddy has no
healthy upstream.

## Required source work (deliberately not included in deployment hardening)

1. Add `/livez` and `/readyz` to every HTTP service. `readyz` must ping its
   database and any dependency that makes the service unable to serve traffic;
   it must not merely report that the process has started.
2. Add a small static HTTP probe binary to `deploy/Dockerfile.go`, or use an
   intentionally chosen base image that contains a verified probe utility.
   Then declare `CMD ["/healthcheck", "http://127.0.0.1:PORT/readyz"]` in
   Compose. Do not use `CMD-SHELL curl` in the current distroless images.
3. Add a dedicated web health route which avoids authentication and expensive
   rendering, and verify it with Node's built-in `fetch` or a reviewed probe.
4. Make Caddy's upstream configuration use `service_healthy` only after those
   probes exist. Until then, Caddy may temporarily return 502 during startup;
   it will not expose a private upstream directly.
5. Add external HTTPS synthetic checks and alerts after DNS/TLS are live.

This document is a release blocker checklist, not an assertion that the stack
currently has production-grade readiness reporting.
