# SănDeal - Task Backlog

**Owner:** Stephen Cheng (Founder / CEO) - **Status:** v0.2.1 - statuses aligned with docs/TASK-COVERAGE.md. 90 task + 90 audit + 10 NFR + 10 audit; DAG acyclic + reciprocal; data model nhất quán (một table một owner). Tasks with no code on disk are `ready_to_implement` (not `done`); see TASK-COVERAGE for the 16 gaps. Spec audit scores (10/10) mean the *spec* passed task-audit, not that implementation shipped. Backlog khởi tạo từ tài liệu nền tảng SănDeal v1.0 (16/06/2026), áp dụng workflow task của CyberOS (engineering-spec@1). **Nguồn sự thật (source of truth):** các file markdown trong thư mục này. Index này được tái tạo khi task được thêm hoặc đổi status. **Tài liệu nguồn:** [`../TÀI LIỆU NỀN TẢNG SẢN PHẨM "SănDeal" - PRD + SRS + CHIẾN LƯỢC KỸ THUẬT : KINH DOANH : RỦI RO.md`](../) **Tài liệu hỗ trợ ship (cho agent triển khai):** [`SHIP-GUIDE.md`](SHIP-GUIDE.md) (conventions + bất biến build) | [`IMPLEMENTATION-ORDER.md`](IMPLEMENTATION-ORDER.md) (thứ tự build theo layer) | [`DATA-MODEL.md`](DATA-MODEL.md) (schema hợp nhất) | [`STATUS-REFERENCE.md`](STATUS-REFERENCE.md) (vòng đời status). Ghi chú: `AGENTS.md` ở gốc repo dành cho giao thức memory CyberOS (BRAIN); conventions build nằm ở SHIP-GUIDE.md. **Playbook tác giả:** workflow `task-author` + `task-audit` của CyberOS (`cyberos/modules/skill/contracts/task/`). Mỗi task đi kèm một file `.audit.md`. **Status enum (10 trạng thái):** `draft | done | implementing | ready_to_review | reviewing | ready_to_test | testing | done | on_hold | closed` (theo [`STATUS-REFERENCE.md`](STATUS-REFERENCE.md)).

---

## §0 - Cách đọc backlog này

Tài liệu này là **nguồn sự thật duy nhất** cho những gì SănDeal sẽ xây, tổ chức theo **phase** (P0 -> P3), rồi theo **module**, rồi theo **slice** trong mỗi module. Mỗi dòng là một task; một task là một yêu cầu nguyên tử, kiểm thử được.

- **Phase** map vào roadmap §8 của tài liệu nguồn. `P0 Nền tảng` dựng hạ tầng xuyên suốt; `P1 MVP` ship extension Shopee + cold-start dữ liệu + SEO + sale ảo; `P2 Mở rộng` thêm TikTok Shop + Lazada + cart optimizer + Premium; `P3 Tăng trưởng` thêm cashback + B2B + mobile + SEA.
- **Slice** là một đơn vị ship gọn trong một module. Slice 1 luôn là bề mặt tối thiểu khả dụng (MVP của module đó).
- **Priority** dùng từ khóa BCP-14: `MUST` (chặn release) - `SHOULD` (nên có) - `COULD` (tốt nếu có) - `MAY` (sau release).
- **Status**: trạng thái hiện tại của task (xem enum 10 trạng thái ở header).
- **Depends on**: danh sách phụ thuộc cross-FR.
- **Effort**: ước lượng thô theo giờ (1h = 30 phút làm tập trung + 30 phút phối hợp/review). Sai số +/-50%. Tính cho một kỹ sư có kinh nghiệm.

**Thứ tự đọc cho người lập kế hoạch:** quét §1 (tổng) -> chọn phase đang làm -> đọc phần phân rã theo module trong phase đó -> đào vào từng file task. **Thứ tự đọc cho người triển khai:** tìm TASK-ID được giao trong phần module -> mở file task markdown để xem chi tiết.

---

## §1 - Tổng quan

