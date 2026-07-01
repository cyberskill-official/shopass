# Contract Evidence Index — SănDeal

This directory holds contract test definitions for API/Integration contract families defined in `specs/001-full-project-plan/contracts/release-contract.md`.

## Contract Families by Phase

### P0
- Gateway health endpoint
- Auth verification hooks
- Migration status
- Observability endpoints (health, readiness, metrics)

### P1
- Auth/session (register, login, refresh, logout)
- Platform account link (link, unlink)
- Product tracking (search, get, list tracked)
- Price history (get prices, price_daily)
- Chart data (sale-realness, chart)
- Alert rules (CRUD)
- Notification registration (register token)
- Consent/DSAR (record consent, subject rights request)

### P2
- Cart snapshot (capture, get)
- Voucher optimizer (compute discounts)
- Affiliate deeplink (create, track click)
- Payment webhook/reconciliation
- Multi-platform scrape adapters (Shopee, TikTok Shop, Lazada)

### P3
- Cashback payout (create, release)
- B2B trend exports (anonymized market trends)
- Mobile auth/push
- Anti-fraud signals
- SEA country gates

## Per-FR Contract Tests

Each FR §5 defines concrete contract tests. Files in this directory will be created when the corresponding FR is implemented.
