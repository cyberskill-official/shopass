# SănDeal - FR implementation coverage (real vs stub vs missing)

Date: 2026-07-03. This maps the 90 FRs to what actually exists in code, separate from "tests pass." Method: parsed each FR's declared `new_files`, checked presence on disk, and inspected the present code for stubs (hardcoded returns, simulated external calls).

Headline: 74 of 90 FRs have code on disk; 16 have none. About 76% of the declared files exist. Of the code that does exist, the internal logic is largely real, but every point that touches an external system (marketplace scraping, CAPTCHA, payment gateways, push channels beyond FCM, ML fitting) is a stub or a baseline. So the scaffolding is real; the outside-world integration is not built yet.

## Per-module status

| Module | FRs | Files present | Status | Notes |
|---|---:|---:|---|---|
| INFRA | 5 | 100% | real | gateway, data model, secrets, observability, region config. Gateway user-auth injection is a placeholder (`track.go`: "giả lập logic gateway"). |
| EXT (extension) | 8 | 100% | real | Strongest module. 87 jest tests, guardrails enforce affiliate/MV3/no-token-leak. |
| DEAL | 6 | 100% | real logic, baseline ML | fake-sale, cold-start, chart, nightly batch all real and tested. The forecast it reads is a baseline. |
| SCRAPE | 8 | 100% files | shopee + lazada + tiktok real | orchestrator/proxy/pacing real; the shopee adapter does real HTTP fetch + JSON parse (micro-VND to BIGINT VND) with a Playwright-farm fallback; lazada and tiktok extract on the farm (embedded-JSON first, DOM fallback, integer VND) and the Go orchestrator adapters now dispatch to the farm instead of returning a canned price; the farm TypeScript type-checks clean and its extraction logic is unit-verified; the CAPTCHA solver is simulated. |
| TRACK | 4 | 95% | real | wishlist, alert_rule, firing engine. Reads user via the gateway placeholder. |
| BILL | 5 | 94% | real, gateway calls stubbed | subscription/reconcile/referral/gating real; the MoMo/ZaloPay/VNPay calls are placeholders. |
| CART | 6 | 90% | real | voucher catalog, snapshot, optimizer, per-country stacking. |
| AFFIL | 5 | 82% | partial | tracking/deeplink/network present; FR-AFFIL-005 (cashback layering) missing. |
| WEB | 5 | 100% | real | scaffold, landing/SEO, chart, wishlist UI, and the GraphQL BFF (FR-WEB-005: read-only, behind gateway, resolvers delegate to REST, DataLoader anti-N+1, depth/cost caps). |
| AUTH | 5 | 100% | real | JWT issue/refresh, password, account model, and social login (FR-AUTH-004: OAuth Authorization Code + PKCE, id_token verify via JWKS, social_identity, takeover-safe email merge). Only the live Google token exchange is gated on a real client secret. |
| PRICE | 5 | 100% | real | tracked_product, snapshot hypertable, history, and cross-platform compare (FR-PRICE-004: GET /v1/compare, latest-per-product, server-side cheapest flag). |
| COMPLY | 8 | 62% | partial | PDPL consent, DPIA, DSAR, breach, ecommerce-obligation present; FR-COMPLY-006/007 (per-country extension, SEA) missing. |
| NOTIF | 7 | 41% | partial | schema, FCM, fan-out, midnight-spike present; FR-NOTIF-005/006/007 (APNs, email, SMS) missing. |
| TRUST | 6 | 41% | partial | open-source, data-minimization, security-audit present; FR-TRUST-004/005/006 (anti-fraud, attribution-gaming, device fingerprint) missing. |
| B2B | 4 | 0% | missing | entire module absent (P3). |
| MOBILE | 3 | 0% | missing | entire module absent (P3). |

## The 16 FRs with no code

FR-AFFIL-005, FR-NOTIF-005, FR-NOTIF-006, FR-NOTIF-007, FR-COMPLY-006, FR-COMPLY-007, FR-TRUST-004, FR-TRUST-005, FR-TRUST-006, FR-B2B-001, FR-B2B-002, FR-B2B-003, FR-B2B-004, FR-MOBILE-001, FR-MOBILE-002, FR-MOBILE-003.

All three P1 gaps (FR-PRICE-004, FR-WEB-005, FR-AUTH-004) are now built (see below). The 16 remaining are P2/P3 - B2B and mobile (whole modules), plus cashback, anti-fraud, SEA compliance, and the extra push channels - consistent with an MVP-first build.

## The stubs that block a working MVP

