# SănDeal - FR implementation coverage (real vs stub vs missing)

Date: 2026-07-03. This maps the 90 FRs to what actually exists in code, separate from "tests pass." Method: parsed each FR's declared `new_files`, checked presence on disk, and inspected the present code for stubs (hardcoded returns, simulated external calls).

Headline: 71 of 90 FRs have code on disk; 19 have none. About 73% of the declared files exist. Of the code that does exist, the internal logic is largely real, but every point that touches an external system (marketplace scraping, CAPTCHA, payment gateways, push channels beyond FCM, ML fitting) is a stub or a baseline. So the scaffolding is real; the outside-world integration is not built yet.

## Per-module status

| Module | FRs | Files present | Status | Notes |
|---|---:|---:|---|---|
| INFRA | 5 | 100% | real | gateway, data model, secrets, observability, region config. Gateway user-auth injection is a placeholder (`track.go`: "giả lập logic gateway"). |
| EXT (extension) | 8 | 100% | real | Strongest module. 87 jest tests, guardrails enforce affiliate/MV3/no-token-leak. |
| DEAL | 6 | 100% | real logic, baseline ML | fake-sale, cold-start, chart, nightly batch all real and tested. The forecast it reads is a baseline. |
| SCRAPE | 8 | 100% files | shopee real, lazada/tiktok stub | orchestrator/proxy/pacing real; the shopee adapter does real HTTP fetch + JSON parse (micro-VND to BIGINT VND) with a Playwright-farm fallback on WAF challenge; lazada and tiktok return hardcoded prices; the CAPTCHA solver is simulated. |
| TRACK | 4 | 95% | real | wishlist, alert_rule, firing engine. Reads user via the gateway placeholder. |
| BILL | 5 | 94% | real, gateway calls stubbed | subscription/reconcile/referral/gating real; the MoMo/ZaloPay/VNPay calls are placeholders. |
| CART | 6 | 90% | real | voucher catalog, snapshot, optimizer, per-country stacking. |
| AFFIL | 5 | 82% | partial | tracking/deeplink/network present; FR-AFFIL-005 (cashback layering) missing. |
| WEB | 5 | 78% | partial | scaffold, landing/SEO, chart, wishlist UI present; FR-WEB-005 missing. |
| AUTH | 5 | 70% | partial | JWT, account linking, social login present; FR-AUTH-004 (account lifecycle) missing. |
| PRICE | 5 | 64% | partial | tracked_product, snapshot hypertable, history present; FR-PRICE-004 missing. |
| COMPLY | 8 | 62% | partial | PDPL consent, DPIA, DSAR, breach, ecommerce-obligation present; FR-COMPLY-006/007 (per-country extension, SEA) missing. |
| NOTIF | 7 | 41% | partial | schema, FCM, fan-out, midnight-spike present; FR-NOTIF-005/006/007 (APNs, email, SMS) missing. |
| TRUST | 6 | 41% | partial | open-source, data-minimization, security-audit present; FR-TRUST-004/005/006 (anti-fraud, attribution-gaming, device fingerprint) missing. |
| B2B | 4 | 0% | missing | entire module absent (P3). |
| MOBILE | 3 | 0% | missing | entire module absent (P3). |

## The 19 FRs with no code

FR-AUTH-004, FR-PRICE-004, FR-WEB-005, FR-AFFIL-005, FR-NOTIF-005, FR-NOTIF-006, FR-NOTIF-007, FR-COMPLY-006, FR-COMPLY-007, FR-TRUST-004, FR-TRUST-005, FR-TRUST-006, FR-B2B-001, FR-B2B-002, FR-B2B-003, FR-B2B-004, FR-MOBILE-001, FR-MOBILE-002, FR-MOBILE-003.

Most are P2/P3 (B2B, mobile, cashback, anti-fraud, SEA compliance, extra push channels), so their absence is consistent with an MVP-first build. The P1 gaps to close first are FR-AUTH-004, FR-PRICE-004, and FR-WEB-005.

## The stubs that block a working MVP

These files exist and compile, but do not do the real thing:

- Marketplace scraping - the shopee adapter does real HTTP fetch and JSON parse with a Playwright-farm fallback; the lazada and tiktok adapters still return canned prices. No live proxy run against real Shopee has been exercised (anti-bot / residential proxy).
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

What genuinely remains is not one-session work: the Lazada/TikTok live adapters, real payment-gateway calls, and live Shopee scraping all need third-party services and credentials to build and verify; and the missing modules (B2B, mobile) plus P2/P3 features (cashback, anti-fraud, extra push channels, SEA compliance) are greenfield. The core value loop, however, is now real, tested end to end, deployable, and self-running.
