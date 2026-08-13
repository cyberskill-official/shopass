# SănDeal production deployment

`docker-compose.production.yml` is the production topology. It is separate
from `docker-compose.yml`, which remains a local development/demo stack and
publishes service ports for convenience. Do not expose that demo stack to the
Internet.

The production manifest has one public container only: Caddy on ports 80 and
443. Database, Redis, Go services, Next.js, BFF, and one-shot jobs are on a
private Docker network with no host port mappings.

## What this hardening does and does not make live

It provides a safe network boundary, TLS edge configuration, image version
pins, private Redis, persistent Caddy/DB volumes, log rotation, and host-owned
scheduling for scrape/forecast jobs. It does **not** turn the current source
tree into a feature-complete public product.

[`Caddyfile`](Caddyfile) sends every `/v1/*` and `/graphql` request to the
private gateway, which removes client-supplied `X-User-*` headers, verifies
AUTH JWKS, applies its rate/WAF middleware, and routes only allowlisted
upstreams. Production API is **same-origin `/v1`** on `APP_DOMAIN` (for
example `https://shopass.cyberskill.world/v1/track`). Do not point browsers or
the extension at `api.shopass.cyberskill.world` unless that name is an
intentional extra Caddy site with DNS. Caddy removes identity headers,
`X-Service-Token`, and `X-Operator-Token` before proxying as defense in depth.
`X-Service-Token` is reserved for private tracksvc↔pricesvc hops and is never
a browser credential. Caddy probes `gateway /healthz` and `web /api/healthz`
so a dead upstream is less likely to be selected during startup; this is not
a substitute for `/readyz` (see [`HEALTHCHECK-PLAN.md`](HEALTHCHECK-PLAN.md)).
The unscoped legacy `price-history` and `compare` routes are deliberately not
published by gateway for this beta; the owner-scoped chart endpoint is the
supported price-read surface. Alert CRUD (`/v1/alerts`) and device registration
(`/v1/devices`) **are** allowlisted. Wishlist, `/v1/ext/sync`, `/v1/ext/ws`,
and `/v1/comply/breach*` are **not** published on the public gateway. Breach
APIs require `X-Operator-Token` on the private complysvc network and fail
closed without `COMPLY_OPERATOR_TOKEN`. `tracksvc`, BFF, auth, price, deal,
Redis, and the database have no public route or host port. WebSocket `/ws` is
not wired by the current gateway and must remain unavailable until it has an
authenticated upgrade implementation.

`/v1/auth/*` is intentionally a Caddy 404. Browser login, registration,
refresh, and logout use same-origin Next.js `/api/auth/*` handlers, which call
the private gateway and store the refresh credential only in an HttpOnly,
host-only cookie. Caddy overwrites `X-Real-IP` with the direct remote address
before proxying; the web handlers forward it to gateway solely for rate-limit
bucketing. If you put a CDN/LB in front of Caddy, configure Caddy's trusted
proxy ranges first or it will rate-limit the proxy rather than the visitor.

Never use the untracked `Caddyfile.demo` as production configuration. It is a
demo-only identity injector and is intentionally not referenced by the
production manifest.

## Prerequisites

- A Linux host with Docker Engine and Docker Compose v2.
- A DNS name whose A/AAAA records point to the host, plus inbound TCP 80/443.
  Caddy obtains and renews TLS certificates after DNS and firewall are correct;
  no DNS-provider credential is needed for ordinary HTTP-01 issuance.
- At least 2 vCPU, 4 GB RAM, and monitored persistent disk. Forecasting and
  Timescale retention will need more capacity as data grows.
- A backup destination and a tested restore procedure before user data exists.
- The legal/compliance work required for production user data, including PDPL
  consent/DPIA and the incident process. See `docs/deployment_guide.md` and
  `docs/feature-requests/SHIP-GUIDE.md`.

## 1. Prepare non-secret and secret configuration

Create a host-owned configuration area. Do not place real files under
`deploy/secrets/`; that directory is only a documented Git-safe contract.

```bash
sudo install -d -m 0700 /etc/shopass/secrets
sudo install -d -m 0700 /etc/shopass
sudo install -m 0600 deploy/.env.production.example /etc/shopass/runtime.env
sudoedit /etc/shopass/runtime.env
```