These files exist and compile, but do not do the real thing:

- Marketplace scraping - the shopee adapter does real HTTP fetch and JSON parse with a Playwright-farm fallback; the lazada and tiktok adapters extract on the farm (embedded-JSON first, DOM fallback, integer VND) and the Go orchestrator dispatches to the farm rather than fabricating a price. What is still not exercised here is a live run: real Shopee behind a residential proxy, and the browser-backed farm adapter tests, which need Playwright Chromium and real proxy credentials.
- CAPTCHA - `scrape/internal/captcha/solver.go` is simulated.
- Payment gateways - `bill/internal/pay/gateway.go` order/callback logic is a placeholder; no real MoMo/ZaloPay/VNPay calls.
- Gateway auth - the API gateway's user-identity injection is a placeholder; services read `X-User-Id` but the real authn path from JWT is not wired end to end.
- Forecasting - the ml service ships a Prophet/LightGBM baseline, not a trained, evaluated model.

## Honest read

What exists is a coherent, well-structured P0/P1 skeleton with strong internal logic and good tests, especially the extension. The Shopee price path is further along than the rest: the adapter reads and parses real Shopee JSON, and the price service applies delta-only writes. What is missing is the connective tissue and the harder integrations - taking payments, the Lazada/TikTok adapters, forecasting the bottom, and running the scrapers behind a residential proxy. The next build effort should finish one vertical slice rather than widening coverage into P2/P3.

## Progress (2026-07-03)

Wired the scrape-to-price slice end to end:

- Price ingest endpoint `POST /v1/price/snapshots` (`services/price/internal/api/ingest.go`) - the internal write API, since price owns `price_snapshot`. Validates money (BIGINT VND, price > 0, list_price >= price) and applies the existing delta-only logic. Three integration tests against Postgres prove write, delta-only skip, and rejection of a non-positive price.
- Price client (`services/scrape/internal/priceclient`) implementing `orchestrator.PriceRepo` by POSTing snapshots to that endpoint, with unit tests.
- A minimal in-memory queue (`services/scrape/internal/memqueue`) and a runnable entrypoint `services/scrape/cmd/scrapesvc` that wires the real Shopee adapter to the price client. It reads `SCRAPE_SEED` (productID:itemID:shopID) and honors `HTTPS_PROXY` for the residential-proxy requirement. Run it with `PRICE_BASE_URL=... SCRAPE_SEED=... go run ./cmd/scrapesvc`.
- An end-to-end test drives a realistic Shopee pdp JSON through adapter, pool, and price client to a price ingest server, asserting the micro-VND price is parsed to BIGINT VND (199000) and list price (250000).

- ML forecast write path (`services/ml/bottom/db.py` + `bottom/run_forecast.py`, runnable via `DATABASE_URL=... python -m bottom.run_forecast`): reads price history from `price_snapshot`, forecasts (Prophet for mature SKUs, cold-start prior otherwise), and upserts `price_forecast` - the table `deal`'s bottom-price alerts read. Two DB-backed tests prove an alertable forecast (p_bottom 0.85) writes and upserts idempotently, and that the runner reads history end to end. Added `psycopg2-binary` to requirements.

With scrape -> price -> `price_snapshot` and ml -> `price_forecast` both wired on real, tested code, the core loop is connected: scrape prices, forecast the bottom, `deal` fires the alert, web reads the chart.

Ran it live end to end on real service processes and a real Postgres (`scripts/smoke_loop.sh`): a new `services/price/cmd/pricesvc` HTTP server was started; `scrapesvc` scraped a fake Shopee endpoint, parsed 199000 VND, and posted it to `pricesvc`, which wrote `price_snapshot`; a mature forecast (p_bottom 0.85) was recorded; and `dealsvc` (given a new `RUN_ONCE` trigger) ran the nightly score, fired the alert into `bottom_alert_log`, and delivered the notification to a notif sink. The smoke asserts the scraped price and the fired alert.

Authored the deploy stack (`deploy/docker-compose.yml` + `deploy/Dockerfile.{go,ml,web}` + `deploy/migrate.sh` + `deploy/README.md`): TimescaleDB, a migration init, notifsvc, pricesvc, dealsvc, and web always-up, plus scrapesvc and mlforecast as on-demand jobs; `scripts/smoke_loop.sh` is the acceptance check. Validating the migration chain against Postgres surfaced two real defects, both fixed: the `platform` lookup was never seeded although everything FKs it (added the canonical shopee/tiktok/lazada seed to `db/migrations/0002`), and `migrate.sh` was applying golang-migrate `.down.sql` files (now skipped). The full chain applies cleanly on a real TimescaleDB (the one remaining error in the sandbox is `price_daily`, the TimescaleDB continuous aggregate, which exists in compose).