| Phase | Module trong phạm vi | Số task | Effort (giờ) | Cổng kiểm tra khi thoát phase |
|---|---|---:|---:|---|
| **P0 - Nền tảng** | INFRA | 5 | ~33 | API Gateway live, data-model migrate được, OTel/Grafana lên, secrets trong Vault |
| **P1 - MVP** | AUTH - SCRAPE - PRICE - EXT - TRACK - DEAL - NOTIF - WEB - COMPLY - TRUST | 46 | ~303 | Extension Shopee đọc giỏ hàng, >=90 ngày lịch sử giá top-SKU, sale ảo + biểu đồ live, push alert, PDPL consent + DPIA nộp, extension open-source |
| **P2 - Mở rộng** | EXT - SCRAPE - CART - AFFIL - BILL - DEAL - NOTIF | 25 | ~177 | TikTok Shop + Lazada đọc được, cart/voucher optimizer đúng stacking 3 nước, Premium thu được tiền, affiliate user-initiated compliant, dự đoán đáy giá ML |
| **P3 - Tăng trưởng** | AFFIL - B2B - MOBILE - COMPLY - TRUST | 14 | ~114 | Cashback hold-then-payout, B2B data ẩn danh bán được, mobile app live, per-country gating cho >=1 nước SEA, anti-fraud engine chặn farming |
| **Tổng** | 16 module - 4 phase | **90** | **~627 giờ (~16 person-week)** | 4 cổng compliance/kỹ thuật |

> Lưu ý: 90 task nằm đúng đỉnh khoảng "~70-90" ước lượng ban đầu. Tài liệu nguồn map ra 16 module với nhiều slice per-sàn (3 sàn x extension + scraping) và 3 dòng doanh thu (affiliate, Premium, B2B). Đây là backlog đầy đủ 3 phase theo yêu cầu.

**Kiểm tra ngân sách effort:** 90 task x ~7h trung bình ~ 627h ~ 15,7 person-week thuần code. Cộng design + review + tích hợp + bất ngờ -> khoảng 9-12 tháng cho 1 kỹ sư full-time, hoặc 5-6 tháng cho 2 kỹ sư - khớp với roadmap 0-12 tháng ở §8 tài liệu nguồn.

---

## §2 - P0 - Nền tảng (INFRA)

**Mục tiêu phase:** dựng hạ tầng xuyên suốt mà mọi module khác phụ thuộc. Khi thoát P0: API Gateway/BFF định tuyến REST+GraphQL+WSS với rate-limit + JWT + WAF; data-model migrate được; OTel + Prometheus + Grafana cho một bề mặt điều tra sự cố; secrets trong Vault (không cleartext).

### P0.1 - INFRA - hạ tầng nền tảng

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-INFRA-001** | API Gateway / BFF - định tuyến REST+GraphQL+WSS, rate-limit, JWT verify, WAF | MUST | done | - | 8h |
| **TASK-INFRA-002** | Data-model foundation - migration framework + bảng `platform` + `app_user` lõi | MUST | done | - | 6h |
| **TASK-INFRA-003** | Secrets management - Vault / AWS Secrets Manager, no-cleartext, rotation | MUST | done | - | 5h |
| **TASK-INFRA-004** | Observability spine - Prometheus + Grafana + OTel tracing + structured logs | MUST | done | TASK-INFRA-001 | 8h |
| **TASK-INFRA-005** | Per-country region config - gating hook (VN/ID/TH/PH/MY/SG/TW) + feature flags | MUST | done | TASK-INFRA-002 | 6h |
| **TASK-INFRA-006** | Wire gateway into local compose; unpublish service ports; kill client X-User-Id trust (improvement) | MUST | done | TASK-INFRA-001, TASK-AUTH-002 | 6h |

---

## §3 - P1 - MVP (extension Shopee + cold-start + SEO + sale ảo)

**Mục tiêu phase:** ship được vòng giá trị lõi cho persona Chi/Huy/Linh trên Shopee: theo dõi giá, phát hiện sale ảo, biểu đồ giá, alert push. Tích lũy >=90 ngày lịch sử giá top-SKU (giải bài toán cold-start). Site SEO kéo traffic organic. PDPL compliant + extension open-source để xây niềm tin hậu-Honey.

