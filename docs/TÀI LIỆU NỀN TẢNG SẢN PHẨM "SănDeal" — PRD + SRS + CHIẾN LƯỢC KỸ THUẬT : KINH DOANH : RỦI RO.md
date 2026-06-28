# TÀI LIỆU NỀN TẢNG SẢN PHẨM "SănDeal" - PRD + SRS + CHIẾN LƯỢC KỸ THUẬT / KINH DOANH / RỦI RO

*Phiên bản 1.0 - Nền tảng SaaS-tiện ích săn deal / theo dõi giá / tối ưu mua sắm đa nền tảng cho Việt Nam & Đông Nam Á. Founder: Stephen Cheng (Trịnh Thái Anh), CyberSkill (TP.HCM). Ngày: 16/06/2026. Tài liệu phục vụ 3 team triển khai song song (Kỹ thuật/Kiến trúc - Kinh doanh/Tài chính - Chiến lược/Rủi ro).*

---

## TL;DR / TÓM TẮT ĐIỀU HÀNH

- **SănDeal là nền tảng đa sàn (Shopee + TikTok Shop + Lazada) ngay từ ngày đầu**, không phải công cụ chỉ-Shopee. Đây là quyết định chiến lược: theo Metric (Online Retail Platform Market Report 2025, công bố 15/01/2026), cả năm 2025 **Shopee chiếm 56,04% và TikTok Shop 41,31% GMV TMĐT Việt Nam** (TikTok Shop tăng từ ~29% năm 2024); Shopee + TikTok Shop = 97% GMV. Công cụ chỉ-Shopee đang nhắm vào thị phần tương đối đang co lại. Moat tự nhiên của thiết kế đa sàn là **so sánh giá chéo 3 sàn** (cùng SKU rẻ hơn ở Lazada hay TikTok Shop hay Shopee).
- **Kiến trúc 3 trụ cột**: (A) Browser Extension Manifest V3 "session piggyback" đọc giỏ hàng/voucher của chính người dùng + backend scraping giá quy mô lớn + thuật toán phát hiện sale ảo / dự đoán đáy giá / tối ưu voucher; (B) Mô hình unit economics free-tier tài trợ bằng affiliate + Premium 29k-79k VND/tháng; (C) Chiến lược cold-start, tuân thủ PDPL (Luật 91/2025/QH15) và né bài học PayPal Honey.
- **Lằn ranh đạo đức = moat niềm tin**: Vụ Honey (MegaLag đăng video 22/12/2024, hơn 13 triệu view trong vài ngày; Honey mất ~3 triệu trong tổng ~20 triệu user trong 2 tuần) và việc Google cập nhật chính sách Chrome Web Store tháng 3/2025 (thực thi từ 10/06/2025) - cấm extension chèn affiliate link/cookie khi không mang lại lợi ích trực tiếp (giảm giá/cashback) và bắt buộc có hành động người dùng - khiến mô hình "cookie-stuffing" trở thành bất hợp pháp. Hệ quả chuỗi: Rakuten Advertising gỡ Honey 12/01/2026 (cắt ~2.000 merchant gồm Walmart); Impact.com đình chỉ 17/01/2026; Awin ngừng thanh toán 21/01/2026. SănDeal phải minh bạch tuyệt đối và biến điều này thành lợi thế cạnh tranh.

---

## 1. TẦM NHÌN SẢN PHẨM & PERSONAS

### 1.1 Tầm nhìn
SănDeal trở thành "trợ lý mua sắm thông minh" số 1 Đông Nam Á: giúp người mua không bao giờ "sập bẫy sale ảo", luôn mua đúng đáy giá, tối ưu voucher/freeship/xu trên cả 3 sàn lớn, và mở rộng sang sản phẩm dữ liệu B2B. Khác BeeCost (chủ yếu Shopee/Tiki, web-first), SănDeal là đa sàn thật sự + tối ưu giỏ hàng client-side + minh bạch đạo đức hậu-Honey.

### 1.2 Personas
- **Chi (28, nhân viên văn phòng, TP.HCM)** - "săn sale" hàng tháng, dùng MoMo/ZaloPay, nhạy giá, sợ mua hớ. Cần: phát hiện sale ảo, lịch sale, freeship tối ưu.
- **Huy (22, sinh viên, Hà Nội)** - săn xu, mã giảm, freeship, mua đồ giá rẻ TikTok Shop. Cần: checklist xu, auto-test mã, thông báo flash sale lúc 0h.
- **Linh (35, mẹ bỉm sữa)** - mua đồ gia dụng/mẹ & bé số lượng lớn, tối ưu giỏ hàng nhiều shop. Cần: cart optimizer, cảnh báo giảm giá sau mua (hoàn tiền).
- **Người bán / Brand (B2B)** - cần dữ liệu xu hướng giá thị trường ẩn danh. Đây là dòng doanh thu B2B margin cao.

---

## 2. BỐI CẢNH THỊ TRƯỜNG (DỮ LIỆU NỀN)

- **GMV TMĐT Việt Nam 2025**: VND 429,7 nghìn tỷ (~16,35 tỷ USD) trên 4 sàn lớn, **+34,75% YoY**; 3,94 tỷ sản phẩm bán ra (+15,23%); số nhà bán có doanh thu **giảm 7,43% còn ~601.800** (Metric, theinvestor.vn 16/01/2026). Shopee 56,04%, TikTok Shop 41,31%, Lazada + Tiki ~3%. Shopee + TikTok Shop = 97% GMV.
- **TikTok Shop H1/2025**: GMV tăng ~148% YoY; lần đầu số nhà bán có doanh thu vượt Shopee (266k+ vs 209k).
- **AOV TMĐT Việt Nam**: ~71 USD (2023, ECDB, sau giảm giá & trả hàng, chưa VAT); add-to-cart ~10,3%, cart abandonment ~83,2%, conversion ~1,7%.
- **SEA tổng GMV TMĐT 2025**: ~157,6 tỷ USD (chính xác **+22,8% YoY**, Momentum Works "Ecommerce in SEA" ấn bản 4, 14/04/2026); Shopee/Lazada/TikTok Shop = 98,8% GMV nền tảng; Thái Lan (+51,8%) và Malaysia (+47,6%) dẫn đầu tăng trưởng.

