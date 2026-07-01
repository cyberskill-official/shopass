# Quickstart: Validate SănDeal Full Project Plan

This guide validates the project from docs-only state through full release. Commands are intentionally high-level until implementation FRs create concrete scripts.

## Prerequisites

- Read `README.md`.
- Read `AGENTS.md`; note it currently points to CyberOS memory protocol setup.
- Read `docs/feature-requests/SHIP-GUIDE.md`.
- Read `docs/feature-requests/IMPLEMENTATION-ORDER.md`.
- Read `docs/feature-requests/DATA-MODEL.md`.
- Read the target FR and its `.audit.md` before each implementation.

## Planning Validation

1. Confirm Spec Kit artifacts exist:

```bash
ls specs/001-full-project-plan
```

Expected: `spec.md`, `plan.md`, `research.md`, `data-model.md`, `quickstart.md`, `contracts/`.

2. Confirm source backlog is still present:

```bash
test -f docs/feature-requests/BACKLOG.md
test -f docs/feature-requests/IMPLEMENTATION-ORDER.md
test -f docs/feature-requests/SHIP-GUIDE.md
```

Expected: all commands exit 0.

## Per-FR Implementation Loop

For each next FR:

1. Select the first un-done FR in the lowest dependency layer where all dependencies are `done`.
2. Read:

```text
docs/feature-requests/SHIP-GUIDE.md
docs/feature-requests/DATA-MODEL.md
docs/feature-requests/<module>/<FR-ID>-<slug>.md
docs/feature-requests/<module>/<FR-ID>-<slug>.audit.md
```

3. Set status to `implementing` in FR frontmatter and BACKLOG.
4. Implement only the files in `new_files`/`modified_files` unless the FR explicitly requires more.
5. Run tests listed in §5 of the FR.
6. Verify every MUST/SHOULD statement in §1 maps to acceptance criteria §4 and passing tests.
7. Move status through review/test gates to `done`.

Expected: no FR reaches `done` without test and acceptance evidence.

## P0 Exit Gate

Run after FR-INFRA-001 through FR-INFRA-005 are done.

Validate:

- Gateway routes REST, GraphQL and WSS where specified.
- Migrations apply from empty DB.
- `platform` and `app_user` foundation exist.
- Secrets are referenced through Vault/AWS Secrets Manager paths, not cleartext.
- OTel traces, Prometheus metrics, Grafana dashboards and structured logs are live.
- Country config returns VN policy and restrictive defaults.

Expected: local stack boots and one health request is observable end-to-end.

## P1 MVP Exit Gate

Run after all P1 MUST FRs are done.

Scenario:

1. Register/login a user.
2. Link Shopee account by anonymous `ext_user_ref` only.
3. Extension reads Shopee product/cart DOM and sends minimal payload.
4. Scrape or ingest a product price.
5. Write `tracked_product` and delta-only `price_snapshot`.
6. Refresh/read `price_daily`.
7. Show web chart.
8. Run sale-realness classifier.
9. Create wishlist and alert rule.
10. Dispatch push notification through FCM path.
11. Verify consent/DSAR/no-cleartext/security audit evidence.

Expected:

- No cookie/token/password appears in backend payloads or logs.
- Chart/history reads do not scan raw snapshots for normal UI.
- Cold-start returns documented fallback behavior.
- P1 compliance and trust evidence is recorded.

## P2 Expansion Exit Gate

Run after all P2 MUST FRs are done.

Scenario:

1. Extension reads Shopee, TikTok Shop and Lazada surfaces.
2. Scraping adapters ingest prices from all supported platforms with pacing and proxy controls.
3. Cart snapshot is captured.
4. Voucher optimizer computes discount using country stacking rules.
5. User explicitly clicks an affiliate action and sees disclosure.
6. System creates deeplink and records click/conversion without PII in `sub_id`.
7. User upgrades to Premium and payment webhook reconciles idempotently.
8. Bottom-price baseline scoring runs and can trigger alert.

Expected:

- No affiliate auto-injection or cookie-stuffing.
- MY/PH no-stack behavior is enforced where configured.
- Payment mismatch does not grant Premium.
- Notification channels handle retry/DLQ.

## P3 Growth Exit Gate

Run after all P3 scoped FRs are done or explicitly deferred.

Scenario:

1. Cashback is created from confirmed conversion.
2. Payout remains held until eligible and passes fraud checks.
3. B2B trend export suppresses rows below K_MIN and contains no product/user/shop identifiers.
4. Mobile app authenticates and receives push.
5. Country gating enables at least one SEA policy with restrictive fallback for unknown countries.
6. Anti-fraud signals flag referral/payment/device abuse patterns.

Expected:

- Cashback cannot be paid before hold and fraud checks.
- B2B export is anonymized by construction.
- Mobile does not bypass consent or notification routing.

## Completion Definition

The full project is complete when:

- 90/90 FR and 10/10 NFR are `done`, or any non-done item has explicit `on_hold`/`closed` rationale.
- All phase exit gates pass.
- All active contracts in `contracts/` pass.
- No SHIP-GUIDE invariant violation remains open.
