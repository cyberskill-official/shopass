# SănDeal - Feature Request Backlog

**Owner:** Stephen Cheng (Founder / CEO) - **Status:** v0.2.0 - SHIP-READY. 90 FR + 90 audit + 10 NFR + 10 audit; DAG acyclic + reciprocal; data model nhất quán (một table một owner); đã qua một vòng review chéo module. Backlog khởi tạo từ tài liệu nền tảng SănDeal v1.0 (16/06/2026), áp dụng workflow feature-request của CyberOS (engineering-spec@1).
**Nguồn sự thật (source of truth):** các file markdown trong thư mục này. Index này được tái tạo khi FR được thêm hoặc đổi status.
**Tài liệu nguồn:** [`../TÀI LIỆU NỀN TẢNG SẢN PHẨM "SănDeal" - PRD + SRS + CHIẾN LƯỢC KỸ THUẬT : KINH DOANH : RỦI RO.md`](../)
**Tài liệu hỗ trợ ship (cho agent triển khai):** [`SHIP-GUIDE.md`](SHIP-GUIDE.md) (conventions + bất biến build) | [`IMPLEMENTATION-ORDER.md`](IMPLEMENTATION-ORDER.md) (thứ tự build theo layer) | [`DATA-MODEL.md`](DATA-MODEL.md) (schema hợp nhất) | [`STATUS-REFERENCE.md`](STATUS-REFERENCE.md) (vòng đời status). Ghi chú: `AGENTS.md` ở gốc repo dành cho giao thức memory CyberOS (BRAIN); conventions build nằm ở SHIP-GUIDE.md.
**Playbook tác giả:** workflow `feature-request-author` + `feature-request-audit` của CyberOS (`cyberos/modules/skill/contracts/feature-request/`). Mỗi FR đi kèm một file `.audit.md`.
**Status enum (10 trạng thái):** `draft | ready_to_implement | implementing | ready_to_review | reviewing | ready_to_test | testing | done | on_hold | closed` (theo [`STATUS-REFERENCE.md`](STATUS-REFERENCE.md)).

---

## §0 - Cách đọc backlog này

Tài liệu này là **nguồn sự thật duy nhất** cho những gì SănDeal sẽ xây, tổ chức theo **phase** (P0 -> P3), rồi theo **module**, rồi theo **slice** trong mỗi module. Mỗi dòng là một FR; một FR là một yêu cầu nguyên tử, kiểm thử được.

- **Phase** map vào roadmap §8 của tài liệu nguồn. `P0 Nền tảng` dựng hạ tầng xuyên suốt; `P1 MVP` ship extension Shopee + cold-start dữ liệu + SEO + sale ảo; `P2 Mở rộng` thêm TikTok Shop + Lazada + cart optimizer + Premium; `P3 Tăng trưởng` thêm cashback + B2B + mobile + SEA.
- **Slice** là một đơn vị ship gọn trong một module. Slice 1 luôn là bề mặt tối thiểu khả dụng (MVP của module đó).
- **Priority** dùng từ khóa BCP-14: `MUST` (chặn release) - `SHOULD` (nên có) - `COULD` (tốt nếu có) - `MAY` (sau release).
- **Status**: trạng thái hiện tại của FR (xem enum 10 trạng thái ở header).
- **Depends on**: danh sách phụ thuộc cross-FR.
- **Effort**: ước lượng thô theo giờ (1h = 30 phút làm tập trung + 30 phút phối hợp/review). Sai số +/-50%. Tính cho một kỹ sư có kinh nghiệm.

**Thứ tự đọc cho người lập kế hoạch:** quét §1 (tổng) -> chọn phase đang làm -> đọc phần phân rã theo module trong phase đó -> đào vào từng file FR.
**Thứ tự đọc cho người triển khai:** tìm FR-ID được giao trong phần module -> mở file FR markdown để xem chi tiết.

---

## §1 - Tổng quan

