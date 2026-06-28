---
id: FR-AFFIL-001
title: "Schema + tracking `affiliate_click` / `affiliate_conversion` - bảng ghi nhấp affiliate (user-initiated) + đối soát hoa hồng theo network; last-click attribution chuẩn bị cho postback"
module: AFFIL
priority: MUST
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-INFRA-002, FR-AFFIL-002, FR-AFFIL-003, FR-AFFIL-004, FR-TRUST-004, FR-TRUST-005, NFR-AFFIL-001]
depends_on: [FR-INFRA-002]
blocks: [FR-AFFIL-002, FR-TRUST-004, FR-TRUST-005]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model: affiliate_click, affiliate_conversion)"
  - "docs/... §4.2 (affiliate reality-check - mô hình hợp lệ duy nhất là user-initiated), §5.3 (gian lận: delay payout)"
source_decisions:
  - "DEC-AFFIL-01: affiliate_click chỉ được tạo khi user CHỦ ĐỘNG bấm 'Mua qua SănDeal'; mỗi bản ghi gắn sub_id duy nhất để đối soát postback (không cookie-stuffing nền)"
  - "DEC-AFFIL-02: sub_id là chuỗi đối soát do SănDeal sinh (định danh click), KHÔNG nhúng định danh người dùng sàn thật; map click -> user qua FK nội bộ"
  - "DEC-AFFIL-03: affiliate_conversion tham chiếu affiliate_click qua click_id (last-click attribution: một conversion thuộc về click gần nhất khớp sub_id)"
  - "DEC-AFFIL-04: tiền (order_value, commission) lưu BIGINT (VND, không thập phân) đồng nhất DEC-PRICE-05; status conversion theo vòng đời pending -> confirmed -> rejected"
  - "DEC-AFFIL-05: conversion bắt đầu ở status 'pending'; chỉ chuyển 'confirmed' khi network xác nhận (postback FR-AFFIL-003); cho phép delay payout chống gian lận (§5.3)"

language: "PostgreSQL 16; service Go 1.22 (affil-svc)"
service: shopass/services/affil/
new_files:
  - services/affil/migrations/0001_affiliate_click.sql
  - services/affil/migrations/0002_affiliate_conversion.sql
  - services/affil/internal/affil/types.go
  - services/affil/internal/affil/repo.go
  - services/affil/internal/affil/subid.go
  - services/affil/internal/affil/repo_test.go
  - services/affil/internal/affil/subid_test.go
