---
id: TASK-B2B-001
title: "Pipeline dữ liệu xu hướng thị trường ẩn danh - aggregate giá theo category x thời gian từ price_daily, cổng k-anonymity (k>=50), KHÔNG lộ user hay SKU đơn lẻ, sinh bảng market_trend_daily nguồn cho mọi sản phẩm B2B"
module: B2B
priority: SHOULD
status: done
verify: T
phase: P3
milestone: P3 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-PRICE-002, TASK-COMPLY-003, TASK-B2B-002, TASK-B2B-003, TASK-B2B-004]
depends_on: [TASK-PRICE-002, TASK-COMPLY-003]
blocks: [TASK-B2B-002, TASK-B2B-003, TASK-B2B-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §6 mục 7 (B2B data/insights - báo cáo xu hướng giá/thị trường ẩn danh)"
  - "docs/... §1 (dòng doanh thu B2B margin cao), §5.5 (PDPL Luật 91/2025), §5.7 (SEA per-country)"
source_decisions:
  - "DEC-B2B-01: aggregate chỉ từ price_daily (đã tổng hợp theo ngày) và tracked_product.category_id - KHÔNG bao giờ chạm price_snapshot raw hay bất kỳ bảng user-level (wishlist, alert, cart) khi sinh dữ liệu B2B"
  - "DEC-B2B-02: cổng k-anonymity - một ô (category_id, platform_id, day) chỉ được phát hành khi có >= K_MIN SKU phân biệt đóng góp (K_MIN = 50); dưới ngưỡng thì nén (suppress), KHÔNG phát hành"
  - "DEC-B2B-03: chỉ phát hành chỉ số tổng hợp (median/p25/p75/count_sku/avg_discount_pct) - KHÔNG bao giờ phát hành giá của một SKU, một shop, hay một user đơn lẻ"
  - "DEC-B2B-04: pipeline là job batch đêm idempotent, ghi vào market_trend_daily với UPSERT theo (category_id, platform_id, day); chạy lại cùng ngày cho kết quả y hệt"
  - "DEC-B2B-05: ô bị suppress được ghi rõ trạng thái suppressed=true + sku_count thật (để audit) nhưng các cột chỉ số để NULL - người tiêu thụ downstream không đọc ô suppressed"
  - "DEC-B2B-06: market_trend_daily KHÔNG chứa product_id, shop_id, user_id hay bất kỳ khóa định danh nào - lược đồ tự nó không thể tái định danh"

language: "PostgreSQL 16 + TimescaleDB 2.x; service Go 1.22 (b2b-svc)"
service: shopass/services/b2b/
new_files:
  - services/b2b/migrations/0001_market_trend_daily.sql
  - services/b2b/internal/trend/aggregate.go
  - services/b2b/internal/trend/kanon.go
  - services/b2b/internal/trend/job.go
  - services/b2b/internal/trend/repo.go
  - services/b2b/internal/trend/aggregate_test.go
  - services/b2b/internal/trend/kanon_test.go
modified_files:
  - services/b2b/internal/trend/types.go            # struct MarketTrendCell
allowed_tools:
  - file_read: services/b2b/**
  - file_write: services/b2b/**
  - bash: cd services/b2b && go test ./...
disallowed_tools:
  - đọc price_snapshot raw, wishlist, alert, cart_snapshot hay bất kỳ bảng user-level nào để sinh dữ liệu B2B (vi phạm DEC-B2B-01)
  - phát hành ô có sku_count < K_MIN (vi phạm DEC-B2B-02, vỡ k-anonymity, rủi ro tái định danh)
  - đưa product_id/shop_id/user_id vào market_trend_daily (vi phạm DEC-B2B-06)

effort_hours: 10
sub_tasks:
  - "1.0h: 0001_market_trend_daily.sql - bảng aggregate + UNIQUE (category_id, platform_id, day) + CHECK sku_count >= 0"
  - "1.5h: aggregate.go - truy vấn nhóm price_daily theo (category_id, platform_id, day), tính median/p25/p75/avg_discount_pct/sku_count"
  - "1.5h: kanon.go - cổng K_MIN: ô đạt ngưỡng phát hành chỉ số; ô dưới ngưỡng suppress (chỉ số NULL, suppressed=true)"
  - "1.5h: job.go - batch đêm idempotent, lặp cửa sổ ngày, UPSERT vào market_trend_daily"
  - "1.0h: repo.go - UPSERT + QueryCells (đọc ô đã phát hành, mặc định lọc suppressed)"
  - "1.5h: kanon_test.go - ô < K_MIN bị suppress; ô = K_MIN được phát hành; ô suppress có chỉ số NULL"
  - "1.0h: aggregate_test.go - median/p25/p75 đúng; không có khóa định danh trong output"
  - "1.0h: OTel metric trend_cells_published_total + trend_cells_suppressed_total + job duration"

risk_if_skipped: "TASK-B2B-001 là nền của toàn bộ dòng doanh thu B2B margin cao (§1, §6 mục 7). Không có pipeline ẩn danh này thì báo cáo B2B (TASK-B2B-002), analytics đối thủ cho seller (TASK-B2B-003) và Premium API (TASK-B2B-004) đều không có nguồn dữ liệu hợp lệ để bán. Nguy hiểm nhất là làm sai bảo mật: nếu sinh dữ liệu B2B từ bảng user-level hoặc phát hành ô có quá ít SKU thì một bên mua dữ liệu có thể tái định danh giá của một shop hay một người dùng cụ thể - vi phạm PDPL Luật 91/2025 (chế tài tới 5% doanh thu, §5.5) và phá tan định vị minh bạch đạo đức hậu-Honey vốn là moat niềm tin của SănDeal. k-anonymity và việc chỉ đọc dữ liệu đã tổng hợp theo ngày là hai lớp phòng vệ bắt buộc, không phải tùy chọn."
---

## §1 - Mô tả (BCP-14 normative)

Service B2B **MUST** sinh bảng dữ liệu xu hướng thị trường ẩn danh `market_trend_daily` bằng cách tổng hợp giá theo `category_id` x `platform_id` x ngày, đọc duy nhất từ continuous aggregate `price_daily` và `tracked_product`, qua cổng k-anonymity, và KHÔNG bao giờ phát hành dữ liệu của một SKU, một shop hay một user đơn lẻ. Hợp đồng:

1. **MUST** định nghĩa bảng `market_trend_daily (category_id, platform_id, day, median_p, p25_p, p75_p, avg_discount_pct, sku_count, suppressed)` với `PRIMARY KEY (category_id, platform_id, day)`. Bảng KHÔNG chứa `product_id`, `shop_id`, `user_id` hay bất kỳ khóa định danh nào (DEC-B2B-06).
2. **MUST** chỉ đọc nguồn từ `price_daily` (continuous aggregate theo ngày) và `tracked_product.category_id` khi sinh dữ liệu (DEC-B2B-01). Pipeline **MUST NOT** đọc `price_snapshot` raw, `wishlist`, `alert`, `cart_snapshot` hay bất kỳ bảng user-level nào.
3. **MUST** tính cho mỗi ô `(category_id, platform_id, day)`: `median_p = median(close_p)`, `p25_p = percentile_cont(0.25)`, `p75_p = percentile_cont(0.75)`, `avg_discount_pct` (trung bình `(max_p - close_p) / max_p` trong ngày), và `sku_count` = số `product_id` phân biệt đóng góp vào ô.
4. **MUST** áp cổng k-anonymity (DEC-B2B-02): chỉ phát hành chỉ số cho ô có `sku_count >= K_MIN` với `K_MIN = 50`. Ô có `sku_count < K_MIN` **MUST** được ghi với `suppressed = true` và mọi cột chỉ số (`median_p, p25_p, p75_p, avg_discount_pct`) để `NULL` (DEC-B2B-05).
5. **MUST** phát hành chỉ số tổng hợp ở mức ô; **MUST NOT** phát hành giá của một SKU, một shop hay một user đơn lẻ qua bất kỳ cột nào (DEC-B2B-03).
6. **MUST** chạy như job batch đêm idempotent (DEC-B2B-04): lặp qua cửa sổ ngày cần tính, UPSERT vào `market_trend_daily` theo khóa `(category_id, platform_id, day)`. Chạy lại cùng cửa sổ ngày **MUST** cho kết quả y hệt (không nhân đôi, không trôi).
7. **MUST** expose hàm repo:
- `UpsertCells(ctx, cells []MarketTrendCell) error` - ghi idempotent.
- `QueryCells(ctx, categoryID int64, platformID int16, from, to time.Time) ([]MarketTrendCell, error)` - đọc ô đã phát hành; mặc định **MUST** lọc bỏ ô `suppressed = true`.
8. **MUST** ghi `sku_count` thật ngay cả khi ô bị suppress (để audit nội bộ), nhưng `QueryCells` mặc định không trả ô suppressed cho người tiêu thụ downstream (DEC-B2B-05).
9. **MUST** đảm bảo `sku_count >= 0` và `p25_p <= median_p <= p75_p` (khi không suppress) qua CHECK constraint hoặc bất biến kiểm trong test.
10. **SHOULD** phát OTel metric: `trend_cells_published_total{platform_id}` (counter), `trend_cells_suppressed_total{platform_id}` (counter), `trend_job_duration_ms` (histogram).
11. **MUST** chạy sau khi `price_daily` đã refresh cho cửa sổ ngày tính (đọc dữ liệu đã ổn định, tránh tính trên ngày chưa chốt).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao chỉ đọc price_daily, không đọc raw hay user-level (DEC-B2B-01)? Dữ liệu B2B là sản phẩm bán ra ngoài. Mọi đường dẫn dữ liệu từ bảng user-level (wishlist, cart, alert) tới đầu ra bán-được là một kênh rò rỉ tiềm tàng. Cắt tận gốc: pipeline chỉ chạm `price_daily` (đã là dữ liệu giá thị trường tổng hợp theo ngày) và `category_id`. Lược đồ đầu ra không có chỗ chứa khóa định danh, nên kể cả lập trình sai cũng không có cột để rò.

Vì sao k-anonymity với K_MIN = 50 (DEC-B2B-02)? Nếu một ô `(category, platform, day)` chỉ có 1-2 SKU, thì "median của ô" gần như chính là giá của một sản phẩm cụ thể - một bên mua dữ liệu biết category hẹp có thể suy ngược ra giá của đối thủ. Ngưỡng 50 SKU phân biệt làm chỉ số tổng hợp đủ "đông" để không quy về một cá thể. Category quá hẹp hoặc ngày quá thưa dữ liệu sẽ bị suppress thay vì phát hành.

Vì sao suppress ghi rõ trạng thái thay vì bỏ hẳn dòng (DEC-B2B-05)? Để audit phân biệt được "ô này chưa từng tính" với "ô này tính rồi nhưng cố ý không phát hành vì dưới ngưỡng". Ghi `suppressed=true` + `sku_count` thật cho phép đội compliance kiểm tra cổng k-anonymity thực sự chạy. Người tiêu thụ downstream không bao giờ thấy ô suppressed vì `QueryCells` lọc mặc định.

Vì sao batch idempotent UPSERT (DEC-B2B-04)? Job đêm có thể chạy lại (retry, backfill, sửa lỗi). UPSERT theo khóa ô làm việc chạy lại an toàn: cùng đầu vào -> cùng đầu ra, không nhân đôi. Dữ liệu B2B bán ra phải tái lập được để chứng minh tính đúng khi khách hàng B2B chất vấn một con số.

Vì sao chạy sau khi price_daily refresh (§1 #11)? `price_daily` là continuous aggregate refresh tăng dần (TASK-PRICE-002). Tính xu hướng trên một ngày chưa chốt sẽ cho số chập chờn. Chờ ngày đã ổn định để mỗi con số B2B là cuối cùng, không đổi về sau.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/b2b/migrations/0001_market_trend_daily.sql
CREATE TABLE market_trend_daily (
  category_id      BIGINT      NOT NULL,
  platform_id      SMALLINT    NOT NULL,
  day              DATE        NOT NULL,
  median_p         BIGINT,                 -- NULL khi suppressed
  p25_p            BIGINT,
  p75_p            BIGINT,
  avg_discount_pct NUMERIC(5,2),           -- 0..100
  sku_count        INTEGER     NOT NULL CHECK (sku_count >= 0),
  suppressed       BOOLEAN     NOT NULL DEFAULT false,
  computed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (category_id, platform_id, day),
  -- bất biến thứ tự phân vị khi không suppress
  CHECK (suppressed OR (p25_p <= median_p AND median_p <= p75_p))
);

CREATE INDEX idx_mtd_published ON market_trend_daily (category_id, platform_id, day)
  WHERE suppressed = false;
```

### Types (Go)

```go
// services/b2b/internal/trend/types.go
type MarketTrendCell struct {
    CategoryID     int64     `db:"category_id"`
    PlatformID     int16     `db:"platform_id"`
    Day            time.Time `db:"day"`
    MedianP        *int64    `db:"median_p"`           // NULL khi suppressed
    P25P           *int64    `db:"p25_p"`
    P75P           *int64    `db:"p75_p"`
    AvgDiscountPct *float64  `db:"avg_discount_pct"`
    SKUCount       int32     `db:"sku_count"`
    Suppressed     bool      `db:"suppressed"`
}

const KMin = 50 // ngưỡng k-anonymity
```

### Cổng k-anonymity (§1 #4)

```go
// services/b2b/internal/trend/kanon.go
// applyKAnon trả về ô phát hành được; nếu dưới ngưỡng thì suppress (chỉ số NULL).
func applyKAnon(raw aggRow) MarketTrendCell {
    c := MarketTrendCell{
        CategoryID: raw.CategoryID,
        PlatformID: raw.PlatformID,
        Day:        raw.Day,
        SKUCount:   raw.SKUCount,
    }
    if raw.SKUCount < KMin {
        c.Suppressed = true // chỉ số để NULL - không phát hành
        return c
    }
    c.MedianP = &raw.MedianP
    c.P25P = &raw.P25P
    c.P75P = &raw.P75P
    c.AvgDiscountPct = &raw.AvgDiscountPct
    return c
}
```

### Truy vấn aggregate (§1 #2, #3) - chỉ chạm price_daily + tracked_product

```sql
-- nguồn DUY NHẤT: price_daily (cagg theo ngày) JOIN tracked_product cho category/platform.
-- KHÔNG có join tới bảng user-level nào.
SELECT tp.category_id,
       tp.platform_id,
       pd.day,
       percentile_cont(0.5)  WITHIN GROUP (ORDER BY pd.close_p) AS median_p,
       percentile_cont(0.25) WITHIN GROUP (ORDER BY pd.close_p) AS p25_p,
       percentile_cont(0.75) WITHIN GROUP (ORDER BY pd.close_p) AS p75_p,
       avg(CASE WHEN pd.max_p > 0
                THEN (pd.max_p - pd.close_p)::numeric / pd.max_p * 100 END) AS avg_discount_pct,
       count(DISTINCT pd.product_id) AS sku_count
FROM price_daily pd
JOIN tracked_product tp ON tp.id = pd.product_id
WHERE pd.day >= $1 AND pd.day < $2
GROUP BY tp.category_id, tp.platform_id, pd.day;
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `market_trend_daily` tồn tại với PK `(category_id, platform_id, day)` và KHÔNG có cột `product_id/shop_id/user_id`.
2. Job sinh một ô có >= 50 SKU phân biệt -> `suppressed=false`, các cột chỉ số khác NULL.
3. Job sinh một ô có 49 SKU -> `suppressed=true`, mọi cột chỉ số (`median_p, p25_p, p75_p, avg_discount_pct`) là NULL.
4. Ô có 50 SKU (đúng ngưỡng) -> được phát hành (biên K_MIN là "lớn hơn hoặc bằng").
5. `QueryCells` mặc định KHÔNG trả ô `suppressed=true`.
6. Chạy job hai lần cho cùng cửa sổ ngày -> số dòng và mọi giá trị y hệt (idempotent, UPSERT).
7. Với ô được phát hành: `p25_p <= median_p <= p75_p` (bất biến phân vị).
8. `avg_discount_pct` nằm trong khoảng 0..100.
9. Quét mã pipeline: không có truy vấn nào chạm `price_snapshot`, `wishlist`, `alert`, `cart_snapshot` (kiểm bằng test grep hoặc review).
10. Output không chứa khóa định danh: mọi dòng `market_trend_daily` chỉ có `category_id/platform_id/day` + chỉ số tổng hợp.
11. Metric `trend_cells_published_total` tăng cho ô phát hành; `trend_cells_suppressed_total` tăng cho ô suppress.

---

## §5 - Kiểm thử (verification)

```go
// services/b2b/internal/trend/kanon_test.go
func TestKAnon_BelowThreshold_Suppressed(t *testing.T) {
    raw := aggRow{CategoryID: 7, PlatformID: 1, Day: d0, SKUCount: 49,
        MedianP: 120_000, P25P: 100_000, P75P: 140_000, AvgDiscountPct: 12.5}
    c := applyKAnon(raw)
    require.True(t, c.Suppressed)
    require.Nil(t, c.MedianP)
    require.Nil(t, c.P25P)
    require.Nil(t, c.P75P)
    require.Nil(t, c.AvgDiscountPct)
    require.EqualValues(t, 49, c.SKUCount) // sku_count thật vẫn ghi để audit
}

func TestKAnon_AtThreshold_Published(t *testing.T) {
    raw := aggRow{CategoryID: 7, PlatformID: 1, Day: d0, SKUCount: 50,
        MedianP: 120_000, P25P: 100_000, P75P: 140_000, AvgDiscountPct: 12.5}
    c := applyKAnon(raw)
    require.False(t, c.Suppressed)
    require.NotNil(t, c.MedianP)
    require.EqualValues(t, 120_000, *c.MedianP)
}

func TestQueryCells_SkipsSuppressed(t *testing.T) {
    r := setupRepo(t)
    r.UpsertCells(ctx, []MarketTrendCell{
        {CategoryID: 7, PlatformID: 1, Day: d0, SKUCount: 60, MedianP: ptr(int64(120_000)),
            P25P: ptr(int64(100_000)), P75P: ptr(int64(140_000))},
        {CategoryID: 7, PlatformID: 1, Day: d1, SKUCount: 10, Suppressed: true},
    })
    cells, _ := r.QueryCells(ctx, 7, 1, d0, d1.AddDate(0, 0, 1))
    require.Len(t, cells, 1) // chỉ ô phát hành
    require.Equal(t, d0.Unix(), cells[0].Day.Unix())
}

func TestJob_Idempotent(t *testing.T) {
    r := setupRepo(t)
    seedPriceDaily(t, /* category 7, platform 1, 60 SKU, 3 ngày */)
    runJob(t, r, d0, d0.AddDate(0, 0, 3))
    snap1 := dumpAll(t, r)
    runJob(t, r, d0, d0.AddDate(0, 0, 3)) // chạy lại
    snap2 := dumpAll(t, r)
    require.Equal(t, snap1, snap2) // y hệt
}

