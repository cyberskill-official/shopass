# DATA-MODEL.md - Catalog dữ liệu hợp nhất cho SănDeal

Tài liệu này là catalog tham chiếu schema CSDL hợp nhất trên toàn bộ 90 task của SănDeal, neo vào tài liệu nền tảng §3.4 (DDL phác thảo PostgreSQL + TimescaleDB). Mỗi bảng có đúng một task sở hữu (task phát lệnh `CREATE TABLE`); các task khác chỉ tham chiếu qua khóa ngoại (FK) hoặc mở rộng cột qua `ALTER TABLE`. Đây là tài liệu tra cứu, không phải migration thực thi: DDL chuẩn nằm trong §3 của từng task sở hữu.

Quy ước: một bảng = một task sở hữu. Mở rộng cột qua `ALTER TABLE` được ghi chú rõ. Kiểu cột viết tắt (BIGINT, TEXT, TS = TIMESTAMPTZ, SMALLINT, JSONB). Tiền tệ luôn là BIGINT theo đơn vị VND, không thập phân (DEC-PRICE-05) - không nơi nào lưu giá dạng float/numeric.

Engine: PostgreSQL 16 cho toàn bộ bảng OLTP; TimescaleDB 2.x chỉ cho lớp time-series giá (`price_snapshot` là hypertable, `price_daily` là continuous aggregate). Chi tiết ở mục "Lớp time-series" bên dưới.

## Người dùng & nền tảng

### platform
- Owner: TASK-INFRA-002. Bảng nền danh mục sàn (shopee, tiktok, lazada).
- Cột chính: id SMALLINT PK, code TEXT UNIQUE (CHECK in shopee|tiktok|lazada), country TEXT (CHECK ISO-3166 alpha-2), base_url TEXT, created_at TS.
- Được FK bởi: TASK-AUTH-003 (platform_account), TASK-PRICE-001 (tracked_product), TASK-CART-001 (voucher_catalog), TASK-CART-002 (cart_snapshot), TASK-CART-006 (user_activity_coin_task), TASK-AFFIL-001 (affiliate_click), TASK-AFFIL-003 (affiliate_network), TASK-SCRAPE-001 (scrape_job).

### app_user
- Owner: TASK-INFRA-002. Bảng người dùng lõi.
- Cột chính: id BIGSERIAL PK, email CITEXT UNIQUE, phone TEXT, display_name TEXT, locale TEXT DEFAULT 'vi-VN', status TEXT DEFAULT 'active', created_at TS.
- Mở rộng: TASK-AUTH-001 `ALTER TABLE app_user ADD COLUMN pwd_hash TEXT, referral_code_id BIGINT` (pwd_hash argon2id PHC; FK referral_code_id -> referral_code(id) thêm khi TASK-BILL-004 tạo bảng).
- Được FK bởi: AUTH-002/003/004/005, TRACK-001/002/003, CART-002/006, AFFIL-001/005, BILL-001/004, NOTIF-001, COMPLY-001/003, TRUST-004/005/006.

### platform_account
- Owner: TASK-AUTH-003. Liên kết tài khoản sàn (cố ý KHÔNG lưu cookie/token/password, DEC-AUTH-12).
- Cột chính: id PK, user_id -> app_user(id), platform_id -> platform(id), ext_user_ref TEXT (CHECK length 1..128), linked_at TS, UNIQUE(user_id, platform_id).

### refresh_token (TASK-AUTH-002), social_identity (TASK-AUTH-004), password_reset (TASK-AUTH-005)
- Phụ trợ auth, tất cả FK app_user(id). Lưu token_hash/secret băm, KHÔNG cleartext.
- refresh_token: token_hash TEXT, family_id UUID, expires_at/revoked_at/used_at TS.
- social_identity: provider TEXT (google|facebook|zalo), subject TEXT, UNIQUE(provider, subject).
- password_reset: token_hash TEXT UNIQUE, expires_at TS, used_at TS.

## Sản phẩm & giá