Set the non-secret values in `runtime.env`, especially `APP_DOMAIN`,
`ACME_EMAIL`, `APP_ORIGIN`, image pins, `POSTGRES_USER`, `POSTGRES_DB`, and
`AUTH_KEY_ID`. `APP_ORIGIN` must be the exact browser origin (for example
`https://deals.example.com`, without a trailing slash); it is used by Next.js
to reject cross-origin login, logout, registration, and refresh requests.
Keep `GATEWAY_INTERNAL_BASE_URL=http://gateway:8080` on the private Compose
network. It is for server-side Next.js route handlers only, never a public API
URL. Generate the initial database password and auth signing key as described
in [`secrets/README.md`](secrets/README.md). Set `DB_PASSWORD_FILE` and
`AUTH_SIGNING_KEY_SECRET_FILE` to those root-owned files.

The current service binaries still require `DATABASE_URL` at runtime. Inject it
from your host secret manager or add it to `/etc/shopass/runtime.env` with mode
`0600` until `_FILE`/Vault loading exists. The same temporary rule applies to
`GOOGLE_CLIENT_SECRET`, credential-bearing `HTTPS_PROXY`, and the high-entropy
`PRICE_INTERNAL_SERVICE_TOKEN`. The latter must have one identical value in
both `tracksvc` and `pricesvc`; `tracksvc` sends it only on its private
`PRICE_INTERNAL_URL` calls as `X-Service-Token`. It must never be configured at
Caddy or exposed to a browser. This limitation is intentional and tracked in
[`secrets/README.md`](secrets/README.md); it must not be represented as full
FR-INFRA-003 compliance.

Google OAuth is deliberately disabled by the production manifest. Its current
callback returns a token pair as JSON, which would expose a refresh token
outside the HttpOnly cookie flow. Do not set `ENABLE_GOOGLE_OAUTH=true` in
production; authsvc refuses to start if it is enabled. Password registration
and login remain the closed-beta sign-in path.

Before any existing database migration, take a backup. `deploy/migrate.sh`
refuses to auto-baseline a database that already has a `platform` table but no
migration ledger. After independently verifying that its schema contains every
migration in this checkout, an operator may run the migration once with
`MIGRATION_BASELINE=acknowledge-existing-schema`. Do not use that acknowledgement
on a partially migrated database; restore or repair it instead.

## 2. Validate without a domain or cloud credential

The production manifest can be parsed without DNS. Supply harmless placeholder
values only for configuration validation:

```bash
APP_DOMAIN=example.invalid \
ACME_EMAIL=ops@example.invalid \
APP_ORIGIN=https://example.invalid \
DATABASE_URL='postgres://placeholder:placeholder@db:5432/shopass?sslmode=disable' \
DB_PASSWORD_FILE=/dev/null \
AUTH_SIGNING_KEY_SECRET_FILE=/dev/null \
AUTH_KEY_ID=test-key \
PRICE_INTERNAL_SERVICE_TOKEN=not-a-real-secret \
BILL_INTERNAL_SERVICE_TOKEN=not-a-real-secret \
COMPLY_OPERATOR_TOKEN=not-a-real-secret \
docker compose --env-file /dev/null -f deploy/docker-compose.production.yml config
```

For an on-host TLS smoke test that does not contact ACME, set the following
non-secret values in a temporary root-owned runtime file:

```dotenv
CADDYFILE=./Caddyfile.local
APP_DOMAIN=localhost
ACME_EMAIL=ops@localhost
HTTP_PORT=8080
HTTPS_PORT=8443
APP_ORIGIN=https://localhost:8443
GATEWAY_INTERNAL_BASE_URL=http://gateway:8080
```

`Caddyfile.local` uses Caddy's internal certificate authority. Test it with
`curl -k https://localhost:8443/`; do not treat that certificate as public TLS.
Before bringing up that local stack, add a randomly generated
`PRICE_INTERNAL_SERVICE_TOKEN` to the same root-owned runtime file; the
placeholder token above is for `config` parsing only.

## 3. First production launch

From the checked-out repository (the systemd units assume `/srv/shopass`):

```bash
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f deploy/docker-compose.production.yml build
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f deploy/docker-compose.production.yml up -d
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f deploy/docker-compose.production.yml ps
```

Expected state: `db` and Redis become healthy, `migrate` exits successfully,
and the long-running services plus gateway and Caddy are up. Only Caddy should
show host port mappings. Verify from a separate network after DNS propagates:

```bash
curl -fsSI https://your-domain.example/
curl -i 'https://your-domain.example/v1/tracked-products'
```

The second request should return `401` without a valid access token, proving
that Caddy did not bypass the gateway. `/v1/auth/login` on the public host
must be `404` (browser auth is `/api/auth/*` only). Repeat the tracked-products
request with a token from the login flow before treating the protected route
as accepted.