**Hàm ý**: Thiết kế đa sàn là bắt buộc; **per-country gating** bắt buộc (luật voucher, affiliate, dữ liệu khác nhau theo nước).

---

## 3. DIMENSION A - KỸ THUẬT / KIẾN TRÚC

### 3.1 Kiến trúc hệ thống tổng thể

```
[Browser Extension MV3]   [Web App (Next.js)]   [Mobile App (giai đoạn sau, React Native/Flutter)]
        |                        |                         |
        +-----------+------------+-------------+-----------+
                    | HTTPS/REST + GraphQL + WSS
              [API Gateway / BFF]  (rate-limit, auth JWT, WAF)
                    |
   +----------------+-----------------+------------------+-------------------+
   |                |                 |                  |                   |
[Auth/User Svc] [Tracking Svc]  [Price Svc]      [Voucher/Cart Optimizer] [Notification Svc]
   |                |                 |                  |                   |
   +-------- [PostgreSQL (OLTP) + TimescaleDB (time-series giá)] -----------+
                    |                 |                                      |
            [Redis (cache/queue)] [Scraping Orchestrator]            [Kafka/Redis Streams]
                                      |                                      |
                  [Playwright headless farm + proxy residential]   [Fan-out workers -> FCM/APNs/Email/SMS]
                                      |
                  [ML Service: Prophet/LightGBM, feature store]
```

Thành phần chính: (1) Extension MV3 (content scripts per-sàn + service worker), (2) Web app + landing SEO, (3) Backend microservices (Go/Node + Python cho ML), (4) Data layer Postgres + TimescaleDB + Redis, (5) Scraping farm, (6) Notification fan-out, (7) ML service.

### 3.2 Extension "session piggyback" - chi tiết theo 3 sàn

**Nguyên tắc cốt lõi**: Content script chạy trong ngữ cảnh tab đã đăng nhập của chính người dùng, đọc DOM / gọi internal JSON endpoint cùng cookie phiên first-party của người dùng. **KHÔNG bao giờ thu thập mật khẩu; token phiên KHÔNG rời khỏi máy client**; chỉ trích xuất dữ liệu giỏ hàng/voucher đã render và gửi về backend dạng tối thiểu hóa (chỉ productId, giá, số lượng - KHÔNG gửi cookie).

- **Shopee**: Content script đọc trang giỏ hàng; có thể gọi internal endpoint dạng `/api/v4/cart/get` cùng cookie first-party. Shopee chạy **thư viện bảo mật JS riêng** nền (device fingerprinting tạo hash thiết bị từ hàng chục data point về trình duyệt/phần cứng/hành vi); endpoint sản phẩm dạng `/api/v4/pdp/get_pc`, `/api/v4/recommend/recommend` (một số truy cập được khi `is_login:false`). DOM giỏ hàng thay đổi theo A/B test.
- **TikTok Shop**: Cấu trúc web/app khác hẳn (content-commerce). Cart/checkout nằm trong webview/SPA -> content script đọc DOM giỏ hàng. TikTok có cơ chế **ký request ở tầng API nội bộ** (tham số dạng `msToken`/`_signature`/`X-Bogus`) cùng app attestation mạnh -> ưu tiên đọc DOM render thay vì gọi API ký.
- **Lazada**: Cấu trúc trang riêng; đọc DOM giỏ hàng; Lazada (Alibaba) thường dùng **Akamai**. Ưu tiên đọc DOM đã render.

**Ràng buộc Manifest V3 & cách kiến trúc vòng quanh**:
- Service worker **ephemeral**: bị kill sau ~30 giây không hoạt động; mỗi sự kiện/API call reset timer; bị kill nếu một sự kiện chạy >5 phút hoặc `fetch()` không phản hồi >30 giây -> KHÔNG giữ state trong biến global; lưu vào `chrome.storage`.
- `chrome.alarms`: tối thiểu **30 giây** (từ Chrome 120; trước đó 1 phút) -> dùng alarm cho polling định kỳ nhẹ, KHÔNG dùng `setInterval`.
- Tác vụ nặng/dài (đọc nhiều tab, đồng bộ) -> đẩy lên backend; extension chỉ làm "đầu đọc" nhẹ.
- `host_permissions` khai báo rõ cho từng domain sàn; dùng `declarativeNetRequest` thay webRequest blocking; dùng **Offscreen API** khi cần DOM parsing/clipboard ngoài service worker.
- WebSocket giữ service worker sống khi cần realtime (nhưng không lạm dụng).

### 3.3 Backend scraping giá quy mô 3 sàn

**Hybrid**: (a) Reverse-engineer internal JSON endpoints (nhanh, rẻ - ví dụ Shopee `/api/v4/...`) khi truy cập được không cần đăng nhập; (b) Headless browser farm (Playwright + anti-fingerprint: spoof Canvas/WebGL/AudioContext, JA3/JA4 TLS, HTTP/2 settings) cho trang có bảo vệ mạnh.

