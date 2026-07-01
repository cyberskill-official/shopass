# Tasks: SănDeal Full Project Plan

**Input**: Design documents from `/specs/001-full-project-plan/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/release-contract.md](contracts/release-contract.md), [quickstart.md](quickstart.md)

**Tests**: Required. Every FR task must run the tests listed in that FR's §5 and prove every §1 MUST/SHOULD maps to §4 acceptance criteria.

**Organization**: Tasks are dependency-ordered for handoff to another AI. Each implementation task points to the source FR doc and expected code area. Mark a task done only after code, tests, acceptance check, status update, and review evidence are complete.

## Phase 1: Setup (Shared Project Controls)

**Purpose**: Create the control surface another AI should follow before touching implementation.

- [x] T001 Create the implementation workspace structure declared in `specs/001-full-project-plan/plan.md` under `services/`, `extension/`, `web/`, `mobile/`, `ml/`, `deploy/`, and `tests/`
- [x] T002 Create a project execution ledger in `specs/001-full-project-plan/execution-ledger.md` with columns Task, FR/NFR, Owner, Status, Evidence, Review Notes
- [x] T003 Create a phase evidence bundle template in `specs/001-full-project-plan/evidence-template.md` using the requirements from `specs/001-full-project-plan/contracts/release-contract.md`
- [x] T004 Create a reusable FR execution checklist in `specs/001-full-project-plan/fr-done-checklist.md` covering read docs, implement, test §5, verify §1->§4, update FR status, update BACKLOG, attach evidence
- [x] T005 [P] Create CI placeholder structure in `.github/workflows/` for backend, extension, web, mobile, ML, compliance, and performance checks
- [x] T006 [P] Create local development placeholder docs in `deploy/docker/README.md` covering PostgreSQL 16, TimescaleDB, Redis/Kafka-Redis Streams, Vault/AWS Secrets Manager, OTel, Prometheus, Grafana
- [x] T007 [P] Create compliance evidence index in `tests/compliance/README.md` listing no-cleartext, token-not-on-server, PDPL consent, affiliate guardrail, and country-gating checks
- [x] T008 [P] Create contract evidence index in `tests/contract/README.md` mapping P0/P1/P2/P3 contract families from `specs/001-full-project-plan/contracts/release-contract.md`

---

## Phase 2: Foundational (Blocking Governance)

**Purpose**: Set the rules that block all FR work.

**CRITICAL**: No implementation task below should be marked done unless these controls are in place.

- [x] T009 Document active constitution rules from `docs/feature-requests/SHIP-GUIDE.md` in `specs/001-full-project-plan/constitution-gates.md`
- [x] T010 Document status transition procedure from `docs/feature-requests/STATUS-REFERENCE.md` in `specs/001-full-project-plan/status-procedure.md`
- [x] T011 Document source-of-truth rule that each FR doc and `.audit.md` controls implementation in `specs/001-full-project-plan/source-of-truth.md`
- [x] T012 Document one-table-one-owner rule from `docs/feature-requests/DATA-MODEL.md` in `specs/001-full-project-plan/schema-ownership.md`
- [x] T013 Create the first release evidence bundle for P0 in `specs/001-full-project-plan/evidence/p0-foundation.md`
- [x] T014 Create the MVP release evidence bundle for P1 in `specs/001-full-project-plan/evidence/p1-mvp.md`
- [x] T015 Create the expansion release evidence bundle for P2 in `specs/001-full-project-plan/evidence/p2-expansion.md`
- [x] T016 Create the growth release evidence bundle for P3 in `specs/001-full-project-plan/evidence/p3-growth.md`

**Checkpoint**: Foundation ready. Start FR implementation from the dependency-ordered tasks below.

---

## Phase 3: User Story 1 - Ship MVP Value Loop (Priority: P1) MVP

**Goal**: Build P0 + P1 so Shopee price tracking, price history, sale-realness, web chart, wishlist and push alert work end-to-end.

**Independent Test**: Complete all P0/P1 tasks below, then run the P1 MVP scenario in `specs/001-full-project-plan/quickstart.md`.

### Layer 0 Implementation

- [x] T017 [P] [US1] Implement FR-EXT-001 MV3 scaffold from `docs/feature-requests/ext/FR-EXT-001-mv3-scaffold-service-worker.md` and audit `docs/feature-requests/ext/FR-EXT-001-mv3-scaffold-service-worker.audit.md` in `extension/`
- [x] T018 [P] [US1] Implement FR-INFRA-001 API Gateway/BFF from `docs/feature-requests/infra/FR-INFRA-001-api-gateway-bff.md` and audit `docs/feature-requests/infra/FR-INFRA-001-api-gateway-bff.audit.md` in `services/gateway/`
- [x] T019 [P] [US1] Implement FR-INFRA-002 data-model foundation from `docs/feature-requests/infra/FR-INFRA-002-data-model-foundation.md` and audit `docs/feature-requests/infra/FR-INFRA-002-data-model-foundation.audit.md` in `services/infra/` and `deploy/migrations/`
- [x] T020 [P] [US1] Implement FR-INFRA-003 secrets management from `docs/feature-requests/infra/FR-INFRA-003-secrets-management.md` and audit `docs/feature-requests/infra/FR-INFRA-003-secrets-management.audit.md` in `services/infra/` and `deploy/secrets/`
- [x] T021 [US1] Review Layer 0 by running each FR §5 test, updating FR frontmatter and `docs/feature-requests/BACKLOG.md`, and recording evidence in `specs/001-full-project-plan/evidence/p0-foundation.md`

### Layer 1 MVP Implementation

- [ ] T022 [P] [US1] Implement FR-AUTH-001 app user schema and registration from `docs/feature-requests/auth/FR-AUTH-001-app-user-schema.md` and audit `docs/feature-requests/auth/FR-AUTH-001-app-user-schema.audit.md` in `services/auth/`
- [ ] T023 [P] [US1] Implement FR-COMPLY-001 PDPL consent framework from `docs/feature-requests/comply/FR-COMPLY-001-pdpl-consent-framework.md` and audit `docs/feature-requests/comply/FR-COMPLY-001-pdpl-consent-framework.audit.md` in `services/comply/`
- [ ] T024 [P] [US1] Implement FR-COMPLY-005 no-cleartext enforcement from `docs/feature-requests/comply/FR-COMPLY-005-no-cleartext-enforcement.md` and audit `docs/feature-requests/comply/FR-COMPLY-005-no-cleartext-enforcement.audit.md` in `services/comply/` and `tests/compliance/`
- [ ] T025 [P] [US1] Implement FR-EXT-002 Shopee content script from `docs/feature-requests/ext/FR-EXT-002-shopee-content-script.md` and audit `docs/feature-requests/ext/FR-EXT-002-shopee-content-script.audit.md` in `extension/src/content/`
- [ ] T026 [P] [US1] Implement FR-INFRA-004 observability spine from `docs/feature-requests/infra/FR-INFRA-004-observability-spine.md` and audit `docs/feature-requests/infra/FR-INFRA-004-observability-spine.audit.md` in `services/infra/` and `deploy/observability/`
- [ ] T027 [P] [US1] Implement FR-INFRA-005 per-country region config from `docs/feature-requests/infra/FR-INFRA-005-per-country-region-config.md` and audit `docs/feature-requests/infra/FR-INFRA-005-per-country-region-config.audit.md` in `services/infra/`
- [ ] T028 [P] [US1] Implement FR-NOTIF-001 notification schema and routing from `docs/feature-requests/notif/FR-NOTIF-001-notification-schema-routing.md` and audit `docs/feature-requests/notif/FR-NOTIF-001-notification-schema-routing.audit.md` in `services/notif/`
- [ ] T029 [P] [US1] Implement FR-PRICE-001 tracked product canonical key from `docs/feature-requests/price/FR-PRICE-001-tracked-product-canonical-key.md` and audit `docs/feature-requests/price/FR-PRICE-001-tracked-product-canonical-key.audit.md` in `services/price/`
- [ ] T030 [P] [US1] Implement FR-TRUST-001 open-source extension trust task from `docs/feature-requests/trust/FR-TRUST-001-open-source-extension.md` and audit `docs/feature-requests/trust/FR-TRUST-001-open-source-extension.audit.md` in `extension/` and `services/trust/`
- [ ] T031 [P] [US1] Implement FR-EXT-004 offscreen and declarativeNetRequest support from `docs/feature-requests/ext/FR-EXT-004-offscreen-declarativenetrequest.md` and audit `docs/feature-requests/ext/FR-EXT-004-offscreen-declarativenetrequest.audit.md` in `extension/src/offscreen/`
- [ ] T032 [US1] Review Layer 1 MVP tasks by running each FR §5 test, checking SHIP-GUIDE invariants, updating statuses, and recording evidence in `specs/001-full-project-plan/evidence/p1-mvp.md`

### Layer 2 MVP Implementation

- [ ] T033 [P] [US1] Implement FR-AUTH-002 JWT issuance from `docs/feature-requests/auth/FR-AUTH-002-jwt-issuance.md` and audit `docs/feature-requests/auth/FR-AUTH-002-jwt-issuance.audit.md` in `services/auth/`
- [ ] T034 [P] [US1] Implement FR-AUTH-003 platform account linking from `docs/feature-requests/auth/FR-AUTH-003-platform-account-linking.md` and audit `docs/feature-requests/auth/FR-AUTH-003-platform-account-linking.audit.md` in `services/auth/`
- [ ] T035 [P] [US1] Implement FR-COMPLY-002 DPIA/TIA register from `docs/feature-requests/comply/FR-COMPLY-002-dpia-tia-register.md` and audit `docs/feature-requests/comply/FR-COMPLY-002-dpia-tia-register.audit.md` in `services/comply/`
- [ ] T036 [P] [US1] Implement FR-COMPLY-003 data subject rights from `docs/feature-requests/comply/FR-COMPLY-003-data-subject-rights.md` and audit `docs/feature-requests/comply/FR-COMPLY-003-data-subject-rights.audit.md` in `services/comply/`
- [ ] T037 [P] [US1] Implement FR-COMPLY-004 breach notification 72h from `docs/feature-requests/comply/FR-COMPLY-004-breach-notification-72h.md` and audit `docs/feature-requests/comply/FR-COMPLY-004-breach-notification-72h.audit.md` in `services/comply/`
- [ ] T038 [P] [US1] Implement FR-EXT-003 extension data minimization pipeline from `docs/feature-requests/ext/FR-EXT-003-data-minimization-pipeline.md` and audit `docs/feature-requests/ext/FR-EXT-003-data-minimization-pipeline.audit.md` in `extension/src/shared/` and `tests/compliance/`
- [ ] T039 [P] [US1] Implement FR-EXT-006 settings and consent UI from `docs/feature-requests/ext/FR-EXT-006-settings-consent-ui.md` and audit `docs/feature-requests/ext/FR-EXT-006-settings-consent-ui.audit.md` in `extension/src/popup/`
- [ ] T040 [P] [US1] Implement FR-NOTIF-002 FCM dispatcher from `docs/feature-requests/notif/FR-NOTIF-002-fcm-dispatcher.md` and audit `docs/feature-requests/notif/FR-NOTIF-002-fcm-dispatcher.audit.md` in `services/notif/`
- [ ] T041 [P] [US1] Implement FR-NOTIF-003 fanout pipeline from `docs/feature-requests/notif/FR-NOTIF-003-fanout-pipeline.md` and audit `docs/feature-requests/notif/FR-NOTIF-003-fanout-pipeline.audit.md` in `services/notif/`
- [ ] T042 [P] [US1] Implement FR-PRICE-002 price snapshot hypertable from `docs/feature-requests/price/FR-PRICE-002-price-snapshot-hypertable.md` and audit `docs/feature-requests/price/FR-PRICE-002-price-snapshot-hypertable.audit.md` in `services/price/` and `deploy/migrations/`
- [ ] T043 [P] [US1] Implement FR-PRICE-005 canonical key matching from `docs/feature-requests/price/FR-PRICE-005-canonical-key-matching.md` and audit `docs/feature-requests/price/FR-PRICE-005-canonical-key-matching.audit.md` in `services/price/`
- [ ] T044 [P] [US1] Implement FR-SCRAPE-001 scraping orchestrator core from `docs/feature-requests/scrape/FR-SCRAPE-001-orchestrator-core.md` and audit `docs/feature-requests/scrape/FR-SCRAPE-001-orchestrator-core.audit.md` in `services/scrape/`
- [ ] T045 [US1] Review Layer 2 MVP tasks by running each FR §5 test, checking no token/cookie/server cleartext evidence, updating statuses, and recording evidence in `specs/001-full-project-plan/evidence/p1-mvp.md`

### Layer 3 MVP Implementation

- [ ] T046 [P] [US1] Implement FR-AUTH-005 account lifecycle from `docs/feature-requests/auth/FR-AUTH-005-account-lifecycle.md` and audit `docs/feature-requests/auth/FR-AUTH-005-account-lifecycle.audit.md` in `services/auth/`
- [ ] T047 [P] [US1] Implement FR-DEAL-001 fake sale detection from `docs/feature-requests/deal/FR-DEAL-001-fake-sale-detection.md` and audit `docs/feature-requests/deal/FR-DEAL-001-fake-sale-detection.audit.md` in `services/deal/` and `ml/deal/`
- [ ] T048 [P] [US1] Implement FR-DEAL-003 chart data API from `docs/feature-requests/deal/FR-DEAL-003-chart-data-api.md` and audit `docs/feature-requests/deal/FR-DEAL-003-chart-data-api.audit.md` in `services/deal/`
- [ ] T049 [P] [US1] Implement FR-EXT-005 extension backend sync from `docs/feature-requests/ext/FR-EXT-005-extension-backend-sync.md` and audit `docs/feature-requests/ext/FR-EXT-005-extension-backend-sync.audit.md` in `extension/` and `services/gateway/`
- [ ] T050 [P] [US1] Implement FR-NOTIF-004 midnight spike scheduler from `docs/feature-requests/notif/FR-NOTIF-004-midnight-spike-scheduler.md` and audit `docs/feature-requests/notif/FR-NOTIF-004-midnight-spike-scheduler.audit.md` in `services/notif/`
- [ ] T051 [P] [US1] Implement FR-PRICE-003 price history API from `docs/feature-requests/price/FR-PRICE-003-price-history-api.md` and audit `docs/feature-requests/price/FR-PRICE-003-price-history-api.audit.md` in `services/price/`
- [ ] T052 [P] [US1] Implement FR-PRICE-004 cross-platform compare from `docs/feature-requests/price/FR-PRICE-004-cross-platform-compare.md` and audit `docs/feature-requests/price/FR-PRICE-004-cross-platform-compare.audit.md` in `services/price/`
- [ ] T053 [P] [US1] Implement FR-SCRAPE-002 Shopee internal API adapter from `docs/feature-requests/scrape/FR-SCRAPE-002-shopee-internal-api-adapter.md` and audit `docs/feature-requests/scrape/FR-SCRAPE-002-shopee-internal-api-adapter.audit.md` in `services/scrape/`
- [ ] T054 [P] [US1] Implement FR-SCRAPE-003 Playwright farm anti-fingerprint from `docs/feature-requests/scrape/FR-SCRAPE-003-playwright-farm-antifingerprint.md` and audit `docs/feature-requests/scrape/FR-SCRAPE-003-playwright-farm-antifingerprint.audit.md` in `services/scrape/`
- [ ] T055 [P] [US1] Implement FR-TRUST-002 data minimization policy from `docs/feature-requests/trust/FR-TRUST-002-data-minimization-policy.md` and audit `docs/feature-requests/trust/FR-TRUST-002-data-minimization-policy.audit.md` in `services/trust/` and `tests/compliance/`
- [ ] T056 [P] [US1] Implement FR-TRUST-003 security audit hooks from `docs/feature-requests/trust/FR-TRUST-003-security-audit-hooks.md` and audit `docs/feature-requests/trust/FR-TRUST-003-security-audit-hooks.audit.md` in `services/trust/` and `tests/compliance/`
- [ ] T057 [P] [US1] Implement FR-WEB-001 Next.js scaffold from `docs/feature-requests/web/FR-WEB-001-nextjs-scaffold.md` and audit `docs/feature-requests/web/FR-WEB-001-nextjs-scaffold.audit.md` in `web/`
- [ ] T058 [P] [US1] Implement FR-AUTH-004 social login from `docs/feature-requests/auth/FR-AUTH-004-social-login.md` and audit `docs/feature-requests/auth/FR-AUTH-004-social-login.audit.md` in `services/auth/`
- [ ] T059 [US1] Review Layer 3 MVP tasks by running each FR §5 test, checking price/extension/compliance paths, updating statuses, and recording evidence in `specs/001-full-project-plan/evidence/p1-mvp.md`

### Layer 4-6 MVP Implementation

- [ ] T060 [P] [US1] Implement FR-DEAL-002 cold-start handling from `docs/feature-requests/deal/FR-DEAL-002-cold-start-handling.md` and audit `docs/feature-requests/deal/FR-DEAL-002-cold-start-handling.audit.md` in `services/deal/` and `ml/deal/`
- [ ] T061 [P] [US1] Implement FR-SCRAPE-004 residential proxy rotation from `docs/feature-requests/scrape/FR-SCRAPE-004-residential-proxy-rotation.md` and audit `docs/feature-requests/scrape/FR-SCRAPE-004-residential-proxy-rotation.audit.md` in `services/scrape/`
- [ ] T062 [P] [US1] Implement FR-SCRAPE-005 delta pacing CAPTCHA from `docs/feature-requests/scrape/FR-SCRAPE-005-delta-pacing-captcha.md` and audit `docs/feature-requests/scrape/FR-SCRAPE-005-delta-pacing-captcha.audit.md` in `services/scrape/`
- [ ] T063 [P] [US1] Implement FR-SCRAPE-006 DOM change monitoring from `docs/feature-requests/scrape/FR-SCRAPE-006-dom-change-monitoring.md` and audit `docs/feature-requests/scrape/FR-SCRAPE-006-dom-change-monitoring.audit.md` in `services/scrape/`
- [ ] T064 [P] [US1] Implement FR-TRACK-001 track product API from `docs/feature-requests/track/FR-TRACK-001-track-product-api.md` and audit `docs/feature-requests/track/FR-TRACK-001-track-product-api.audit.md` in `services/track/`
- [ ] T065 [P] [US1] Implement FR-WEB-002 SEO landing from `docs/feature-requests/web/FR-WEB-002-seo-landing.md` and audit `docs/feature-requests/web/FR-WEB-002-seo-landing.audit.md` in `web/`
- [ ] T066 [P] [US1] Implement FR-WEB-003 price chart UI from `docs/feature-requests/web/FR-WEB-003-price-chart-ui.md` and audit `docs/feature-requests/web/FR-WEB-003-price-chart-ui.audit.md` in `web/`
- [ ] T067 [P] [US1] Implement FR-WEB-005 GraphQL BFF from `docs/feature-requests/web/FR-WEB-005-graphql-bff.md` and audit `docs/feature-requests/web/FR-WEB-005-graphql-bff.audit.md` in `web/` and `services/gateway/`
- [ ] T068 [P] [US1] Implement FR-TRACK-002 wishlist schema from `docs/feature-requests/track/FR-TRACK-002-wishlist-schema.md` and audit `docs/feature-requests/track/FR-TRACK-002-wishlist-schema.audit.md` in `services/track/`
- [ ] T069 [P] [US1] Implement FR-TRACK-003 alert rule schema from `docs/feature-requests/track/FR-TRACK-003-alert-rule-schema.md` and audit `docs/feature-requests/track/FR-TRACK-003-alert-rule-schema.audit.md` in `services/track/`
- [ ] T070 [P] [US1] Implement FR-TRACK-004 alert firing engine from `docs/feature-requests/track/FR-TRACK-004-alert-firing-engine.md` and audit `docs/feature-requests/track/FR-TRACK-004-alert-firing-engine.audit.md` in `services/track/`
- [ ] T071 [P] [US1] Implement FR-WEB-004 wishlist alert UI from `docs/feature-requests/web/FR-WEB-004-wishlist-alert-ui.md` and audit `docs/feature-requests/web/FR-WEB-004-wishlist-alert-ui.audit.md` in `web/`
- [ ] T072 [US1] Run P0 foundation exit gate from `specs/001-full-project-plan/quickstart.md` and record output in `specs/001-full-project-plan/evidence/p0-foundation.md`
- [ ] T073 [US1] Run P1 MVP exit gate from `specs/001-full-project-plan/quickstart.md` and record output in `specs/001-full-project-plan/evidence/p1-mvp.md`

**Checkpoint**: User Story 1 complete. MVP can be reviewed independently.

---

## Phase 4: User Story 2 - Expand To Monetizable Multi-Platform Product (Priority: P2)

**Goal**: Add TikTok Shop, Lazada, cart/voucher optimizer, Premium billing, affiliate, ML bottom-price baseline, and non-push notification channels.

**Independent Test**: Complete all P2 tasks below, then run the P2 Expansion scenario in `specs/001-full-project-plan/quickstart.md`.

### P2 Foundation and Platform Expansion

- [ ] T074 [P] [US2] Implement FR-AFFIL-001 affiliate tracking schema from `docs/feature-requests/affil/FR-AFFIL-001-affiliate-tracking-schema.md` and audit `docs/feature-requests/affil/FR-AFFIL-001-affiliate-tracking-schema.audit.md` in `services/affil/`
- [ ] T075 [P] [US2] Implement FR-CART-001 voucher catalog schema from `docs/feature-requests/cart/FR-CART-001-voucher-catalog-schema.md` and audit `docs/feature-requests/cart/FR-CART-001-voucher-catalog-schema.audit.md` in `services/cart/`
- [ ] T076 [P] [US2] Implement FR-AFFIL-002 user-initiated deeplink from `docs/feature-requests/affil/FR-AFFIL-002-user-initiated-deeplink.md` and audit `docs/feature-requests/affil/FR-AFFIL-002-user-initiated-deeplink.audit.md` in `services/affil/`
- [ ] T077 [P] [US2] Implement FR-BILL-001 subscription schema from `docs/feature-requests/bill/FR-BILL-001-subscription-schema.md` and audit `docs/feature-requests/bill/FR-BILL-001-subscription-schema.audit.md` in `services/bill/`
- [ ] T078 [P] [US2] Implement FR-CART-005 auto-test codes from `docs/feature-requests/cart/FR-CART-005-auto-test-codes.md` and audit `docs/feature-requests/cart/FR-CART-005-auto-test-codes.audit.md` in `services/cart/` and `extension/`
- [ ] T079 [P] [US2] Implement FR-EXT-007 TikTok Shop content script from `docs/feature-requests/ext/FR-EXT-007-tiktok-shop-content-script.md` and audit `docs/feature-requests/ext/FR-EXT-007-tiktok-shop-content-script.audit.md` in `extension/src/content/`
- [ ] T080 [P] [US2] Implement FR-EXT-008 Lazada content script from `docs/feature-requests/ext/FR-EXT-008-lazada-content-script.md` and audit `docs/feature-requests/ext/FR-EXT-008-lazada-content-script.audit.md` in `extension/src/content/`
- [ ] T081 [P] [US2] Implement FR-CART-006 coin checklist from `docs/feature-requests/cart/FR-CART-006-coin-checklist.md` and audit `docs/feature-requests/cart/FR-CART-006-coin-checklist.audit.md` in `services/cart/` and `extension/`
- [ ] T082 [US2] Review P2 foundation tasks by running each FR §5 test, checking affiliate/extension guardrails, updating statuses, and recording evidence in `specs/001-full-project-plan/evidence/p2-expansion.md`

### P2 Monetization and Cart Implementation

- [ ] T083 [P] [US2] Implement FR-AFFIL-003 network integration from `docs/feature-requests/affil/FR-AFFIL-003-network-integration.md` and audit `docs/feature-requests/affil/FR-AFFIL-003-network-integration.audit.md` in `services/affil/`
- [ ] T084 [P] [US2] Implement FR-AFFIL-004 Honey avoidance guardrails from `docs/feature-requests/affil/FR-AFFIL-004-honey-avoidance-guardrails.md` and audit `docs/feature-requests/affil/FR-AFFIL-004-honey-avoidance-guardrails.audit.md` in `services/affil/` and `tests/compliance/`
- [ ] T085 [P] [US2] Implement FR-BILL-002 payment gateway from `docs/feature-requests/bill/FR-BILL-002-payment-gateway.md` and audit `docs/feature-requests/bill/FR-BILL-002-payment-gateway.audit.md` in `services/bill/`
- [ ] T086 [P] [US2] Implement FR-CART-002 cart snapshot schema from `docs/feature-requests/cart/FR-CART-002-cart-snapshot-schema.md` and audit `docs/feature-requests/cart/FR-CART-002-cart-snapshot-schema.audit.md` in `services/cart/`
- [ ] T087 [P] [US2] Implement FR-NOTIF-005 APNs dispatcher from `docs/feature-requests/notif/FR-NOTIF-005-apns-dispatcher.md` and audit `docs/feature-requests/notif/FR-NOTIF-005-apns-dispatcher.audit.md` in `services/notif/`
- [ ] T088 [P] [US2] Implement FR-NOTIF-006 email dispatcher from `docs/feature-requests/notif/FR-NOTIF-006-email-dispatcher.md` and audit `docs/feature-requests/notif/FR-NOTIF-006-email-dispatcher.audit.md` in `services/notif/`
- [ ] T089 [P] [US2] Implement FR-BILL-004 referral code from `docs/feature-requests/bill/FR-BILL-004-referral-code.md` and audit `docs/feature-requests/bill/FR-BILL-004-referral-code.audit.md` in `services/bill/`
- [ ] T090 [P] [US2] Implement FR-BILL-005 upgrade triggers gating from `docs/feature-requests/bill/FR-BILL-005-upgrade-triggers-gating.md` and audit `docs/feature-requests/bill/FR-BILL-005-upgrade-triggers-gating.audit.md` in `services/bill/`
- [ ] T091 [P] [US2] Implement FR-NOTIF-007 SMS dispatcher VN from `docs/feature-requests/notif/FR-NOTIF-007-sms-dispatcher-vn.md` and audit `docs/feature-requests/notif/FR-NOTIF-007-sms-dispatcher-vn.audit.md` in `services/notif/`
- [ ] T092 [P] [US2] Implement FR-BILL-003 payment reconciliation from `docs/feature-requests/bill/FR-BILL-003-payment-reconciliation.md` and audit `docs/feature-requests/bill/FR-BILL-003-payment-reconciliation.audit.md` in `services/bill/`
- [ ] T093 [P] [US2] Implement FR-CART-003 cart voucher optimizer from `docs/feature-requests/cart/FR-CART-003-cart-voucher-optimizer.md` and audit `docs/feature-requests/cart/FR-CART-003-cart-voucher-optimizer.audit.md` in `services/cart/`
- [ ] T094 [P] [US2] Implement FR-SCRAPE-007 TikTok Shop adapter from `docs/feature-requests/scrape/FR-SCRAPE-007-tiktok-shop-adapter.md` and audit `docs/feature-requests/scrape/FR-SCRAPE-007-tiktok-shop-adapter.audit.md` in `services/scrape/`
- [ ] T095 [P] [US2] Implement FR-SCRAPE-008 Lazada adapter from `docs/feature-requests/scrape/FR-SCRAPE-008-lazada-adapter.md` and audit `docs/feature-requests/scrape/FR-SCRAPE-008-lazada-adapter.audit.md` in `services/scrape/`
- [ ] T096 [P] [US2] Implement FR-CART-004 per-country stacking rules from `docs/feature-requests/cart/FR-CART-004-per-country-stacking-rules.md` and audit `docs/feature-requests/cart/FR-CART-004-per-country-stacking-rules.audit.md` in `services/cart/`
- [ ] T097 [P] [US2] Implement FR-DEAL-004 bottom price Prophet baseline from `docs/feature-requests/deal/FR-DEAL-004-bottom-price-prophet.md` and audit `docs/feature-requests/deal/FR-DEAL-004-bottom-price-prophet.audit.md` in `ml/deal/`
- [ ] T098 [P] [US2] Implement FR-DEAL-006 nightly scoring alert from `docs/feature-requests/deal/FR-DEAL-006-nightly-scoring-alert.md` and audit `docs/feature-requests/deal/FR-DEAL-006-nightly-scoring-alert.audit.md` in `services/deal/` and `ml/deal/`
- [ ] T099 [P] [US2] Implement FR-DEAL-005 LightGBM bottom price model from `docs/feature-requests/deal/FR-DEAL-005-bottom-price-lightgbm.md` and audit `docs/feature-requests/deal/FR-DEAL-005-bottom-price-lightgbm.audit.md` in `ml/deal/`
- [ ] T100 [US2] Run P2 expansion exit gate from `specs/001-full-project-plan/quickstart.md` and record output in `specs/001-full-project-plan/evidence/p2-expansion.md`

**Checkpoint**: User Story 2 complete. Multi-platform monetizable product can be reviewed independently.

---

## Phase 5: User Story 3 - Complete Growth and Compliance Surface (Priority: P3)

**Goal**: Add cashback, B2B anonymized analytics, mobile app, SEA compliance, and anti-fraud.

**Independent Test**: Complete all P3 tasks below, then run the P3 Growth scenario in `specs/001-full-project-plan/quickstart.md`.

### P3 Analytics, Mobile, Trust, and SEA Compliance

- [ ] T101 [P] [US3] Implement FR-COMPLY-008 VN ecommerce law from `docs/feature-requests/comply/FR-COMPLY-008-vn-ecommerce-law.md` and audit `docs/feature-requests/comply/FR-COMPLY-008-vn-ecommerce-law.audit.md` in `services/comply/`
- [ ] T102 [P] [US3] Implement FR-B2B-001 anonymized trend pipeline from `docs/feature-requests/b2b/FR-B2B-001-anonymized-trend-pipeline.md` and audit `docs/feature-requests/b2b/FR-B2B-001-anonymized-trend-pipeline.audit.md` in `services/b2b/`
- [ ] T103 [P] [US3] Implement FR-MOBILE-001 mobile scaffold from `docs/feature-requests/mobile/FR-MOBILE-001-mobile-scaffold.md` and audit `docs/feature-requests/mobile/FR-MOBILE-001-mobile-scaffold.audit.md` in `mobile/`
- [ ] T104 [P] [US3] Implement FR-TRUST-004 anti-fraud engine from `docs/feature-requests/trust/FR-TRUST-004-anti-fraud-engine.md` and audit `docs/feature-requests/trust/FR-TRUST-004-anti-fraud-engine.audit.md` in `services/trust/`
- [ ] T105 [P] [US3] Implement FR-B2B-002 B2B insights reports from `docs/feature-requests/b2b/FR-B2B-002-b2b-insights-reports.md` and audit `docs/feature-requests/b2b/FR-B2B-002-b2b-insights-reports.audit.md` in `services/b2b/`
- [ ] T106 [P] [US3] Implement FR-B2B-003 seller competitor analytics from `docs/feature-requests/b2b/FR-B2B-003-seller-competitor-analytics.md` and audit `docs/feature-requests/b2b/FR-B2B-003-seller-competitor-analytics.audit.md` in `services/b2b/`
- [ ] T107 [P] [US3] Implement FR-B2B-004 premium API access from `docs/feature-requests/b2b/FR-B2B-004-premium-api-access.md` and audit `docs/feature-requests/b2b/FR-B2B-004-premium-api-access.audit.md` in `services/b2b/`
- [ ] T108 [P] [US3] Implement FR-MOBILE-003 deeplink virality from `docs/feature-requests/mobile/FR-MOBILE-003-deeplink-virality.md` and audit `docs/feature-requests/mobile/FR-MOBILE-003-deeplink-virality.audit.md` in `mobile/`
- [ ] T109 [P] [US3] Implement FR-TRUST-005 attribution gaming detection from `docs/feature-requests/trust/FR-TRUST-005-attribution-gaming-detection.md` and audit `docs/feature-requests/trust/FR-TRUST-005-attribution-gaming-detection.audit.md` in `services/trust/`
- [ ] T110 [P] [US3] Implement FR-MOBILE-002 mobile tracking checkout from `docs/feature-requests/mobile/FR-MOBILE-002-mobile-tracking-checkout.md` and audit `docs/feature-requests/mobile/FR-MOBILE-002-mobile-tracking-checkout.audit.md` in `mobile/`
- [ ] T111 [P] [US3] Implement FR-TRUST-006 device fingerprint multiaccount from `docs/feature-requests/trust/FR-TRUST-006-device-fingerprint-multiaccount.md` and audit `docs/feature-requests/trust/FR-TRUST-006-device-fingerprint-multiaccount.audit.md` in `services/trust/`
- [ ] T112 [P] [US3] Implement FR-COMPLY-006 per-country gating from `docs/feature-requests/comply/FR-COMPLY-006-per-country-gating.md` and audit `docs/feature-requests/comply/FR-COMPLY-006-per-country-gating.audit.md` in `services/comply/`
- [ ] T113 [P] [US3] Implement FR-AFFIL-005 cashback layering from `docs/feature-requests/affil/FR-AFFIL-005-cashback-layering.md` and audit `docs/feature-requests/affil/FR-AFFIL-005-cashback-layering.audit.md` in `services/affil/`
- [ ] T114 [P] [US3] Implement FR-COMPLY-007 SEA data protection from `docs/feature-requests/comply/FR-COMPLY-007-sea-data-protection.md` and audit `docs/feature-requests/comply/FR-COMPLY-007-sea-data-protection.audit.md` in `services/comply/`
- [ ] T115 [US3] Run P3 growth exit gate from `specs/001-full-project-plan/quickstart.md` and record output in `specs/001-full-project-plan/evidence/p3-growth.md`

**Checkpoint**: User Story 3 complete. Full product can be reviewed independently.

---

## Phase 6: NFR Validation and Cross-Cutting Release Hardening

**Purpose**: Verify the 10 NFRs and perform final project review.

- [ ] T116 [P] Validate NFR-INFRA-001 API performance from `docs/non-functional-requirements/infra/NFR-INFRA-001-api-performance.md` and audit `docs/non-functional-requirements/infra/NFR-INFRA-001-api-performance.audit.md` in `tests/performance/`
- [ ] T117 [P] Validate NFR-INFRA-002 availability SLA from `docs/non-functional-requirements/infra/NFR-INFRA-002-availability-sla.md` and audit `docs/non-functional-requirements/infra/NFR-INFRA-002-availability-sla.audit.md` in `tests/performance/`
- [ ] T118 [P] Validate NFR-COMPLY-001 PDPL compliance from `docs/non-functional-requirements/comply/NFR-COMPLY-001-pdpl-compliance.md` and audit `docs/non-functional-requirements/comply/NFR-COMPLY-001-pdpl-compliance.audit.md` in `tests/compliance/`
- [ ] T119 [P] Validate NFR-AFFIL-001 affiliate compliance from `docs/non-functional-requirements/affil/NFR-AFFIL-001-affiliate-compliance.md` and audit `docs/non-functional-requirements/affil/NFR-AFFIL-001-affiliate-compliance.audit.md` in `tests/compliance/`
- [ ] T120 [P] Validate NFR-SCRAPE-001 anti-bot resilience from `docs/non-functional-requirements/scrape/NFR-SCRAPE-001-anti-bot-resilience.md` and audit `docs/non-functional-requirements/scrape/NFR-SCRAPE-001-anti-bot-resilience.audit.md` in `tests/performance/`
- [ ] T121 [P] Validate NFR-SCRAPE-002 scraping cost from `docs/non-functional-requirements/scrape/NFR-SCRAPE-002-scraping-cost.md` and audit `docs/non-functional-requirements/scrape/NFR-SCRAPE-002-scraping-cost.audit.md` in `tests/performance/`
- [ ] T122 [P] Validate NFR-PRICE-001 timeseries scale from `docs/non-functional-requirements/price/NFR-PRICE-001-timeseries-scale.md` and audit `docs/non-functional-requirements/price/NFR-PRICE-001-timeseries-scale.audit.md` in `tests/performance/`
- [ ] T123 [P] Validate NFR-EXT-001 Manifest V3 constraints from `docs/non-functional-requirements/ext/NFR-EXT-001-manifest-v3-constraints.md` and audit `docs/non-functional-requirements/ext/NFR-EXT-001-manifest-v3-constraints.audit.md` in `tests/compliance/`
- [ ] T124 [P] Validate NFR-NOTIF-001 midnight spike scale from `docs/non-functional-requirements/notif/NFR-NOTIF-001-midnight-spike-scale.md` and audit `docs/non-functional-requirements/notif/NFR-NOTIF-001-midnight-spike-scale.audit.md` in `tests/performance/`
- [ ] T125 [P] Validate NFR-TRUST-001 security trust from `docs/non-functional-requirements/trust/NFR-TRUST-001-security-trust.md` and audit `docs/non-functional-requirements/trust/NFR-TRUST-001-security-trust.audit.md` in `tests/compliance/`
- [ ] T126 Run the full project completion checklist from `specs/001-full-project-plan/quickstart.md` and record final evidence in `specs/001-full-project-plan/evidence/final-release.md`
- [ ] T127 Verify all 90 FR and 10 NFR are `done` or explicitly `on_hold`/`closed` with rationale in `docs/feature-requests/BACKLOG.md` and `specs/001-full-project-plan/execution-ledger.md`
- [ ] T128 Verify no SHIP-GUIDE invariant violation remains open and record result in `specs/001-full-project-plan/evidence/final-release.md`
- [ ] T129 Prepare final release notes in `specs/001-full-project-plan/release-notes.md` with included FRs, deferred items, test summary, migration version, and compliance notes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 Setup**: No dependencies.
- **Phase 2 Foundational**: Depends on Phase 1.
- **US1 MVP**: Depends on Phase 2. Within US1, execute Layer 0 -> Layer 1 -> Layer 2 -> Layer 3 -> Layer 4-6 MVP tasks.
- **US2 Expansion**: Depends on US1 checkpoint and the individual FR dependencies listed in `docs/feature-requests/IMPLEMENTATION-ORDER.md`.
- **US3 Growth**: Depends on US2 checkpoint and the individual FR dependencies listed in `docs/feature-requests/IMPLEMENTATION-ORDER.md`.
- **NFR and Final Release**: Depends on all desired FR tasks being complete.

### User Story Dependencies

- **User Story 1 (P1)**: MVP first. Delivers P0/P1 and proves core product loop.
- **User Story 2 (P2)**: Starts after US1. Adds multi-platform monetization.
- **User Story 3 (P3)**: Starts after US2. Adds growth, SEA and anti-fraud surfaces.

### Per-FR Done Rule

Before ticking any FR task:

1. Read `docs/feature-requests/SHIP-GUIDE.md`.
2. Read the FR file and companion `.audit.md`.
3. Read relevant `docs/non-functional-requirements/` files.
4. Implement only within the FR's declared `new_files` and `modified_files` unless a deviation is recorded.
5. Run tests from FR §5.
6. Verify §1 normative statements trace to §4 acceptance criteria.
7. Update FR frontmatter status and `docs/feature-requests/BACKLOG.md`.
8. Add evidence to the phase evidence file.

## Parallel Opportunities

- T005-T008 can run in parallel after T001-T004.
- T017-T020 can run in parallel.
- Tasks marked `[P]` inside the same dependency layer can run in parallel if they touch different files and their FR dependencies are done.
- NFR validation tasks T116-T125 can run in parallel after the related FR implementations are complete.

## Parallel Example: Layer 0

```text
Task: "T017 Implement FR-EXT-001 in extension/"
Task: "T018 Implement FR-INFRA-001 in services/gateway/"
Task: "T019 Implement FR-INFRA-002 in services/infra/ and deploy/migrations/"
Task: "T020 Implement FR-INFRA-003 in services/infra/ and deploy/secrets/"
```

## Implementation Strategy

### MVP First

1. Complete T001-T016.
2. Complete T017-T073.
3. Stop and review P0/P1 evidence before starting P2.

### Incremental Delivery

1. P0/P1 MVP: T017-T073.
2. P2 expansion: T074-T100.
3. P3 growth: T101-T115.
4. Final NFR/release: T116-T129.

### Review Control

Use `specs/001-full-project-plan/execution-ledger.md` as the handoff board. Another AI can take one unchecked task, implement it, attach evidence, and stop at each checkpoint for review.