modified_files: []
allowed_tools:
  - file_read: services/affil/**
  - file_write: services/affil/**
  - bash: cd services/affil && go test ./...
disallowed_tools:
  - tạo affiliate_click ngoài luồng user bấm (vi phạm DEC-AFFIL-01, NFR-AFFIL-001 - cookie-stuffing kiểu Honey)
  - lưu tiền dạng float/numeric thập phân (vi phạm DEC-AFFIL-04)
  - nhúng định danh người dùng sàn thật vào sub_id (vi phạm DEC-AFFIL-02, rò rỉ PII)
  - đặt conversion thẳng 'confirmed' khi chưa có postback network (vi phạm DEC-AFFIL-05)

effort_hours: 6
sub_tasks:
  - "0.5h: 0001_affiliate_click.sql - bảng affiliate_click + FK user/platform/product + UNIQUE(sub_id) + index clicked_at"
  - "0.5h: 0002_affiliate_conversion.sql - bảng affiliate_conversion + FK click_id + CHECK status + CHECK tiền >= 0"
  - "1.0h: subid.go - sinh sub_id ngẫu nhiên không đoán được, không nhúng PII; parse/validate"
  - "1.0h: types.go + repo.go - RecordClick (chỉ khi caller xác nhận user-initiated) + RecordConversion + ConfirmConversion"
  - "1.0h: repo_test.go - insert click, sub_id unique, conversion gắn click, vòng đời pending->confirmed, CHECK tiền âm"
  - "0.5h: subid_test.go - sub_id duy nhất qua nhiều lần sinh, không chứa PII, độ dài/entropy đủ"
  - "1.0h: OTel metric affiliate_click_recorded_total{platform,network} + affiliate_conversion_total{status}"

risk_if_skipped: "affiliate_click và affiliate_conversion là sổ cái của dòng doanh thu affiliate - dòng tiền chính tài trợ free-tier (§4.1). Không có chúng thì deep link user-initiated (FR-AFFIL-002) không có chỗ ghi nhận, postback network (FR-AFFIL-003) không có bảng để đối soát, và anti-fraud (FR-TRUST-004/005) không có dữ liệu để phát hiện gaming attribution. Nếu sub_id nhúng định danh người dùng sàn thật thì rò rỉ PII (vi phạm PDPL). Nếu tiền lưu float thì sai số tích lũy trên đối soát hoa hồng. Nếu conversion vào thẳng 'confirmed' không chờ postback thì SănDeal trả cashback (FR-AFFIL-005) cho đơn chưa được network xác nhận - mở cửa cho gian lận và lỗ. Đây là nền dữ liệu cho toàn bộ kiếm tiền affiliate compliant hậu-Honey."
---

## §1 - Mô tả (BCP-14 normative)

Service AFFIL **MUST** định nghĩa hai bảng sổ cái affiliate: `affiliate_click` (ghi mỗi lần user chủ động bấm mua qua SănDeal) và `affiliate_conversion` (ghi đơn hàng quy về một click, theo last-click attribution), với tiền lưu BIGINT VND và vòng đời status có kiểm soát. Hợp đồng:

1. **MUST** định nghĩa bảng `affiliate_click (id, user_id, platform_id, product_id, sub_id, clicked_at, network)` với `user_id` REFERENCES `app_user(id)`, `platform_id` REFERENCES `platform(id)`, `product_id` REFERENCES `tracked_product(id)`.
2. **MUST** đặt `sub_id TEXT NOT NULL UNIQUE`: mỗi click có một `sub_id` đối soát duy nhất, do SănDeal sinh (FR-AFFIL-001 `subid.go`), dùng để network postback ánh xạ ngược conversion về click (DEC-AFFIL-01).
3. `sub_id` **MUST NOT** nhúng định danh người dùng sàn thật hay PII (email, số điện thoại, tên); nó là một token đối soát ngẫu nhiên (DEC-AFFIL-02). Quan hệ click -> user nằm ở cột `user_id` nội bộ, không ở trong `sub_id`.
4. `affiliate_click` **MUST** chỉ được ghi qua hàm `RecordClick` mà caller (FR-AFFIL-002) khẳng định là user-initiated; KHÔNG có đường ghi click tự động nền (NFR-AFFIL-001). `network` ghi nguồn (`'involve_asia'`, `'accesstrade'`, ...) để đối soát đúng kênh.
5. **MUST** định nghĩa bảng `affiliate_conversion (id, click_id, order_value, commission, status, confirmed_at)` với `click_id` REFERENCES `affiliate_click(id)`.
6. **MUST** lưu `order_value` và `commission` dạng `BIGINT` (VND, không thập phân) - KHÔNG float/numeric (DEC-AFFIL-04) - và ràng buộc `order_value >= 0`, `commission >= 0` qua CHECK.
7. `affiliate_conversion.status` **MUST** thuộc tập `{'pending','confirmed','rejected'}` qua CHECK; conversion mới **MUST** bắt đầu ở `'pending'` (DEC-AFFIL-05). `confirmed_at` chỉ được set khi chuyển sang `'confirmed'`.
8. **MUST** áp last-click attribution ở mức dữ liệu (DEC-AFFIL-03): một conversion gắn vào đúng một `click_id`; khi postback (FR-AFFIL-003) mang `sub_id`, repo tra `affiliate_click` theo `sub_id` để lấy `click_id`. Nếu `sub_id` không khớp click nào -> conversion bị từ chối ghi (không tạo conversion mồ côi).
9. **MUST** expose hàm repo:
    - `RecordClick(ctx, c AffiliateClick) (int64, error)` - chèn một click user-initiated, trả `click_id`.
    - `RecordConversion(ctx, subID string, orderValue, commission int64, network string) (int64, error)` - tra click theo `sub_id`, chèn conversion `pending`.
    - `ConfirmConversion(ctx, conversionID int64) error` - chuyển `pending` -> `confirmed`, set `confirmed_at = now()`.
    - `RejectConversion(ctx, conversionID int64, reason string) error` - chuyển sang `rejected`.
10. **MUST** đánh index `idx_click_subid` trên `affiliate_click(sub_id)` (đã UNIQUE) cho tra cứu postback nhanh, và `idx_click_user_time` trên `(user_id, clicked_at DESC)` cho báo cáo và anti-fraud (FR-TRUST-004 velocity).
11. **MUST** idempotent với postback lặp: cùng một `sub_id` postback hai lần (network retry) **MUST NOT** tạo hai conversion - dùng `UNIQUE` trên `affiliate_conversion(click_id)` hoặc kiểm tồn tại trước khi chèn (một click sinh tối đa một conversion).
12. **SHOULD** phát OTel: `affiliate_click_recorded_total{platform_id, network}` (counter), `affiliate_conversion_total{status}` (counter), `affiliate_conversion_value_vnd` (histogram order_value).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao sub_id ngẫu nhiên không nhúng PII (DEC-AFFIL-02)?** `sub_id` đi ra ngoài hệ thống: nó nằm trong deep link affiliate và quay lại qua postback của network. Nếu nhúng `user_id` thật hay email vào đó, ta rò rỉ định danh người dùng cho bên thứ ba và vi phạm tối thiểu hóa dữ liệu của PDPL. Một token ngẫu nhiên đủ entropy làm khóa đối soát mà không tiết lộ ai; ánh xạ về user giữ kín ở cột `user_id` nội bộ.

**Vì sao click chỉ ghi khi user-initiated (DEC-AFFIL-01, NFR-AFFIL-001)?** Đây là lằn ranh đạo đức trung tâm của SănDeal hậu-Honey (§4.2). Honey âm thầm thay cookie affiliate nền, không cần user bấm - chính điều khiến nó bị gỡ và bị Chrome cấm. Mô hình affiliate hợp lệ duy nhất là user chủ động bấm "Mua qua SănDeal". Vì vậy ở mức schema, không tồn tại đường ghi click tự động: bảng chỉ nhận click qua một hàm mà caller phải là luồng người dùng bấm thật (FR-AFFIL-002).

**Vì sao tiền BIGINT VND (DEC-AFFIL-04)?** `order_value` và `commission` là tiền tệ VN, luôn số nguyên đồng. Float gây sai số tích lũy khi cộng dồn hoa hồng qua hàng nghìn conversion để đối soát với báo cáo network. BIGINT chính xác tuyệt đối và đủ lớn cho mọi giá trị đơn hàng thực tế.

**Vì sao conversion bắt đầu 'pending' và chỉ 'confirmed' khi postback (DEC-AFFIL-05)?** Một đơn hàng được tạo không có nghĩa hoa hồng đã chắc chắn: người mua có thể hủy đơn, trả hàng, hoặc network từ chối do nghi gian lận. Nếu ta coi mọi conversion là chắc ngay và trả cashback (FR-AFFIL-005) liền, SănDeal lỗ khi đơn bị đảo. Vòng đời `pending -> confirmed/rejected` với delay payout là best practice ngành (§5.3) để chống gaming attribution.

**Vì sao một click tối đa một conversion (§1 #11)?** Network có thể postback lặp do retry. Nếu mỗi postback tạo một conversion, sổ cái phình và tiền hoa hồng bị đếm trùng. Ràng buộc một-một giữa click và conversion (qua UNIQUE click_id) làm postback idempotent, giữ đối soát đúng.

**Vì sao last-click attribution ở mức dữ liệu (DEC-AFFIL-03)?** Cả ba sàn dùng last-click (§4.2): conversion thuộc về click affiliate gần nhất. Bằng cách postback mang `sub_id` và tra ngược ra đúng click đã sinh `sub_id` đó, ta thực thi last-click một cách xác định, không đoán.

---

## §3 - Hợp đồng API / DDL

### Migrations

```sql
-- services/affil/migrations/0001_affiliate_click.sql
CREATE TABLE affiliate_click (
  id          BIGSERIAL   PRIMARY KEY,
  user_id     BIGINT      NOT NULL REFERENCES app_user(id),
  platform_id SMALLINT    NOT NULL REFERENCES platform(id),
  product_id  BIGINT      REFERENCES tracked_product(id),
  sub_id      TEXT        NOT NULL UNIQUE,        -- token đối soát; KHÔNG nhúng PII
  network     TEXT        NOT NULL,               -- 'involve_asia','accesstrade',...
  clicked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_click_user_time ON affiliate_click (user_id, clicked_at DESC);
-- sub_id đã UNIQUE (B-tree); postback tra theo sub_id dùng chính unique index đó.

-- services/affil/migrations/0002_affiliate_conversion.sql
CREATE TABLE affiliate_conversion (
  id           BIGSERIAL   PRIMARY KEY,
  click_id     BIGINT      NOT NULL UNIQUE REFERENCES affiliate_click(id),  -- 1 click <= 1 conversion
  order_value  BIGINT      NOT NULL CHECK (order_value >= 0),   -- VND
  commission   BIGINT      NOT NULL CHECK (commission  >= 0),   -- VND
  status       TEXT        NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending','confirmed','rejected')),
  confirmed_at TIMESTAMPTZ
);
```

### Types (Go)

```go
// services/affil/internal/affil/types.go
type AffiliateClick struct {
    ID         int64     `db:"id"`
    UserID     int64     `db:"user_id"`
    PlatformID int16     `db:"platform_id"`
    ProductID  *int64    `db:"product_id"`
    SubID      string    `db:"sub_id"`
    Network    string    `db:"network"`
    ClickedAt  time.Time `db:"clicked_at"`
}

type AffiliateConversion struct {
    ID          int64      `db:"id"`
    ClickID     int64      `db:"click_id"`
    OrderValue  int64      `db:"order_value"`  // VND
    Commission  int64      `db:"commission"`   // VND
    Status      string     `db:"status"`       // pending|confirmed|rejected
    ConfirmedAt *time.Time `db:"confirmed_at"`
}
```

### Repo - last-click + vòng đời (Go)

```go
// services/affil/internal/affil/repo.go

// RecordConversion tra click theo sub_id (last-click), chèn conversion 'pending'.
// Trả ErrUnknownSubID nếu sub_id không khớp click nào -> KHÔNG tạo conversion mồ côi (§1 #8).
func (r *Repo) RecordConversion(ctx context.Context, subID string, orderValue, commission int64, network string) (int64, error) {
    var clickID int64
    err := r.pool.QueryRow(ctx,
        `SELECT id FROM affiliate_click WHERE sub_id = $1`, subID).Scan(&clickID)
    if errors.Is(err, pgx.ErrNoRows) {
        return 0, ErrUnknownSubID
    } else if err != nil {
        return 0, err
    }
    var id int64
    err = r.pool.QueryRow(ctx,
        `INSERT INTO affiliate_conversion (click_id, order_value, commission)
         VALUES ($1,$2,$3)
         ON CONFLICT (click_id) DO NOTHING                 -- postback lặp idempotent (§1 #11)
         RETURNING id`,
        clickID, orderValue, commission).Scan(&id)
    if errors.Is(err, pgx.ErrNoRows) {
        return 0, ErrConversionExists // click đã có conversion
    }
    metrics.ConversionRecorded("pending")
    return id, err
}

func (r *Repo) ConfirmConversion(ctx context.Context, id int64) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE affiliate_conversion
            SET status='confirmed', confirmed_at=now()
          WHERE id=$1 AND status='pending'`, id)  // chỉ pending -> confirmed
    return err
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `affiliate_click` và `affiliate_conversion` tồn tại với FK đúng tới `app_user`, `platform`, `tracked_product`.
2. `RecordClick` chèn một click -> 1 dòng; đọc lại đúng `user_id/platform_id/sub_id/network`.
3. Chèn hai click cùng `sub_id` -> lần hai vi phạm UNIQUE (sub_id duy nhất).
4. `sub_id` sinh ra không chứa chuỗi nào khớp PII (email/phone/`user_id` thật) - kiểm `subid_test`.
5. `RecordConversion(subID_hợp_lệ, ...)` -> tạo conversion `status='pending'`, `confirmed_at IS NULL`, gắn đúng `click_id`.
6. `RecordConversion(subID_không_tồn_tại, ...)` -> trả `ErrUnknownSubID`, KHÔNG tạo dòng conversion.
7. `RecordConversion` cùng `sub_id` (cùng click) hai lần -> chỉ một conversion (ON CONFLICT click_id), lần hai trả `ErrConversionExists`.
8. `INSERT affiliate_conversion` với `order_value < 0` hoặc `commission < 0` -> lỗi CHECK.
9. `INSERT` với `status='paid'` (ngoài tập) -> lỗi CHECK.
10. `ConfirmConversion(id)` trên conversion `pending` -> `status='confirmed'`, `confirmed_at` được set.
11. `ConfirmConversion(id)` trên conversion đã `confirmed` -> không đổi (chỉ pending mới chuyển).
12. Metric `affiliate_click_recorded_total` tăng khi RecordClick; `affiliate_conversion_total{status="pending"}` tăng khi RecordConversion.

---

## §5 - Kiểm thử (verification)

```go
// services/affil/internal/affil/repo_test.go
func TestRecordClick_Insert(t *testing.T) {
    r, uid, pid := setupWithUserProduct(t)
    id, err := r.RecordClick(ctx, AffiliateClick{
        UserID: uid, PlatformID: 1, ProductID: &pid,
        SubID: "sd_ab12cd34", Network: "involve_asia"})
    require.NoError(t, err)
    require.Greater(t, id, int64(0))
}

func TestRecordClick_SubIDUnique(t *testing.T) {
    r, uid, _ := setupWithUserProduct(t)
    c := AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_dup", Network: "accesstrade"}
    _, err1 := r.RecordClick(ctx, c)
    require.NoError(t, err1)
    _, err2 := r.RecordClick(ctx, c)
    require.Error(t, err2) // UNIQUE(sub_id)
}

func TestConversion_LastClickBySubID(t *testing.T) {
    r, uid, _ := setupWithUserProduct(t)
    r.RecordClick(ctx, AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_x", Network: "involve_asia"})
    cid, err := r.RecordConversion(ctx, "sd_x", 250_000, 12_000, "involve_asia")
    require.NoError(t, err)
    require.Greater(t, cid, int64(0))
    require.Equal(t, "pending", statusOf(t, r, cid))
}

func TestConversion_UnknownSubID_NoOrphan(t *testing.T) {
    r, _, _ := setupWithUserProduct(t)
    _, err := r.RecordConversion(ctx, "sd_unknown", 100_000, 5_000, "accesstrade")
    require.ErrorIs(t, err, ErrUnknownSubID)
    require.Equal(t, 0, countConversions(t, r))
}

func TestConversion_PostbackIdempotent(t *testing.T) {
    r, uid, _ := setupWithUserProduct(t)
    r.RecordClick(ctx, AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_y", Network: "involve_asia"})
    _, err1 := r.RecordConversion(ctx, "sd_y", 250_000, 12_000, "involve_asia")
    require.NoError(t, err1)
    _, err2 := r.RecordConversion(ctx, "sd_y", 250_000, 12_000, "involve_asia")
    require.ErrorIs(t, err2, ErrConversionExists) // một click một conversion
    require.Equal(t, 1, countConversions(t, r))
}

func TestConversion_NegativeMoney_Rejected(t *testing.T) {
    r, uid, _ := setupWithUserProduct(t)
    r.RecordClick(ctx, AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_z", Network: "involve_asia"})
    _, err := r.RecordConversion(ctx, "sd_z", -1, 0, "involve_asia")
    require.Error(t, err) // CHECK order_value >= 0
}

func TestConfirm_PendingToConfirmed(t *testing.T) {
    r, uid, _ := setupWithUserProduct(t)
    r.RecordClick(ctx, AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_c", Network: "involve_asia"})
    cid, _ := r.RecordConversion(ctx, "sd_c", 250_000, 12_000, "involve_asia")
    require.NoError(t, r.ConfirmConversion(ctx, cid))
    require.Equal(t, "confirmed", statusOf(t, r, cid))
}
```

```go
// services/affil/internal/affil/subid_test.go
func TestSubID_UniqueAndNoPII(t *testing.T) {
    seen := map[string]bool{}
    for i := 0; i < 1000; i++ {
        s := NewSubID()
        require.False(t, seen[s]) // duy nhất
        seen[s] = true
        require.NotContains(t, s, "@")        // không giống email
        require.GreaterOrEqual(t, len(s), 12) // đủ entropy
    }
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0001_affiliate_click.sql` -> `0002_affiliate_conversion.sql` -> `subid.go` (sinh token ngẫu nhiên qua `crypto/rand`, tiền tố `sd_` cho dễ nhận diện trong postback) -> `types.go` + `repo.go` (`RecordClick`, `RecordConversion` tra last-click, `ConfirmConversion`/`RejectConversion`) -> tests. `RecordClick` được FR-AFFIL-002 gọi từ handler `POST /v1/affiliate/link` (luồng user bấm). `RecordConversion`/`ConfirmConversion` được FR-AFFIL-003 gọi từ webhook postback của network. Driver `pgx`.

---

## §7 - Phụ thuộc

- **FR-INFRA-002** - bảng `app_user`, `platform`, `tracked_product` phải tồn tại trước (mọi FK của `affiliate_click`).
- **FR-AFFIL-002 (downstream)** - deep link user-initiated gọi `RecordClick` khi user bấm "Mua qua SănDeal".
- **FR-AFFIL-003 (downstream)** - tích hợp network + webhook postback gọi `RecordConversion`/`ConfirmConversion`.
- **FR-AFFIL-005 (downstream, P3)** - cashback layering đọc conversion `confirmed` để chia % cho user (hold tới khi confirm).
- **FR-TRUST-004 / FR-TRUST-005 (downstream)** - anti-fraud đọc `affiliate_click(user_id, clicked_at)` + conversion để phát hiện velocity/gaming attribution.
- Lib: `pgx`, `crypto/rand`.

---

## §8 - Payload ví dụ

### Ghi click (nội bộ, do FR-AFFIL-002 gọi sau khi user bấm)

```go
clickID, err := affilRepo.RecordClick(ctx, affil.AffiliateClick{
    UserID:     7,
    PlatformID: 1,            // shopee
    ProductID:  ptr(int64(90112)),
    SubID:      "sd_ab12cd34ef56",   // token đối soát ngẫu nhiên
    Network:    "involve_asia",
})
```

### Ghi conversion từ postback (nội bộ, do FR-AFFIL-003 gọi)

```sql
-- network postback mang sub_id; repo tra ngược ra click_id rồi chèn pending
SELECT id FROM affiliate_click WHERE sub_id = 'sd_ab12cd34ef56';
INSERT INTO affiliate_conversion (click_id, order_value, commission)
VALUES (1001, 250000, 12000)        -- VND
ON CONFLICT (click_id) DO NOTHING;
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Cookie window per-sàn/per-network (7 ngày vs 30/7 desktop/mobile - cần xác minh §10 tài liệu nguồn) - lưu ở bảng cấu hình network của FR-AFFIL-003, không đổi schema này.
- Đa conversion cho một click (đơn nhiều shop tách nhiều conversion) - hiện một-một; nới khi network thực tế trả tách dòng, đổi UNIQUE(click_id) sang khóa khác.
- Lưu raw payload postback để truy vết tranh chấp - thêm bảng phụ `affiliate_postback_log` ở FR-AFFIL-003.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Ghi click tự động nền (cookie-stuffing) | review + NFR-AFFIL-001 | bị Chrome gỡ, network đình chỉ | Chỉ ghi qua RecordClick user-initiated (DEC-AFFIL-01) |
| sub_id nhúng PII | subid test | rò rỉ định danh, vi phạm PDPL | Token ngẫu nhiên, map user nội bộ (DEC-AFFIL-02) |
| sub_id postback không khớp click | ErrUnknownSubID | conversion mồ côi | Từ chối ghi, không tạo dòng (§1 #8) |
| Postback lặp tạo conversion trùng | UNIQUE click_id | đếm hoa hồng trùng | ON CONFLICT DO NOTHING (§1 #11) |
| Tiền lưu float | review + types int64 | sai số đối soát | BIGINT VND (DEC-AFFIL-04) |
| order_value/commission âm | DB CHECK | sổ cái sai | CHECK >= 0 (§1 #6) |
| Conversion vào thẳng 'confirmed' | CHECK + repo path | trả cashback đơn chưa xác nhận | Bắt đầu pending, confirm qua postback (DEC-AFFIL-05) |
| status ngoài tập | DB CHECK | trạng thái rác | CHECK status IN (...) (§1 #7) |
| FK user/product không tồn tại | lỗi pgx | từ chối ghi | Tạo user/track trước (FR-INFRA-002/FR-PRICE-001) |

---

## §11 - Ghi chú

- `affiliate_click` + `affiliate_conversion` là sổ cái của dòng doanh thu affiliate - dòng tiền chính tài trợ free-tier (§4.1).
- Click chỉ tồn tại khi user chủ động bấm: ở mức schema không có đường tự động nền, đó là cam kết compliant hậu-Honey (NFR-AFFIL-001).
- `sub_id` là token đối soát ngẫu nhiên, không nhúng PII - đi ra network rồi quay về qua postback mà không lộ ai là người dùng.
- Vòng đời `pending -> confirmed/rejected` + một-click-một-conversion cho phép delay payout chống gaming (§5.3) và giữ đối soát đúng khi network retry.
- Tiền BIGINT VND tránh sai số tích lũy khi cộng dồn hoa hồng để đối chiếu báo cáo network.
- Đây là nền cho FR-AFFIL-002 (ghi click), FR-AFFIL-003 (postback), FR-AFFIL-005 (cashback) và FR-TRUST-004/005 (anti-fraud).

---

*Hết FR-AFFIL-001. Status: ready_to_implement (mục tiêu audit 10/10).*