**Proxy residential xoay vòng (giá tham khảo 2026)**:
- Bright Data: ~$8,40/GB (entry 10GB), giảm còn ~$3,30/GB ở 10TB; PAYG ~$25/GB tier thấp nhất.
- Oxylabs: từ ~$4/GB (PAYG), gói từ $99/tháng; 100M+ IP, ~99,95% success, ~0,6s response.
- Decodo (ex-Smartproxy): ~$2,20/GB residential, 65M+ IP, nhanh nhất (~0,54s).
- IPRoyal: từ ~$1,75/GB PAYG (rẻ nhất).
- SOAX: từ $90/25GB; NetNut cạnh tranh tầm trung.
- **Phân tầng 2026**: Enterprise (Bright Data, Oxylabs) $8,5-12/GB; Mid (Decodo, SOAX, NetNut) $3-6/GB; Budget (IPRoyal) từ $1,75/GB.
- Datacenter proxy gần như vô dụng với target có Cloudflare/Akamai -> **bắt buộc residential** cho scraping nghiêm túc.

**CAPTCHA, request pacing, scan-frequency tiering, delta-only writes**: Phân tầng tần suất quét theo độ "hot" của SKU (flash sale: phút; SKU thường: vài giờ/ngày). Chỉ ghi DB khi giá thay đổi (delta-only) để tiết kiệm storage time-series. Pacing ngẫu nhiên + jitter để tránh rate-limit. Một số target có puzzle slider/CAPTCHA khi nghi ngờ -> cần dịch vụ giải CAPTCHA hoặc headless có hành vi giống người.

### 3.4 DATA MODEL (PostgreSQL + TimescaleDB) - DDL phác thảo

```sql
-- NỀN TẢNG & NGƯỜI DÙNG
CREATE TABLE platform (
  id           SMALLINT PRIMARY KEY,
  code         TEXT UNIQUE NOT NULL,  -- 'shopee','tiktok','lazada'
  country      TEXT NOT NULL,         -- 'VN','ID','TH'...
  base_url     TEXT, created_at TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE app_user (
  id           BIGSERIAL PRIMARY KEY,
  email        CITEXT UNIQUE, phone TEXT,
  display_name TEXT, locale TEXT DEFAULT 'vi-VN',
  pwd_hash     TEXT,                  -- argon2id; KHÔNG lưu cleartext
  referral_code_id BIGINT,
  created_at   TIMESTAMPTZ DEFAULT now(), status TEXT DEFAULT 'active'
);
CREATE TABLE platform_account (   -- liên kết tài khoản sàn (không lưu mật khẩu)
  id           BIGSERIAL PRIMARY KEY,
  user_id      BIGINT REFERENCES app_user(id),
  platform_id  SMALLINT REFERENCES platform(id),
  ext_user_ref TEXT,                 -- định danh ẩn danh hóa, KHÔNG token phiên
  linked_at    TIMESTAMPTZ DEFAULT now(),
  UNIQUE(user_id, platform_id)
);
-- SẢN PHẨM & GIÁ
CREATE TABLE tracked_product (
  id           BIGSERIAL PRIMARY KEY,
  platform_id  SMALLINT REFERENCES platform(id),
  platform_item_id TEXT NOT NULL,    -- itemid/shopid hoặc product id sàn
  shop_id      TEXT, title TEXT, category_id BIGINT,
  canonical_key TEXT,                -- để so sánh chéo sàn (chuẩn hóa)
  first_seen   TIMESTAMPTZ DEFAULT now(),
  UNIQUE(platform_id, platform_item_id)
);
CREATE INDEX idx_tp_canonical ON tracked_product(canonical_key);

-- TIME-SERIES GIÁ (TimescaleDB hypertable)
CREATE TABLE price_snapshot (
  product_id   BIGINT NOT NULL REFERENCES tracked_product(id),
  ts           TIMESTAMPTZ NOT NULL,
  price        BIGINT NOT NULL,       -- VND
  list_price   BIGINT,                -- giá niêm yết (gốc)
  stock        INTEGER, sold INTEGER,
  flash_sale   BOOLEAN DEFAULT false,
  PRIMARY KEY (product_id, ts)
);
SELECT create_hypertable('price_snapshot','ts', chunk_time_interval => INTERVAL '7 days');
ALTER TABLE price_snapshot SET (timescaledb.compress, timescaledb.compress_segmentby='product_id');
SELECT add_compression_policy('price_snapshot', INTERVAL '30 days');
CREATE MATERIALIZED VIEW price_daily WITH (timescaledb.continuous) AS
  SELECT product_id, time_bucket('1 day', ts) AS day,
         min(price) min_p, max(price) max_p, last(price, ts) close_p
  FROM price_snapshot GROUP BY product_id, day;

-- WISHLIST / ALERT
CREATE TABLE wishlist (id BIGSERIAL PRIMARY KEY, user_id BIGINT, name TEXT, created_at TIMESTAMPTZ DEFAULT now());
CREATE TABLE wishlist_item (id BIGSERIAL PRIMARY KEY, wishlist_id BIGINT, product_id BIGINT, target_price BIGINT, added_at TIMESTAMPTZ DEFAULT now());
CREATE TABLE alert_rule (
  id BIGSERIAL PRIMARY KEY, user_id BIGINT, product_id BIGINT,
  rule_type TEXT,            -- 'price_below','drop_pct','real_sale','bottom_predicted'
  threshold NUMERIC, channel TEXT[],  -- ['push','email','sms']
  active BOOLEAN DEFAULT true, created_at TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE alert (id BIGSERIAL PRIMARY KEY, alert_rule_id BIGINT, fired_at TIMESTAMPTZ, payload JSONB, status TEXT);

-- VOUCHER & CART
CREATE TABLE voucher_catalog (
  id BIGSERIAL PRIMARY KEY, platform_id SMALLINT, code TEXT, type TEXT, -- 'shop','platform','freeship'
  discount_type TEXT, discount_value BIGINT, min_spend BIGINT, cap BIGINT,
  shop_id TEXT, valid_from TIMESTAMPTZ, valid_to TIMESTAMPTZ, stack_group TEXT
);
CREATE TABLE cart_snapshot (id BIGSERIAL PRIMARY KEY, user_id BIGINT, platform_id SMALLINT, captured_at TIMESTAMPTZ DEFAULT now());
CREATE TABLE cart_item (id BIGSERIAL PRIMARY KEY, cart_snapshot_id BIGINT, product_id BIGINT, shop_id TEXT, qty INT, unit_price BIGINT);

-- AFFILIATE
CREATE TABLE affiliate_click (id BIGSERIAL PRIMARY KEY, user_id BIGINT, platform_id SMALLINT, product_id BIGINT, sub_id TEXT, clicked_at TIMESTAMPTZ, network TEXT);
CREATE TABLE affiliate_conversion (id BIGSERIAL PRIMARY KEY, click_id BIGINT, order_value BIGINT, commission BIGINT, status TEXT, confirmed_at TIMESTAMPTZ);

-- BILLING
CREATE TABLE subscription (id BIGSERIAL PRIMARY KEY, user_id BIGINT, tier TEXT, price BIGINT, started_at TIMESTAMPTZ, renews_at TIMESTAMPTZ, status TEXT);
CREATE TABLE payment (id BIGSERIAL PRIMARY KEY, subscription_id BIGINT, gateway TEXT, amount BIGINT, fee BIGINT, paid_at TIMESTAMPTZ, status TEXT);
CREATE TABLE referral_code (id BIGSERIAL PRIMARY KEY, user_id BIGINT, code TEXT UNIQUE, uses INT DEFAULT 0);

-- NOTIFICATION & GAMIFICATION
CREATE TABLE notification (id BIGSERIAL PRIMARY KEY, user_id BIGINT, channel TEXT, template TEXT, payload JSONB, scheduled_at TIMESTAMPTZ, sent_at TIMESTAMPTZ, status TEXT);
CREATE TABLE user_activity_coin_task (id BIGSERIAL PRIMARY KEY, user_id BIGINT, platform_id SMALLINT, task_type TEXT, due_date DATE, done BOOLEAN DEFAULT false);
```