| Phase | Module trong phạm vi | Số FR | Effort (giờ) | Cổng kiểm tra khi thoát phase |
|---|---|---:|---:|---|
| **P0 - Nền tảng** | INFRA | 5 | ~33 | API Gateway live, data-model migrate được, OTel/Grafana lên, secrets trong Vault |
| **P1 - MVP** | AUTH - SCRAPE - PRICE - EXT - TRACK - DEAL - NOTIF - WEB - COMPLY - TRUST | 46 | ~303 | Extension Shopee đọc giỏ hàng, >=90 ngày lịch sử giá top-SKU, sale ảo + biểu đồ live, push alert, PDPL consent + DPIA nộp, extension open-source |
| **P2 - Mở rộng** | EXT - SCRAPE - CART - AFFIL - BILL - DEAL - NOTIF | 25 | ~177 | TikTok Shop + Lazada đọc được, cart/voucher optimizer đúng stacking 3 nước, Premium thu được tiền, affiliate user-initiated compliant, dự đoán đáy giá ML |
| **P3 - Tăng trưởng** | AFFIL - B2B - MOBILE - COMPLY - TRUST | 14 | ~114 | Cashback hold-then-payout, B2B data ẩn danh bán được, mobile app live, per-country gating cho >=1 nước SEA, anti-fraud engine chặn farming |
| **Tổng** | 16 module - 4 phase | **90** | **~627 giờ (~16 person-week)** | 4 cổng compliance/kỹ thuật |

> Lưu ý: 90 FR nằm đúng đỉnh khoảng "~70-90" ước lượng ban đầu. Tài liệu nguồn map ra 16 module với nhiều slice per-sàn (3 sàn x extension + scraping) và 3 dòng doanh thu (affiliate, Premium, B2B). Đây là backlog đầy đủ 3 phase theo yêu cầu.

**Kiểm tra ngân sách effort:** 90 FR x ~7h trung bình ~ 627h ~ 15,7 person-week thuần code. Cộng design + review + tích hợp + bất ngờ -> khoảng 9-12 tháng cho 1 kỹ sư full-time, hoặc 5-6 tháng cho 2 kỹ sư - khớp với roadmap 0-12 tháng ở §8 tài liệu nguồn.

---

## §2 - P0 - Nền tảng (INFRA)

**Mục tiêu phase:** dựng hạ tầng xuyên suốt mà mọi module khác phụ thuộc. Khi thoát P0: API Gateway/BFF định tuyến REST+GraphQL+WSS với rate-limit + JWT + WAF; data-model migrate được; OTel + Prometheus + Grafana cho một bề mặt điều tra sự cố; secrets trong Vault (không cleartext).

### P0.1 - INFRA - hạ tầng nền tảng

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-INFRA-001** | API Gateway / BFF - định tuyến REST+GraphQL+WSS, rate-limit, JWT verify, WAF | MUST | done | - | 8h |
| **FR-INFRA-002** | Data-model foundation - migration framework + bảng `platform` + `app_user` lõi | MUST | done | - | 6h |
| **FR-INFRA-003** | Secrets management - Vault / AWS Secrets Manager, no-cleartext, rotation | MUST | done | - | 5h |
| **FR-INFRA-004** | Observability spine - Prometheus + Grafana + OTel tracing + structured logs | MUST | done | FR-INFRA-001 | 8h |
| **FR-INFRA-005** | Per-country region config - gating hook (VN/ID/TH/PH/MY/SG/TW) + feature flags | MUST | done | FR-INFRA-002 | 6h |

---

## §3 - P1 - MVP (extension Shopee + cold-start + SEO + sale ảo)

**Mục tiêu phase:** ship được vòng giá trị lõi cho persona Chi/Huy/Linh trên Shopee: theo dõi giá, phát hiện sale ảo, biểu đồ giá, alert push. Tích lũy >=90 ngày lịch sử giá top-SKU (giải bài toán cold-start). Site SEO kéo traffic organic. PDPL compliant + extension open-source để xây niềm tin hậu-Honey.

**Cổng compliance:** PDPL (Luật 91/2025) consent + DPIA nộp trong 60 ngày - no-cleartext + token-not-on-server - extension open-source + disclosure Chrome Web Store.

**Phụ thuộc tới hạn:** mọi module P1 cần P0 xong (gateway, data-model, observability, secrets).