Built the durable scrape queue (FR-SCRAPE-001, `services/scrape/internal/pgqueue`): a Postgres-backed `orchestrator.Queue` over the existing `scrape_job` table using `FOR UPDATE SKIP LOCKED` and a `locked_until` lease, joining `platform_item_id` from `tracked_product`. Three integration tests prove enqueue/claim/ack, that concurrent claims return distinct jobs, and lease reclaim after a crashed worker. `scrapesvc` now uses it whenever `DATABASE_URL` is set (enqueue the seed, then drain all due jobs), falling back to the in-memory queue otherwise; compose passes `DATABASE_URL` so the stack uses the durable path.

Built the scheduler that makes the loop self-running: `services/scrape/internal/feeder.SyncJobs` registers a `scrape_job` for every `tracked_product` that lacks one, and the orchestrator now persists tier-based rescheduling after each scrape (`Pool.commit` -> `Rescheduler`, implemented by `pgqueue.Reschedule`), so hot products re-scrape in minutes and cold ones in about a day. `scrapesvc` runs the feeder then drains the queue, so a periodic invocation keeps the whole loop going. Verified: all 14 runnable Go modules green with per-module integration databases (the CI `go` job now provisions `shopass_{deal,price,scrape}_test` so the DROP/CREATE integration suites do not collide).

What genuinely remains is not one-session work: real payment-gateway calls and live Shopee scraping need third-party services and credentials to build and verify; and the missing modules (B2B, mobile) plus P2/P3 features (cashback, anti-fraud, extra push channels, SEA compliance) are greenfield. The core value loop, however, is now real, tested end to end, deployable, and self-running.

## Progress (2026-07-03, feature build)

Closed FR-PRICE-004 and hardened the Lazada/TikTok scrape path.

- FR-PRICE-004 cross-platform compare: `GET /v1/compare?canonical_key=...` in price-svc (`internal/price/compare_query.go` + `internal/api/compare.go`, registered in the router). It joins tracked_product by canonical_key, takes the latest snapshot per product with a per-product LATERAL (real ts, not price_daily), derives the platform display name server-side, and computes the cheapest flag server-side (ties all flagged, single platform is cheapest by itself). Returns 400 for a missing key, 404 for an unknown key (not an empty 200). Five integration tests against Postgres prove the missing-key, unknown-key, cheapest-across-platforms (including latest-per-product), single-platform, and tie cases; the full price package stays green.
- Lazada adapter (FR-SCRAPE-008): the Go orchestrator adapter no longer fabricates a price. It now dispatches every job to the Playwright farm (Akamai fingerprints at the TLS layer, so a raw HTTP client is the wrong tool), and returns an error when no farm is configured rather than writing a fake value. Four Go unit tests cover dispatch, pass-through, the no-farm error, and farm-error propagation.
- Farm TypeScript harness: the farm package had a tsconfig that never type-checked (`module: nodenext` + `verbatimModuleSyntax` against a commonjs package, and `types: []` hiding the jest globals) - 189 errors across every adapter and test. Fixed the config and two real null-safety bugs (the Lazada script scan and the TikTok ItemModule index) so `tsc --noEmit` is clean (0 errors). The Lazada and TikTok extraction logic (embedded-JSON first, DOM fallback, integer VND) is verified against its own test cases (4/4 each) on the compiled real source.
- CI reliability: the per-module Go test step now runs with `-p 1` so a module's integration suites (which share one database) execute serially and cannot race on shared tables (the scrape feeder vs pgqueue race on tracked_product).

Environment-gated, not done: the browser-backed farm adapter tests (`adapter.test.ts`) need Playwright Chromium; FR-AUTH-004 social login needs a real OAuth client secret to verify end to end.

## Progress (2026-07-03, GraphQL BFF)