**Lý do hypertable/partitioning**: `price_snapshot` là bảng lớn nhất (hàng tỷ dòng với hàng triệu SKU) -> hypertable chunk 7 ngày + nén sau 30 ngày + continuous aggregate cho biểu đồ. Index trên `canonical_key` phục vụ so sánh chéo sàn.

### 3.5 THUẬT TOÁN (pseudo-code)

**(1) Phát hiện sale ảo (statistical)**
```
function detectFakeSale(product_id, current_price, list_price):
  hist = getPriceHistory(product_id, window=90d)   # từ price_snapshot
  if len(hist) < 14: return UNKNOWN                 # cold-start
  median90 = percentile(hist.price, 50)
  p10 = percentile(hist.price, 10)
  trailing_min = min(hist.price)
  inflated = list_price > median90 * 1.15           # giá gốc bị thổi
  not_real_discount = current_price >= median90 * 0.97
  if inflated and not_real_discount: return "SALE_AO"
  if current_price <= p10 and current_price <= trailing_min*1.02: return "SALE_XIN"
  return "TAM_DUOC"
```

**(2) Dự đoán đáy giá (AI)** - feature engineering từ lịch sale (double-date 1.1, 2.2 ... 12.12, payday, flash sale), baseline Prophet -> nâng cấp LightGBM, cold-start dùng category priors.
```
Features: day_of_month, is_double_date(d==m), days_to_next_double_date,
          is_payday_window, trailing_min_30/60/90, price_vs_median90,
          volatility_30d, category_seasonality, flash_sale_flag, platform_id
Pipeline:
  baseline = Prophet(seasonality=yearly+monthly+double_date_regressors)
  if product has >=180d history: model = LightGBM(target = future_min_price_14d)
  else: model = category_prior_model(category_id)   # cold-start
Serving: batch nightly score -> alert nếu P(bottom within 14d) > 0.7
```

**(3) Tối ưu giỏ hàng/voucher/freeship (knapsack-like, có ràng buộc stacking)**
```
function optimizeCart(cart_items, vouchers, platform_rules):
  best = {discount: 0, combo: []}
  applicable_shop = filterByMinSpend(shopVouchers, cart_items)
  for pv in platformVouchers + [none]:
    for fs in freeshipVouchers + [none]:
      if not validStack(pv, fs, platform_rules): continue
      for shopCombo in chooseBestShopVoucherPerShop(applicable_shop, cart_items):
        if not meetsMinSpend(shopCombo, pv, fs, cart_items): continue
        total = sum(discount(shopCombo)) + discount(pv) + freeshipValue(fs)
        total = applyCaps(total, pv.cap, fs.cap)
        if total > best.discount: best = {discount: total, combo: shopCombo+[pv,fs]}
  return best
```
*Ví dụ minh họa*: Giỏ 2 shop A (300k), B (250k); voucher shop A "-30k cho đơn >=250k", platform "-50k cho đơn >=500k", freeship "<=30k". **VN** cho stack 1 shop + 1 platform + freeship -> tổng giảm = 30k + 50k + 30k = **110k**. Tại **MY/PH (2025 đã bỏ stacking)** freeship gộp vào platform -> chỉ chọn max(50k, 30k) = 50k + shop 30k = **80k**. (Shopee MY/PH 2025: tối đa 1 shop voucher + 1 platform voucher, freeship & cashback gộp chung nhóm platform.)

