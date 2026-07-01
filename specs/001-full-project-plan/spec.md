# Feature Specification: SănDeal Full Project Plan

**Feature Branch**: `001-full-project-plan`

**Created**: 2026-06-30

**Status**: Draft

**Input**: User description: "dùng skill spec kit, lập plan hoàn thành full dự án này đi đã có docs sẵn r '/home/jay/Desktop/shopass/docs' '/home/jay/Desktop/shopass/AGENTS.md' '/home/jay/Desktop/shopass/README.md'"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ship MVP value loop (Priority: P1)

Một agent hoặc kỹ sư có thể đi từ repo chỉ có docs đến MVP SănDeal chạy được: extension Shopee đọc dữ liệu tối thiểu, backend ghi lịch sử giá, phát hiện sale ảo, web hiển thị biểu đồ, wishlist và push alert.

**Why this priority**: Đây là vòng giá trị lõi được nêu trong README và BACKLOG: theo dõi giá, sale thật/ảo, biểu đồ, alert. Nếu chưa có vòng này thì các module P2/P3 chưa có nền dữ liệu và chưa chứng minh sản phẩm.

**Independent Test**: Hoàn thành toàn bộ FR P0 và P1 MUST theo thứ tự topo, chạy test §5 của từng FR, và demo được Shopee product -> price_snapshot delta-only -> price_daily -> chart/alert.

**Acceptance Scenarios**:

1. **Given** repo chưa có code và 90 FR ở `ready_to_implement`, **When** team build các layer P0/P1 theo `IMPLEMENTATION-ORDER.md`, **Then** các FR P0/P1 MUST đi tới `done` và gateway, auth, price, scrape, extension, web, notification, comply, trust cùng chạy qua quickstart MVP.
2. **Given** sản phẩm Shopee đã có dữ liệu giá, **When** user mở web hoặc extension, **Then** hệ thống hiển thị lịch sử giá từ `price_daily`, phân loại sale ảo hoặc sale thật, và cho phép tạo alert.

---

### User Story 2 - Expand to monetizable multi-platform product (Priority: P2)

Một agent hoặc kỹ sư có thể mở rộng MVP sang TikTok Shop, Lazada, tối ưu voucher/giỏ hàng, Premium billing và affiliate compliant.

**Why this priority**: Đây là phase mở rộng doanh thu và moat chéo sàn. P2 chỉ nên bắt đầu khi P1 đã đủ dữ liệu, đủ niềm tin, và đủ compliance.

**Independent Test**: Hoàn thành P2 MUST theo dependencies, chạy contract tests cho content scripts 3 sàn, cart optimizer, payment webhooks, affiliate link user-initiated và notification channels.

**Acceptance Scenarios**:

1. **Given** P1 đã `done`, **When** user duyệt Shopee/TikTok Shop/Lazada, **Then** extension chỉ gửi productId/price/qty hoặc cart snapshot tối thiểu, không gửi cookie/token.
2. **Given** user chủ động bấm mua qua SănDeal, **When** hệ thống tạo affiliate deep link, **Then** disclosure hiện rõ, click được ghi bằng `sub_id` không PII, và không có cookie-stuffing/auto-redirect.
3. **Given** user cần tối ưu giỏ, **When** optimizer chạy, **Then** luật stacking theo quốc gia được áp dụng qua CountryPolicy, mặc định restrictive cho nước chưa cấu hình.

---

### User Story 3 - Complete growth and compliance surface (Priority: P3)

Một agent hoặc kỹ sư có thể hoàn tất cashback, B2B anonymized analytics, mobile app, per-country SEA compliance và anti-fraud.

**Why this priority**: P3 tăng trưởng nhưng có blast radius lớn về fraud, data protection và vận hành đa quốc gia, nên phải nằm sau foundation compliance và billing.

**Independent Test**: Hoàn thành P3 theo dependencies, chạy các scenario k-anonymity, cashback hold-then-release, mobile push/auth, country gating và anti-fraud.