Do not seed demo data or run `make smoke` against a production database.
`make seed` / `make smoke` refuse unless `APP_ENV=dev` (or `development`) or
`ALLOW_SEED=1` is set. Demo row IDs: `app_user.id = 999`,
`tracked_product.id = 100` (plus matching `alert_rule` / `price_forecast`).

## 4. Scheduled work

`dealsvc` already runs its nightly score at 02:00 Asia/Ho_Chi_Minh. Install the
root-owned scrape and forecast timers from [`systemd/README.md`](systemd/README.md).
They run `scrapesvc` every five minutes and `mlforecast` at 01:30, respectively,
with `flock` to prevent overlap. They intentionally do not mount the Docker
socket into an application container.

## 5. Operations

Use the production manifest consistently:

```bash
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f deploy/docker-compose.production.yml logs -f --tail=100
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f deploy/docker-compose.production.yml ps
```

The production DB has no published port. Use `docker compose exec` for a
controlled administrative session rather than opening 5432:

```bash
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f deploy/docker-compose.production.yml exec db \
  psql -U shopass -d shopass
```

Schedule encrypted, off-host backups and test restore regularly. A local export
is useful only as a first step, not as a disaster-recovery strategy:

```bash
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f deploy/docker-compose.production.yml exec -T db \
  pg_dump -U shopass shopass | gzip > shopass_$(date +%F).sql.gz
```

See [`HEALTHCHECK-PLAN.md`](HEALTHCHECK-PLAN.md) for why application readiness
checks are not yet declared in Compose and the exact source work needed before
adding them.

## Web data path (R18)

Default: browser → Next.js → gateway `/v1/*` REST (see
`docs/architecture/web-data-path.md`). The GraphQL BFF is **experimental** —
`docker compose … --profile bff up -d` plus `BFF_UPSTREAM_URL=http://bff:8085`.

## Observability (R13 Prometheus + R14 Loki)

Local/staging overlay (does not change the default compose profile):

```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.observability.yml \
  --env-file deploy/.env --profile observability up -d
```

Loopback UI: Grafana `http://127.0.0.1:3001` (admin/admin from `.env.example`),
Prometheus `:9090`, Alertmanager `:9093`, Loki `:3100`, Pushgateway `:9091`,
gateway metrics `:9094`.

- Alert rules: `deploy/prometheus/rules/shopass.yml` (see `obs/README.md` SLO table).
- Logs: Promtail → Loki; Grafana dashboard **Shopass — errors across services**.
- Loki retention: 14 days (`deploy/loki/loki.yml`).
- Telegram delivery needs host Alertmanager credentials (Stephen ask; noop by default).

## Job scheduling (R17)

Host systemd timers under `deploy/systemd/` run `scrapesvc` (every 5m) and
`mlforecast` (01:30 Asia/Ho_Chi_Minh). After each successful run,
`deploy/scripts/job-heartbeat.sh` pushes `shopass_job_last_success_unixtime` to
Pushgateway so Prometheus can alert on missing runs. Install notes:
[`systemd/README.md`](systemd/README.md).

Demo seed IDs (never production): `app_user` 999, `tracked_product` 100 — see
Makefile `seed`/`smoke` guards (`APP_ENV=dev` or `ALLOW_SEED=1`).

## Release blockers still outside this deployment change

1. Run gateway/auth proxy integration and adversarial JWT tests against the
   built containers, including JWKS outage, header spoofing, rate limits, and
   auth key restart/rotation. `/ws` still requires a verified implementation.
2. Implement controlled auth-key rotation overlap and a human-run recovery
   procedure. The production secret mount prevents boot-time key regeneration,
   but rotation is not automated.
3. Implement `_FILE`/Vault/AWS Secret Manager loading in every service, then
   remove `DATABASE_URL`, third-party credentials, and
   `PRICE_INTERNAL_SERVICE_TOKEN` from process environments.
4. Add real `/livez` + dependency-aware `/readyz` endpoints and probe support
   as listed in [`HEALTHCHECK-PLAN.md`](HEALTHCHECK-PLAN.md).
5. Wire a real notification provider, alert delivery monitoring, metrics,
   tracing, off-host backups, alerting, and restore drills.
6. Obtain residential proxy credentials and verify scraping/legal/compliance
   controls before collecting real marketplace or personal data.

Until those blockers are closed and human-tested, this is a hardened deployment
foundation and staging/public-read edge, not approval to accept real users for
private tracking or alert functionality.