**Cổng compliance:** PDPL (Luật 91/2025) consent + DPIA nộp trong 60 ngày - no-cleartext + token-not-on-server - extension open-source + disclosure Chrome Web Store.

**Phụ thuộc tới hạn:** mọi module P1 cần P0 xong (gateway, data-model, observability, secrets).

### P1.1 - AUTH - người dùng + liên kết tài khoản sàn

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-AUTH-001** | Schema `app_user` (argon2id pwd_hash, CITEXT email, phone, locale) + đăng ký | MUST | done | TASK-INFRA-002 | 6h |
| **TASK-AUTH-002** | Phát hành JWT + refresh + phiên (BFF auth) | MUST | done | TASK-AUTH-001, TASK-INFRA-001 | 6h |
| **TASK-AUTH-003** | Liên kết `platform_account` - ext_user_ref ẩn danh, KHÔNG lưu token phiên | MUST | done | TASK-AUTH-001 | 5h |
| **TASK-AUTH-004** | Social login (Google / Facebook / Zalo OAuth) | SHOULD | done | TASK-AUTH-002 | 6h |
| **TASK-AUTH-005** | Vòng đời tài khoản - reset mật khẩu, status, xóa tài khoản (DSAR PDPL) | MUST | done | TASK-AUTH-001, TASK-COMPLY-003 | 5h |

### P1.2 - SCRAPE - scraping giá (Shopee trước)

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-SCRAPE-001** | Scraping orchestrator lõi - scheduler + scan-frequency tiering (hot/thường) | MUST | done | TASK-INFRA-003, TASK-PRICE-001 | 10h |
| **TASK-SCRAPE-002** | Shopee internal-API adapter (`/api/v4/...`, hybrid, `is_login:false`) | MUST | done | TASK-SCRAPE-001 | 8h |
| **TASK-SCRAPE-003** | Playwright headless farm + anti-fingerprint (Canvas/WebGL/JA3/JA4/HTTP2) | MUST | done | TASK-SCRAPE-001 | 12h |
| **TASK-SCRAPE-004** | Residential proxy rotation + tiering + cost-guard (Bright Data/Oxylabs/Decodo/IPRoyal) | MUST | done | TASK-SCRAPE-003 | 8h |
| **TASK-SCRAPE-005** | Delta-only writes + pacing/jitter + CAPTCHA handling | MUST | done | TASK-SCRAPE-002, TASK-PRICE-002 | 6h |
| **TASK-SCRAPE-006** | DOM-change monitoring + adapter health (resilient A/B test DOM) | MUST | done | TASK-SCRAPE-002 | 6h |

### P1.3 - PRICE - time-series giá + so sánh chéo sàn

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-PRICE-001** | `tracked_product` registry chuẩn hóa (UNIQUE platform_id, item_id) | MUST | done | TASK-INFRA-002 | 6h |
| **TASK-PRICE-002** | `price_snapshot` TimescaleDB hypertable + nén + continuous aggregate | MUST | done | TASK-PRICE-001 | 8h |
| **TASK-PRICE-003** | API lịch sử giá (`GET /v1/products/{id}/price-history?range=90d`) | MUST | done | TASK-PRICE-002 | 5h |
| **TASK-PRICE-004** | So sánh giá chéo 3 sàn (`GET /v1/compare?canonical_key=...`) | MUST | done | TASK-PRICE-005 | 6h |
| **TASK-PRICE-005** | Thuật toán matching `canonical_key` (dedup sản phẩm chéo sàn) | MUST | done | TASK-PRICE-001 | 8h |