**Acceptance Scenarios**:

1. **Given** conversion affiliate đã confirmed, **When** cashback được tạo, **Then** payout chỉ available sau hold window và có hook anti-fraud.
2. **Given** B2B report được xuất, **When** category/day không đạt K_MIN, **Then** dữ liệu bị suppress và không chứa product_id/shop_id/user_id.
3. **Given** user ở country chưa cấu hình rõ, **When** checkout/gating logic chạy, **Then** hệ thống dùng mặc định restrictive.

### Edge Cases

- Scraping bị anti-bot, CAPTCHA, DOM drift hoặc proxy cost spike: orchestrator phải backoff, ghi health signal, không làm hỏng price history, và không vượt cost guard.
- Chưa đủ 90 ngày lịch sử giá: deal engine trả `UNKNOWN` hoặc dùng category priors theo FR-DEAL-002, không tuyên bố sale thật/ảo chắc chắn.
- Consent bị từ chối hoặc rút lại: các luồng xử lý dữ liệu cá nhân tương ứng bị chặn, DSAR vẫn chạy được.
- Payment webhook trùng hoặc lệch tiền: reconciliation idempotent, status `mismatch` hoặc `failed`, không cấp Premium sai.
- Notification spike lúc 00:00: scheduler jitter/bucket để flatten-the-curve, FCM 429 phải backoff.
- Country unsupported: áp dụng no-stack/restrictive policy, không mở rộng behavior lạc quan.
- Extension service worker bị suspend: state nằm trong `chrome.storage`, không phụ thuộc biến global.
- Các FR song song cùng cần một table: chỉ FR owner tạo table, FR khác FK hoặc ALTER theo DATA-MODEL.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Project MUST be implemented as the 90 FR + 10 NFR backlog under `docs/feature-requests/` and `docs/non-functional-requirements/`; each original FR remains the atomic implementation and test unit.
- **FR-002**: Delivery MUST follow `docs/feature-requests/IMPLEMENTATION-ORDER.md` layer order; a FR may start only when all `depends_on` are `done`, except explicit human override recorded in backlog or issue.
- **FR-003**: Every FR implementation MUST read its own FR file, companion `.audit.md`, `SHIP-GUIDE.md`, `DATA-MODEL.md`, and relevant NFRs before code changes.
- **FR-004**: Every MUST/SHOULD statement in §1 of each FR MUST trace to acceptance criteria §4 and passing tests §5 before that FR can become `done`.
- **FR-005**: The repository MUST grow into the documented product structure: `services/`, `extension/`, `web/`, `mobile/`, deployment/infra files, and tests, using each FR's `new_files`/`modified_files`.
- **FR-006**: Backend services MUST use Go 1.22, PostgreSQL 16, TimescaleDB for price time-series, Redis/Kafka-Redis Streams for queues, and OTel/Prometheus/Grafana observability unless the FR specifies a narrower choice.
- **FR-007**: Deal ML MUST use Python with Prophet baseline and LightGBM follow-up as specified by FR-DEAL-004/005.
- **FR-008**: Extension MUST use TypeScript and Chrome Manifest V3 rules: ephemeral service worker, `chrome.storage`, `chrome.alarms >= 30s`, Offscreen API where needed, and no cookie/token exfiltration.
- **FR-009**: Web MUST use Next.js + TypeScript App Router and GraphQL BFF where specified.
- **FR-010**: Mobile SHOULD be implemented after MVP/P2 dependencies using React Native or Flutter, with final choice recorded before FR-MOBILE-001 starts.
- **FR-011**: All money columns MUST be BIGINT in VND, never float/numeric for prices or commission values unless an existing audited FR explicitly says otherwise for a non-money analytic percentage.
- **FR-012**: `price_snapshot` MUST be delta-only TimescaleDB hypertable, compressed after the configured window, and chart/history reads MUST use `price_daily`.
- **FR-013**: Platform session tokens, cookies, passwords, and secrets MUST NOT be stored cleartext or sent from extension to server; passwords use argon2id PHC, secrets use Vault/AWS Secrets Manager references.
- **FR-014**: Affiliate links MUST be generated only after explicit user action with disclosure; cookie-stuffing, pop-under, auto-redirect, and extension-driven affiliate injection are forbidden.
- **FR-015**: PDPL consent, DPIA/TIA, DSAR, breach notification, and per-country gating MUST be implemented before any dependent data-processing or country expansion feature is released.
- **FR-016**: Notification delivery MUST prioritize push > email > sms and protect midnight spikes with jitter, buckets, quota backoff, DLQ and idempotent fan-out.
- **FR-017**: Scraping MUST use residential proxy, pacing/jitter, adapter health checks and DOM drift monitoring; automation of coins/vouchers MUST stay user-initiated or checklist-only as specified.
- **FR-018**: Status changes MUST follow `STATUS-REFERENCE.md` and update both FR frontmatter and `BACKLOG.md` when a FR moves through lifecycle.
- **FR-019**: Release gates MUST be evaluated at P0, P1, P2 and P3 exits using the quickstart validation scenarios in this spec package.
- **FR-020**: The implementation plan MUST include operational work for local dev, CI, migrations, secrets bootstrap, observability, compliance evidence, and production readiness.