### tracked_product
- Owner: TASK-PRICE-001. Registry chuẩn hóa một dòng cho mỗi (platform_id, platform_item_id); đích FK của toàn bộ time-series giá và mọi thực thể neo theo sản phẩm.
- Cột chính: id BIGSERIAL PK, platform_id -> platform(id), platform_item_id TEXT NOT NULL, shop_id TEXT, title TEXT, category_id BIGINT (id category native của sàn, không có bảng FK), canonical_key TEXT (NULL tới khi TASK-PRICE-005 so khớp), first_seen TS bất biến, UNIQUE(platform_id, platform_item_id), INDEX idx_tp_canonical(canonical_key).
- Được FK bởi: PRICE-002 (price_snapshot), PRICE-005 (canonical_review_queue), TRACK-001/002/003, CART-002 (cart_item), AFFIL-001 (affiliate_click), SCRAPE-001 (scrape_job), DEAL-004/005/006.

### price_snapshot
- Owner: TASK-PRICE-002. TimescaleDB hypertable time-series giá, ghi delta-only.
- Cột chính: product_id -> tracked_product(id), ts TS, price BIGINT (CHECK > 0, VND), list_price BIGINT (CHECK >= price), stock INT, sold INT, flash_sale BOOLEAN, PRIMARY KEY(product_id, ts). Hypertable chunk 7 ngày, nén sau 30 ngày, retention raw 18 tháng.

### price_daily (continuous aggregate)
- Owner: TASK-PRICE-002. Không phải bảng thường: continuous aggregate trên price_snapshot, bucket 1 ngày, cột min_p/max_p/close_p, GROUP BY product_id, day. Giữ vô thời hạn.
- Được đọc bởi: TASK-WEB-003 (biểu đồ), TASK-DEAL-001 (sale ảo), TASK-DEAL-004/005, TASK-B2B-001.

### canonical_review_queue
- Owner: TASK-PRICE-005. Hàng đợi duyệt gộp canonical_key độ tin cậy thấp.
- Cột chính: id PK, product_id -> tracked_product(id), candidate_key TEXT, confidence REAL (0..1), status TEXT (pending|approved|rejected), UNIQUE(product_id, candidate_key).

## Wishlist & alert

### user_tracked_product (TASK-TRACK-001)
- Bảng nối user theo dõi sản phẩm. PK(user_id, product_id), cả hai là FK (app_user, tracked_product), tracked_at TS.

### wishlist / wishlist_item (TASK-TRACK-002)
- wishlist: id PK, user_id -> app_user(id), name TEXT, created_at TS.
- wishlist_item: id PK, wishlist_id -> wishlist(id) ON DELETE CASCADE, product_id -> tracked_product(id), target_price BIGINT (CHECK > 0, VND), UNIQUE(wishlist_id, product_id).

### alert_rule / alert (TASK-TRACK-003)
- alert_rule: id PK, user_id -> app_user(id), product_id -> tracked_product(id), rule_type TEXT (price_below|drop_pct|real_sale|bottom_predicted), threshold BIGINT (VND hoặc %; NULL cho real_sale/bottom_predicted), channel TEXT[], active BOOLEAN. Lưu ý: `threshold` là BIGINT có chủ đích, khác NUMERIC ở §3.4 tài liệu nguồn - threshold luôn là số nguyên (VND cho price_below, phần trăm nguyên 1..99 cho drop_pct), không có giá trị thập phân; giữ BIGINT đồng nhất với DEC-PRICE-05 và tránh sai số float khi so giá (DEC-TRACK-22). Đây là quyết định đã chốt, KHÔNG đổi về NUMERIC.
- alert: id PK, alert_rule_id -> alert_rule(id) ON DELETE CASCADE, fired_at TS, payload JSONB, status TEXT.

### alert_fired_state (TASK-TRACK-004)
- Trạng thái edge-trigger chống spam. PK alert_rule_id -> alert_rule(id) ON DELETE CASCADE, last_condition_met BOOLEAN, last_fired_at TS.