**(4) Auto-test mã giảm (an toàn, client-side, tuần tự, nhịp người)**
```
function testCodes(candidate_codes, cart):
  results = []
  for code in candidate_codes:
    sleep(random(2.5s, 5s))             # nhịp người, tránh anti-bot & ToS
    apply = userInitiatedApply(code)    # CHỈ khi user bấm "thử mã" (tuân thủ Chrome policy)
    if apply.valid: results.append((code, apply.discount))
    revert()                            # không tự chốt đơn
  return sortDesc(results)
```

**(5) Notification fan-out / midnight-spike (flatten-the-curve, jitter, batching)**
```
function scheduleMidnightAlerts(alerts):
  # FCM khuyến cáo tránh dồn ±2 phút quanh các mốc :00/:15/:30/:45
  for a in alerts:
    jitter = random(-90s, +180s)        # rải đều quanh thời điểm sự kiện
    a.dispatch_at = a.event_time + jitter
  buckets = groupBy(alerts, key = floor(dispatch_at / 60s))
  for b in buckets:
    if size(b) > FCM_RATE_LIMIT_PER_MIN: spreadAcrossNextMinutes(b)
    enqueue(b)                          # Kafka/Redis Streams -> fan-out workers
```

### 3.6 Notification fan-out cho đỉnh 00:00

Kiến trúc: producer -> **Kafka/Redis Streams** -> fan-out workers -> per-channel dispatchers:
- **FCM** (Android + Web Push): quota mặc định **600.000 messages/phút/project** (HTTP v1, token bucket refill mỗi phút; "The default quota of 600k messages per minute covers over 99% of FCM developers"); vượt -> HTTP 429 `RESOURCE_EXHAUSTED`. FCM khuyến cáo "flatten the curve", tránh dồn quanh mốc :00 (lưu lượng tăng >2 lần trong 30s-2 phút đầu mỗi giờ). Xin tăng quota trước >=15 ngày (>18M/phút cần >=30 ngày; tối đa 2 sự kiện temp quota/năm).
- **APNs** (iOS): nhiều kết nối song song; xử lý lỗi 410 (token hết hạn), retry backoff cho 500/503.
- **Email**: Amazon SES / SendGrid / Postmark (rất rẻ, ~$0,10/1.000 với SES).
- **SMS Việt Nam + SEA**: SpeedSMS, eSMS (ViHAT), VietGuys, Mobifone; Twilio fallback. Brandname cần đăng ký định danh với Cục An toàn thông tin; phí duy trì brandname ~50.000 VND/tháng/nhà mạng; SMS Long Code rẻ hơn ~50% so brandname; giá SMS nội địa ~200-500 VND/tin tùy nhà mạng. Twilio ~$0,1552/SMS tới VN (đắt -> chỉ fallback OTP/giá-trị-cao).
- Backoff + jitter + **dead-letter queue** cho tin thất bại.

**Mô hình chi phí thông báo**: Push gần như miễn phí (FCM/APNs) -> kênh chính. Email rất rẻ. SMS đắt -> chỉ dùng cho alert giá trị cao hoặc Premium. **Ưu tiên: push > email > SMS.**

### 3.7 API design (REST/GraphQL - phác thảo)
- `POST /v1/track` `{platform, item_url}` -> tạo tracked_product.
- `GET /v1/products/{id}/price-history?range=90d` -> time-series.
- `POST /v1/alerts` `{product_id, rule_type, threshold, channel[]}`.
- `POST /v1/cart/optimize` `{platform, items[], vouchers[]}` -> best combo.
- `GET /v1/compare?canonical_key=...` -> giá 3 sàn.
- `POST /v1/affiliate/link` `{product_id}` -> deep link affiliate (CHỈ khi user bấm).
- GraphQL cho web app (truy vấn linh hoạt wishlist/biểu đồ).

### 3.8 Yêu cầu phi chức năng (NFR)
- **Hiệu năng**: p95 API < 300ms (đọc cache); biểu đồ giá < 500ms.
- **Khả năng mở rộng**: hàng triệu SKU, hàng tỷ dòng price_snapshot -> TimescaleDB nén/aggregate.
- **Bảo mật**: KHÔNG lưu cleartext credential; token phiên không rời client; secrets trong vault (HashiCorp Vault / AWS Secrets Manager); argon2id cho mật khẩu.
- **Khả dụng**: SLA 99,5%.
- **Observability**: Prometheus + Grafana + structured logs + tracing (OpenTelemetry).

### 3.9 Phân tích anti-bot per-sàn & xếp hạng rủi ro ban

| Cơ chế | Shopee | TikTok Shop | Lazada |
|---|---|---|---|
| WAF/anti-bot | Thư viện bảo mật JS riêng + fingerprinting; nghi dùng CDN/WAF | Hệ ByteDance, ký request (msToken/_signature/X-Bogus) | Akamai (Alibaba) |
| Internal API | `/api/v4/...` (một số không cần login) | API ký, khó | API ký |
| Fingerprinting | Canvas/WebGL/device hash | App attestation mạnh | TLS/HTTP |
| CAPTCHA | Slider/puzzle khi nghi ngờ | Có | Có |

**Xếp hạng rủi ro ban (Low/Medium/High)**:
- (a) Scraping giá read-only quy mô lớn: Shopee **Medium-High**, TikTok Shop **High**, Lazada **Medium-High**.
- (b) Đọc giỏ hàng client-side qua extension/cookie của chính user: cả 3 **Low** (first-party, hành vi giống người dùng) - nhưng vẫn vi phạm ToS tiềm tàng.
- (c) Tự động hóa xu/voucher: cả 3 **High** (dễ bị coi là bot/abuse) -> SănDeal chỉ làm checklist nhắc nhở + auto-test mã do người dùng khởi tạo, KHÔNG tự động click xu.

