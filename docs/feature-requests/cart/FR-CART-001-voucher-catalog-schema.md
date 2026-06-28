---
id: FR-CART-001
title: "Schema + ingest voucher_catalog - danh mục voucher 3 loại (shop/platform/freeship) với discount_type, min_spend, cap, stack_group, valid window; BIGINT VND; nguồn cho optimizer giỏ hàng"
module: CART
priority: MUST
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-INFRA-002, FR-CART-002, FR-CART-003, FR-CART-004, FR-CART-005, FR-EXT-002]
depends_on: [FR-INFRA-002]
blocks: [FR-CART-003, FR-CART-005]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model voucher_catalog: code, type, discount_type, discount_value, min_spend, cap, shop_id, valid_from/to, stack_group)"
  - "docs/... §3.5(3) (optimizer cần min_spend/cap/stack_group để chọn combo), §3.7 (POST /v1/cart/optimize dùng vouchers[])"
source_decisions:
  - "DEC-CART-01: voucher_catalog.type thuộc enum {shop, platform, freeship} - ba lớp voucher mà optimizer (FR-CART-003) xếp chồng theo luật stacking"
  - "DEC-CART-02: discount_type thuộc {amount, percent}; discount_value là BIGINT (VND cho amount; phần trăm nguyên cho percent) - không float"
  - "DEC-CART-03: min_spend và cap lưu BIGINT VND (đồng nhất DEC-PRICE-05); cap là trần giảm tuyệt đối (NULL = không trần), min_spend là ngưỡng đơn tối thiểu (NULL/0 = không ngưỡng)"
  - "DEC-CART-04: stack_group là nhãn nhóm xếp chồng - hai voucher cùng stack_group KHÔNG được dùng cùng nhau (loại trừ lẫn nhau); luật stacking per-country (FR-CART-004) đọc nhãn này"
  - "DEC-CART-05: voucher có cửa sổ hiệu lực [valid_from, valid_to]; optimizer chỉ xét voucher còn hiệu lực tại thời điểm tính - ingest giữ cả voucher tương lai/quá hạn, lọc lúc đọc"
  - "DEC-CART-06: shop voucher gắn shop_id (NOT NULL với type=shop); platform/freeship voucher shop_id NULL (áp toàn sàn)"
  - "DEC-CART-35: ingest validate trước khi ghi (type/discount_type trong enum, percent trong [1,100], shop voucher có shop_id, valid_from <= valid_to) - vi phạm trả lỗi rõ, không ghi rác"
  - "DEC-CART-36: ingest idempotent theo (platform_id, code) qua UNIQUE + ON CONFLICT DO UPDATE - nạp lại cùng voucher cập nhật thay vì nhân bản"

language: "PostgreSQL 16; migration golang-migrate (FR-INFRA-002); service Go 1.22 (cart-svc) cho ingest + repo"
service: shopass/services/cart/
new_files:
  - db/migrations/0010_voucher_catalog.up.sql
  - db/migrations/0010_voucher_catalog.down.sql
  - services/cart/internal/voucher/types.go
  - services/cart/internal/voucher/repo.go
  - services/cart/internal/voucher/ingest.go
  - services/cart/internal/voucher/repo_test.go
  - services/cart/internal/voucher/ingest_test.go
modified_files:
  - services/cart/internal/cart/errors.go        # thêm lỗi voucher không hợp lệ
