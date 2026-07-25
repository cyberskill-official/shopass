---
id: TASK-B2B-004
title: "Premium API access cho dev/doanh nghiệp - API key có tier (free/pro/enterprise), rate-limit per-key qua gateway, phục vụ market_trend_daily đã ẩn danh, KHÔNG mở raw hay user-level qua API"
module: B2B
priority: COULD
status: ready_to_implement
verify: T
phase: P3
milestone: P3 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-INFRA-001, TASK-B2B-001, TASK-B2B-002]
depends_on: [TASK-INFRA-001, TASK-B2B-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §6 mục 11 (Premium API access cho dev/doanh nghiệp, API tiers)"
  - "docs/... §3.1 (API Gateway/BFF rate-limit + auth), §3.7 (API design), §5.5 (PDPL)"
source_decisions:
  - "DEC-B2B-30: xác thực API public bằng API key (header X-API-Key) có prefix định danh + phần bí mật băm tại nghỉ (KHÔNG lưu key cleartext) - tách khỏi JWT người dùng cuối"
  - "DEC-B2B-31: mỗi api_key gắn tier (free/pro/enterprise); tier quyết định rate-limit (req/phút) + quota tháng + phạm vi endpoint cho phép"
  - "DEC-B2B-32: rate-limit thực thi per-key tại API Gateway (TASK-INFRA-001); 429 + header Retry-After khi vượt; KHÔNG để rate-limit ở mỗi service riêng lẻ né được"
  - "DEC-B2B-33: API public CHỈ phục vụ dữ liệu đã ẩn danh từ market_trend_daily (qua TASK-B2B-001) - KHÔNG bao giờ expose price_snapshot raw, price_daily raw, hay bất kỳ endpoint user-level nào ra API key"
  - "DEC-B2B-34: thu hồi key tức thời (revoked=true) phải có hiệu lực ngay ở lớp xác thực - key bị lộ vô hiệu trong vòng một chu kỳ cache ngắn"
  - "DEC-B2B-35: mọi lời gọi API ghi usage (api_key_id, endpoint, ts, status) để tính quota + hóa đơn; usage KHÔNG ghi nội dung dữ liệu trả về"

language: "PostgreSQL 16 + Go 1.22 (b2b-svc) + cấu hình API Gateway (TASK-INFRA-001)"
service: shopass/services/b2b/
new_files:
  - services/b2b/migrations/0004_api_key.sql
  - services/b2b/internal/apikey/auth.go
  - services/b2b/internal/apikey/ratelimit.go
  - services/b2b/internal/apikey/usage.go
  - services/b2b/internal/api/public_trend_handler.go
  - services/b2b/internal/apikey/auth_test.go
  - services/b2b/internal/apikey/ratelimit_test.go
modified_files:
  - services/b2b/internal/api/router.go            # đăng ký /public/v1/trends sau middleware api-key
allowed_tools:
  - file_read: services/b2b/**
  - file_write: services/b2b/**
  - bash: cd services/b2b && go test ./...
disallowed_tools:
  - lưu API key dạng cleartext (vi phạm DEC-B2B-30)
  - expose price_snapshot/price_daily raw hay endpoint user-level qua API key (vi phạm DEC-B2B-33)
  - bỏ qua rate-limit per-key hoặc đặt rate-limit chỉ ở service lẻ (vi phạm DEC-B2B-32)

effort_hours: 6
sub_tasks:
  - "1.0h: 0004_api_key.sql - bảng api_key (prefix, secret_hash, tier, rate_per_min, monthly_quota, revoked) + api_usage"
  - "1.0h: auth.go - parse X-API-Key, tách prefix, so secret_hash (argon2id/sha256+salt), chặn revoked"
  - "1.0h: ratelimit.go - token bucket per-key theo rate_per_min; trả 429 + Retry-After; tích hợp điểm thực thi gateway"
  - "1.0h: public_trend_handler.go - phục vụ market_trend_daily (ô đã phát hành) cho key hợp lệ; 200/401/403/429"
  - "0.5h: usage.go - ghi api_usage mỗi lời gọi (không ghi nội dung dữ liệu)"
  - "1.0h: auth_test.go - key hợp lệ pass; sai/revoked -> 401; tier không có quyền endpoint -> 403"
  - "0.5h: ratelimit_test.go - vượt rate_per_min -> 429 + Retry-After; reset theo cửa sổ"

risk_if_skipped: "TASK-B2B-004 mở dòng doanh thu API tiers (§6 mục 11) cho lập trình viên và doanh nghiệp tiêu thụ dữ liệu xu hướng giá theo chương trình. Đây là bề mặt công khai nhất của hệ thống nên rủi ro cao nhất: (1) nếu API public vô tình expose price_snapshot raw hay endpoint user-level thì toàn bộ moat dữ liệu độc quyền và k-anonymity của TASK-B2B-001 bị mở toang ra Internet qua một API key - thảm họa PDPL (§5.5) và mất tài sản dữ liệu; (2) nếu rate-limit per-key không thực thi ở gateway thì một key có thể rút cạn dữ liệu hoặc làm quá tải hệ thống; (3) nếu lưu key cleartext hoặc thu hồi key không có hiệu lực ngay thì một key lộ trở thành cửa hậu lâu dài. API key auth + rate-limit + giới hạn endpoint chỉ-ẩn-danh là ba điều kiện bắt buộc trước khi mở API ra ngoài."
---

## §1 - Mô tả (BCP-14 normative)

Service B2B **MUST** cung cấp API public xác thực bằng API key có tier, rate-limit per-key tại API Gateway, và CHỈ phục vụ dữ liệu đã ẩn danh từ `market_trend_daily`. Hợp đồng:

1. **MUST** xác thực API public bằng API key qua header `X-API-Key` (DEC-B2B-30). Key có dạng `prefix.secret`; lưu trữ chỉ `prefix` (tra cứu) + `secret_hash` (băm có salt). **MUST NOT** lưu phần bí mật dạng cleartext.
2. **MUST** định nghĩa bảng `api_key (id, prefix, secret_hash, org_name, tier, rate_per_min, monthly_quota, revoked, created_at)`; `tier` thuộc `{free, pro, enterprise}` (DEC-B2B-31).
3. **MUST** từ chối key sai, không tồn tại, hoặc `revoked=true` với `401`. Thu hồi (`revoked=true`) **MUST** có hiệu lực trong vòng một chu kỳ cache ngắn (<= 60 giây) (DEC-B2B-34).
4. **MUST** thực thi rate-limit per-key tại API Gateway (TASK-INFRA-001) theo `rate_per_min` của tier (DEC-B2B-32); vượt -> `429` + header `Retry-After`. **MUST NOT** đặt rate-limit chỉ ở service lẻ để có thể né.
5. **MUST** chỉ phục vụ dữ liệu đã ẩn danh từ `market_trend_daily` (ô `suppressed=false`, qua TASK-B2B-001) (DEC-B2B-33). API public **MUST NOT** expose `price_snapshot` raw, `price_daily` raw, hay bất kỳ endpoint user-level nào (wishlist/alert/cart/auth).
6. **MUST** gating phạm vi endpoint theo tier: tier không có quyền endpoint yêu cầu -> `403`. (vd `free` chỉ đọc dải tổng hợp giới hạn; `enterprise` mở phạm vi rộng hơn.)
7. **MUST** ghi `api_usage (api_key_id, endpoint, ts, status_code)` cho mỗi lời gọi để tính quota + hóa đơn (DEC-B2B-35); **MUST NOT** ghi nội dung dữ liệu trả về.
8. **MUST** chặn khi vượt `monthly_quota` của tier với `429` (hoặc `402` nếu chính sách yêu cầu nâng cấp) - phân biệt rõ với rate-limit per-phút.
9. **MUST** expose endpoint public `GET /public/v1/trends?category_id=...&platform_id=...&from=...&to=...` trả chuỗi ô `market_trend_daily` đã phát hành.
10. **MUST** phân biệt mã trạng thái: `200` (dữ liệu), `401` (key sai/revoked), `403` (tier không có quyền endpoint), `429` (vượt rate/quota), `400` (tham số sai).
11. **SHOULD** phát OTel metric: `api_request_total{tier,status}` (counter), `api_rate_limited_total{tier}` (counter), `api_request_duration_ms` (histogram).
12. **MUST** đảm bảo so sánh `secret_hash` dùng so khớp hằng-thời-gian (constant-time) để tránh lộ qua timing.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao API key tách khỏi JWT người dùng (DEC-B2B-30)? Người tiêu thụ API public là server/script của doanh nghiệp, không phải trình duyệt người dùng. JWT có vòng đời ngắn + refresh hợp cho phiên người dùng; API key dài hạn, gắn tổ chức + tier + quota hợp cho tích hợp máy-máy. Hai cơ chế khác mục đích nên tách.

Vì sao chỉ lưu hash + prefix (DEC-B2B-30, §1 #1, #12)? Bảng `api_key` là mục tiêu giá trị cao. Lưu cleartext nghĩa là rò DB = rò mọi key. Lưu `secret_hash` + so khớp hằng-thời-gian làm DB lộ vẫn không cho kẻ tấn công khôi phục key dùng được. `prefix` công khai chỉ để tra cứu nhanh dòng nào trước khi so hash.

Vì sao rate-limit ở gateway, không ở service lẻ (DEC-B2B-32, §1 #4)? Nếu mỗi service tự giới hạn thì một key gọi vòng qua nhiều service có thể vượt tổng. Gateway là điểm vào duy nhất - đặt rate-limit per-key ở đó đảm bảo trần là trần thật, không né được. Đây cũng đúng vai trò của API Gateway đã dựng ở TASK-INFRA-001.

Vì sao API public chỉ mở market_trend_daily (DEC-B2B-33, §1 #5)? Đây là điểm dễ sai nhất và hậu quả lớn nhất. API public là bề mặt ra Internet. Nếu nó chạm được raw hay user-level thì một API key biến thành ống dẫn rút toàn bộ dữ liệu chi tiết - phá k-anonymity và moat. Quy tắc cứng: API public chỉ đọc ô đã phát hành của TASK-B2B-001, không có route nào khác.

Vì sao thu hồi key phải có hiệu lực ngay (DEC-B2B-34, §1 #3)? Khi key lộ (rò vào repo công khai, log), phải vô hiệu nhanh. Nếu cache tier/quyền quá lâu thì key thu hồi vẫn dùng được hàng giờ. Giới hạn cache <= 60 giây cân bằng giữa hiệu năng và độ trễ thu hồi.

Vì sao tách rate-limit (phút) khỏi quota (tháng) (§1 #8)? Hai trục kiểm soát khác nhau: rate-limit chống đột biến tức thời (bảo vệ hệ thống), quota chống dùng quá gói (bảo vệ doanh thu). Một key có thể chưa chạm trần phút nhưng đã hết quota tháng - cần báo lý do khác nhau để khách hiểu đúng.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/b2b/migrations/0004_api_key.sql
CREATE TABLE api_key (
  id            BIGSERIAL   PRIMARY KEY,
  prefix        TEXT        NOT NULL UNIQUE,     -- phần công khai để tra cứu
  secret_hash   TEXT        NOT NULL,            -- băm có salt; KHÔNG cleartext
  org_name      TEXT        NOT NULL,
  tier          TEXT        NOT NULL CHECK (tier IN ('free','pro','enterprise')),
  rate_per_min  INTEGER     NOT NULL CHECK (rate_per_min > 0),
  monthly_quota INTEGER     NOT NULL CHECK (monthly_quota > 0),
  revoked       BOOLEAN     NOT NULL DEFAULT false,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE api_usage (
  id          BIGSERIAL   PRIMARY KEY,
  api_key_id  BIGINT      NOT NULL REFERENCES api_key(id),
  endpoint    TEXT        NOT NULL,
  ts          TIMESTAMPTZ NOT NULL DEFAULT now(),
  status_code SMALLINT    NOT NULL
  -- KHÔNG ghi nội dung dữ liệu trả về
);
CREATE INDEX idx_usage_key_month ON api_usage (api_key_id, ts);
```

### Auth (Go)

```go
// services/b2b/internal/apikey/auth.go
// authenticate tách prefix.secret, tra dòng theo prefix, so hash hằng-thời-gian,
// chặn revoked. Trả APIKey (kèm tier) hoặc lỗi 401.
func (a *Auth) authenticate(ctx context.Context, raw string) (*APIKey, error) {
    prefix, secret, ok := splitKey(raw) // "pfx_live_ab12.SECRET"
    if !ok {
        return nil, ErrUnauthorized
    }
    k, err := a.cache.GetByPrefix(ctx, prefix) // cache <= 60s để thu hồi nhanh
    if err != nil || k == nil || k.Revoked {
        return nil, ErrUnauthorized
    }
    if !verifyHashConstantTime(secret, k.SecretHash) {
        return nil, ErrUnauthorized
    }
    return k, nil
}
```

### Rate-limit + endpoint gating (§1 #4, #6)

```go
// services/b2b/internal/apikey/ratelimit.go
// allow thực thi token-bucket per-key; điểm này nằm ở lớp gateway (TASK-INFRA-001).
func (rl *RateLimiter) allow(keyID int64, ratePerMin int) (ok bool, retryAfter time.Duration) {
    b := rl.bucket(keyID, ratePerMin)
    if b.take() {
        return true, 0
    }
    return false, b.untilNextToken() // -> header Retry-After khi 429
}

// allowedEndpoint kẹp phạm vi theo tier; tier thiếu quyền -> 403.
func allowedEndpoint(tier, endpoint string) bool {
    switch tier {
    case "enterprise":
        return true
    case "pro":
        return endpoint == "/public/v1/trends"
    default: // free
        return endpoint == "/public/v1/trends" // phạm vi/độ sâu giới hạn thêm ở handler
    }
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `api_key` + `api_usage` tồn tại; `api_key.secret_hash` không phải cleartext.
2. Gọi `/public/v1/trends` với key hợp lệ -> `200` + chuỗi ô `market_trend_daily` đã phát hành.
3. Key sai/không tồn tại -> `401`.
4. Key `revoked=true` -> `401`; sau khi set revoked, có hiệu lực trong <= 60 giây (qua hết cache).
5. Vượt `rate_per_min` -> `429` + header `Retry-After` > 0.
6. Vượt `monthly_quota` -> `429` (hoặc `402` theo chính sách), phân biệt với rate-limit phút.
7. Tier không có quyền endpoint -> `403`.
8. Không có route public nào trả `price_snapshot` raw, `price_daily` raw, hay dữ liệu user-level (review router + test).
9. Mỗi lời gọi ghi một dòng `api_usage` với `endpoint` + `status_code`, KHÔNG có nội dung dữ liệu.
10. So khớp secret dùng hàm hằng-thời-gian (review + test không phụ thuộc timing để pass/fail logic).
11. Phản hồi `/public/v1/trends` chỉ chứa chỉ số tổng hợp (không khóa định danh).

---

## §5 - Kiểm thử (verification)

```go
// services/b2b/internal/apikey/auth_test.go
func TestAuth_ValidKey_OK(t *testing.T) {
    a, raw := setupKeyWith(t, "pro", false) // tier pro, chưa revoke
    k, err := a.authenticate(ctx, raw)
    require.NoError(t, err)
    require.Equal(t, "pro", k.Tier)
}

func TestAuth_WrongSecret_401(t *testing.T) {
    a, raw := setupKeyWith(t, "pro", false)
    bad := mangleSecret(raw)
    _, err := a.authenticate(ctx, bad)
    require.ErrorIs(t, err, ErrUnauthorized)
}

func TestAuth_Revoked_401(t *testing.T) {
    a, raw := setupKeyWith(t, "pro", true) // revoked
    _, err := a.authenticate(ctx, raw)
    require.ErrorIs(t, err, ErrUnauthorized)
}

func TestAuth_RevocationTakesEffect(t *testing.T) {
    a, raw := setupKeyWith(t, "pro", false)
    _, err := a.authenticate(ctx, raw)
    require.NoError(t, err)
    a.revoke(ctx, raw)
    a.advanceCache(61 * time.Second) // qua chu kỳ cache <= 60s
    _, err = a.authenticate(ctx, raw)
    require.ErrorIs(t, err, ErrUnauthorized)
}

// services/b2b/internal/apikey/ratelimit_test.go
func TestRate_ExceedsLimit_429(t *testing.T) {
    rl := newRateLimiter()
    for i := 0; i < 60; i++ {
        ok, _ := rl.allow(7, 60)
        require.True(t, ok)
    }
    ok, retry := rl.allow(7, 60) // request thứ 61 trong phút
    require.False(t, ok)
    require.Greater(t, retry, time.Duration(0))
}

func TestQuota_ExceedsMonthly_429(t *testing.T) {
    rl := newRateLimiter()
    const monthly = 10
    for i := 0; i < monthly; i++ {
        ok := rl.allowMonthly(7, monthly) // trong hạn tháng
        require.True(t, ok)
    }
    ok := rl.allowMonthly(7, monthly) // request thứ 11 trong tháng
    require.False(t, ok)              // vượt monthly_quota -> 429 (phân biệt rate phút, §4 #6)
}

func TestEndpoint_FreeTierDenied_403(t *testing.T) {
    require.False(t, allowedEndpoint("free", "/public/v1/admin"))
    require.True(t, allowedEndpoint("free", "/public/v1/trends"))
    require.True(t, allowedEndpoint("enterprise", "/public/v1/anything"))
}

func TestPublic_NoRawRoutes(t *testing.T) {
    routes := registeredPublicRoutes(t)
    for _, r := range routes {
        require.NotContains(t, r, "price-history")
        require.NotContains(t, r, "snapshot")
        require.NotContains(t, r, "wishlist")
        require.NotContains(t, r, "cart")
    }
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0004 (api_key + api_usage) -> auth.go (xác thực key + cache thu hồi) -> ratelimit.go (token-bucket + endpoint gating) -> public_trend_handler.go (phục vụ market_trend_daily) -> usage.go -> tests. Rate-limit per-key cắm vào điểm thực thi của API Gateway (TASK-INFRA-001) - task này cung cấp logic + cấu hình, gateway là nơi chạy. Phát key (sinh secret, băm, trả cleartext một lần duy nhất cho khách) là quy trình quản trị; task chỉ xác thực key đã phát.

---

## §7 - Phụ thuộc

- TASK-INFRA-001 - API Gateway/BFF là điểm thực thi rate-limit per-key + định tuyến `/public/v1/*`.
- TASK-B2B-001 - `market_trend_daily` + `QueryCells` (ô đã phát hành) là nguồn dữ liệu duy nhất API public phục vụ.
- TASK-B2B-002 (cùng module) - chia sẻ khái niệm tier/tổ chức B2B; api_key có thể đồng bộ tier với subscription.
- Extension/lib: thư viện băm (argon2id hoặc sha256+salt), `crypto/subtle` cho so khớp hằng-thời-gian; driver `pgx`.

---

## §8 - Payload ví dụ

### Gọi API public (header)

```
GET /public/v1/trends?category_id=7&platform_id=1&from=2026-05-22&to=2026-06-20
X-API-Key: pfx_live_ab12.s3cr3t_no_log
```

### Phản hồi (200) - chỉ chỉ số tổng hợp

```json
{
  "category_id": 7,
  "platform_id": 1,
  "cells": [
    { "day": "2026-06-19", "median_p": 318000, "p25_p": 249000, "p75_p": 405000, "avg_discount_pct": 13.80 },
    { "day": "2026-06-20", "median_p": 320000, "p25_p": 250000, "p75_p": 410000, "avg_discount_pct": 14.30 }
  ]
}
```

### Vượt rate-limit (429)

```
HTTP/1.1 429 Too Many Requests
Retry-After: 23
```

```json
{ "error": "rate_limited", "retry_after_seconds": 23, "message": "Vượt giới hạn req/phút của gói pro." }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Webhook đẩy dữ liệu xu hướng thay vì poll - thêm khi có nhu cầu khách hàng enterprise.
- API key scoping theo category cụ thể (key chỉ đọc category được mua) - mở rộng mô hình entitlement ở slice sau.
- OpenAPI spec công bố + portal tự phục vụ phát key - hạ tầng developer-experience giai đoạn sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Lưu key cleartext | review schema + migration | rò DB = rò mọi key | secret_hash băm có salt (DEC-B2B-30) |
| API public chạm raw/user-level | TestPublic_NoRawRoutes + review router | rò toàn bộ dữ liệu chi tiết ra Internet | Chỉ route /public/v1/trends -> market_trend_daily (DEC-B2B-33) |
| Rate-limit né được ở service lẻ | review điểm thực thi | một key rút cạn/quá tải | Rate-limit per-key tại gateway (DEC-B2B-32) |
| Key lộ dùng được lâu | TestAuth_RevocationTakesEffect | cửa hậu | Thu hồi có hiệu lực <= 60s (DEC-B2B-34) |
| Timing attack lên so secret | review dùng crypto/subtle | đoán key qua timing | So khớp hằng-thời-gian (§1 #12) |
| Vượt quota tháng nhầm thành lỗi khác | AC #6 | khách hiểu sai lý do | Tách 429 quota vs 429 rate (§1 #8) |
| Usage ghi nội dung dữ liệu | review usage.go | log phình + rủi ro lộ | api_usage chỉ endpoint+status (DEC-B2B-35) |
| Tier free gọi endpoint nhạy cảm | TestEndpoint_FreeTierDenied_403 | vượt quyền gói | allowedEndpoint gating theo tier |

---

## §11 - Ghi chú

- API public là bề mặt ra Internet rủi ro cao nhất - quy tắc cứng: chỉ phục vụ ô đã phát hành của market_trend_daily, không route nào khác.
- API key tách khỏi JWT người dùng: máy-máy dài hạn + tier + quota, khác phiên trình duyệt ngắn hạn.
- Lưu hash + so khớp hằng-thời-gian làm DB lộ vẫn không cho khôi phục key dùng được.
- Rate-limit ở gateway (không ở service lẻ) đảm bảo trần là trần thật; đúng vai trò TASK-INFRA-001.
- Thu hồi key phải nhanh (<= 60s) vì key lộ là sự cố cần phản ứng tức thì.
- Hai trục kiểm soát: rate-limit phút (bảo vệ hệ thống) tách khỏi quota tháng (bảo vệ doanh thu).

---

*Hết TASK-B2B-004. Status: ready_to_implement (mục tiêu audit 10/10).*