## Voucher & cart

### voucher_catalog (TASK-CART-001)
- id PK, platform_id -> platform(id), code TEXT, type TEXT (shop|platform|freeship), discount_type TEXT (amount|percent), discount_value BIGINT, min_spend BIGINT, cap BIGINT, shop_id TEXT (CHECK: shop -> NOT NULL, khác -> NULL), valid_from/valid_to TS, stack_group TEXT, UNIQUE(platform_id, code).

### cart_snapshot / cart_item (TASK-CART-002)
- cart_snapshot: id PK, user_id -> app_user(id), platform_id -> platform(id), snapshot_ref UUID, captured_at TS, UNIQUE(user_id, snapshot_ref).
- cart_item: id PK, cart_snapshot_id -> cart_snapshot(id) ON DELETE CASCADE, product_id -> tracked_product(id) nullable, platform_item_id TEXT, shop_id TEXT, qty INT (CHECK > 0), unit_price BIGINT (CHECK > 0, VND), CHECK item_identified.

### user_activity_coin_task (TASK-CART-006)
- id PK, user_id -> app_user(id), platform_id -> platform(id), task_type TEXT, due_date DATE, done BOOLEAN, UNIQUE(user_id, platform_id, task_type, due_date).

## Affiliate & cashback

### affiliate_click / affiliate_conversion (TASK-AFFIL-001)
- affiliate_click: id PK, user_id -> app_user(id), platform_id -> platform(id), product_id -> tracked_product(id), sub_id TEXT UNIQUE (token đối soát, không PII), network TEXT, clicked_at TS.
- affiliate_conversion: id PK, click_id -> affiliate_click(id) UNIQUE (1 click <= 1 conversion), order_value BIGINT, commission BIGINT (cả hai CHECK >= 0, VND), status TEXT (pending|confirmed|rejected), confirmed_at TS.
- Được FK bởi: TASK-AFFIL-005 (cashback_entry), TASK-TRUST-005 (payout_hold).

### affiliate_network / affiliate_postback_log (TASK-AFFIL-003)
- affiliate_network: id SMALLSERIAL PK, code TEXT UNIQUE (involve_asia|accesstrade), platform_id -> platform(id), base_url/target_param/sub_id_param TEXT, postback_secret_ref TEXT (Vault key path, không cleartext), active BOOLEAN.
- affiliate_postback_log: id PK, network TEXT, raw_payload JSONB, signature TEXT, verified BOOLEAN, received_at TS.

### cashback_entry / payout_request (TASK-AFFIL-005)
- cashback_entry: id PK, user_id -> app_user(id), conversion_id -> affiliate_conversion(id), commission/user_share/kept_margin BIGINT (CHECK >= 0, VND), status TEXT (pending|available|paid|clawed_back), available_at TS, UNIQUE(conversion_id).
- payout_request: id PK, user_id -> app_user(id), amount BIGINT (CHECK > 0), status TEXT (queued|sent|failed), gateway_ref TEXT.

## Billing

### plan_catalog / subscription (TASK-BILL-001)
- plan_catalog: id SMALLSERIAL PK, tier TEXT UNIQUE (free|premium_basic|premium_plus|premium_pro), price BIGINT (CHECK >= 0, VND), billing_period TEXT, active BOOLEAN.
- subscription: id PK, user_id -> app_user(id), plan_id -> plan_catalog(id), started_at/renews_at TS (CHECK renews_at > started_at), status TEXT (active|past_due|canceled|expired). Lưu ý: chuẩn hóa so với §3.4 - giá nằm ở plan_catalog, subscription tham chiếu plan_id thay vì lưu cột price/tier trực tiếp.
- Được FK bởi: TASK-BILL-003 (payment).

### payment (TASK-BILL-003)
- id PK, order_ref TEXT UNIQUE (idempotency), subscription_id -> subscription(id), gateway TEXT, amount BIGINT (CHECK >= 0), fee BIGINT (CHECK >= 0), status TEXT (pending|paid|failed|mismatch), transaction_id TEXT, paid_at TS.