func TestAggregate_NoIdentifierColumns(t *testing.T) {
    cols := tableColumns(t, "market_trend_daily")
    require.NotContains(t, cols, "product_id")
    require.NotContains(t, cols, "shop_id")
    require.NotContains(t, cols, "user_id")
}

func TestAggregate_PercentileOrder(t *testing.T) {
    r := setupRepo(t)
    seedPriceDaily(t, /* 60 SKU phân bố giá */)
    runJob(t, r, d0, d0.AddDate(0, 0, 1))
    cells, _ := r.QueryCells(ctx, 7, 1, d0, d0.AddDate(0, 0, 1))
    require.LessOrEqual(t, *cells[0].P25P, *cells[0].MedianP)
    require.LessOrEqual(t, *cells[0].MedianP, *cells[0].P75P)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0001 (bảng) -> aggregate.go (truy vấn nhóm từ price_daily) -> kanon.go (cổng K_MIN) -> job.go (batch idempotent) -> repo.go -> tests. Job lập lịch chạy đêm sau cửa sổ refresh của `price_daily`. Cửa sổ ngày tính nên trễ một ngày so với hôm nay để chỉ chạm ngày đã chốt. Backfill lịch sử chạy job với cửa sổ ngày rộng, an toàn nhờ UPSERT idempotent.

---

## §7 - Phụ thuộc

- TASK-PRICE-002 - `price_daily` (continuous aggregate) phải tồn tại và đã refresh; là nguồn dữ liệu duy nhất của pipeline.
- TASK-COMPLY-003 - quyền chủ thể dữ liệu (DSAR); khi user xóa dữ liệu, đường dẫn B2B đã ẩn danh + tổng hợp nên không chứa dữ liệu cá nhân của họ, nhưng pipeline phải nhất quán với chính sách xử lý dữ liệu PDPL.
- TASK-B2B-002 / TASK-B2B-003 / TASK-B2B-004 (downstream) - đọc `market_trend_daily` qua `QueryCells`; không bao giờ đọc nguồn raw.
- Extension/lib: PostgreSQL `percentile_cont`; driver `pgx`.

---

## §8 - Payload ví dụ

### Một ô đã phát hành (đọc nội bộ qua QueryCells)

```json
{
  "category_id": 7,
  "platform_id": 1,
  "day": "2026-06-20",
  "median_p": 320000,
  "p25_p": 250000,
  "p75_p": 410000,
  "avg_discount_pct": 14.30,
  "sku_count": 412,
  "suppressed": false
}
```

### Một ô bị suppress (chỉ thấy trong audit nội bộ, downstream không đọc)

```json
{
  "category_id": 991,
  "platform_id": 2,
  "day": "2026-06-20",
  "median_p": null,
  "p25_p": null,
  "p75_p": null,
  "avg_discount_pct": null,
  "sku_count": 18,
  "suppressed": true
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- K_MIN có nên khác nhau theo độ nhạy của category (mỹ phẩm vs điện tử) - giữ 50 cho mọi category ở slice này, hiệu chỉnh sau nếu compliance yêu cầu.
- Thêm chiều `region`/`country` cho ô khi mở SEA - gắn khi TASK-COMPLY-007 chốt mô hình per-country (giữ k-anonymity per-country để không gộp chéo nước).
- Bổ sung nhiễu vi sai (differential privacy) lên trên k-anonymity - cân nhắc nếu khách hàng B2B yêu cầu độ chi tiết cao hơn mà vẫn an toàn.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Ô dưới ngưỡng được phát hành | kanon_test + audit suppressed | rủi ro tái định danh, vi phạm PDPL | Cổng K_MIN ép suppress; test biên 49/50 |
| Pipeline đọc bảng user-level | review + test grep truy vấn | rò rỉ dữ liệu cá nhân | Chỉ join price_daily + tracked_product (DEC-B2B-01) |
| Output lọt khóa định danh | TestAggregate_NoIdentifierColumns | tái định danh trực tiếp | Lược đồ không có cột product_id/shop_id/user_id |
| Job chạy lại nhân đôi dòng | TestJob_Idempotent | dữ liệu B2B sai, mất uy tín | UPSERT theo khóa ô |
| Tính trên ngày chưa chốt | so số đổi giữa các lần chạy | con số B2B chập chờn | Chỉ tính ngày trễ >= 1 sau refresh price_daily |
| p25 > median do lỗi tính phân vị | CHECK constraint + test | dữ liệu vô lý | percentile_cont đúng thứ tự + CHECK |
| avg_discount_pct ngoài 0..100 | test biên | chỉ số phi lý | CASE max_p > 0 + kẹp khoảng |
| category_id null trong tracked_product | join loại bỏ dòng | thiếu ô | Chấp nhận; SKU chưa phân loại không vào aggregate |
| K_MIN đặt quá thấp do cấu hình | review hằng số | rò rỉ | K_MIN là hằng số mã, đổi phải qua review |

---

## §11 - Ghi chú

- `market_trend_daily` là tài sản dữ liệu bán-được của SănDeal; lược đồ cố ý không chứa khóa định danh để bản thân nó không tái định danh được.
- Hai lớp phòng vệ độc lập: (1) chỉ đọc dữ liệu đã tổng hợp theo ngày, (2) cổng k-anonymity K_MIN = 50. Một lớp hỏng thì lớp kia vẫn chặn.
- Suppress ghi rõ trạng thái + sku_count thật là để audit chứng minh cổng k-anonymity thực sự chạy, không phải bị bỏ qua.
- Idempotent UPSERT làm dữ liệu B2B tái lập được - điều kiện để bán dữ liệu cho khách hàng đòi hỏi tính đúng kiểm chứng được.
- Khi mở SEA, giữ k-anonymity riêng từng nước; KHÔNG gộp SKU chéo nước để lấp ô thưa (sẽ trộn thị trường khác nhau và lách ngưỡng).

---

*Hết TASK-B2B-001. Status: ready_to_review (awaiting HITL).*

*HITL accept (operator merge-then-continue): feature PR #104 merge `1bc583fce6be2eda10737dfbfe79c43d20157ea5` → done.*