### Key Entities *(include if feature involves data)*

- **Implementation Program**: The full SănDeal delivery initiative, containing phases P0-P3, release gates, risks, and completion evidence.
- **Feature Request**: Atomic implementation unit from docs, with id, module, phase, priority, dependencies, status, effort, new files, tests and acceptance criteria.
- **Module**: Product or service area: infra, auth, scrape, price, ext, track, deal, notif, web, comply, trust, cart, affil, bill, b2b, mobile.
- **Release Phase**: P0 foundation, P1 MVP, P2 expansion, P3 growth.
- **Dependency Layer**: Topological layer from `IMPLEMENTATION-ORDER.md`; all FRs in a layer can be parallelized after prior layers are done.
- **Gate Evidence**: Test reports, contract results, migration status, observability dashboards, compliance records and release notes proving a phase can exit.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 90/90 FR and 10/10 NFR have implementation evidence and terminal status `done` or explicitly approved `closed/on_hold` with rationale.
- **SC-002**: P1 demo completes in one run: track Shopee product, ingest price, write delta-only snapshot, refresh daily aggregate, show chart, detect deal status, create alert, dispatch push.
- **SC-003**: P2 demo completes in one run: read cart on 3 platforms, optimize voucher stack with country rules, create user-initiated affiliate link, process Premium payment webhook.
- **SC-004**: P3 demo completes in one run: cashback hold/release, anonymized B2B trend export with K_MIN suppression, mobile auth/push, country gate for at least one SEA market.
- **SC-005**: Contract, unit and integration tests referenced by every implemented FR pass in CI before release.
- **SC-006**: No compliance invariant violations are present in security/compliance audit: no server-side platform tokens, no cleartext secrets, no non-consensual processing, no affiliate injection.
- **SC-007**: API performance and time-series paths meet documented NFR SLOs, including chart/history p95 targets and notification spike handling.
- **SC-008**: The critical path data collection begins as soon as FR-SCRAPE-002 and FR-SCRAPE-005 are ready, so sale-realness history can accumulate during UI work.

## Assumptions

- The existing docs are authoritative and audited; this Spec Kit package coordinates delivery rather than rewriting individual FR specs.
- `AGENTS.md` is currently a placeholder for CyberOS memory protocol; build conventions are sourced from `SHIP-GUIDE.md`.
- The repo intentionally starts without code; service directories will be created by FR implementations.
- VN is first release market; SEA expansion follows per-country gating in P3.
- Human review is optional but status and evidence must still be recorded.
- Mobile framework final selection remains open until FR-MOBILE-001; React Native is the default planning assumption because the repo already uses TypeScript for extension/web.
