# Staging runbook — monetized vertical slice

Proves: live (or fixture) Shopee scrape → `price_snapshot` → forecast → deal alert → notif enqueue/FCM → web chart/alerts, plus sandbox pay → subscription → gating.

Password login only. Google OAuth stays disabled (`ENABLE_GOOGLE_OAUTH=false`).

## Prerequisites (operator)

| Need | Env / secret |
|------|----------------|
| Compose stack | `deploy/.env` from `.env.example` |
| Firebase (FCM) | `FCM_PROJECT_ID` + `FCM_SERVICE_ACCOUNT_JSON` (or `GOOGLE_APPLICATION_CREDENTIALS`) |
| Residential proxy (live Shopee) | `HTTPS_PROXY` (+ legal OK) |
| Payment sandboxes | `MOMO_*` / `ZALOPAY_*` / `VNPAY_*` + public HTTPS IPN host |
| Internal tokens | `PRICE_INTERNAL_SERVICE_TOKEN`, `BILL_INTERNAL_SERVICE_TOKEN` |

Without Firebase, notifications stay **queued** (deal→notif path still validates). Without proxy, use fixture smoke for scrape.

## Bring-up

```bash
cp deploy/.env.example deploy/.env   # fill secrets
docker compose -f deploy/docker-compose.yml --env-file deploy/.env up -d --build
# Production topology: docker-compose.production.yml + Caddy + gateway
```

Health: `pricesvc`, `tracksvc`, `dealsvc`, `notifsvc`, `billsvc`, `authsvc`, `web`.

## Path 1 — Price → alert → notify

1. **Register** / login (password). Note JWT.
2. **Track** a Shopee product (`POST /v1/track` or dashboard).
3. **Scrape** into snapshot:
   - Fixture: `make smoke` / `scripts/smoke_loop.sh` (sets `SMOKE_ALLOW_DESTRUCTIVE=1` + smoke DB).
   - Live: `HTTPS_PROXY=... SCRAPE_SEED=<productID>:<itemID>:<shopID> docker compose --profile jobs run --rm scrapesvc`
4. Confirm row: `SELECT price FROM price_snapshot WHERE product_id = … ORDER BY captured_at DESC LIMIT 1;`
5. **Forecast** (mature SKU): `docker compose --profile jobs run --rm mlforecast` (or wait for timer).
6. **Register push token**: `POST /v1/devices` with `{ "fcm_token", "platform": "web" }` (or Alerts page paste form).
7. **Alert rule**: create `bottom_predicted` on `/alerts` (requires Premium — see Path 2 if free).
8. Trigger deal nightly / call dealsvc batch; expect `POST` to notifsvc `/notify` → `notification.status` → `queued` → `sent` when FCM configured.
9. Open `/products/{id}/chart`.

**DoD:** enqueue succeeds only with verified push token; failed enqueue does **not** write `bottom_alert_log`.

## Path 2 — Pay → gate

1. As free user, `POST /v1/alerts` with `bottom_predicted` → **402** + `upgrade_path: /billing`.
2. Open `/billing`, choose plan + gateway → `POST /v1/billing/checkout` → pay URL / VietQR payload.
3. Simulate IPN (sandbox or signed local body):

```bash
# body must match pending payment amount; X-Signature = HMAC-SHA256 hex of raw body
curl -X POST "$GATEWAY/v1/billing/ipn/momo" \
  -H "Content-Type: application/json" \
  -H "X-Signature: $SIG" \
  -d '{"order_ref":"order_<uid>_premium_basic","transaction_id":"t1","amount":29000,"status":"paid"}'
```

4. Confirm `subscription.status = 'active'`.
5. Retry `bottom_predicted` create → **201**.
6. Bad signature / amount mismatch → payment not paid; sub not activated.

## Path 3 — Wishlist limit

Free tier `wishlist_items = 20`. Add items until **402** `wishlist_limit_reached`; after Premium, higher/unlimited per `plan_feature`.

## Explicitly out of this runbook

B2B, mobile, APNs/email/SMS, cashback, antifraud, SEA comply, Lazada/TikTok live, Vault `_FILE` full migration.

## HITL

Agent must not self-set CyberOS task status to `done`. Operator accepts after Paths 1–2 succeed on staging.