### P1.4 - EXT - browser extension MV3 (Shopee)

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-EXT-001** | Scaffold extension Manifest V3 + service worker ephemeral, chrome.alarms, host_permissions | MUST | done | - | 8h |
| **TASK-EXT-002** | Shopee content script đọc giỏ/voucher (session piggyback) | MUST | done | TASK-EXT-001 | 10h |
| **TASK-EXT-003** | Pipeline tối thiểu hóa dữ liệu client (chỉ productId/price/qty; KHÔNG cookie/token) | MUST | done | TASK-EXT-002 | 6h |
| **TASK-EXT-004** | Offscreen API cho DOM parsing/clipboard + declarativeNetRequest | SHOULD | done | TASK-EXT-001 | 5h |
| **TASK-EXT-005** | Đồng bộ extension <-> backend (auth bridge, WSS keep-alive khi cần realtime) | MUST | done | TASK-EXT-003, TASK-AUTH-002 | 6h |
| **TASK-EXT-006** | UI settings + consent (PDPL consent lúc cài, disclosure dữ liệu) | MUST | done | TASK-EXT-001, TASK-COMPLY-001 | 5h |

### P1.5 - TRACK - theo dõi sản phẩm + wishlist + alert rules

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-TRACK-001** | API theo dõi sản phẩm (`POST /v1/track {platform, item_url}`) | MUST | done | TASK-PRICE-001, TASK-SCRAPE-002 | 5h |
| **TASK-TRACK-002** | Schema + API `wishlist` / `wishlist_item` (target_price) | MUST | done | TASK-TRACK-001 | 5h |
| **TASK-TRACK-003** | Schema + API `alert_rule` (price_below/drop_pct/real_sale/bottom_predicted) | MUST | done | TASK-TRACK-001 | 6h |
| **TASK-TRACK-004** | Engine kích hoạt alert (đánh giá rule trên `price_snapshot`) | MUST | done | TASK-TRACK-003, TASK-PRICE-002, TASK-NOTIF-001 | 6h |

### P1.6 - DEAL - phát hiện sale ảo + cold-start

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-DEAL-001** | Phát hiện sale ảo (statistical: median90/p10/trailing_min -> SALE_AO/SALE_XIN/TAM_DUOC) | MUST | done | TASK-PRICE-002 | 8h |
| **TASK-DEAL-002** | Xử lý cold-start (category priors, <14d -> UNKNOWN, cổng baseline 90 ngày) | MUST | done | TASK-DEAL-001 | 6h |
| **TASK-DEAL-003** | API dữ liệu biểu đồ giá (daily aggregate cho chart, p95 <500ms) | MUST | done | TASK-PRICE-002 | 5h |

### P1.7 - NOTIF - thông báo (push trước)

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-NOTIF-001** | Schema notification + routing theo cost model (push > email > sms) | MUST | done | TASK-INFRA-002 | 6h |
| **TASK-NOTIF-002** | FCM Web/Android dispatcher (token mgmt, quota 600k/phút, 429 backoff) | MUST | done | TASK-NOTIF-001 | 8h |
| **TASK-NOTIF-003** | Fan-out pipeline (Kafka/Redis Streams -> workers -> per-channel) + DLQ | MUST | done | TASK-NOTIF-001 | 8h |
| **TASK-NOTIF-004** | Scheduler flatten-the-curve cho đỉnh 00:00 (jitter, bucketing, FCM rate-limit) | MUST | done | TASK-NOTIF-003 | 6h |

### P1.8 - WEB - web app + SEO

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-WEB-001** | Scaffold Next.js 14, cấu trúc shell UI, API client `lib/api.ts` (xử lý 401 refresh), middleware guard | MUST | done | TASK-AUTH-002, TASK-INFRA-001 | 8h |
| **TASK-WEB-002** | SEO landing page (SSG), Metadata API, JSON-LD (FAQ, Article, ItemList) + auto sitemap/robots | MUST | done | - | 6h |
| **TASK-WEB-003** | UI biểu đồ lịch sử giá (render p95 <500ms) | MUST | done | TASK-WEB-001, TASK-DEAL-003 | 6h |
| **TASK-WEB-004** | UI quản lý wishlist + alert | MUST | done | TASK-WEB-001, TASK-TRACK-002, TASK-TRACK-003 | 6h |
| **TASK-WEB-005** | GraphQL BFF cho web (truy vấn linh hoạt wishlist/biểu đồ) | SHOULD | done | TASK-INFRA-001, TASK-WEB-001 | 6h |