### referral_code (TASK-BILL-004)
- id PK, user_id -> app_user(id) UNIQUE (một mã/người), code TEXT UNIQUE, uses INT DEFAULT 0.
- Là đích của app_user.referral_code_id (FK thêm sau khi bảng này tồn tại).

### plan_feature (TASK-BILL-005)
- id PK, tier TEXT (free|premium_basic|premium_plus|premium_pro), feature_key TEXT, limit_value BIGINT (-1 = unlimited; 0 = không quyền; >0 = giới hạn), UNIQUE(tier, feature_key).

## Notification

### notification (TASK-NOTIF-001)
- id PK, user_id -> app_user(id), channel TEXT (push|email|sms), template TEXT, payload JSONB, scheduled_at/sent_at TS, status TEXT (pending|queued|sent|failed).
- Mở rộng: TASK-NOTIF-003 `ALTER TABLE notification ADD COLUMN attempts INT, lease_until TS, last_error TEXT` (claim/lease cho fan-out at-least-once idempotent).
- Được FK bởi: TASK-NOTIF-003 (notification_dlq).

### user_channel_token (TASK-NOTIF-001)
- PK(user_id, channel, platform), user_id -> app_user(id), channel TEXT (push|email|sms), platform TEXT (ios|android|web|email|sms), address TEXT (FCM/APNs token | email | phone), verified BOOLEAN, updated_at TS.
- Cột platform tách kênh push: FCM (TASK-NOTIF-002) nhặt platform IN ('android','web'), APNs (TASK-NOTIF-005) nhặt platform='ios'; nhờ platform trong PK, một user có token nhiều thiết bị.

### notification_dlq (TASK-NOTIF-003)
- id PK, notification_id -> notification(id), channel TEXT, payload JSONB, attempts INT, last_error TEXT, reason TEXT (permanent|max_attempts), dead_at TS.

## Compliance / PDPL

### consent_policy / consent_record (TASK-COMPLY-001)
- consent_policy: id PK, purpose_key TEXT, version INT, title_vi/body_vi TEXT, effective_from TS, UNIQUE(purpose_key, version).
- consent_record: id PK, user_id -> app_user(id), purpose_key TEXT, policy_version INT, granted BOOLEAN, source TEXT (web|extension|mobile), ts TS, ip INET, user_agent TEXT, composite FK (purpose_key, policy_version) -> consent_policy(purpose_key, version).

### processing_activity / dpia / tia (TASK-COMPLY-002)
- processing_activity: id PK, name TEXT, purpose_key TEXT, data_categories TEXT[], cross_border BOOLEAN, recipient_country TEXT (CHECK cross_border -> NOT NULL).
- dpia: id PK, activity_id -> processing_activity(id), version INT, risk_level TEXT (low|medium|high), status TEXT, UNIQUE(activity_id, version).
- tia: id PK, dpia_id -> dpia(id), recipient_country TEXT, safeguard_vi TEXT.

### dsar_request (TASK-COMPLY-003)
- id PK, user_id -> app_user(id), kind TEXT (access|rectify|erase|portability), status TEXT, requested_at/sla_due_at/completed_at TS, note TEXT.

### breach_incident (TASK-COMPLY-004)
- id PK, summary TEXT, severity TEXT (low|medium|high|critical), status TEXT (detected|triaged|notified_authority|notified_subjects|closed), acknowledged_at TS (đồng hồ 72h đếm từ đây), source_ref TEXT (alert/trace của TASK-INFRA-004).

### country_rule (TASK-COMPLY-006)
- id PK, country TEXT (CHECK len = 2), gate_key TEXT, value JSONB, version INT, effective_from TS, UNIQUE(country, gate_key, version).

