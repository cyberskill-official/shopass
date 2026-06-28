---
id: FR-PRICE-001
title: "tracked_product registry chuẩn hóa - khóa duy nhất (platform_id, platform_item_id) + cột canonical_key + index so sánh chéo sàn cho FK của price_snapshot"
module: PRICE
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-PRICE-002, FR-PRICE-005, FR-INFRA-002, FR-SCRAPE-002, FR-TRACK-001]
depends_on: [FR-INFRA-002]
blocks: [FR-PRICE-002, FR-PRICE-005, FR-SCRAPE-001, FR-TRACK-001]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model tracked_product)"
  - "docs/... §5.6 (moat so sánh chéo sàn), §3.7 (chuẩn hóa định danh sản phẩm)"
source_decisions:
  - "DEC-PRICE-06: tracked_product là registry chuẩn hóa, đúng 1 dòng cho mỗi (platform_id, platform_item_id); không nhân bản theo lần quét"
  - "DEC-PRICE-07: canonical_key tách khỏi thuật toán matching - ở FR này chỉ là cột TEXT nullable + index; giá trị do FR-PRICE-005 điền sau"
  - "DEC-PRICE-08: UNIQUE(platform_id, platform_item_id) chống trùng SKU trong cùng một sàn; là khóa idempotent cho upsert"
  - "DEC-PRICE-09: index idx_tp_canonical phục vụ JOIN gom các SKU cùng sản phẩm trên nhiều sàn (moat so sánh chéo §5.6)"
  - "DEC-PRICE-10: first_seen TIMESTAMPTZ DEFAULT now() là mốc cold-start - biết SKU vào hệ thống từ khi nào để cắt cửa sổ lịch sử giá"

language: "PostgreSQL 16; service Go 1.22 (price-svc)"
service: shopass/services/price/
new_files:
  - services/price/migrations/0001_tracked_product.sql
  - services/price/internal/price/product.go
  - services/price/internal/price/product_repo.go
  - services/price/internal/price/product_repo_test.go