---

## 4. DIMENSION B - KINH DOANH / TÀI CHÍNH

### 4.1 Unit economics (mô hình chi phí)

**Chi phí per 1.000 users/tháng (ước tính)**:
- Proxy/scraping: ~$30-60 (chia sẻ; người dùng share dữ liệu giúp giảm scraping).
- Notification: push ~$0; email ~$1; SMS hạn chế.
- Cloud/DB/server: ~$50-100 phân bổ.
- Headless farm: ~$20-40 phân bổ.
-> Tổng biến phí ~$100-200/1.000 users (~0,1-0,2 USD/user/tháng).

**Per 100.000 users/tháng**: proxy/scraping tăng nhưng cận biên giảm (chia sẻ SKU); DB/Timescale tốn hơn (~$2.000-5.000); tổng ~$10.000-20.000 (~0,1-0,2 USD/user). Quy mô có lợi thế chia sẻ dữ liệu giá.

**Doanh thu**:
- **Affiliate (CPS)**: Shopee VN base ~1-2% (một số nguồn), category 2,5-12%; cookie window xung đột (7 ngày theo nhiều nguồn / lên tới 8 ngày qua Skimlinks; có nguồn nói ~30 ngày desktop / 7 ngày mobile - **CẦN XÁC MINH**). EPC Shopee qua Skimlinks ~$0,03, conversion ~18,19%, basket ~$4,12. Lazada: 1-10% (new customer tới ~12%), extra commission tới ~34% bonus. TikTok Shop: seller tự đặt 1-80% (thực tế 5-30%), last-click.
- **Premium**: 29k-79k VND/tháng; conversion free->paid giả định 2-5%.
- **Phí thanh toán** MoMo/ZaloPay/VNPay: QR rẻ hơn thẻ; ước ~1,5-2,5%/giao dịch.

**Contribution margin, breakeven, LTV/CAC (kịch bản)**:
- **Base**: 100k users, 3% Premium (3.000 x 39k = 117tr VND/tháng) + affiliate (giả định 5% user mua/tháng x AOV ~71 USD x ~2% commission). LTV Premium ~12 tháng x 39k ~ 468k VND. CAC qua SEO/viral thấp (~20-50k VND) -> payback < 2 tháng nếu organic.
- **Pessimistic**: 1% Premium, affiliate bị reject nhiều do compliance -> cần >50k users để hòa vốn.
- **Optimistic**: 7% Premium + B2B data + cashback layering -> margin dương sớm.

### 4.2 Affiliate reality-check 3 sàn

**Điều cấm then chốt (Shopee Affiliate ToS)**: cấm "use robots or other automated query tools, computer-generated search requests"; cấm "any automated means or form of scraping, or other data extraction methods"; cookie chỉ được set khi link hiển thị và user click "voluntarily and consciously"; cấm cookie dropping, pop-under, auto-redirect, forced install. Một số nước hạn chế kênh "Software/Coupon/Mobile App", "browser extension/toolbar", "incentivized traffic" (ví dụ Shopee Indonesia bar kênh Software/Coupon/Mobile App).

-> **Hàm ý compliance**: Mô hình affiliate hợp lệ DUY NHẤT cho SănDeal là người dùng **chủ động bấm** nút "Mua qua SănDeal" -> tạo deep link affiliate hiển thị rõ ràng (disclosure). KHÔNG được tự chèn/thay cookie nền (Honey-style). KHÔNG dùng extension scraping để gắn affiliate.

**Vụ PayPal Honey (lằn ranh đạo đức)**: MegaLag đăng video 22/12/2024 ("Exposing the Honey Influencer Scam", >13 triệu view), phơi bày Honey âm thầm thay cookie affiliate của creator (last-click, cơ chế "Selective Stand Down"); Honey mất ~3 triệu trong ~20 triệu user trong 2 tuần. Nhiều class action (gồm vụ do LegalEagle, Gamers Nexus). Google cập nhật chính sách Chrome Web Store **tháng 3/2025, thực thi 10/06/2025**: cấm chèn affiliate link/code/cookie khi không có lợi ích trực tiếp (giảm giá/cashback/donation) + bắt buộc disclosure + bắt buộc hành động người dùng. Chuỗi gỡ bỏ: Rakuten 12/01/2026, Impact.com 17/01/2026, Awin 21/01/2026. -> SănDeal biến minh bạch thành moat niềm tin.

### 4.3 Pricing strategy & cold-start economics
- Free tier (tài trợ affiliate) + Premium 29k/49k/79k VND/tháng (theo tier tính năng). Benchmark: Keepa ~$17-21/tháng (€19), CamelCamelCamel free (affiliate + ads). Tại VN willingness-to-pay thấp -> free mạnh, Premium nhẹ + gamified upgrade trigger.
- Payment rails: MoMo (31tr+ user), ZaloPay, VNPay, bank QR (VietQR).
- **Cold-start data**: cần lịch sử giá trước khi tính năng (sale ảo, dự đoán, biểu đồ) hữu ích. Cách seed: (1) backfill scraping ngay từ đầu cho top SKU; (2) crowdsource từ extension users (đọc trang họ xem); (3) mua/đối tác dữ liệu lịch sử nếu có. Chi phí/thời gian: ~2-3 tháng scraping tích lũy để có baseline 90 ngày.

---

## 5. DIMENSION C - CHIẾN LƯỢC / RỦI RO