### P1.1 - AUTH - người dùng + liên kết tài khoản sàn

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-AUTH-001** | Schema `app_user` (argon2id pwd_hash, CITEXT email, phone, locale) + đăng ký | MUST | done | FR-INFRA-002 | 6h |
| **FR-AUTH-002** | Phát hành JWT + refresh + phiên (BFF auth) | MUST | done | FR-AUTH-001, FR-INFRA-001 | 6h |
| **FR-AUTH-003** | Liên kết `platform_account` - ext_user_ref ẩn danh, KHÔNG lưu token phiên | MUST | done | FR-AUTH-001 | 5h |
| **FR-AUTH-004** | Social login (Google / Facebook / Zalo OAuth) | SHOULD | ready_to_implement | FR-AUTH-002 | 6h |
| **FR-AUTH-005** | Vòng đời tài khoản - reset mật khẩu, status, xóa tài khoản (DSAR PDPL) | MUST | ready_to_implement | FR-AUTH-001, FR-COMPLY-003 | 5h |

### P1.2 - SCRAPE - scraping giá (Shopee trước)

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-SCRAPE-001** | Scraping orchestrator lõi - scheduler + scan-frequency tiering (hot/thường) | MUST | done | FR-INFRA-003, FR-PRICE-001 | 10h |
| **FR-SCRAPE-002** | Shopee internal-API adapter (`/api/v4/...`, hybrid, `is_login:false`) | MUST | ready_to_implement | FR-SCRAPE-001 | 8h |
| **FR-SCRAPE-003** | Playwright headless farm + anti-fingerprint (Canvas/WebGL/JA3/JA4/HTTP2) | MUST | ready_to_implement | FR-SCRAPE-001 | 12h |
| **FR-SCRAPE-004** | Residential proxy rotation + tiering + cost-guard (Bright Data/Oxylabs/Decodo/IPRoyal) | MUST | ready_to_implement | FR-SCRAPE-003 | 8h |
| **FR-SCRAPE-005** | Delta-only writes + pacing/jitter + CAPTCHA handling | MUST | ready_to_implement | FR-SCRAPE-002, FR-PRICE-002 | 6h |
| **FR-SCRAPE-006** | DOM-change monitoring + adapter health (resilient A/B test DOM) | MUST | ready_to_implement | FR-SCRAPE-002 | 6h |

### P1.3 - PRICE - time-series giá + so sánh chéo sàn

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-PRICE-001** | `tracked_product` registry chuẩn hóa (UNIQUE platform_id, item_id) | MUST | done | FR-INFRA-002 | 6h |
| **FR-PRICE-002** | `price_snapshot` TimescaleDB hypertable + nén + continuous aggregate | MUST | done | FR-PRICE-001 | 8h |
| **FR-PRICE-003** | API lịch sử giá (`GET /v1/products/{id}/price-history?range=90d`) | MUST | done | FR-PRICE-002 | 5h |
| **FR-PRICE-004** | So sánh giá chéo 3 sàn (`GET /v1/compare?canonical_key=...`) | MUST | pending | FR-PRICE-005 | 6h |
| **FR-PRICE-005** | Thuật toán matching `canonical_key` (dedup sản phẩm chéo sàn) | MUST | done | FR-PRICE-001 | 8h |

### P1.4 - EXT - browser extension MV3 (Shopee)

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-EXT-001** | Scaffold extension Manifest V3 + service worker ephemeral, chrome.alarms, host_permissions | MUST | done | - | 8h |
| **FR-EXT-002** | Shopee content script đọc giỏ/voucher (session piggyback) | MUST | done | FR-EXT-001 | 10h |
| **FR-EXT-003** | Pipeline tối thiểu hóa dữ liệu client (chỉ productId/price/qty; KHÔNG cookie/token) | MUST | done | FR-EXT-002 | 6h |
| **FR-EXT-004** | Offscreen API cho DOM parsing/clipboard + declarativeNetRequest | SHOULD | done | FR-EXT-001 | 5h |
| **FR-EXT-005** | Đồng bộ extension <-> backend (auth bridge, WSS keep-alive khi cần realtime) | MUST | ready_to_implement | FR-EXT-003, FR-AUTH-002 | 6h |
| **FR-EXT-006** | UI settings + consent (PDPL consent lúc cài, disclosure dữ liệu) | MUST | done | FR-EXT-001, FR-COMPLY-001 | 5h |