### P1.9 - COMPLY - PDPL nền tảng

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-COMPLY-001** | Khung consent PDPL (Luật 91/2025/QH15) - đơn mục đích, tái lập | MUST | done | TASK-INFRA-002 | 8h |
| **TASK-COMPLY-002** | Sổ đăng ký DPIA/TIA (nộp trong 60 ngày, cập nhật mỗi 6 tháng) | MUST | done | TASK-COMPLY-001 | 6h |
| **TASK-COMPLY-003** | Quyền chủ thể dữ liệu (truy cập/sửa/xóa/di chuyển - DSAR) | MUST | done | TASK-COMPLY-001 | 8h |
| **TASK-COMPLY-004** | Quy trình thông báo vi phạm 72 giờ | MUST | done | TASK-COMPLY-001, TASK-INFRA-004 | 5h |
| **TASK-COMPLY-005** | Cưỡng chế no-cleartext + token-not-on-server - audit gate CI | MUST | done | TASK-INFRA-003 | 5h |

### P1.10 - TRUST - niềm tin (open-source + audit)

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-TRUST-001** | Open-source extension + reproducible build + disclosure | MUST | done | TASK-EXT-001 | 6h |
| **TASK-TRUST-002** | Chính sách minh bạch tối thiểu hóa dữ liệu + xử lý local-first | MUST | done | TASK-EXT-003 | 5h |
| **TASK-TRUST-003** | Hook security audit độc lập (chứng minh không gửi cookie/mật khẩu) | MUST | done | TASK-EXT-003, TASK-COMPLY-005 | 6h |

---

## §4 - P2 - Mở rộng (TikTok Shop + Lazada + cart optimizer + Premium)

**Mục tiêu phase:** mở đa sàn thật (thêm TikTok Shop 41,31% GMV + Lazada), bật moat so sánh chéo 3 sàn, ship cart/voucher optimizer client-side, và bắt đầu thu tiền qua Premium + affiliate compliant.

**Cổng compliance:** affiliate chỉ user-initiated + disclosure (né Honey) - stacking voucher đúng luật per-country (VN stack / MY-PH no-stack 2025) - PCI-lite cho payment gateway.

### P2.1 - EXT - TikTok Shop + Lazada content scripts

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-EXT-007** | TikTok Shop content script (webview/SPA DOM reader, tránh API ký msToken/X-Bogus) | MUST | done | TASK-EXT-002 | 10h |
| **TASK-EXT-008** | Lazada content script (Akamai-aware, đọc DOM render) | MUST | done | TASK-EXT-002 | 8h |

### P2.2 - SCRAPE - TikTok Shop + Lazada adapters

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-SCRAPE-007** | TikTok Shop scraping adapter (ưu tiên DOM-render, né API ký) | MUST | done | TASK-SCRAPE-003 | 10h |
| **TASK-SCRAPE-008** | Lazada scraping adapter (Akamai, residential bắt buộc) | MUST | done | TASK-SCRAPE-003 | 8h |

### P2.3 - CART - voucher + tối ưu giỏ hàng

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-CART-001** | Schema + ingest `voucher_catalog` (shop/platform/freeship, stack_group, cap) | MUST | done | TASK-INFRA-002 | 6h |
| **TASK-CART-002** | Schema `cart_snapshot` + `cart_item` (nhận từ extension) | MUST | done | TASK-EXT-003 | 5h |
| **TASK-CART-003** | Optimizer giỏ/voucher/freeship (knapsack, ràng buộc stacking, applyCaps) | MUST | done | TASK-CART-001, TASK-CART-002 | 12h |
| **TASK-CART-004** | Engine luật stacking per-country (VN stack vs MY/PH bỏ stacking 2025) | MUST | done | TASK-CART-003, TASK-INFRA-005 | 6h |
| **TASK-CART-005** | `testCodes`: thử mã an toàn client-side (sleep nhịp người, user-initiated, revert) | MUST | done | TASK-EXT-002, TASK-CART-001 | 6h |
| **TASK-CART-006** | Checklist xu/coin (nhắc nhở, KHÔNG auto-click - chống abuse) | SHOULD | done | TASK-EXT-002 | 5h |