### 5.1 Cold-start (chicken-and-egg)
Tính năng lõi cần lịch sử giá -> nhưng lịch sử cần thời gian. Giải: Phase 1 chạy site SEO + scraping nền tích lũy dữ liệu; ra mắt tính năng "sale ảo/biểu đồ" khi đủ ~90 ngày dữ liệu cho top SKU; crowdsource mở rộng phủ SKU.

### 5.2 Rủi ro phụ thuộc nền tảng (existential)
Nếu Shopee/TikTok/Lazada gửi C&D, chặn extension kỹ thuật, đổi DOM/API, hoặc Chrome gỡ extension (kiểu Honey): -> Mitigation: (1) **đa sàn = phân tán rủi ro**; (2) kiến trúc đọc DOM resilient + giám sát thay đổi DOM; (3) tuân thủ ToS/affiliate để giảm cớ C&D; (4) web app độc lập extension; (5) đa kênh phân phối (Edge, Cốc Cốc, Firefox).

### 5.3 Gian lận người dùng
Referral abuse, fake account farming xu, gaming affiliate attribution, multi-account. -> Phát hiện: device fingerprint, velocity checks, đồ thị quan hệ, **delay payout affiliate** để điều tra (best practice ngành).

### 5.4 Trust & security (chống bị hiểu lầm là malware)
Extension đọc cookie phiên Shopee/TikTok/Lazada dễ bị nghi là scam. -> Concrete: (1) **open-source extension**; (2) security audit độc lập; (3) chính sách dữ liệu minh bạch (không gửi cookie/mật khẩu); (4) xử lý dữ liệu tối thiểu hóa, local-first; (5) disclosure rõ trên Chrome Web Store. ~45% người tiêu dùng VN lo ngại lừa đảo/lộ dữ liệu (Ken Research) -> niềm tin là yếu tố sống còn.

### 5.5 Pháp lý / compliance VN + SEA
- **PDPL - Luật 91/2025/QH15**: Quốc hội thông qua 26/06/2025, **hiệu lực 01/01/2026** (KHÔNG phải 01/07/2026; **đề bài ghi nhầm - đã đính chính**); **Nghị định 356/2025/NĐ-CP** ban hành 31/12/2025, thay Nghị định 13/2023 (DLA Piper; Tilleke & Gibbins). Đồng thuận phải tự nguyện, cụ thể, đơn mục đích, tái lập được; im lặng != đồng thuận. Quyền chủ thể dữ liệu; DPIA/TIA (nộp trong 60 ngày từ khi bắt đầu xử lý; cập nhật mỗi 6 tháng); chuyển dữ liệu xuyên biên giới. **Chế tài**: tới 5% doanh thu năm trước cho vi phạm xuyên biên giới; tới 10 lần lợi nhuận bất chính cho mua bán dữ liệu; tới 3 tỷ VND cho vi phạm nghiêm trọng. Thông báo vi phạm trong **72 giờ**.
- Cookie/session token: lưu trữ phải tuân PDPL; SănDeal chủ trương KHÔNG lưu token phiên trên server.
- SEA khác biệt: Indonesia PDP Law, Thailand PDPA -> per-country compliance.
- TMĐT VN: Nghị định 52/2013 + 85/2021 (đăng ký MOIT, trách nhiệm sàn, ngưỡng >100.000 giao dịch/năm cho foreign platform); Dự thảo Luật TMĐT 2025 lần đầu quản livestream & affiliate marketing.
- ToS Shopee/TikTok/Lazada: rủi ro phơi bày - tuân thủ chặt.

### 5.6 Cạnh tranh
- **BeeCost** (VN): xây thương hiệu trên "sale ảo/sale xịn", theo dõi giá Shopee/Tiki/Lazada/Sendo, web + extension, theo dõi 350tr+ sản phẩm. Điểm yếu: tính năng rải rác theo sàn, mạnh nhất Shopee/Tiki, chưa tối ưu giỏ hàng đa sàn mạnh.
- Websosanh, Sosanhgia (so sánh giá VN).
- Quốc tế: Keepa ($17-21/tháng, sâu cho seller), CamelCamelCamel (free, Amazon-only, affiliate + ads), Honey/PayPal Honey (tai tiếng), Karma, Cently, Capital One Shopping.
- **Khoảng trống khai thác**: đa sàn thật + cart/voucher optimizer client-side + so sánh giá chéo 3 sàn + minh bạch đạo đức hậu-Honey + B2B data. **Moat**: dữ liệu lịch sử giá độc quyền tích lũy, network effect crowdsource, thương hiệu niềm tin.

### 5.7 GTM & funnel
- **Phase 1**: site SEO (keyword: "cách săn xu Shopee", "lịch sale", "mã freeship", "sale thật hay sale ảo") -> **Phase 2**: signup free -> **Phase 3**: Premium upsell.
- Virality: share-on-sale, referral code.
- Seeding cộng đồng VN (Facebook groups, Telegram) KHÔNG vi phạm affiliate spam (không spam link affiliate nền).
- **SEA sequencing**: VN -> ID/TH -> PH/MY -> SG/TW. Per-country gating bắt buộc: MY & PH bỏ stacking voucher 2025; quyền kênh affiliate khác; cookie window khác; chế độ bảo vệ dữ liệu khác.

---

## 6. CATALOG TÍNH NĂNG MỞ RỘNG (hiện có + đề xuất mới tối đa hóa lợi nhuận)

**Hiện có (founder đã lên kế hoạch)**: theo dõi giá, quét giỏ hàng, phát hiện sale ảo, dự đoán đáy AI, tối ưu voucher/giỏ, auto-test mã, checklist xu, cảnh báo hoàn tiền sau mua.