### P1.5 - TRACK - theo dõi sản phẩm + wishlist + alert rules

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-TRACK-001** | API theo dõi sản phẩm (`POST /v1/track {platform, item_url}`) | MUST | ready_to_implement | FR-PRICE-001, FR-SCRAPE-002 | 5h |
| **FR-TRACK-002** | Schema + API `wishlist` / `wishlist_item` (target_price) | MUST | ready_to_implement | FR-TRACK-001 | 5h |
| **FR-TRACK-003** | Schema + API `alert_rule` (price_below/drop_pct/real_sale/bottom_predicted) | MUST | ready_to_implement | FR-TRACK-001 | 6h |
| **FR-TRACK-004** | Engine kích hoạt alert (đánh giá rule trên `price_snapshot`) | MUST | ready_to_implement | FR-TRACK-003, FR-PRICE-002, FR-NOTIF-001 | 6h |

### P1.6 - DEAL - phát hiện sale ảo + cold-start

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-DEAL-001** | Phát hiện sale ảo (statistical: median90/p10/trailing_min -> SALE_AO/SALE_XIN/TAM_DUOC) | MUST | ready_to_implement | FR-PRICE-002 | 8h |
| **FR-DEAL-002** | Xử lý cold-start (category priors, <14d -> UNKNOWN, cổng baseline 90 ngày) | MUST | ready_to_implement | FR-DEAL-001 | 6h |
| **FR-DEAL-003** | API dữ liệu biểu đồ giá (daily aggregate cho chart, p95 <500ms) | MUST | ready_to_implement | FR-PRICE-002 | 5h |

### P1.7 - NOTIF - thông báo (push trước)

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-NOTIF-001** | Schema notification + routing theo cost model (push > email > sms) | MUST | done | FR-INFRA-002 | 6h |
| **FR-NOTIF-002** | FCM Web/Android dispatcher (token mgmt, quota 600k/phút, 429 backoff) | MUST | ready_to_implement | FR-NOTIF-001 | 8h |
| **FR-NOTIF-003** | Fan-out pipeline (Kafka/Redis Streams -> workers -> per-channel) + DLQ | MUST | ready_to_implement | FR-NOTIF-001 | 8h |
| **FR-NOTIF-004** | Scheduler flatten-the-curve cho đỉnh 00:00 (jitter, bucketing, FCM rate-limit) | MUST | ready_to_implement | FR-NOTIF-003 | 6h |

### P1.8 - WEB - web app + SEO

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-WEB-001** | Scaffold Next.js + auth + shell dashboard | MUST | ready_to_implement | FR-AUTH-002 | 8h |
| **FR-WEB-002** | Landing SEO (keyword: săn xu Shopee, lịch sale, mã freeship, sale thật/ảo) | MUST | ready_to_implement | FR-WEB-001 | 8h |
| **FR-WEB-003** | UI biểu đồ lịch sử giá (render p95 <500ms) | MUST | ready_to_implement | FR-WEB-001, FR-DEAL-003 | 6h |
| **FR-WEB-004** | UI quản lý wishlist + alert | MUST | ready_to_implement | FR-WEB-001, FR-TRACK-002, FR-TRACK-003 | 6h |
| **FR-WEB-005** | GraphQL BFF cho web (truy vấn linh hoạt wishlist/biểu đồ) | SHOULD | ready_to_implement | FR-INFRA-001, FR-WEB-001 | 6h |

### P1.9 - COMPLY - PDPL nền tảng

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-COMPLY-001** | Khung consent PDPL (Luật 91/2025/QH15) - đơn mục đích, tái lập | MUST | done | FR-INFRA-002 | 8h |
| **FR-COMPLY-002** | Sổ đăng ký DPIA/TIA (nộp trong 60 ngày, cập nhật mỗi 6 tháng) | MUST | done | FR-COMPLY-001 | 6h |
| **FR-COMPLY-003** | Quyền chủ thể dữ liệu (truy cập/sửa/xóa/di chuyển - DSAR) | MUST | done | FR-COMPLY-001 | 8h |
| **FR-COMPLY-004** | Quy trình thông báo vi phạm 72 giờ | MUST | done | FR-COMPLY-001, FR-INFRA-004 | 5h |
| **FR-COMPLY-005** | Cưỡng chế no-cleartext + token-not-on-server - audit gate CI | MUST | done | FR-INFRA-003 | 5h |