### ecommerce_obligation / yearly_transaction_count / compliance_threshold (TASK-COMPLY-008)
- ecommerce_obligation: id PK, obligation_key TEXT, description_vi TEXT, status TEXT, due_at/completed_at TS, source_law TEXT (ND_52_2013|ND_85_2021|DRAFT_2025), version INT, UNIQUE(obligation_key, version).
- yearly_transaction_count: year INT PK, count BIGINT.
- compliance_threshold: key TEXT, value BIGINT, version INT, UNIQUE(key, version).

## Trust / anti-fraud

### fraud_signal / account_link_edge (TASK-TRUST-004)
- fraud_signal: id PK, subject_user_id -> app_user(id), kind TEXT (velocity|graph|rule), risk_score SMALLINT (CHECK 0..100), reasons JSONB, status TEXT (open|investigating|confirmed_fraud|cleared), UNIQUE(subject_user_id, kind).
- account_link_edge: a_user/b_user BIGINT (cả hai -> app_user(id)), link_type TEXT (referral|payment|device), weight REAL, PK(a_user, b_user, link_type), CHECK(a_user < b_user).

### payout_hold (TASK-TRUST-005)
- PK conversion_id -> affiliate_conversion(id), beneficiary -> app_user(id), payout_amount BIGINT (CHECK >= 0, VND), confirmed_at/eligible_at TS, hold_reason TEXT, status TEXT (holding|under_investigation|released|denied), released_at TS.

### device_fingerprint (TASK-TRUST-006)
- device_hash TEXT (one-way hash salt server, không thuộc tính thô), user_id -> app_user(id), first_seen/last_seen TS, PK(device_hash, user_id).

## B2B

### market_trend_daily (TASK-B2B-001)
- Bảng xu hướng thị trường ẩn danh, qua cổng k-anonymity (K_MIN = 50). Cố ý KHÔNG chứa product_id/shop_id/user_id; chỉ đọc từ price_daily + tracked_product.category_id.
- Cột chính: category_id BIGINT, platform_id SMALLINT, day DATE, median_p/p25_p/p75_p BIGINT (NULL khi suppressed), avg_discount_pct NUMERIC(5,2), sku_count INT, suppressed BOOLEAN, PRIMARY KEY(category_id, platform_id, day). Dùng soft-reference (không FK cứng) để cách ly pipeline phân tích khỏi bảng OLTP.

### b2b_subscription (TASK-B2B-002)
- id PK, org_name TEXT, tier TEXT (basic|pro|enterprise), max_categories INT, history_days INT, can_export BOOLEAN, status TEXT (active|past_due|canceled), expires_at TS.

### seller_owned_sku (TASK-B2B-003)
- id PK, seller_org_id BIGINT (soft-ref tổ chức B2B), shop_id TEXT, product_id BIGINT (soft-ref tracked_product.id, không FK cứng), verified BOOLEAN, UNIQUE(seller_org_id, product_id).

### api_key / api_usage (TASK-B2B-004)
- api_key: id PK, prefix TEXT UNIQUE, secret_hash TEXT (băm có salt, không cleartext), org_name TEXT, tier TEXT (free|pro|enterprise), rate_per_min INT, monthly_quota INT, revoked BOOLEAN.
- api_usage: id PK, api_key_id -> api_key(id), endpoint TEXT, ts TS, status_code SMALLINT (không ghi nội dung dữ liệu trả về).

## Scrape & ML

### scrape_job (TASK-SCRAPE-001)
- PK product_id -> tracked_product(id), platform_id -> platform(id), tier scrape_tier (ENUM, DEFAULT 'cold'), next_run_at TS, attempts INT, last_status TEXT, locked_until TS.

### proxy_usage (TASK-SCRAPE-004)
- day DATE, provider TEXT, tier TEXT, country TEXT, bytes_used BIGINT, cost_micro_usd BIGINT (micro-USD số nguyên), PRIMARY KEY(day, provider, tier, country).