### P2.4 - AFFIL - affiliate compliant

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-AFFIL-001** | Schema + tracking affiliate (user-initiated) + đối soát mạng | MUST | done | TASK-INFRA-002 | 6h |
| **TASK-AFFIL-002** | Deep-link generator user-initiated (`POST /v1/affiliate/link`, disclosure, no auto-cookie) | MUST | done | TASK-AFFIL-001 | 6h |
| **TASK-AFFIL-003** | Tích hợp affiliate network (Involve Asia / Accesstrade) compliant | MUST | done | TASK-AFFIL-002 | 8h |
| **TASK-AFFIL-004** | Guardrails né Honey (no cookie-stuffing, bắt buộc user-action, tuân Chrome policy 10/06/2025) | MUST | done | TASK-AFFIL-002, TASK-EXT-003 | 5h |

### P2.5 - BILL - Premium + thanh toán

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-BILL-001** | Schema `subscription` + tier (Premium 29k/49k/79k) + vòng đời | MUST | done | TASK-AUTH-001 | 6h |
| **TASK-BILL-002** | Tích hợp cổng thanh toán (MoMo / ZaloPay / VNPay / VietQR) | MUST | done | TASK-BILL-001 | 10h |
| **TASK-BILL-003** | Bản ghi `payment` + reconciliation + webhook | MUST | done | TASK-BILL-002 | 6h |
| **TASK-BILL-004** | `referral_code` + attribution + hook chống abuse | SHOULD | done | TASK-BILL-001 | 5h |
| **TASK-BILL-005** | Trigger upgrade free->Premium (gamified) + feature gating | SHOULD | done | TASK-BILL-001 | 6h |

### P2.6 - DEAL - dự đoán đáy giá (ML)

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-DEAL-004** | Baseline dự đoán đáy (Prophet, regressor double-date/payday) | MUST | done | TASK-PRICE-002, TASK-DEAL-002 | 10h |
| **TASK-DEAL-005** | Model LightGBM (>=180d history, target future_min_price_14d) + feature store | SHOULD | done | TASK-DEAL-004 | 12h |
| **TASK-DEAL-006** | Batch scoring đêm + alert (P(bottom trong 14d) > 0.7) | MUST | done | TASK-DEAL-004, TASK-TRACK-003 | 6h |

### P2.7 - NOTIF - email + SMS + APNs

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-NOTIF-005** | APNs iOS dispatcher (xử lý 410, retry backoff 500/503) | MUST | done | TASK-NOTIF-003 | 5h |
| **TASK-NOTIF-006** | Email dispatcher (SES / SendGrid / Postmark) | MUST | done | TASK-NOTIF-003 | 4h |
| **TASK-NOTIF-007** | SMS dispatcher VN (SpeedSMS/eSMS/VietGuys + Twilio fallback, brandname, chỉ high-value) | SHOULD | done | TASK-NOTIF-003 | 6h |

---

## §5 - P3 - Tăng trưởng (cashback + B2B + mobile + SEA)

**Mục tiêu phase:** layer cashback trên affiliate, mở dòng doanh thu B2B margin cao (dữ liệu xu hướng giá ẩn danh), ra mobile app, và mở rộng SEA với per-country gating (ID/TH/PH/MY/SG/TW).

**Cổng compliance:** cashback hold-then-payout + delay payout chống gian lận - B2B data ẩn danh (k-anonymity) - per-country data-protection (Indonesia PDP, Thailand PDPA) - VN e-commerce law (NĐ 52/85, MOIT).

### P3.1 - AFFIL - cashback layering

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-AFFIL-005** | Cashback layering (chia % cho user, hold tới khi affiliate confirm, delay payout) | SHOULD | done | TASK-AFFIL-003, TASK-BILL-002, TASK-TRUST-005 | 10h |