Built FR-WEB-005, the read-only GraphQL BFF, as a new `services/bff` package (graphql-js + @graphql-tools/schema + dataloader, run as a small Node HTTP service). It sits behind the gateway and takes the caller identity from a forwarded `x-user-id` header - it does not verify the token itself (DEC-WEB-21). The schema is Query-only (`me`, `wishlists`, `productChart`) with no Mutation (DEC-WEB-25). Resolvers never touch the database; they delegate to the REST services through a `RestClient`, so ownership checks stay in track-svc (DEC-WEB-22). A per-request DataLoader batches and de-duplicates chart loads so a wishlist of N items makes one round of chart calls, not N (DEC-WEB-23), and a depth cap plus an additive cost cap reject over-large queries before any resolver runs (DEC-WEB-24). Upstream 403/404 map to a neutral `NOT_FOUND` and 5xx/network to a generic error, so one user's request cannot leak another's resource (§1 #10). Ten tests on Node's built-in runner cover the anonymous rejection, REST delegation and the forwarded chart-feed shape, DataLoader batching (3 distinct calls for 5 field requests), the neutral error mapping, header forwarding, and the depth and cost caps; `tsc --noEmit` is clean. A `deploy/Dockerfile.node` builds it. Wiring it into compose behind the gateway is a follow-up, tracked with the gateway itself, which is not in the core stack yet.

## Progress (2026-07-03, social login)

Built the verifiable core of FR-AUTH-004 (social login) in `services/auth`. The one part that needs live Google credentials - the real token exchange over the network - is isolated; everything around it is built and tested.

- Migration `0007_social_identity` links a provider subject to an app_user with `UNIQUE(provider, subject)`.
- PKCE (`internal/auth/pkce`, S256 only), a one-time TTL state store (`MemTmpStore`), and random state + nonce back the Authorization Code + PKCE flow (DEC-AUTH-17).
- The id_token verifier (`oidc.go`) checks the RS256 signature against the provider JWKS, the issuer, the audience (client id), expiry, and nonce before any claim is trusted (§1 #4). Its tests cover the valid path plus every rejection: wrong audience, wrong issuer, nonce mismatch, expired, wrong signing key, unknown kid, and the alg-none / HS256 confusion attacks.
- `OAuthService` runs `StartOAuth` and `OAuthCallback`, and `resolveUser` implements the account-linking safety matrix (DEC-AUTH-19): an existing social identity wins; a verified email merges into an existing account; an unverified email never merges and instead creates a new account with a NULL email so it cannot collide with the victim. State is single-use, so a replayed callback is rejected.
- `GoogleProvider` builds the authorize URL (PKCE S256, state, nonce) and does the code exchange with an injectable HTTP client, so the exchange-then-verify path is tested end to end against a fake token endpoint plus a self-signed id_token; only a real Google call needs the client secret (loaded via the secrets manager, never a code/env literal - §1 #9).
- `pgRepo` gained `FindBySocial`, `LinkSocial`, and `CreateSocialUser` (a separate `SocialRepo` interface, so the main `Repo` interface and its fakes are untouched), verified against Postgres. Fixing this surfaced a real latent bug: `FindByEmail` scanned `phone` and `pwd_hash` into plain strings and so could not read a social-only account (both NULL); it now COALESCEs them, which also matters on the second-social-login merge path.

The full auth module - the new social-login code plus every pre-existing suite - is green (`go vet` clean, unit and Postgres integration). What remains for a live sign-in is supplying a Google (or Zalo) OAuth client id and secret and standing up the two HTTP handlers that call `StartOAuth` and `OAuthCallback`.

## Progress (2026-07-03, runnable servers + compose wiring)

To make the BFF and auth deployable, three services that previously had no runnable HTTP server were given one, and all three plus the BFF are now wired into `docker-compose.yml`.

- `authsvc` (`services/auth/cmd/authsvc`): an HTTP server wiring the existing register/login/refresh plus the new social-login start/callback and a JWKS endpoint. It generates an RS256 signing key at boot and enables social login only when `GOOGLE_CLIENT_ID` is set (the secret comes from the environment, not a literal). It also gained an `HTTPKeySet` that fetches and caches a provider JWKS (unit-tested with a fake fetch). The whole auth module stays green.
- `dealsvc` now serves its existing chart handler (FR-DEAL-003) over HTTP alongside the nightly cron, backed by a new real chart repository (daily series from `price_daily`, maturity from `first_seen`, fake-sale verdict from the series). Deal's full suite stays green with the composed test database.
- `tracksvc`'s stub main was replaced with a real server serving the existing wishlist handlers, with a middleware that reads the gateway-forwarded `X-User-Id`.
- `docker-compose.yml` gained `authsvc`, `tracksvc`, and `bff` service entries (and exposed the deal chart port), with matching `deploy/.env.example` variables. The compose file validates and every new main builds; `docker compose up` itself is still not run here (no Docker in the build environment).

Two honest edges remain: the BFF returns wishlists without their items until `tracksvc` embeds items in its list response, and social login needs real Google credentials. Both are noted in `deploy/README.md`.