allowed_tools:
  - file_read: services/cart/**
  - file_read: db/migrations/**
  - file_write: services/cart/**
  - file_write: db/migrations/**
  - bash: cd services/cart && go test ./...
disallowed_tools:
  - lưu discount_value/min_spend/cap dạng float/numeric thập phân (vi phạm DEC-CART-02/03, sai số tiền tệ)
  - chấp nhận type ngoài {shop,platform,freeship} hoặc discount_type ngoài {amount,percent} (vi phạm DEC-CART-01/02, optimizer không hiểu)
  - cho shop voucher thiếu shop_id (vi phạm DEC-CART-06, không biết áp cho shop nào)
  - trả voucher quá hạn/chưa hiệu lực cho optimizer (vi phạm DEC-CART-05, tính combo sai)

effort_hours: 6
sub_tasks:
  - "0.75h: 0010_voucher_catalog.up/down.sql - bảng + CHECK type/discount_type + index (platform_id, type, valid_to)"
  - "1.0h: types.go - struct Voucher (BIGINT fields) + enum type/discount_type"
  - "1.0h: repo.go - InsertVoucher, ListActive(platformId, shopIds, now) lọc theo cửa sổ hiệu lực + shop"
  - "1.0h: ingest.go - nạp voucher từ nguồn (extension đọc + feed), validate trước khi ghi (type/discount_type/shop_id/cửa sổ)"
  - "1.0h: repo_test.go - insert 3 loại, ListActive lọc hết hạn/tương lai, shop voucher cần shop_id, BIGINT round-trip"
  - "1.25h: ingest_test.go - reject type lạ, reject percent>100, reject shop voucher thiếu shop_id, idempotent theo (platform_id, code)"

risk_if_skipped: "voucher_catalog là danh mục voucher mà optimizer giỏ hàng (FR-CART-003) và auto-test mã (FR-CART-005) đọc để chọn tổ hợp giảm tốt nhất - không có nó thì hai tính năng lõi của Phase 2 (tối ưu giỏ + thử mã) không có dữ liệu đầu vào. Nếu lưu discount_value/min_spend/cap dạng float thì optimizer tính sai tổng giảm (sai số tích lũy khi cộng nhiều voucher), trả combo không tối ưu hoặc vượt cap - người dùng mất tiền hoặc nhận gợi ý sai. Nếu chấp nhận type/discount_type ngoài enum thì optimizer gặp loại nó không có nhánh xử lý. Nếu trả voucher quá hạn cho optimizer thì nó tính combo dựa voucher không dùng được - gợi ý vô nghĩa lúc thanh toán. stack_group sai làm luật stacking per-country (FR-CART-004) không loại trừ đúng các voucher xung khắc - tính ra tổng giảm cao hơn thực tế, sai lệch trực tiếp con số người dùng thấy."
---

## §1 - Mô tả (BCP-14 normative)

Service CART **MUST** định nghĩa schema và ingest cho `voucher_catalog` - danh mục voucher ba loại (shop/platform/freeship) với các trường tiền tệ BIGINT VND, nhãn nhóm xếp chồng, và cửa sổ hiệu lực, làm nguồn dữ liệu cho optimizer giỏ hàng. Hợp đồng:

1. **MUST** định nghĩa bảng `voucher_catalog (id BIGSERIAL PK, platform_id SMALLINT REFERENCES platform(id), code TEXT, type TEXT, discount_type TEXT, discount_value BIGINT, min_spend BIGINT, cap BIGINT, shop_id TEXT, valid_from TIMESTAMPTZ, valid_to TIMESTAMPTZ, stack_group TEXT)`.
2. **MUST** ràng buộc `type` - `{shop, platform, freeship}` qua CHECK (DEC-CART-01); loại ngoài enum bị DB từ chối.
3. **MUST** ràng buộc `discount_type` - `{amount, percent}` qua CHECK (DEC-CART-02); `discount_value` là `BIGINT` - với `amount` là số tiền VND, với `percent` là phần trăm nguyên `[1, 100]`. KHÔNG dùng float/numeric.
4. **MUST** lưu `min_spend` và `cap` dạng `BIGINT` VND (DEC-CART-03): `min_spend` là ngưỡng giá trị đơn tối thiểu để áp voucher (NULL hoặc 0 = không ngưỡng); `cap` là trần giảm giá tuyệt đối (NULL = không trần). Đồng nhất DEC-PRICE-05.
5. **MUST** xử lý `shop_id` theo loại (DEC-CART-06): với `type = 'shop'` thì `shop_id` **MUST** NOT NULL (voucher gắn một shop); với `type` - `{platform, freeship}` thì `shop_id` **MUST** NULL (áp toàn sàn). Ràng buộc bằng CHECK.
6. **MUST** lưu `stack_group` là nhãn nhóm xếp chồng (DEC-CART-04): hai voucher cùng `stack_group` không được dùng cùng nhau (loại trừ lẫn nhau). Luật stacking per-country (FR-CART-004) đọc nhãn này. `stack_group` NULL nghĩa voucher không thuộc nhóm loại trừ nào.
7. **MUST** lưu cửa sổ hiệu lực `[valid_from, valid_to]`; ingest giữ cả voucher tương lai và quá hạn (DEC-CART-05).
8. **MUST** expose `repo.ListActive(ctx, platformID, shopIDs, now)` trả CHỈ các voucher còn hiệu lực tại `now` (`valid_from <= now <= valid_to`), khớp `platform_id`, và (với shop voucher) khớp một trong `shopIDs` của giỏ - đây là tập đầu vào cho optimizer (FR-CART-003).
9. **MUST** expose `ingest.Upsert(ctx, v Voucher)` validate trước khi ghi: type/discount_type trong enum, `percent` trong `[1,100]`, shop voucher có `shop_id`, `valid_from <= valid_to`; vi phạm trả lỗi rõ, không ghi rác.
10. **MUST** đặt khóa idempotent cho ingest theo `(platform_id, code)` (UNIQUE) để nạp lại cùng voucher không tạo bản trùng (`ON CONFLICT (platform_id, code) DO UPDATE`).
11. **MUST** tạo index hỗ trợ `ListActive`: `idx_vc_active ON voucher_catalog (platform_id, type, valid_to)` (lọc theo sàn + loại + còn hạn).
12. **MUST** đảm bảo `discount_value > 0`, và nếu `discount_type = 'percent'` thì `discount_value <= 100`, qua CHECK.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao ba loại voucher enum (DEC-CART-01)?** Sàn VN có ba lớp giảm giá xếp chồng: voucher shop (giảm tại một shop), voucher platform (giảm toàn sàn), và freeship (giảm phí vận chuyển). Optimizer (FR-CART-003) xếp chồng chúng theo luật stacking. Mỗi loại có ngữ nghĩa và ràng buộc xếp chồng khác nhau - enum CHECK đảm bảo chỉ ba loại đã biết, để optimizer có nhánh xử lý đúng.

**Vì sao tiền tệ BIGINT (DEC-CART-02/03)?** Optimizer cộng nhiều voucher (shop + platform + freeship) để ra tổng giảm. Lưu float gây sai số tích lũy - cộng ba số float rồi so với cap có thể lệch vài đồng, đủ để trả combo sai hoặc vượt cap. BIGINT VND chính xác tuyệt đối. `percent` cũng là số nguyên (sàn không có "10,5%"); optimizer nhân `giá * percent / 100` bằng số nguyên.

**Vì sao stack_group nhãn loại trừ (DEC-CART-04)?** Luật stacking phức tạp và đổi theo nước (VN cho stack, MY/PH 2025 bỏ stack - §3.9). Thay vì mã hóa luật vào từng voucher, ta gắn nhãn `stack_group`: voucher cùng nhóm loại trừ lẫn nhau. Luật per-country (FR-CART-004) diễn giải các nhãn này. Đây là tách dữ liệu (voucher thuộc nhóm nào) khỏi luật (nhóm nào loại trừ nhóm nào theo nước).

**Vì sao shop_id theo loại (DEC-CART-06)?** Voucher shop phải biết áp cho shop nào - thiếu `shop_id` là vô nghĩa (giảm ở đâu?). Voucher platform/freeship áp toàn sàn nên `shop_id` NULL. CHECK ràng buộc đúng quan hệ này để không có voucher shop "mồ côi" không biết áp đâu.

**Vì sao giữ voucher quá hạn nhưng lọc lúc đọc (DEC-CART-05)?** Ingest chạy định kỳ, nạp cả voucher sắp có hiệu lực và vừa hết hạn (để có lịch sử + voucher tương lai). Nhưng optimizer chỉ được tính trên voucher còn hiệu lực TẠI thời điểm tính - dùng voucher quá hạn là gợi ý không dùng được lúc thanh toán. `ListActive` lọc theo `now` tách lưu trữ khỏi tính toán.

**Vì sao idempotent theo (platform_id, code) (DEC-CART-36)?** Ingest nạp lại cùng voucher nhiều lần (refresh danh mục). UNIQUE + ON CONFLICT DO UPDATE làm nạp lại cập nhật thay vì nhân bản, giữ danh mục sạch.

---

## §3 - Hợp đồng API / DDL

### Migration (golang-migrate)

```sql
-- db/migrations/0010_voucher_catalog.up.sql
CREATE TABLE voucher_catalog (
  id             BIGSERIAL   PRIMARY KEY,
  platform_id    SMALLINT    NOT NULL REFERENCES platform(id),
  code           TEXT        NOT NULL,
  type           TEXT        NOT NULL CHECK (type IN ('shop','platform','freeship')),
  discount_type  TEXT        NOT NULL CHECK (discount_type IN ('amount','percent')),
  discount_value BIGINT      NOT NULL CHECK (discount_value > 0
                              AND (discount_type <> 'percent' OR discount_value <= 100)),
  min_spend      BIGINT      CHECK (min_spend IS NULL OR min_spend >= 0),  -- VND
  cap            BIGINT      CHECK (cap IS NULL OR cap > 0),               -- VND, trần giảm
  shop_id        TEXT,
  valid_from     TIMESTAMPTZ NOT NULL,
  valid_to       TIMESTAMPTZ NOT NULL CHECK (valid_to >= valid_from),
  stack_group    TEXT,
  -- shop voucher cần shop_id; platform/freeship thì shop_id NULL (DEC-CART-06)
  CONSTRAINT shop_id_by_type CHECK (
    (type = 'shop' AND shop_id IS NOT NULL) OR
    (type <> 'shop' AND shop_id IS NULL)
  ),
  UNIQUE (platform_id, code)   -- idempotent ingest (DEC-CART-36)
);

CREATE INDEX idx_vc_active ON voucher_catalog (platform_id, type, valid_to);

-- db/migrations/0010_voucher_catalog.down.sql
DROP TABLE voucher_catalog;
```

### Types (Go)

```go
// services/cart/internal/voucher/types.go
type VoucherType string
const ( TypeShop VoucherType = "shop"; TypePlatform = "platform"; TypeFreeship = "freeship" )

type DiscountType string
const ( DiscountAmount DiscountType = "amount"; DiscountPercent = "percent" )

type Voucher struct {
    ID            int64        `db:"id"`
    PlatformID    int16        `db:"platform_id"`
    Code          string       `db:"code"`
    Type          VoucherType  `db:"type"`
    DiscountType  DiscountType `db:"discount_type"`
    DiscountValue int64        `db:"discount_value"` // VND (amount) hoặc % nguyên (percent)
    MinSpend      *int64       `db:"min_spend"`      // VND
    Cap           *int64       `db:"cap"`            // VND, trần giảm
    ShopID        *string      `db:"shop_id"`
    ValidFrom     time.Time    `db:"valid_from"`
    ValidTo       time.Time    `db:"valid_to"`
    StackGroup    *string      `db:"stack_group"`
}
```

### Ingest validate (ingest.go)

```go
// services/cart/internal/voucher/ingest.go
func (i *Ingestor) Upsert(ctx context.Context, v Voucher) error {
    if err := validate(v); err != nil {
        return err // không ghi rác (DEC-CART-35)
    }
    _, err := i.pool.Exec(ctx,
        `INSERT INTO voucher_catalog
           (platform_id, code, type, discount_type, discount_value, min_spend, cap, shop_id, valid_from, valid_to, stack_group)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
         ON CONFLICT (platform_id, code) DO UPDATE SET
           discount_type=EXCLUDED.discount_type, discount_value=EXCLUDED.discount_value,
           min_spend=EXCLUDED.min_spend, cap=EXCLUDED.cap, valid_from=EXCLUDED.valid_from,
           valid_to=EXCLUDED.valid_to, stack_group=EXCLUDED.stack_group`,
        v.PlatformID, v.Code, v.Type, v.DiscountType, v.DiscountValue,
        v.MinSpend, v.Cap, v.ShopID, v.ValidFrom, v.ValidTo, v.StackGroup)
    return err
}

func validate(v Voucher) error {
    switch v.Type {
    case TypeShop:
        if v.ShopID == nil { return ErrShopVoucherNeedsShopID } // DEC-CART-06
    case TypePlatform, TypeFreeship:
        if v.ShopID != nil { return ErrPlatformVoucherHasShopID }
    default:
        return ErrUnknownVoucherType
    }
    if v.DiscountType == DiscountPercent && (v.DiscountValue < 1 || v.DiscountValue > 100) {
        return ErrPercentOutOfRange
    }
    if v.DiscountValue <= 0 { return ErrNonPositiveDiscount }
    if v.ValidTo.Before(v.ValidFrom) { return ErrInvalidWindow }
    return nil
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch (up/down); `voucher_catalog` tồn tại với CHECK `type`, `discount_type`, và ràng buộc `shop_id_by_type`.
2. INSERT `type` ngoài `{shop,platform,freeship}` -> lỗi CHECK; `discount_type` ngoài `{amount,percent}` -> lỗi CHECK.
3. INSERT shop voucher thiếu `shop_id` -> lỗi CHECK `shop_id_by_type`; platform/freeship voucher kèm `shop_id` -> lỗi CHECK.
4. INSERT `discount_type='percent'` với `discount_value=120` -> lỗi CHECK (>100); `discount_value=0` -> lỗi CHECK.
5. `discount_value`, `min_spend`, `cap` round-trip int64 đúng (không float).
6. `ListActive(platformID, shopIDs, now)` trả voucher còn hiệu lực tại `now`, đúng `platform_id`; bỏ voucher hết hạn (`valid_to < now`) và voucher tương lai (`valid_from > now`).
7. `ListActive` với shop voucher chỉ trả voucher có `shop_id` thuộc `shopIDs` của giỏ; platform/freeship voucher luôn được xét (shop_id NULL).
8. `Upsert` reject type lạ, percent ngoài `[1,100]`, shop voucher thiếu shop_id, `valid_to < valid_from` - trả lỗi rõ, không ghi.
9. `Upsert` cùng `(platform_id, code)` hai lần -> DO UPDATE (một dòng, giá trị mới), không nhân bản.
10. `stack_group` lưu và đọc đúng (NULL được phép); là nhãn cho FR-CART-004 đọc.
11. `go test ./...` xanh.

---

## §5 - Kiểm thử (verification)

```go
// services/cart/internal/voucher/repo_test.go
func TestListActive_FiltersByWindow(t *testing.T) {
    r := setupVoucher(t)
    now := time.Now()
    insert(t, r, Voucher{PlatformID: 1, Code: "EXPIRED", Type: TypePlatform, DiscountType: DiscountAmount,
        DiscountValue: 50_000, ValidFrom: now.AddDate(0, 0, -10), ValidTo: now.AddDate(0, 0, -1)}) // hết hạn
    insert(t, r, Voucher{PlatformID: 1, Code: "ACTIVE", Type: TypePlatform, DiscountType: DiscountAmount,
        DiscountValue: 50_000, ValidFrom: now.AddDate(0, 0, -1), ValidTo: now.AddDate(0, 0, 5)})   // còn hạn
    out, _ := r.ListActive(ctx, 1, nil, now)
    codes := codesOf(out)
    require.Contains(t, codes, "ACTIVE")
    require.NotContains(t, codes, "EXPIRED")
}

func TestListActive_ShopVoucherScopedToCartShops(t *testing.T) {
    r := setupVoucher(t)
    now := time.Now()
    insert(t, r, shopVoucher(1, "SHOPA", "shopA", now))
    insert(t, r, shopVoucher(1, "SHOPB", "shopB", now))
    out, _ := r.ListActive(ctx, 1, []string{"shopA"}, now) // giỏ chỉ có shopA
    require.Contains(t, codesOf(out), "SHOPA")
    require.NotContains(t, codesOf(out), "SHOPB")
}

func TestBigintRoundTrip(t *testing.T) {
    r := setupVoucher(t)
    insert(t, r, Voucher{PlatformID: 1, Code: "CAP", Type: TypePlatform, DiscountType: DiscountAmount,
        DiscountValue: 70_000, Cap: ptr(int64(70_000)), MinSpend: ptr(int64(500_000)),
        ValidFrom: t0, ValidTo: t0.AddDate(0, 1, 0)})
    out, _ := r.ListActive(ctx, 1, nil, t0.AddDate(0, 0, 1))
    require.Equal(t, int64(70_000), *out[0].Cap)
    require.Equal(t, int64(500_000), *out[0].MinSpend)
}
```

```go
// services/cart/internal/voucher/ingest_test.go
func TestUpsert_RejectShopVoucherWithoutShopID(t *testing.T) {
    i := setupIngest(t)
    err := i.Upsert(ctx, Voucher{PlatformID: 1, Code: "X", Type: TypeShop, DiscountType: DiscountAmount,
        DiscountValue: 30_000, ValidFrom: t0, ValidTo: t0.AddDate(0, 1, 0)}) // thiếu ShopID
    require.ErrorIs(t, err, ErrShopVoucherNeedsShopID)
}

func TestUpsert_RejectPercentOver100(t *testing.T) {
    i := setupIngest(t)
    err := i.Upsert(ctx, Voucher{PlatformID: 1, Code: "P", Type: TypePlatform, DiscountType: DiscountPercent,
        DiscountValue: 120, ValidFrom: t0, ValidTo: t0.AddDate(0, 1, 0)})
    require.ErrorIs(t, err, ErrPercentOutOfRange)
}

func TestUpsert_Idempotent(t *testing.T) {
    i := setupIngest(t)
    v := Voucher{PlatformID: 1, Code: "DUP", Type: TypePlatform, DiscountType: DiscountAmount,
        DiscountValue: 50_000, ValidFrom: t0, ValidTo: t0.AddDate(0, 1, 0)}
    require.NoError(t, i.Upsert(ctx, v))
    v.DiscountValue = 60_000
    require.NoError(t, i.Upsert(ctx, v)) // cùng (platform_id, code) → UPDATE
    require.Equal(t, 1, countVouchers(t, i, 1, "DUP"))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `0010_voucher_catalog.up/down.sql` (bảng + CHECK + UNIQUE + index) -> `types.go` (struct Voucher BIGINT + enum) -> `repo.go` (`InsertVoucher`, `ListActive` lọc cửa sổ + shop) -> `ingest.go` (`Upsert` validate + ON CONFLICT) -> tests. Migration đánh số 0010 nối tiếp dãy nền tảng của FR-INFRA-002 (0001-0003) và các module khác; cặp up/down bất biến sau merge (DEC-INFRA-06). Nguồn voucher đến từ extension đọc (FR-EXT-002 voucher-reader) và feed danh mục; `Upsert` là cổng validate trước khi vào catalog. `ListActive` là API duy nhất optimizer (FR-CART-003) dùng để lấy tập voucher ứng viên.

---

## §7 - Phụ thuộc

- **FR-INFRA-002** - migration framework (golang-migrate) + bảng `platform` (FK `platform_id`) phải có trước (depends_on cứng).
- **FR-CART-003 (downstream)** - optimizer đọc `ListActive` làm tập voucher ứng viên để chọn combo tối ưu.
- **FR-CART-004 (downstream)** - luật stacking per-country đọc `stack_group` để quyết voucher nào loại trừ nhau.
- **FR-CART-005 (downstream)** - auto-test mã thử các voucher trong catalog (user-initiated).
- **FR-EXT-002 (nguồn dữ liệu)** - voucher-reader của content script đọc voucher hiển thị; cùng feed danh mục nuôi ingest.
- Extension/lib: driver `pgx`; golang-migrate.

---

## §8 - Payload ví dụ

### Ingest một platform voucher (giảm 50k cho đơn >= 500k, trần 50k)

```go
err := ingestor.Upsert(ctx, voucher.Voucher{
    PlatformID:    1,            // shopee
    Code:          "GIAM50K",
    Type:          voucher.TypePlatform,
    DiscountType:  voucher.DiscountAmount,
    DiscountValue: 50_000,       // VND
    MinSpend:      ptr(int64(500_000)),
    Cap:           ptr(int64(50_000)),
    ValidFrom:     time.Now(),
    ValidTo:       time.Now().AddDate(0, 0, 7),
    StackGroup:    ptr("platform-main"), // FR-CART-004 đọc nhãn này
})
```

### Voucher còn hiệu lực cho optimizer (đọc qua ListActive)

```sql
SELECT code, type, discount_type, discount_value, min_spend, cap, shop_id, stack_group
FROM voucher_catalog
WHERE platform_id = 1
  AND valid_from <= now() AND valid_to >= now()
  AND (type <> 'shop' OR shop_id = ANY($1));  -- $1 = shop_id của giỏ
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Voucher điều kiện bậc thang (giảm tăng theo mức chi) - mô hình hóa khi gặp dữ liệu thật; slice này một ngưỡng `min_spend`.
- Voucher giới hạn lượt dùng / per-user - thêm trường khi cần chống lạm dụng.
- Đa tiền tệ khi mở SEA (THB/IDR) - giữ BIGINT theo minor unit từng nước, gắn currency theo platform.country.
- Phân loại con của freeship (freeship xtra vs thường) - thêm khi luật stacking cần phân biệt.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| discount_value/min_spend/cap float | review + test | optimizer tính sai tổng | BIGINT VND (DEC-CART-02/03) |
| type/discount_type lạ | DB CHECK + ingest | optimizer không có nhánh | enum CHECK (DEC-CART-01/02) |
| shop voucher thiếu shop_id | CHECK shop_id_by_type | không biết áp shop nào | NOT NULL với type=shop (DEC-CART-06) |
| percent > 100 | CHECK + ingest validate | giảm vô lý | percent <= 100 (§1 #3) |
| voucher quá hạn lọt vào optimizer | ListActive lọc now | gợi ý không dùng được | Lọc [valid_from, valid_to] (DEC-CART-05) |
| stack_group sai/thiếu | review FR-CART-004 | loại trừ sai, tổng giảm lệch | Nhãn nhóm chuẩn (DEC-CART-04) |
| Nạp trùng voucher | ingest_test idempotent | catalog rác | ON CONFLICT (platform_id, code) DO UPDATE |
| shop voucher của shop khác lọt vào giỏ | ListActive scope shopIDs | combo sai shop | Lọc shop_id ANY(shopIDs) (§1 #8) |
| valid_to < valid_from | CHECK + validate | cửa sổ vô nghĩa | CHECK valid_to >= valid_from |

---

## §11 - Ghi chú

- `voucher_catalog` là nguồn voucher cho optimizer giỏ (FR-CART-003) và auto-test mã (FR-CART-005) - hai tính năng lõi Phase 2.
- Tiền tệ BIGINT VND tránh sai số khi optimizer cộng nhiều voucher rồi so với cap; percent là số nguyên.
- stack_group tách dữ liệu (voucher thuộc nhóm nào) khỏi luật (nhóm nào loại trừ theo nước) - luật per-country ở FR-CART-004.
- shop_id ràng buộc theo loại để không có voucher shop mồ côi không biết áp đâu.
- Giữ voucher quá hạn/tương lai trong catalog nhưng `ListActive` chỉ trả voucher còn hiệu lực tại thời điểm tính - tách lưu trữ khỏi tính toán.
- Ingest idempotent theo (platform_id, code) giữ danh mục sạch khi refresh định kỳ.

---

*Hết FR-CART-001. Status: ready_to_implement (mục tiêu audit 10/10).*