**Đề xuất MỚI (mô tả / kiếm tiền / độ khả thi build / tác động user-love)**:
1. **So sánh giá chéo 3 sàn** - cùng SKU rẻ hơn ở đâu. Kiếm tiền: affiliate sàn rẻ nhất. Build: Medium. User-love: Rất cao (moat đa sàn tự nhiên).
2. **Cashback layering** trên affiliate - chia lại % cho user. Kiếm tiền: giữ phần chênh affiliate. Build: Medium. User-love: Rất cao.
3. **Lịch "thời điểm mua tốt nhất"** - dựa double-date/payday. Kiếm tiền: Premium. Build: Low. User-love: Cao.
4. **Group-buy / mua chung điều phối** - gom đơn đạt min-spend voucher. Kiếm tiền: affiliate volume + Premium. Build: High. User-love: Cao.
5. **Subscription dự đoán giảm giá** - alert "sắp xuống đáy". Kiếm tiền: Premium. Build: Medium. User-love: Cao.
6. **Universal checkout assistant** (browser/mobile) - tối ưu voucher khi thanh toán. Kiếm tiền: Premium + affiliate (user-initiated). Build: High. User-love: Rất cao.
7. **B2B data/insights** - báo cáo xu hướng giá/thị trường ẩn danh bán cho brand/seller. Kiếm tiền: B2B subscription (margin cao). Build: Medium. User-love: Trung tính (ẩn danh).
8. **Seller-facing analytics** - theo dõi giá đối thủ. Kiếm tiền: SaaS B2B. Build: Medium-High. User-love: N/A (B2B).
9. **Affiliate content auto-generation** cho KOL. Kiếm tiền: Premium/KOL tools. Build: Medium. User-love: Cao (KOL).
10. **Loyalty/points program** nội bộ SănDeal. Kiếm tiền: tăng retention -> affiliate. Build: Medium. User-love: Cao.
11. **Premium API access** cho dev/doanh nghiệp. Kiếm tiền: API tiers. Build: Low-Medium. User-love: N/A.
12. **Deal-alert communities** + **KOL/influencer tools**. Kiếm tiền: affiliate + Premium. Build: Medium. User-love: Cao.

---

## 7. SRS (yêu cầu chức năng & phi chức năng - tóm tắt)
- **FR**: track sản phẩm; lịch sử giá; alert đa kênh; phát hiện sale ảo; dự đoán đáy; tối ưu giỏ/voucher; auto-test mã; checklist xu; so sánh chéo sàn; affiliate deep link (user-initiated); cashback; B2B reports.
- **NFR**: xem mục 3.8 (p95 < 300ms, 99,5% SLA, bảo mật no-cleartext, PDPL compliance, observability).
- **Architecture/Data model/Algorithms/API**: mục 3.1-3.7.

## 8. MVP & ROADMAP (map 3 team song song)
- **Team Kỹ thuật/Kiến trúc**: extension MV3 (3 sàn đọc DOM) + backend scraping + data model + thuật toán sale ảo + notification push. MVP: theo dõi giá + sale ảo + biểu đồ (Shopee trước, TikTok/Lazada nối tiếp).
- **Team Kinh doanh/Tài chính**: tích hợp affiliate (Involve Asia/Accesstrade) compliant + Premium billing (MoMo/ZaloPay/VNPay) + unit economics dashboard.
- **Team Chiến lược/Rủi ro**: site SEO + PDPL/legal + trust (open-source, audit) + GTM + per-country gating.
- **Phase 1 (0-3 tháng)**: cold-start data + SEO + extension Shopee. **Phase 2 (3-6)**: TikTok + Lazada + cart optimizer + Premium. **Phase 3 (6-12)**: cashback, B2B data, mobile app, SEA expansion.

## 9. RISK REGISTER (Low/Medium/High)
| Rủi ro | Mức | Mitigation |
|---|---|---|
| Sàn C&D / chặn extension | High | Đa sàn, web app độc lập, tuân ToS |
| Scraping bị ban | Medium-High | Residential proxy, pacing, hybrid |
| Affiliate reject do compliance | Medium | Chỉ user-initiated, disclosure |
| Chrome gỡ extension (Honey-style) | Medium | Minh bạch, không cookie-stuffing |
| PDPL vi phạm | High | DPIA, consent, no-cleartext, 72h breach |
| Cold-start dữ liệu | Medium | Backfill + crowdsource |
| Hiểu lầm malware | Medium | Open-source, audit, minh bạch |
| Gian lận user | Medium | Fingerprint, velocity, delay payout |

## 10. GIẢ ĐỊNH & CÂU HỎI CẦN XÁC MINH
- Cookie window Shopee VN chính xác (7 ngày vs 30/7 desktop/mobile) - **xác minh với Involve Asia/Accesstrade**.
- Commission rate cụ thể từng category 3 sàn VN 2026.
- Phí chính xác MoMo/ZaloPay/VNPay theo hợp đồng merchant.
- Giá SMS brandname/OTP chính xác (VND/tin) theo nhà mạng.
- **PDPL hiệu lực 01/01/2026** (đã đính chính so với đề bài ghi 01/07/2026).
- Chính sách Chrome Web Store: cập nhật 3/2025, **thực thi 10/06/2025**.
- Quyền kênh affiliate "browser extension" có bị cấm ở từng nước không - xác minh ToS từng quốc gia trước khi mở rộng.

---
*Hết tài liệu nền tảng v1.0. Tài liệu này phục vụ đồng thời 3 team triển khai song song và cần được cập nhật khi các giả định ở Mục 10 được xác minh.*
