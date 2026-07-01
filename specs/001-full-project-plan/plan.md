# Implementation Plan: SănDeal Full Project Plan

**Branch**: `001-full-project-plan` | **Date**: 2026-06-30 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-full-project-plan/spec.md`

## Summary

Build SănDeal from an audited docs-only backlog into a complete multi-platform product. The plan preserves the original 90 FR + 10 NFR as implementation units, follows the verified 8-layer dependency order, and adds Spec Kit artifacts for orchestration, gates, project structure, contracts and validation.

Technical approach: create the product as a multi-surface monorepo with Go backend services, PostgreSQL 16/TimescaleDB, TypeScript MV3 extension, Next.js web, Python ML, and later mobile. The critical path prioritizes data foundation and price history early so the sale-realness features can accumulate history while UI and monetization layers are built.

## Technical Context

**Language/Version**: Go 1.22 for backend services; TypeScript for extension and web; Python 3.x for deal ML; React Native TypeScript as default mobile assumption pending FR-MOBILE-001.

**Primary Dependencies**: PostgreSQL 16, TimescaleDB 2.x, Redis or Kafka-Redis Streams, OpenTelemetry, Prometheus, Grafana, Vault/AWS Secrets Manager, Chrome Manifest V3 APIs, Playwright, Next.js App Router, Prophet, LightGBM.

**Storage**: PostgreSQL 16 for OLTP; TimescaleDB hypertable `price_snapshot` and continuous aggregate `price_daily`; Redis/Kafka-Redis Streams for queue/fan-out; Vault/AWS Secrets Manager for secret material.

**Testing**: Per-FR unit, contract and integration tests from §5 of each FR; migration tests for DB; MV3 extension tests; Playwright/e2e tests for web and extension flows; load/performance tests for API/time-series/notification NFRs; compliance/security tests for no-cleartext/token-not-on-server/affiliate guardrails.

**Target Platform**: Linux server/container runtime; Chrome/Chromium MV3 extension; Next.js web; iOS/Android mobile in P3; VN first, SEA expansion behind country gates.

**Project Type**: Multi-service SaaS plus browser extension, web app, ML jobs, and later mobile app.

**Performance Goals**: API/chart/history p95 targets from NFR docs; chart reads from `price_daily`; FCM quota handling up to 600,000 messages/minute/project with backoff; price writes delta-only; scraping paced with jitter and cost guard.

**Constraints**: PDPL consent/DPIA/DSAR/breach 72h; no cleartext secrets; no platform token/cookie leaves client; all money BIGINT VND; MV3 service worker ephemeral; affiliate only user-initiated with disclosure; residential proxy required for scraping; default restrictive country policy.

**Scale/Scope**: 90 FR, 10 NFR, 16 modules, 4 phases, 8 dependency layers, estimated 627h direct engineering before review/integration buffer.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Spec Kit constitution file is still the upstream template. For this repo, enforce the 9 non-negotiable invariants from `docs/feature-requests/SHIP-GUIDE.md` as the active constitution:

1. Security and PDPL: no cleartext, no server-side platform tokens, argon2id PHC, consent, DPIA, DSAR, breach 72h.
2. Post-Honey affiliate ethics: user-initiated only, disclosure, no cookie-stuffing/auto-redirect/pop-under.
3. Money: BIGINT VND only.
4. Price: delta-only `price_snapshot`, TimescaleDB hypertable, charts from `price_daily`.
5. Extension MV3: ephemeral service worker, `chrome.storage`, alarms >=30s, Offscreen/DNR where needed, minimal payload only.
6. Notification: push > email > sms, FCM quota/backoff, flatten midnight spikes.
7. Data model: one table one owner FR, extensions by ALTER/FK only.
8. Per-country gating: VN first, restrictive default, MY/PH no-stack rules.
9. Scraping: residential proxy, pacing/jitter, DOM drift monitoring, no risky voucher/coin automation beyond user-initiated/checklist behavior.

**Gate Result**: PASS for planning. No violations are required by this plan. Any implementation PR that violates one invariant must fail review unless the original audited FR is amended and re-audited.

## Project Structure

### Documentation (this feature)

```text
specs/001-full-project-plan/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── release-contract.md
└── tasks.md
```

`tasks.md` is intentionally not created by this plan command; generate it with `speckit-tasks` after review.

### Source Code (repository root)

```text
services/
├── gateway/
├── auth/
├── infra/
├── price/
├── scrape/
├── track/
├── deal/
├── notif/
├── cart/
├── affil/
├── bill/
├── b2b/
├── comply/
└── trust/

extension/
├── manifest.json
├── src/
│   ├── service-worker/
│   ├── content/
│   ├── offscreen/
│   ├── popup/
│   └── shared/
└── tests/

web/
├── app/
├── components/
├── graphql/
└── tests/

mobile/
├── app/
└── tests/

ml/
├── deal/
└── tests/

deploy/
├── docker/
├── k8s/
├── migrations/
├── observability/
└── secrets/