### P1.10 - TRUST - niềm tin (open-source + audit)

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-TRUST-001** | Open-source extension + reproducible build + disclosure | MUST | done | FR-EXT-001 | 6h |
| **FR-TRUST-002** | Chính sách minh bạch tối thiểu hóa dữ liệu + xử lý local-first | MUST | ready_to_implement | FR-EXT-003 | 5h |
| **FR-TRUST-003** | Hook security audit độc lập (chứng minh không gửi cookie/mật khẩu) | MUST | ready_to_implement | FR-EXT-003, FR-COMPLY-005 | 6h |

---

## §4 - P2 - Mở rộng (TikTok Shop + Lazada + cart optimizer + Premium)

**Mục tiêu phase:** mở đa sàn thật (thêm TikTok Shop 41,31% GMV + Lazada), bật moat so sánh chéo 3 sàn, ship cart/voucher optimizer client-side, và bắt đầu thu tiền qua Premium + affiliate compliant.

**Cổng compliance:** affiliate chỉ user-initiated + disclosure (né Honey) - stacking voucher đúng luật per-country (VN stack / MY-PH no-stack 2025) - PCI-lite cho payment gateway.

### P2.1 - EXT - TikTok Shop + Lazada content scripts

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-EXT-007** | TikTok Shop content script (webview/SPA DOM reader, tránh API ký msToken/X-Bogus) | MUST | ready_to_implement | FR-EXT-002 | 10h |
| **FR-EXT-008** | Lazada content script (Akamai-aware, đọc DOM render) | MUST | ready_to_implement | FR-EXT-002 | 8h |

### P2.2 - SCRAPE - TikTok Shop + Lazada adapters

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-SCRAPE-007** | TikTok Shop scraping adapter (ưu tiên DOM-render, né API ký) | MUST | ready_to_implement | FR-SCRAPE-003 | 10h |
| **FR-SCRAPE-008** | Lazada scraping adapter (Akamai, residential bắt buộc) | MUST | ready_to_implement | FR-SCRAPE-003 | 8h |

### P2.3 - CART - voucher + tối ưu giỏ hàng

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-CART-001** | Schema + ingest `voucher_catalog` (shop/platform/freeship, stack_group, cap) | MUST | done | FR-INFRA-002 | 6h |
| **FR-CART-002** | Schema `cart_snapshot` + `cart_item` (nhận từ extension) | MUST | ready_to_implement | FR-EXT-003 | 5h |
| **FR-CART-003** | Optimizer giỏ/voucher/freeship (knapsack, ràng buộc stacking, applyCaps) | MUST | ready_to_implement | FR-CART-001, FR-CART-002 | 12h |
| **FR-CART-004** | Engine luật stacking per-country (VN stack vs MY/PH bỏ stacking 2025) | MUST | ready_to_implement | FR-CART-003, FR-INFRA-005 | 6h |
| **FR-CART-005** | `testCodes`: thử mã an toàn client-side (sleep nhịp người, user-initiated, revert) | MUST | done | FR-EXT-002, FR-CART-001 | 6h |
| **FR-CART-006** | Checklist xu/coin (nhắc nhở, KHÔNG auto-click - chống abuse) | SHOULD | ready_to_implement | FR-EXT-002 | 5h |

### P2.4 - AFFIL - affiliate compliant

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-AFFIL-001** | Schema + tracking affiliate (user-initiated) + đối soát mạng | MUST | done | FR-INFRA-002 | 6h |
| **FR-AFFIL-002** | Deep-link generator user-initiated (`POST /v1/affiliate/link`, disclosure, no auto-cookie) | MUST | done | FR-AFFIL-001 | 6h |
| **FR-AFFIL-003** | Tích hợp affiliate network (Involve Asia / Accesstrade) compliant | MUST | ready_to_implement | FR-AFFIL-002 | 8h |
| **FR-AFFIL-004** | Guardrails né Honey (no cookie-stuffing, bắt buộc user-action, tuân Chrome policy 10/06/2025) | MUST | ready_to_implement | FR-AFFIL-002, FR-EXT-003 | 5h |