### P3.2 - B2B - dữ liệu + analytics

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-B2B-001** | Pipeline dữ liệu xu hướng thị trường ẩn danh (k-anonymity, aggregate) | SHOULD | ready_to_review | TASK-PRICE-002, TASK-COMPLY-003 | 10h |
| **TASK-B2B-002** | Báo cáo B2B insights + subscription | SHOULD | ready_to_review | TASK-B2B-001, TASK-BILL-001 | 8h |
| **TASK-B2B-003** | Seller-facing competitor price analytics | COULD | ready_to_review | TASK-B2B-001 | 8h |
| **TASK-B2B-004** | Premium API access (tiers, rate-limited) | COULD | ready_to_review | TASK-INFRA-001, TASK-B2B-001 | 6h |

### P3.3 - MOBILE - mobile app + SEA virality

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-MOBILE-001** | Scaffold mobile (React Native/Flutter) + auth + push | SHOULD | ready_to_review | TASK-AUTH-002, TASK-NOTIF-002 | 12h |
| **TASK-MOBILE-002** | Mobile theo dõi giá + alert + universal checkout assistant | SHOULD | ready_to_review | TASK-MOBILE-001, TASK-CART-003 | 10h |
| **TASK-MOBILE-003** | Deep-link + share-on-sale virality + referral | COULD | ready_to_review | TASK-MOBILE-001, TASK-BILL-004 | 6h |

### P3.4 - COMPLY - per-country gating + SEA + e-commerce law

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-COMPLY-006** | Khung per-country gating (luật voucher/affiliate/dữ liệu theo nước) | MUST | done | TASK-INFRA-005, TASK-CART-004 | 8h |
| **TASK-COMPLY-007** | Adapter bảo vệ dữ liệu SEA (Indonesia PDP, Thailand PDPA) | SHOULD | done | TASK-COMPLY-001, TASK-COMPLY-006 | 8h |
| **TASK-COMPLY-008** | Tuân thủ luật TMĐT VN (NĐ 52/2013 + 85/2021, MOIT, dự thảo livestream/affiliate 2025) | SHOULD | done | TASK-COMPLY-001 | 6h |

### P3.5 - TRUST - anti-fraud ở quy mô

| TASK-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **TASK-TRUST-004** | Anti-fraud engine (referral abuse, fake-account farming, velocity, relationship graph) | MUST | done | TASK-BILL-004, TASK-AFFIL-001 | 10h |
| **TASK-TRUST-005** | Phát hiện gaming affiliate attribution + delay payout | MUST | done | TASK-AFFIL-001, TASK-AFFIL-003, TASK-BILL-002, TASK-TRUST-004 | 6h |
| **TASK-TRUST-006** | Device-fingerprint + phát hiện multi-account | SHOULD | done | TASK-TRUST-004 | 6h |

---

## §6 - Index NFR (Non-Functional Requirements)

NFR sống dưới [`../non-functional-requirements/<module>/`](../non-functional-requirements/). Mỗi NFR có file `.audit.md` đi kèm.

| NFR-ID | Tiêu đề | Module | Nguồn (tài liệu §) |
|---|---|---|---|
| **NFR-INFRA-001** | Hiệu năng API - p95 < 300ms (đọc cache), biểu đồ < 500ms | infra | §3.8 |
| **NFR-INFRA-002** | Khả dụng - SLA 99,5% | infra | §3.8 |
| **NFR-PRICE-001** | Khả năng mở rộng time-series - hàng tỷ dòng `price_snapshot`, nén + aggregate | price | §3.8, §3.4 |
| **NFR-SCRAPE-001** | Resilience anti-bot per-sàn + xếp hạng rủi ro ban (Low/Medium/High) | scrape | §3.9 |
| **NFR-SCRAPE-002** | Chi phí scraping/proxy - biến phí ~0,1-0,2 USD/user/tháng | scrape | §4.1 |
| **NFR-EXT-001** | Ràng buộc Manifest V3 (service worker ephemeral, alarms >=30s, no global state) | ext | §3.2 |
| **NFR-NOTIF-001** | Scale đỉnh 00:00 - flatten-the-curve, FCM 600k/phút, không vượt 429 | notif | §3.6 |
| **NFR-AFFIL-001** | Compliance affiliate - chỉ user-initiated, disclosure, né Honey/Chrome policy | affil | §4.2 |
| **NFR-COMPLY-001** | PDPL Luật 91/2025 - consent, DPIA, 72h breach, no-cleartext, chế tài tới 5% doanh thu | comply | §5.5 |
| **NFR-TRUST-001** | Bảo mật & niềm tin - no cleartext credential, token không rời client, argon2id, Vault | trust | §3.8, §5.4 |