modified_files: []
allowed_tools:
  - file_read: services/price/**
  - file_write: services/price/**
  - bash: cd services/price && go test ./...
disallowed_tools:
  - điền canonical_key trong FR này (thuộc FR-PRICE-005; insert phải để canonical_key NULL)
  - tạo nhiều dòng tracked_product cho cùng (platform_id, platform_item_id) (vi phạm DEC-PRICE-06/08)
  - dùng SERIAL/INT cho id (phải BIGSERIAL/BIGINT theo quy ước FR-INFRA-002)

effort_hours: 6
sub_tasks:
  - "0.5h: 0001_tracked_product.sql - CREATE TABLE + FK platform + UNIQUE(platform_id, platform_item_id)"
  - "0.5h: 0001_tracked_product.sql - CREATE INDEX idx_tp_canonical ON tracked_product(canonical_key)"
  - "1.0h: product.go - struct TrackedProduct với db tags + kiểu pointer cho cột nullable"
  - "1.5h: product_repo.go - Upsert (ON CONFLICT), GetByID, GetByCanonicalKey, FindByPlatformItem"
  - "2.0h: product_repo_test.go - 5 test (new, conflict idempotent, unique, canonical lookup, canonical NULL on insert)"
  - "0.5h: OTel metric tracked_product_upsert_total{platform_id, action=insert|update}"

risk_if_skipped: "tracked_product là registry chuẩn hóa mà toàn bộ module PRICE neo vào. Không có nó thì price_snapshot (FR-PRICE-002) không có đích FK product_id, thuật toán so khớp chéo sàn (FR-PRICE-005) không có cột canonical_key để điền, và API theo dõi giá (FR-TRACK-001) không có thực thể sản phẩm để gắn người dùng. Thiếu UNIQUE(platform_id, platform_item_id) thì mỗi lần quét lại tạo một dòng mới, registry phình và price_snapshot trỏ tới nhiều id rác cho cùng một SKU. Thiếu index idx_tp_canonical thì JOIN gom SKU cùng sản phẩm trên nhiều sàn phải quét toàn bảng, vỡ moat so sánh chéo sàn (§5.6)."
---

## §1 - Mô tả (BCP-14 normative)

Service PRICE **MUST** định nghĩa registry chuẩn hóa `tracked_product`: đúng một dòng cho mỗi cặp (platform_id, platform_item_id), có cột `canonical_key` làm khóa gom chéo sàn (điền sau bởi FR-PRICE-005), và là đích FK cho `price_snapshot`. Hợp đồng:

1. **MUST** định nghĩa bảng `tracked_product (id BIGSERIAL PK, platform_id SMALLINT, platform_item_id TEXT NOT NULL, shop_id TEXT, title TEXT, category_id BIGINT, canonical_key TEXT, first_seen TIMESTAMPTZ DEFAULT now())`.
2. **MUST** ràng buộc `platform_id` REFERENCES `platform(id)` (FK tới bảng nền của FR-INFRA-002). Một SKU luôn thuộc đúng một sàn đã đăng ký.
3. **MUST** đặt `UNIQUE (platform_id, platform_item_id)` (DEC-PRICE-08): cùng một sàn không được có hai dòng cho cùng `platform_item_id`. Đây là khóa idempotent cho upsert.
4. **MUST** để `canonical_key` là `TEXT` nullable (DEC-PRICE-07): FR này chỉ tạo cột; FR-PRICE-005 mới chạy thuật toán so khớp và UPDATE giá trị. Insert ở FR này **MUST** để `canonical_key = NULL`.
5. **MUST** tạo `CREATE INDEX idx_tp_canonical ON tracked_product (canonical_key)` (DEC-PRICE-09) phục vụ JOIN gom các SKU cùng sản phẩm trên nhiều sàn. Dòng có `canonical_key IS NULL` không cản index (B-tree bỏ qua NULL khi tra cứu theo giá trị).
6. **MUST** đặt `first_seen TIMESTAMPTZ DEFAULT now()` (DEC-PRICE-10): mốc cold-start ghi nhận thời điểm SKU vào hệ thống, dùng để cắt cửa sổ lịch sử giá cho sale ảo.
7. **MUST** expose hàm repo `Upsert(ctx, p TrackedProduct) (TrackedProduct, error)`: chèn dòng mới hoặc cập nhật metadata (`shop_id`, `title`, `category_id`) khi cặp (platform_id, platform_item_id) đã tồn tại; trả về dòng đã ghi kèm `id`.
8. **MUST** thực hiện upsert idempotent bằng `INSERT ... ON CONFLICT (platform_id, platform_item_id) DO UPDATE`: gọi nhiều lần với cùng cặp khóa **MUST** giữ nguyên một dòng (cùng `id`, cùng `first_seen`).
9. **MUST** giữ `first_seen` bất biến qua các lần upsert: nhánh `DO UPDATE` **MUST NOT** ghi đè `first_seen` (mốc cold-start phải là lần thấy đầu tiên, không phải lần gần nhất).
10. **MUST** expose `GetByID(ctx, id int64) (TrackedProduct, error)` và `FindByPlatformItem(ctx, platformID int16, platformItemID string) (TrackedProduct, error)` để tra cứu theo khóa kỹ thuật và theo khóa tự nhiên.
11. **MUST** expose `GetByCanonicalKey(ctx, key string) ([]TrackedProduct, error)`: trả về mọi SKU chia sẻ cùng `canonical_key` (các biến thể của một sản phẩm trên nhiều sàn). Tra cứu bằng `key = NULL` không hợp lệ, hàm **MUST** trả lỗi tham số.
12. **SHOULD** phát OTel metric `tracked_product_upsert_total{platform_id, action}` với `action - {insert, update}` để theo dõi nhịp tăng trưởng registry và tỷ lệ SKU mới.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao tracked_product là registry chuẩn hóa một dòng (DEC-PRICE-06)?** Cùng một SKU được scraper quét lại nhiều lần mỗi ngày. Nếu mỗi lần quét tạo một dòng, registry phình vô tận và `price_snapshot` trỏ tới hàng loạt `id` rác cho cùng một sản phẩm, làm biểu đồ giá vỡ. Một dòng cho mỗi (platform_id, platform_item_id) cho ta một danh tính ổn định để mọi snapshot giá neo vào.

**Vì sao tách canonical_key khỏi thuật toán matching (DEC-PRICE-07)?** So khớp chéo sàn (cùng một chiếc tai nghe bán trên Shopee, TikTok, Lazada) là bài toán khó: tên khác nhau, ảnh khác nhau, không có mã chung. Nhồi thuật toán đó vào FR schema này làm phình phạm vi và chặn các FR phụ thuộc. FR này chỉ dựng cột `canonical_key` (nullable) và index; FR-PRICE-005 chạy so khớp và điền giá trị sau. Schema sẵn sàng trước, thuật toán đến sau mà không phải đổi bảng.

**Vì sao UNIQUE(platform_id, platform_item_id) (DEC-PRICE-08)?** `platform_item_id` chỉ duy nhất trong phạm vi một sàn, không phải toàn cục: Shopee và Lazada có thể trùng chuỗi item_id cho hai sản phẩm khác nhau. Ràng buộc kép (sàn, item) là khóa tự nhiên đúng, đồng thời là khóa `ON CONFLICT` cho upsert idempotent với retry của scraper.

**Vì sao index idx_tp_canonical (DEC-PRICE-09)?** Moat sản phẩm của SănDeal là so sánh giá cùng một món trên nhiều sàn (§5.6). Truy vấn lõi là "cho canonical_key X, lấy mọi SKU của mọi sàn". Không index thì JOIN này quét toàn bảng tracked_product (hàng triệu dòng). Index trên `canonical_key` biến nó thành tra cứu range nhanh.

**Vì sao first_seen DEFAULT now() và bất biến (DEC-PRICE-10, §1 #9)?** Sale ảo cần biết lịch sử giá đủ dài. `first_seen` đánh dấu khi SKU vào hệ thống để biết cửa sổ dữ liệu có bao nhiêu. Nếu nhánh upsert ghi đè `first_seen` mỗi lần quét, mốc cold-start luôn nhảy về hiện tại và mọi SKU trông như "mới", phá logic cắt cửa sổ. Vì vậy `DO UPDATE` chỉ chạm metadata, không chạm `first_seen`.

**Vì sao canonical_key để NULL khi insert (§1 #4)?** Lúc scraper phát hiện SKU, ta chưa biết nó là biến thể của sản phẩm nào. NULL nghĩa là "chưa so khớp". FR-PRICE-005 quét các dòng NULL và gán nhóm. Điền sẵn một giá trị đoán mò ở FR này sẽ tạo nhóm sai mà sau khó dọn.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/price/migrations/0001_tracked_product.sql
CREATE TABLE tracked_product (
  id               BIGSERIAL   PRIMARY KEY,
  platform_id      SMALLINT    NOT NULL REFERENCES platform(id),
  platform_item_id TEXT        NOT NULL,
  shop_id          TEXT,
  title            TEXT,
  category_id      BIGINT,
  canonical_key    TEXT,                                  -- NULL cho tới khi FR-PRICE-005 so khớp
  first_seen       TIMESTAMPTZ NOT NULL DEFAULT now(),    -- mốc cold-start, bất biến
  UNIQUE (platform_id, platform_item_id)
);

-- Phục vụ JOIN gom SKU cùng sản phẩm trên nhiều sàn (moat so sánh chéo §5.6)
CREATE INDEX idx_tp_canonical ON tracked_product (canonical_key);
```

### Types (Go)

```go
// services/price/internal/price/product.go
type TrackedProduct struct {
    ID             int64     `db:"id"`
    PlatformID     int16     `db:"platform_id"`
    PlatformItemID string    `db:"platform_item_id"`
    ShopID         *string   `db:"shop_id"`
    Title          *string   `db:"title"`
    CategoryID     *int64    `db:"category_id"`
    CanonicalKey   *string   `db:"canonical_key"` // nil khi chưa so khớp (FR-PRICE-005 điền)
    FirstSeen      time.Time `db:"first_seen"`
}
```

### Upsert idempotent (§1 #7, #8, #9)

```go
// services/price/internal/price/product_repo.go
// Upsert chèn SKU mới hoặc cập nhật metadata khi (platform_id, platform_item_id) đã có.
// canonical_key KHÔNG được ghi ở đây - để NULL, FR-PRICE-005 điền sau.
// first_seen KHÔNG bị nhánh DO UPDATE ghi đè (giữ mốc cold-start).
func (r *Repo) Upsert(ctx context.Context, p TrackedProduct) (TrackedProduct, error) {
    var out TrackedProduct
    err := r.pool.QueryRow(ctx,
        `INSERT INTO tracked_product
            (platform_id, platform_item_id, shop_id, title, category_id)
         VALUES ($1, $2, $3, $4, $5)
         ON CONFLICT (platform_id, platform_item_id) DO UPDATE
            SET shop_id     = EXCLUDED.shop_id,
                title       = EXCLUDED.title,
                category_id = EXCLUDED.category_id
         RETURNING id, platform_id, platform_item_id, shop_id, title,
                   category_id, canonical_key, first_seen`,
        p.PlatformID, p.PlatformItemID, p.ShopID, p.Title, p.CategoryID).
        Scan(&out.ID, &out.PlatformID, &out.PlatformItemID, &out.ShopID,
            &out.Title, &out.CategoryID, &out.CanonicalKey, &out.FirstSeen)
    if err != nil {
        return TrackedProduct{}, err
    }
    metrics.ProductUpsert(out.PlatformID)
    return out, nil
}

func (r *Repo) GetByCanonicalKey(ctx context.Context, key string) ([]TrackedProduct, error) {
    if key == "" {
        return nil, fmt.Errorf("canonical_key rỗng không hợp lệ")
    }
    rows, err := r.pool.Query(ctx,
        `SELECT id, platform_id, platform_item_id, shop_id, title,
                category_id, canonical_key, first_seen
         FROM tracked_product WHERE canonical_key = $1`, key)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    return scanProducts(rows)
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> bảng `tracked_product` tồn tại với đủ cột, PK `id`, và `first_seen` DEFAULT now().
2. FK `platform_id` REFERENCES `platform(id)` tồn tại; INSERT với `platform_id` không có trong `platform` -> lỗi FK.
3. INSERT hai dòng cùng `(platform_id, platform_item_id)` qua đường thẳng (không upsert) -> lỗi UNIQUE.
4. `canonical_key` chấp nhận NULL; sau insert qua `Upsert`, giá trị đọc lại là NULL (chưa so khớp).
5. Index `idx_tp_canonical` tồn tại (`SELECT ... FROM pg_indexes WHERE indexname='idx_tp_canonical'` trả 1 dòng).
6. `Upsert` SKU mới -> trả về dòng có `id > 0` và `first_seen` được DB gán.
7. `Upsert` lần hai cùng `(platform_id, platform_item_id)` với metadata khác -> trả về CÙNG `id`, `title`/`shop_id`/`category_id` đã cập nhật, tổng số dòng vẫn là 1.
8. `Upsert` lần hai -> `first_seen` KHÔNG đổi so với lần đầu (mốc cold-start bất biến).
9. `GetByID` trả đúng dòng đã upsert; `GetByID` với id không tồn tại -> lỗi no-rows.
10. `FindByPlatformItem(platformID, itemID)` trả đúng dòng theo khóa tự nhiên.
11. `GetByCanonicalKey(key)` trả mọi SKU chia sẻ `key`; gọi với chuỗi rỗng -> lỗi tham số, không truy vấn DB.
12. Metric `tracked_product_upsert_total` tăng mỗi lần `Upsert` thành công.

---

## §5 - Kiểm thử (verification)

```go
// services/price/internal/price/product_repo_test.go
func TestUpsert_New(t *testing.T) {
    r, plat := setupRepoWithPlatform(t) // seed platform(id=1,'shopee','VN')
    p := TrackedProduct{PlatformID: plat, PlatformItemID: "i-123", Title: ptr("Tai nghe X")}
    out, err := r.Upsert(ctx, p)
    require.NoError(t, err)
    require.Greater(t, out.ID, int64(0))
    require.False(t, out.FirstSeen.IsZero())
    require.Equal(t, 1, countProducts(t, r))
}

func TestUpsert_Conflict_Idempotent(t *testing.T) {
    r, plat := setupRepoWithPlatform(t)
    a, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-123", Title: ptr("Tên cũ")})
    b, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-123", Title: ptr("Tên mới")})
    require.Equal(t, a.ID, b.ID)              // cùng một dòng
    require.Equal(t, "Tên mới", *b.Title)     // metadata đã cập nhật
    require.Equal(t, a.FirstSeen, b.FirstSeen) // first_seen bất biến
    require.Equal(t, 1, countProducts(t, r))   // không nhân bản
}

func TestUnique_PlatformItem(t *testing.T) {
    r, plat := setupRepoWithPlatform(t)
    _, err := r.pool.Exec(ctx,
        `INSERT INTO tracked_product (platform_id, platform_item_id) VALUES ($1,$2),($1,$2)`,
        plat, "dup")
    require.Error(t, err) // vi phạm UNIQUE(platform_id, platform_item_id)
}

func TestGetByCanonicalKey(t *testing.T) {
    r, plat := setupRepoWithPlatform(t)
    a, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-1"})
    c, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-2"})
    // mô phỏng FR-PRICE-005 gán cùng canonical_key
    r.pool.Exec(ctx, `UPDATE tracked_product SET canonical_key='k-xyz' WHERE id = ANY($1)`,
        []int64{a.ID, c.ID})
    rows, err := r.GetByCanonicalKey(ctx, "k-xyz")
    require.NoError(t, err)
    require.Len(t, rows, 2)

    _, err = r.GetByCanonicalKey(ctx, "")
    require.Error(t, err) // chuỗi rỗng → lỗi tham số
}

func TestCanonicalKey_NullOnInsert(t *testing.T) {
    r, plat := setupRepoWithPlatform(t)
    out, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-9"})
    require.Nil(t, out.CanonicalKey) // FR này KHÔNG điền canonical_key
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0001_tracked_product.sql` (bảng + UNIQUE + index) -> `product.go` (struct) -> `product_repo.go` (Upsert + 3 hàm tra cứu) -> tests. Migration chạy qua golang-migrate (cùng runner của FR-INFRA-002). Index `idx_tp_canonical` tạo cùng migration, trên bảng còn rỗng nên không khóa lâu. Driver dùng `pgx`. Hàm `Upsert` dùng `RETURNING` để lấy lại `id` và `first_seen` do DB sinh, tránh một round-trip đọc thêm.

---

## §7 - Phụ thuộc

- **FR-INFRA-002** - bảng `platform` và framework migration phải có trước (FK `platform_id`, runner golang-migrate).
- **FR-PRICE-002 (downstream)** - `price_snapshot.product_id` REFERENCES `tracked_product(id)`; cần bảng này tồn tại trước.
- **FR-PRICE-005 (downstream)** - thuật toán so khớp chéo sàn UPDATE `canonical_key` cho các dòng NULL; đọc qua index `idx_tp_canonical`.
- **FR-SCRAPE-002 (upstream)** - scraper gọi `Upsert` khi phát hiện hoặc làm mới một SKU.
- **FR-TRACK-001 (downstream)** - API theo dõi giá gắn người dùng vào `tracked_product.id`.
- Lib: driver `pgx`.

---

## §8 - Payload ví dụ

### Scraper đăng ký một SKU (nội bộ)

```go
tp, err := productRepo.Upsert(ctx, price.TrackedProduct{
    PlatformID:     1,                       // shopee (theo platform seed FR-INFRA-002)
    PlatformItemID: "20114455667",
    ShopID:         ptr("88123"),
    Title:          ptr("Tai nghe Bluetooth ABC Pro"),
    CategoryID:     ptr(int64(7021)),
    // CanonicalKey để trống - FR-PRICE-005 điền sau
})
// tp.ID dùng làm product_id khi ghi price_snapshot (FR-PRICE-002)
```

### So sánh chéo sàn - gom mọi SKU cùng một sản phẩm

```sql
-- Dùng index idx_tp_canonical: với một canonical_key, lấy giá mới nhất trên từng sàn
SELECT tp.platform_id, tp.platform_item_id, tp.title, tp.id
FROM tracked_product tp
WHERE tp.canonical_key = 'ck-earbud-abc-pro'
ORDER BY tp.platform_id;
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Định dạng nội bộ của `canonical_key` (hash chuẩn hóa hay id cụm) - thuộc FR-PRICE-005, không ảnh hưởng schema ở đây.
- Lưu lịch sử đổi `title`/`category_id` (audit metadata) - chỉ thêm nếu cần truy vết, slice sau.
- Đánh dấu SKU "đã ngừng bán" (cột `delisted_at`) - bổ sung khi scraper báo 404; không cản FR này.
- Soft-delete vs hard-delete dòng tracked_product - giữ đơn giản tới khi có yêu cầu xóa dữ liệu (FR-COMPLY).

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| platform_id không có trong platform | lỗi FK pgx | từ chối ghi | Seed platform (FR-INFRA-002) chạy trước; scraper map đúng id |
| Trùng (platform_id, platform_item_id) qua insert thẳng | DB UNIQUE | lỗi ghi | Luôn dùng Upsert (ON CONFLICT), không INSERT trần |
| Nhánh DO UPDATE ghi đè first_seen | test idempotent | mốc cold-start nhảy về hiện tại | DO UPDATE chỉ chạm shop_id/title/category_id |
| Insert kèm canonical_key (vi phạm ranh giới) | code review + test NullOnInsert | nhóm chéo sàn sai | Upsert không nhận canonical_key; chỉ FR-PRICE-005 UPDATE |
| GetByCanonicalKey gọi với chuỗi rỗng | guard tham số | quét nhầm dòng NULL | Trả lỗi tham số trước khi truy vấn |
| platform_item_id NULL hoặc rỗng | NOT NULL (NULL) / validate (rỗng) | dòng vô danh | NOT NULL ở DB; scraper validate rỗng trước khi gửi |
| Index idx_tp_canonical thiếu | EXPLAIN seq scan | JOIN chéo sàn chậm | Migration tạo index; AC #5 kiểm tồn tại |
| Registry phình do quét tạo dòng mới | đếm dòng / metric upsert | price_snapshot trỏ id rác | UNIQUE + Upsert đảm bảo một dòng mỗi SKU |
| Race hai scraper upsert cùng SKU | ON CONFLICT | một insert, một update | Theo thiết kế (idempotent); không vỡ |

---

## §11 - Ghi chú

- `tracked_product` là điểm neo của module PRICE: mọi snapshot giá, mọi nhóm chéo sàn, mọi theo dõi của người dùng đều trỏ về `id` của bảng này.
- Tách `canonical_key` (cột + index ở FR này) khỏi thuật toán so khớp (FR-PRICE-005) cho phép schema sẵn sàng trước, không chặn các FR phụ thuộc chờ thuật toán khó.
- `UNIQUE(platform_id, platform_item_id)` vừa là khóa tự nhiên đúng (item_id chỉ duy nhất trong một sàn), vừa là khóa idempotent cho upsert với retry của scraper.
- `first_seen` bất biến là chi tiết dễ sai nhất: nhánh `DO UPDATE` phải bỏ qua nó, nếu không mốc cold-start hỏng và logic cắt cửa sổ lịch sử giá lệch.
- Index `idx_tp_canonical` là hạ tầng cho moat so sánh chéo sàn (§5.6); dòng `canonical_key IS NULL` (chưa so khớp) không cản tra cứu theo giá trị.
- Khi mở SEA, `platform` thêm dòng cho sàn nước khác; `tracked_product` không đổi cấu trúc, chỉ thêm SKU mới với `platform_id` tương ứng.

---

*Hết FR-PRICE-001. Status: ready_to_implement (mục tiêu audit 10/10).*