### P2.5 - BILL - Premium + thanh toán

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-BILL-001** | Schema `subscription` + tier (Premium 29k/49k/79k) + vòng đời | MUST | done | FR-AUTH-001 | 6h |
| **FR-BILL-002** | Tích hợp cổng thanh toán (MoMo / ZaloPay / VNPay / VietQR) | MUST | ready_to_implement | FR-BILL-001 | 10h |
| **FR-BILL-003** | Bản ghi `payment` + reconciliation + webhook | MUST | ready_to_implement | FR-BILL-002 | 6h |
| **FR-BILL-004** | `referral_code` + attribution + hook chống abuse | SHOULD | ready_to_implement | FR-BILL-001 | 5h |
| **FR-BILL-005** | Trigger upgrade free->Premium (gamified) + feature gating | SHOULD | ready_to_implement | FR-BILL-001 | 6h |

### P2.6 - DEAL - dự đoán đáy giá (ML)

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-DEAL-004** | Baseline dự đoán đáy (Prophet, regressor double-date/payday) | MUST | ready_to_implement | FR-PRICE-002, FR-DEAL-002 | 10h |
| **FR-DEAL-005** | Model LightGBM (>=180d history, target future_min_price_14d) + feature store | SHOULD | ready_to_implement | FR-DEAL-004 | 12h |
| **FR-DEAL-006** | Batch scoring đêm + alert (P(bottom trong 14d) > 0.7) | MUST | ready_to_implement | FR-DEAL-004, FR-TRACK-003 | 6h |

### P2.7 - NOTIF - email + SMS + APNs

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-NOTIF-005** | APNs iOS dispatcher (xử lý 410, retry backoff 500/503) | MUST | ready_to_implement | FR-NOTIF-003 | 5h |
| **FR-NOTIF-006** | Email dispatcher (SES / SendGrid / Postmark) | MUST | ready_to_implement | FR-NOTIF-003 | 4h |
| **FR-NOTIF-007** | SMS dispatcher VN (SpeedSMS/eSMS/VietGuys + Twilio fallback, brandname, chỉ high-value) | SHOULD | ready_to_implement | FR-NOTIF-003 | 6h |

---

## §5 - P3 - Tăng trưởng (cashback + B2B + mobile + SEA)

**Mục tiêu phase:** layer cashback trên affiliate, mở dòng doanh thu B2B margin cao (dữ liệu xu hướng giá ẩn danh), ra mobile app, và mở rộng SEA với per-country gating (ID/TH/PH/MY/SG/TW).

**Cổng compliance:** cashback hold-then-payout + delay payout chống gian lận - B2B data ẩn danh (k-anonymity) - per-country data-protection (Indonesia PDP, Thailand PDPA) - VN e-commerce law (NĐ 52/85, MOIT).

### P3.1 - AFFIL - cashback layering

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-AFFIL-005** | Cashback layering (chia % cho user, hold tới khi affiliate confirm, delay payout) | SHOULD | ready_to_implement | FR-AFFIL-003, FR-BILL-002, FR-TRUST-005 | 10h |

### P3.2 - B2B - dữ liệu + analytics

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-B2B-001** | Pipeline dữ liệu xu hướng thị trường ẩn danh (k-anonymity, aggregate) | SHOULD | ready_to_implement | FR-PRICE-002, FR-COMPLY-003 | 10h |
| **FR-B2B-002** | Báo cáo B2B insights + subscription | SHOULD | ready_to_implement | FR-B2B-001, FR-BILL-001 | 8h |
| **FR-B2B-003** | Seller-facing competitor price analytics | COULD | ready_to_implement | FR-B2B-001 | 8h |
| **FR-B2B-004** | Premium API access (tiers, rate-limited) | COULD | ready_to_implement | FR-INFRA-001, FR-B2B-001 | 6h |

### P3.3 - MOBILE - mobile app + SEA virality

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-MOBILE-001** | Scaffold mobile (React Native/Flutter) + auth + push | SHOULD | ready_to_implement | FR-AUTH-002, FR-NOTIF-002 | 12h |
| **FR-MOBILE-002** | Mobile theo dõi giá + alert + universal checkout assistant | SHOULD | ready_to_implement | FR-MOBILE-001, FR-CART-003 | 10h |
| **FR-MOBILE-003** | Deep-link + share-on-sale virality + referral | COULD | ready_to_implement | FR-MOBILE-001, FR-BILL-004 | 6h |