---

## §7 - Thứ tự build (topo) & ghi chú phụ thuộc

```
  P0 ------------> P1 -------------------> P2 --------------> P3
  INFRA            AUTH | SCRAPE | PRICE    EXT(TikTok/Lazada)  AFFIL(cashback)
  (gateway,        EXT(Shopee) | TRACK      SCRAPE(2 san)       B2B | MOBILE
   data-model,     DEAL(sale ao) | NOTIF    CART | AFFIL        COMPLY(SEA)
   secrets, obs)   WEB | COMPLY | TRUST     BILL | DEAL(ML)     TRUST(anti-fraud)
```

**Chuỗi tới hạn (critical path) cho MVP:** TASK-INFRA-002 -> TASK-PRICE-001 -> TASK-PRICE-002 -> TASK-SCRAPE-001 -> TASK-SCRAPE-002 -> TASK-PRICE-003 -> TASK-DEAL-001 -> TASK-DEAL-003 -> TASK-WEB-003. Song song: TASK-EXT-001 -> TASK-EXT-002 -> TASK-EXT-003. Cold-start: bắt đầu TASK-SCRAPE-002 sớm nhất có thể để tích lũy 90 ngày dữ liệu (§5.1).

**Cổng tiến phase:**
1. task phase N pass hết acceptance criteria (§4 của mỗi task) trước khi phase N+1 bắt đầu.
2. Coverage audit-row PDPL = 100% trên mọi bề mặt xử lý dữ liệu cá nhân (NFR-COMPLY-001).
3. Rò rỉ cross-tenant/cross-user = 0 (property test).
4. P1 exit cần >=90 ngày lịch sử giá cho top-SKU (giải cold-start) + extension open-source live.

---

## §8 - Tham chiếu chéo Risk Register (§9 tài liệu nguồn)

| Rủi ro (§9) | Mức | task/NFR giảm thiểu |
|---|---|---|
| Sàn C&D / chặn extension | High | TASK-SCRAPE-006, TASK-EXT-007/008 (đa sàn), TASK-WEB-001 (web độc lập), NFR-SCRAPE-001 |
| Scraping bị ban | Medium-High | TASK-SCRAPE-003/004/005, NFR-SCRAPE-001 |
| Affiliate reject do compliance | Medium | TASK-AFFIL-002/004, NFR-AFFIL-001 |
| Chrome gỡ extension (Honey-style) | Medium | TASK-AFFIL-004, TASK-TRUST-001/002, NFR-AFFIL-001 |
| PDPL vi phạm | High | TASK-COMPLY-001..005, TASK-COMPLY-007, NFR-COMPLY-001 |
| Cold-start dữ liệu | Medium | TASK-DEAL-002, TASK-SCRAPE-002 (backfill sớm), §7 ghi chú |
| Hiểu lầm malware | Medium | TASK-TRUST-001/002/003, TASK-EXT-006 |
| Gian lận user | Medium | TASK-TRUST-004/005/006, TASK-BILL-004 |

---

*Hết BACKLOG SănDeal v0.1.0. Index này phục vụ 3 team song song (Kỹ thuật/Kiến trúc - Kinh doanh/Tài chính - Chiến lược/Rủi ro) và cần cập nhật khi task đổi status hoặc khi các giả định ở §10 tài liệu nguồn được xác minh.*

## Conventions (CyberOS)

One backlog for both classes: rows are `- [status] TASK-ID-slug - title`; `class: improvement` rows carry an `(improvement)` suffix, product rows are untagged. task frontmatter `status` is the record of truth; this file is the index.

- improvement programs: see `improvement/` (moved from `docs/improvement/`; class: improvement work - convert items to tasks on pickup)
