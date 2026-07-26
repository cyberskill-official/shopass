# Staging runbook — monetized vertical slice

Proves: price into `price_snapshot` → forecast → deal alert → notif enqueue → web chart/alerts, plus Premium entitlement → gating.

Password login only. Google OAuth stays disabled (`ENABLE_GOOGLE_OAUTH=false`).

## Prerequisites (operator)

| Need | Env / secret | Status |
|------|----------------|--------|
| Compose stack | `deploy/.env` from `.env.example` | Required |
| Internal tokens | `PRICE_INTERNAL_SERVICE_TOKEN`, `BILL_INTERNAL_SERVICE_TOKEN` | Required |
| Firebase (FCM) | `FCM_PROJECT_ID` + `FCM_SERVICE_ACCOUNT_JSON` (or `GOOGLE_APPLICATION_CREDENTIALS`) | **Deferred** — Path 1 DoD is enqueue (`queued`) |
| Residential proxy (live Shopee) | `HTTPS_PROXY` (+ legal OK) | **Deferred** — use `make simulate-prices` or fixture smoke |
| Payment sandboxes | `MOMO_*` / `ZALOPAY_*` / `VNPAY_*` + public HTTPS IPN host | **Deferred** — use `make grant-premium` for review |

Without Firebase, notifications stay **queued** (deal→notif path still validates). Without proxy, simulate ingest or fixture smoke. Without pay sandboxes, grant Premium via SQL helper (gating still real).

## Bring-up

```bash
cp deploy/.env.example deploy/.env   # fill secrets
docker compose -f deploy/docker-compose.yml --env-file deploy/.env up -d --build
# Production topology: docker-compose.production.yml + Caddy + gateway
```

Health: `gateway` (`:8080`), `web` (`:3000`); private services (`pricesvc`, `tracksvc`, `dealsvc`, `notifsvc`, `billsvc`, `authsvc`, `bff`) have no host ports (TASK-INFRA-006).

## Path 1 — Price → alert → notify

1. **Register** / login (password). Note JWT.
2. **Track** a Shopee product (`POST /v1/track` or dashboard), or use seeded product `100`.
3. **Price into snapshot** (prefer simulate while proxy is deferred):
   - **Simulate drop (no scrape):** `make simulate-prices` → posts `199000 → 179000 → 149000` via pricesvc ingest for product `100`.
   - Fixture scrape→notif: `scripts/smoke_loop.sh` (`SMOKE_ALLOW_DESTRUCTIVE=1` + smoke DB).
   - Live (when proxy ready): `HTTPS_PROXY=... SCRAPE_SEED=<productID>:<itemID>:<shopID> docker compose --profile jobs run --rm scrapesvc`
4. Confirm rows: `SELECT product_id, ts, price FROM price_snapshot WHERE product_id = 100 ORDER BY ts;`
5. **Forecast** (mature SKU): `make seed` includes a forecast fixture, or `docker compose --profile jobs run --rm mlforecast`.
6. **Register push token**: `POST /v1/devices` with `{ "fcm_token", "platform": "web" }` (or Alerts page paste form).
7. **Alert rule**: create `bottom_predicted` on `/alerts` (requires Premium — see Path 2 if free).
8. Trigger deal nightly / `make deal-once`; expect `POST` to notifsvc `/notify` → `notification.status` → `queued` (→ `sent` only when FCM is configured later).
9. Open `/products/100/chart`.

**DoD:** enqueue succeeds only with verified push token; failed enqueue does **not** write `bottom_alert_log`. Live FCM `sent` is deferred.

## Path 2 — Pay → gate

### Primary (review while sandboxes deferred)

1. As free user, `POST http://127.0.0.1:8080/v1/alerts` with Bearer JWT and `bottom_predicted` → **402** + `upgrade_path: /billing`. (Forged `X-User-Id` without JWT → **401**.)
2. Temporary bypass (keeps billsvc gating real; skips checkout/IPN):

```bash
make grant-premium USER_ID=<uid>
```

3. Confirm `subscription.status = 'active'` for that user.
4. Retry `bottom_predicted` create via gateway with JWT → **201**.

### Optional full pay path (when sandboxes / local HMAC ready)

1. Open `/billing`, choose plan + gateway → `POST /v1/billing/checkout` (via web → gateway) → pay URL / VietQR payload.
2. Simulate IPN (sandbox or signed local body with `MOMO_SECRET_KEY`):

```bash
# body must match pending payment amount; X-Signature = HMAC-SHA256 hex of raw body
# IPN is public (no JWT); still goes through gateway host port.
curl -X POST "http://127.0.0.1:8080/v1/billing/ipn/momo" \
  -H "Content-Type: application/json" \
  -H "X-Signature: $SIG" \
  -d '{"order_ref":"order_<uid>_premium_basic","transaction_id":"t1","amount":29000,"status":"paid"}'
```

3. Confirm `subscription.status = 'active'`.
4. Retry `bottom_predicted` create → **201**.
5. Bad signature / amount mismatch → payment not paid; sub not activated.

`make grant-premium` is local/staging review only (compose `db` on loopback). Do not use in production.

## Path 3 — Wishlist limit

Free tier `wishlist_items = 20`. Add items until **402** `wishlist_limit_reached`; after Premium, higher/unlimited per `plan_feature`.

## Explicitly out of this runbook

B2B, mobile, APNs/email/SMS, cashback, antifraud, SEA comply, Lazada/TikTok live, Vault `_FILE` full migration, live FCM send, residential proxy scrape, real MoMo/ZaloPay/VNPay sandboxes.

## HITL

**Accepted (operator, 2026-07-26):** Paths 1–2 under current DoD — `make simulate-prices` + `make grant-premium` (FCM send, `HTTPS_PROXY` live scrape, and real payment sandboxes remain deferred).

**Reaffirmed (operator “Suggested next steps” / implement, 2026-07-26):** same DoD; helpers already on `main` (PR #56); Wave‑1 R1/R2/R4/R5/R17 already `done`. Next engineering is post‑R1 product-goal pick (Trust / Comply / Channels) or secrets when provisioned — not B2B/mobile first.

Agent must not self-set CyberOS product/improvement task status to `done` without a separate human verdict at the review and final-acceptance gates.