### P3.4 - COMPLY - per-country gating + SEA + e-commerce law

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-COMPLY-006** | Khung per-country gating (luật voucher/affiliate/dữ liệu theo nước) | MUST | ready_to_implement | FR-INFRA-005, FR-CART-004 | 8h |
| **FR-COMPLY-007** | Adapter bảo vệ dữ liệu SEA (Indonesia PDP, Thailand PDPA) | SHOULD | ready_to_implement | FR-COMPLY-001, FR-COMPLY-006 | 8h |
| **FR-COMPLY-008** | Tuân thủ luật TMĐT VN (NĐ 52/2013 + 85/2021, MOIT, dự thảo livestream/affiliate 2025) | SHOULD | ready_to_implement | FR-COMPLY-001 | 6h |

### P3.5 - TRUST - anti-fraud ở quy mô

| FR-ID | Tiêu đề | Pri | Status | Depends on | Effort |
|---|---|:-:|:-:|---|---:|
| **FR-TRUST-004** | Anti-fraud engine (referral abuse, fake-account farming, velocity, relationship graph) | MUST | ready_to_implement | FR-BILL-004, FR-AFFIL-001 | 10h |
| **FR-TRUST-005** | Phát hiện gaming affiliate attribution + delay payout | MUST | ready_to_implement | FR-AFFIL-001, FR-AFFIL-003, FR-BILL-002, FR-TRUST-004 | 6h |
| **FR-TRUST-006** | Device-fingerprint + phát hiện multi-account | SHOULD | ready_to_implement | FR-TRUST-004 | 6h |

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

**Chuỗi tới hạn (critical path) cho MVP:** FR-INFRA-002 -> FR-PRICE-001 -> FR-PRICE-002 -> FR-SCRAPE-001 -> FR-SCRAPE-002 -> FR-PRICE-003 -> FR-DEAL-001 -> FR-DEAL-003 -> FR-WEB-003. Song song: FR-EXT-001 -> FR-EXT-002 -> FR-EXT-003. Cold-start: bắt đầu FR-SCRAPE-002 sớm nhất có thể để tích lũy 90 ngày dữ liệu (§5.1).

**Cổng tiến phase:**
1. FR phase N pass hết acceptance criteria (§4 của mỗi FR) trước khi phase N+1 bắt đầu.
2. Coverage audit-row PDPL = 100% trên mọi bề mặt xử lý dữ liệu cá nhân (NFR-COMPLY-001).
3. Rò rỉ cross-tenant/cross-user = 0 (property test).
4. P1 exit cần >=90 ngày lịch sử giá cho top-SKU (giải cold-start) + extension open-source live.

---

## §8 - Tham chiếu chéo Risk Register (§9 tài liệu nguồn)

| Rủi ro (§9) | Mức | FR/NFR giảm thiểu |
|---|---|---|
| Sàn C&D / chặn extension | High | FR-SCRAPE-006, FR-EXT-007/008 (đa sàn), FR-WEB-001 (web độc lập), NFR-SCRAPE-001 |
| Scraping bị ban | Medium-High | FR-SCRAPE-003/004/005, NFR-SCRAPE-001 |
| Affiliate reject do compliance | Medium | FR-AFFIL-002/004, NFR-AFFIL-001 |
| Chrome gỡ extension (Honey-style) | Medium | FR-AFFIL-004, FR-TRUST-001/002, NFR-AFFIL-001 |
| PDPL vi phạm | High | FR-COMPLY-001..005, FR-COMPLY-007, NFR-COMPLY-001 |
| Cold-start dữ liệu | Medium | FR-DEAL-002, FR-SCRAPE-002 (backfill sớm), §7 ghi chú |
| Hiểu lầm malware | Medium | FR-TRUST-001/002/003, FR-EXT-006 |
| Gian lận user | Medium | FR-TRUST-004/005/006, FR-BILL-004 |

---

*Hết BACKLOG SănDeal v0.1.0. Index này phục vụ 3 team song song (Kỹ thuật/Kiến trúc - Kinh doanh/Tài chính - Chiến lược/Rủi ro) và cần cập nhật khi FR đổi status hoặc khi các giả định ở §10 tài liệu nguồn được xác minh.*