tests/
├── contract/
├── integration/
├── e2e/
├── performance/
└── compliance/
```

**Structure Decision**: Use a monorepo because the audited FRs span multiple services and clients but share schemas, contracts, release gates and compliance evidence. Each FR's `service`, `new_files`, and `modified_files` remain the source of truth for exact file creation.

## Phase Plan

### Phase 0 - Foundation, 5 FR, about 33h

Goal: API gateway, data model/migrations, secrets, observability and country config. Exit only when services can run locally, migrations apply cleanly, secrets are referenced through secret manager paths, and observability shows traces/logs/metrics.

Start immediately with layer 0: FR-EXT-001, FR-INFRA-001, FR-INFRA-002, FR-INFRA-003. Although extension scaffold is in P1, it has no dependencies and should run in parallel with infra.

### Phase 1 - MVP, 46 FR, about 303h

Goal: Shopee MVP: auth, extension, scrape, price history, tracking, deal classification, push notifications, web chart/wishlist, PDPL and trust evidence.

Critical path:

```text
FR-INFRA-002 -> FR-PRICE-001 -> FR-PRICE-002 -> FR-DEAL-001/FR-DEAL-003 -> FR-WEB-003
```

Parallel data cold-start path:

```text
FR-INFRA-003 -> FR-SCRAPE-001 -> FR-SCRAPE-002 -> FR-SCRAPE-005
```

Extension trust path:

```text
FR-EXT-001 -> FR-EXT-002 -> FR-EXT-003 -> FR-EXT-005/FR-TRUST-002/FR-TRUST-003
```

Exit only when the MVP demo in `quickstart.md` passes and all P1 MUST FRs are `done`.

### Phase 2 - Expansion, 25 FR, about 177h

Goal: TikTok Shop + Lazada support, cart/voucher optimizer, affiliate, Premium billing, ML bottom-price prediction and additional notification channels.

Prioritize monetization dependencies after platform support: affiliate schema/deeplink/network, billing schema/payment/reconciliation, and cart optimizer with country stacking policy. Exit only when the P2 demo passes and compliance guardrails prove user-initiated affiliate behavior.

### Phase 3 - Growth, 14 FR, about 114h

Goal: Cashback, B2B anonymized trend products, mobile app, SEA compliance and anti-fraud. Exit only when B2B k-anonymity suppression, cashback hold/release, mobile auth/push and country-gated expansion all pass.

## Dependency Layer Execution

Layer 0: FR-EXT-001, FR-INFRA-001, FR-INFRA-002, FR-INFRA-003.

Layer 1: FR-AFFIL-001, FR-AUTH-001, FR-CART-001, FR-COMPLY-001, FR-COMPLY-005, FR-EXT-002, FR-INFRA-004, FR-INFRA-005, FR-NOTIF-001, FR-PRICE-001, FR-TRUST-001, FR-EXT-004.

Layer 2: FR-AFFIL-002, FR-AUTH-002, FR-AUTH-003, FR-BILL-001, FR-CART-005, FR-COMPLY-002, FR-COMPLY-003, FR-COMPLY-004, FR-EXT-003, FR-EXT-006, FR-EXT-007, FR-EXT-008, FR-NOTIF-002, FR-NOTIF-003, FR-PRICE-002, FR-PRICE-005, FR-SCRAPE-001, FR-CART-006, FR-COMPLY-008.

Layer 3: FR-AFFIL-003, FR-AFFIL-004, FR-AUTH-005, FR-BILL-002, FR-CART-002, FR-DEAL-001, FR-DEAL-003, FR-EXT-005, FR-NOTIF-004, FR-NOTIF-005, FR-NOTIF-006, FR-PRICE-003, FR-PRICE-004, FR-SCRAPE-002, FR-SCRAPE-003, FR-TRUST-002, FR-TRUST-003, FR-WEB-001, FR-AUTH-004, FR-B2B-001, FR-BILL-004, FR-BILL-005, FR-MOBILE-001, FR-NOTIF-007.

Layer 4: FR-BILL-003, FR-CART-003, FR-DEAL-002, FR-SCRAPE-004, FR-SCRAPE-005, FR-SCRAPE-006, FR-SCRAPE-007, FR-SCRAPE-008, FR-TRACK-001, FR-TRUST-004, FR-WEB-002, FR-WEB-003, FR-B2B-002, FR-WEB-005, FR-B2B-003, FR-B2B-004, FR-MOBILE-003.

Layer 5: FR-CART-004, FR-DEAL-004, FR-TRACK-002, FR-TRACK-003, FR-TRUST-005, FR-MOBILE-002, FR-TRUST-006.

Layer 6: FR-COMPLY-006, FR-DEAL-006, FR-TRACK-004, FR-WEB-004, FR-AFFIL-005, FR-DEAL-005.

Layer 7: FR-COMPLY-007.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |

## Post-Design Constitution Check

PASS. The design artifacts preserve FR ownership, avoid new data-model conflicts, keep audited docs authoritative, and define validation gates for the 9 invariants.