### price_forecast (TASK-DEAL-004; ghi bởi DEAL-004 prophet/category_prior và DEAL-005 lgbm; đọc bởi DEAL-006)
- product_id -> tracked_product(id), run_date DATE, horizon_day SMALLINT (CHECK 1..14), yhat/yhat_lower/yhat_upper BIGINT (VND), p_bottom_14d REAL (CHECK 0..1), model_kind TEXT (prophet|category_prior|lgbm), scored_at TIMESTAMPTZ (DEAL-006 lọc độ tươi), PRIMARY KEY(product_id, run_date, horizon_day).

### feature_store (TASK-DEAL-005)
- product_id -> tracked_product(id), as_of_date DATE, các cột feature số (day_of_month, trailing_min_30/60/90 BIGINT, price_vs_median90 REAL, volatility_30d REAL, category_seasonality REAL, flash_sale_flag BOOLEAN, platform_id SMALLINT), future_min_price_14d BIGINT (NHÃN, nullable, chỉ điền lúc train), PRIMARY KEY(product_id, as_of_date).

### bottom_alert_log (TASK-DEAL-006)
- user_id BIGINT, product_id -> tracked_product(id), fired_on DATE (Asia/Ho_Chi_Minh), p_bottom DOUBLE PRECISION (CHECK > 0.7), fired_at TS, UNIQUE(user_id, product_id, fired_on) - idempotent 1 alert/cặp/ngày.

## Bảng lõi từ §3.4

Xác nhận mỗi bảng lõi trong DDL phác thảo §3.4 ánh xạ tới đúng một task sở hữu:

- platform -> TASK-INFRA-002
- app_user -> TASK-INFRA-002 (mở rộng pwd_hash/referral_code_id bởi TASK-AUTH-001)
- platform_account -> TASK-AUTH-003
- tracked_product -> TASK-PRICE-001
- price_snapshot -> TASK-PRICE-002 (hypertable)
- price_daily -> TASK-PRICE-002 (continuous aggregate, không phải bảng thường)
- wishlist -> TASK-TRACK-002
- wishlist_item -> TASK-TRACK-002
- alert_rule -> TASK-TRACK-003
- alert -> TASK-TRACK-003
- voucher_catalog -> TASK-CART-001
- cart_snapshot -> TASK-CART-002
- cart_item -> TASK-CART-002
- affiliate_click -> TASK-AFFIL-001
- affiliate_conversion -> TASK-AFFIL-001
- subscription -> TASK-BILL-001 (chuẩn hóa: giá tách sang plan_catalog)
- payment -> TASK-BILL-003
- referral_code -> TASK-BILL-004
- notification -> TASK-NOTIF-001 (mở rộng attempts/lease_until/last_error bởi TASK-NOTIF-003)
- user_activity_coin_task -> TASK-CART-006

Cả 20 bảng lõi §3.4 đều có task sở hữu. Các task phát triển schema sâu hơn §3.4 ở vài chỗ (thêm cột CHECK, tách normalization, thêm bảng phụ trợ) nhưng không bỏ sót bảng lõi nào.

## Postgres vs TimescaleDB

Toàn bộ bảng là PostgreSQL 16 thường, trừ lớp time-series giá:

- `price_snapshot` (TASK-PRICE-002): TimescaleDB hypertable. `create_hypertable('price_snapshot','ts', chunk_time_interval => INTERVAL '7 days')`. Nén `compress_segmentby='product_id'` + `add_compression_policy(30 days)`. Retention raw `add_retention_policy(18 months)`. Giá BIGINT (VND), ghi delta-only - chỉ INSERT khi (price, list_price, stock, flash_sale) đổi.
- `price_daily` (TASK-PRICE-002): continuous aggregate (`WITH (timescaledb.continuous)`) trên price_snapshot, bucket 1 ngày (min_p/max_p/close_p), refresh policy mỗi giờ (`add_continuous_aggregate_policy`), giữ vô thời hạn. Đây là nguồn cho biểu đồ giá và pipeline B2B, không phải bảng vật lý ghi trực tiếp.

Index time-series chính là PK (product_id, ts) cộng index canonical_key trên tracked_product phục vụ JOIN gom SKU chéo sàn (§3.4, §5.6).
