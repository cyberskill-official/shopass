---
id: TASK-PRICE-002
title: "price_snapshot TimescaleDB hypertable - nén + continuous aggregate + delta-only + chunk 7 ngày cho time-series giá tỷ-dòng"
module: PRICE
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-PRICE-001, TASK-PRICE-003, TASK-PRICE-004, TASK-DEAL-001, TASK-SCRAPE-005]
depends_on: [TASK-PRICE-001]
blocks: [TASK-B2B-001, TASK-DEAL-001, TASK-DEAL-003, TASK-DEAL-004, TASK-PRICE-003, TASK-SCRAPE-005, TASK-TRACK-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model price_snapshot)"
  - "docs/... §3.8 (NFR khả năng mở rộng), §5.1 (cold-start 90 ngày)"
source_decisions:
  - "DEC-PRICE-01: price_snapshot là hypertable, chunk_time_interval = 7 ngày"
  - "DEC-PRICE-02: nén (compress) sau 30 ngày, segmentby = product_id"
  - "DEC-PRICE-03: continuous aggregate price_daily (min/max/last theo ngày) cho biểu đồ"
  - "DEC-PRICE-04: ghi delta-only - chỉ INSERT khi giá thay đổi so với snapshot gần nhất"
  - "DEC-PRICE-05: price lưu BIGINT (VND, không thập phân) để tránh sai số float"

language: "PostgreSQL 16 + TimescaleDB 2.x; service Go 1.22 (price-svc)"
service: shopass/services/price/
new_files:
  - services/price/migrations/0002_price_snapshot.sql
  - services/price/migrations/0003_price_daily_cagg.sql
  - services/price/migrations/0004_compression_policy.sql
  - services/price/internal/price/snapshot.go
  - services/price/internal/price/delta.go
  - services/price/internal/price/repo.go
  - services/price/internal/price/repo_test.go
  - services/price/internal/price/delta_test.go
modified_files:
  - services/price/internal/price/types.go            # thêm struct PriceSnapshot
allowed_tools:
  - file_read: services/price/**
  - file_write: services/price/**
  - bash: cd services/price && go test ./...
disallowed_tools:
  - lưu price dạng float/numeric thập phân (vi phạm DEC-PRICE-05)
  - ghi snapshot mỗi lần quét dù giá không đổi (vi phạm DEC-PRICE-04, đốt storage)
  - tạo bảng thường thay vì hypertable (vi phạm DEC-PRICE-01)

effort_hours: 8
sub_tasks:
  - "0.5h: 0002_price_snapshot.sql - bảng + create_hypertable(chunk 7d)"
  - "0.5h: 0003_price_daily_cagg.sql - continuous aggregate min/max/last"
  - "0.5h: 0004_compression_policy.sql - compress segmentby product_id + add_compression_policy 30d"
  - "1.0h: types.go (PriceSnapshot) + repo.go (insert + query range)"
  - "1.0h: delta.go - so sánh snapshot gần nhất, chỉ ghi khi (price, list_price, stock, flash_sale) đổi"
  - "0.5h: index + retention policy (giữ raw 18 tháng, aggregate vô hạn)"
  - "1.5h: repo_test.go - insert, query range 90d, hypertable tồn tại, cagg refresh"
  - "1.5h: delta_test.go - giá không đổi -> 0 row; giá đổi -> 1 row; flash_sale flip -> 1 row"
  - "1.0h: OTel metric price_snapshot_written_total + delta_skipped_total"
risk_if_skipped: "price_snapshot là bảng lớn nhất hệ thống (hàng tỷ dòng với hàng triệu SKU). Không hypertable -> query lịch sử giá quét toàn bảng, p95 vỡ NFR <500ms. Không nén -> chi phí storage time-series bùng nổ (vi phạm unit economics §4.1). Không delta-only -> ghi mỗi lần quét đốt storage 10-100x. Không continuous aggregate -> biểu đồ giá (TASK-WEB-003) và sale ảo (TASK-DEAL-001) không có nguồn dữ liệu đã tổng hợp. Đây là nền tảng cho toàn bộ tính năng lõi của SănDeal."
---

## §1 - Mô tả (BCP-14 normative)

Service PRICE **MUST** lưu chuỗi thời gian giá sản phẩm vào TimescaleDB hypertable `price_snapshot`, ghi theo nguyên tắc delta-only, nén sau 30 ngày, và phục vụ biểu đồ qua continuous aggregate. Hợp đồng:

1. **MUST** định nghĩa bảng `price_snapshot (product_id, ts, price, list_price, stock, sold, flash_sale)` với `PRIMARY KEY (product_id, ts)`, `product_id` REFERENCES `tracked_product(id)`.
2. **MUST** biến `price_snapshot` thành hypertable: `create_hypertable('price_snapshot','ts', chunk_time_interval => INTERVAL '7 days')` (DEC-PRICE-01).
3. **MUST** lưu `price` và `list_price` dạng `BIGINT` (đơn vị VND, không thập phân) - KHÔNG dùng float/numeric (DEC-PRICE-05). Sai số float trên tiền tệ là không chấp nhận được.
4. **MUST** ghi delta-only (DEC-PRICE-04): trước khi INSERT, đọc snapshot gần nhất của `product_id`; chỉ ghi dòng mới khi ÍT NHẤT một trong `(price, list_price, stock, flash_sale)` thay đổi. Nếu mọi trường bằng snapshot gần nhất -> bỏ qua, tăng metric `delta_skipped_total`.
5. **MUST** bật nén (DEC-PRICE-02): `ALTER TABLE price_snapshot SET (timescaledb.compress, timescaledb.compress_segmentby='product_id')` + `add_compression_policy('price_snapshot', INTERVAL '30 days')`. Chunk cũ hơn 30 ngày được nén tự động.
6. **MUST** tạo continuous aggregate `price_daily` (DEC-PRICE-03): bucket 1 ngày, cột `min_p = min(price)`, `max_p = max(price)`, `close_p = last(price, ts)`, GROUP BY `product_id, day`. Đăng ký refresh policy (refresh cửa sổ gần nhất mỗi giờ).
7. **MUST** đặt retention policy: giữ raw `price_snapshot` 18 tháng (`add_retention_policy('price_snapshot', INTERVAL '18 months')`); continuous aggregate `price_daily` giữ vô thời hạn (lịch sử dài phục vụ dự đoán đáy giá TASK-DEAL-004).
8. **MUST** expose hàm repo:
- `InsertSnapshot(ctx, snap PriceSnapshot) (written bool, err error)` - áp dụng delta-only; `written=false` khi bị bỏ qua.
- `QueryRange(ctx, productID int64, from, to time.Time) ([]PriceSnapshot, error)` - đọc raw trong khoảng.
- `QueryDaily(ctx, productID int64, from, to time.Time) ([]DailyBucket, error)` - đọc từ `price_daily`.
9. **MUST** đánh index hỗ trợ truy vấn theo `product_id` + thời gian (hypertable đã index `ts`; thêm index phụ nếu cần cho `flash_sale` filter).
10. **SHOULD** phát OTel metric: `price_snapshot_written_total{platform_id}` (counter), `price_snapshot_delta_skipped_total{platform_id}` (counter), `price_query_duration_ms` (histogram).
11. **MUST** xử lý ghi đồng thời cùng `(product_id, ts)`: dùng `INSERT ... ON CONFLICT (product_id, ts) DO NOTHING` để idempotent với retry của scraper.
12. **MUST** đảm bảo `price > 0` và `list_price IS NULL OR list_price >= price` qua CHECK constraint (giá bán không cao hơn giá niêm yết).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao hypertable chunk 7 ngày (DEC-PRICE-01)?** `price_snapshot` là bảng lớn nhất hệ thống - hàng triệu SKU x nhiều snapshot/ngày = hàng tỷ dòng. Hypertable tự chia chunk theo thời gian: query "90 ngày gần nhất" chỉ chạm ~13 chunk thay vì quét toàn bảng. Chunk 7 ngày cân bằng giữa số chunk (không quá nhiều) và độ mịn của nén/retention.

**Vì sao delta-only (DEC-PRICE-04)?** Scraper quét SKU theo tần suất (flash sale: phút; SKU thường: vài giờ). Phần lớn lần quét giá KHÔNG đổi. Ghi mỗi lần quét đốt storage 10-100x vô ích. Delta-only chỉ ghi khi có thay đổi thực - vừa tiết kiệm storage (giữ unit economics §4.1 ~0,1-0,2 USD/user), vừa làm chuỗi giá "sạch" cho thuật toán sale ảo.

**Vì sao price là BIGINT VND (DEC-PRICE-05)?** Giá VN luôn là số nguyên đồng (không có "0,5 đồng"). Lưu float gây sai số tích lũy khi so sánh `current_price >= median90 * 0.97` trong sale ảo. BIGINT là chính xác tuyệt đối và đủ lớn (giá tới hàng nghìn tỷ vẫn vừa int64).

**Vì sao nén sau 30 ngày (DEC-PRICE-02)?** Dữ liệu cũ hơn 30 ngày hiếm khi ghi mới, chủ yếu đọc. Nén columnar segmentby `product_id` đạt tỷ lệ nén cao (cùng product_id, giá tương quan) -> giảm storage 10-20x cho phần "đuôi" lịch sử dài. 30 ngày để chunk "nóng" (đang ghi) không bị nén sớm.

**Vì sao continuous aggregate (DEC-PRICE-03)?** Biểu đồ giá (TASK-WEB-003, NFR <500ms) và sale ảo (TASK-DEAL-001) không cần độ phân giải giây - cần min/max/close theo ngày. Tính realtime từ raw mỗi lần render là chậm. Continuous aggregate tính sẵn, refresh tăng dần -> query biểu đồ chỉ đọc bảng nhỏ đã tổng hợp.

**Vì sao retention raw 18 tháng nhưng aggregate vô hạn (§1 #7)?** Raw giây/phút chỉ hữu ích gần đây (sale ảo cần 90 ngày). Aggregate ngày nhẹ và là input cho dự đoán đáy giá LightGBM (TASK-DEAL-005, cần >=180 ngày). Giữ aggregate dài, vứt raw cũ - cân bằng chi phí vs giá trị.

**Vì sao ON CONFLICT DO NOTHING (§1 #11)?** Scraper có retry; cùng `(product_id, ts)` có thể được gửi 2 lần. PK + DO NOTHING làm INSERT idempotent, không vỡ khi trùng.

---

## §3 - Hợp đồng API / DDL

### Migrations

```sql
-- services/price/migrations/0002_price_snapshot.sql
CREATE TABLE price_snapshot (
  product_id   BIGINT      NOT NULL REFERENCES tracked_product(id),
  ts           TIMESTAMPTZ NOT NULL,
  price        BIGINT      NOT NULL CHECK (price > 0),          -- VND, không thập phân
  list_price   BIGINT      CHECK (list_price IS NULL OR list_price >= price),
  stock        INTEGER,
  sold         INTEGER,
  flash_sale   BOOLEAN     NOT NULL DEFAULT false,
  PRIMARY KEY (product_id, ts)
);

SELECT create_hypertable('price_snapshot', 'ts',
  chunk_time_interval => INTERVAL '7 days');

CREATE INDEX idx_ps_flash ON price_snapshot (product_id, ts DESC)
  WHERE flash_sale = true;

-- services/price/migrations/0003_price_daily_cagg.sql
CREATE MATERIALIZED VIEW price_daily
  WITH (timescaledb.continuous) AS
  SELECT product_id,
         time_bucket('1 day', ts) AS day,
         min(price)      AS min_p,
         max(price)      AS max_p,
         last(price, ts) AS close_p
  FROM price_snapshot
  GROUP BY product_id, day
  WITH NO DATA;

SELECT add_continuous_aggregate_policy('price_daily',
  start_offset => INTERVAL '3 days',
  end_offset   => INTERVAL '1 hour',
  schedule_interval => INTERVAL '1 hour');

-- services/price/migrations/0004_compression_policy.sql
ALTER TABLE price_snapshot SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'product_id'
);
SELECT add_compression_policy('price_snapshot', INTERVAL '30 days');
SELECT add_retention_policy('price_snapshot',  INTERVAL '18 months');
```

### Types (Go)

```go
// services/price/internal/price/types.go
type PriceSnapshot struct {
    ProductID  int64     `db:"product_id"`
    TS         time.Time `db:"ts"`
    Price      int64     `db:"price"`       // VND
    ListPrice  *int64    `db:"list_price"`
    Stock      *int32    `db:"stock"`
    Sold       *int32    `db:"sold"`
    FlashSale  bool      `db:"flash_sale"`
}

type DailyBucket struct {
    ProductID int64     `db:"product_id"`
    Day       time.Time `db:"day"`
    MinP      int64     `db:"min_p"`
    MaxP      int64     `db:"max_p"`
    CloseP    int64     `db:"close_p"`
}
```

### Delta-only (§1 #4)

```go
// services/price/internal/price/delta.go
// InsertSnapshot áp dụng delta-only: chỉ ghi khi khác snapshot gần nhất.
func (r *Repo) InsertSnapshot(ctx context.Context, s PriceSnapshot) (bool, error) {
    last, err := r.latest(ctx, s.ProductID)
    if err != nil && !errors.Is(err, pgx.ErrNoRows) {
        return false, err
    }
    if err == nil && !changed(last, s) {
        metrics.DeltaSkipped(s.ProductID)
        return false, nil // không đổi → bỏ qua
    }
    _, err = r.pool.Exec(ctx,
        `INSERT INTO price_snapshot (product_id, ts, price, list_price, stock, sold, flash_sale)
         VALUES ($1,$2,$3,$4,$5,$6,$7)
         ON CONFLICT (product_id, ts) DO NOTHING`,
        s.ProductID, s.TS, s.Price, s.ListPrice, s.Stock, s.Sold, s.FlashSale)
    if err != nil {
        return false, err
    }
    metrics.SnapshotWritten(s.ProductID)
    return true, nil
}

func changed(a, b PriceSnapshot) bool {
    return a.Price != b.Price ||
        !eqPtr(a.ListPrice, b.ListPrice) ||
        a.FlashSale != b.FlashSale ||
        !eqPtr32(a.Stock, b.Stock)
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `price_snapshot` tồn tại VÀ là hypertable (`SELECT * FROM timescaledb_information.hypertables WHERE hypertable_name='price_snapshot'` trả 1 dòng).
2. `chunk_time_interval` = 7 ngày (kiểm qua `timescaledb_information.dimensions`).
3. INSERT snapshot mới (product chưa có dữ liệu) -> `written=true`, 1 dòng trong bảng.
4. INSERT snapshot trùng giá/stock/flash với dòng gần nhất -> `written=false`, KHÔNG thêm dòng.
5. INSERT khi `price` đổi -> `written=true`, thêm 1 dòng.
6. INSERT khi `flash_sale` flip (false->true) cùng giá -> `written=true` (flash là tín hiệu có ý nghĩa).
7. INSERT `price <= 0` -> lỗi CHECK constraint.
8. INSERT `list_price < price` -> lỗi CHECK constraint.
9. INSERT trùng `(product_id, ts)` 2 lần -> DO NOTHING, không lỗi, 1 dòng.
10. `QueryRange(productID, now-90d, now)` trả đúng các snapshot trong khoảng, sắp theo `ts`.
11. Continuous aggregate `price_daily` tồn tại; sau refresh, `QueryDaily` trả min/max/close đúng theo ngày.
12. Sau khi nén thủ công 1 chunk cũ (`compress_chunk`), query vẫn trả đúng dữ liệu (nén trong suốt với đọc).
13. Metric `price_snapshot_written_total` tăng khi `written=true`; `delta_skipped_total` tăng khi bỏ qua.

---

## §5 - Kiểm thử (verification)

```go
// services/price/internal/price/delta_test.go
func TestDelta_NoChange_Skips(t *testing.T) {
    r, pid := setupWithProduct(t)
    s := PriceSnapshot{ProductID: pid, TS: t0, Price: 100_000}
    w1, _ := r.InsertSnapshot(ctx, s)
    require.True(t, w1)

    s2 := PriceSnapshot{ProductID: pid, TS: t0.Add(time.Hour), Price: 100_000}
    w2, _ := r.InsertSnapshot(ctx, s2)
    require.False(t, w2) // giá không đổi → bỏ qua

    n := countRows(t, r, pid)
    require.Equal(t, 1, n)
}

func TestDelta_PriceChange_Writes(t *testing.T) {
    r, pid := setupWithProduct(t)
    r.InsertSnapshot(ctx, PriceSnapshot{ProductID: pid, TS: t0, Price: 100_000})
    w, _ := r.InsertSnapshot(ctx, PriceSnapshot{ProductID: pid, TS: t0.Add(time.Hour), Price: 89_000})
    require.True(t, w)
    require.Equal(t, 2, countRows(t, r, pid))
}

func TestDelta_FlashFlip_Writes(t *testing.T) {
    r, pid := setupWithProduct(t)
    r.InsertSnapshot(ctx, PriceSnapshot{ProductID: pid, TS: t0, Price: 100_000, FlashSale: false})
    w, _ := r.InsertSnapshot(ctx, PriceSnapshot{ProductID: pid, TS: t0.Add(time.Minute), Price: 100_000, FlashSale: true})
    require.True(t, w) // flash flip là tín hiệu, dù giá bằng
}

func TestHypertable_Exists(t *testing.T) {
    r, _ := setupWithProduct(t)
    var n int
    r.pool.QueryRow(ctx,
        `SELECT count(*) FROM timescaledb_information.hypertables
         WHERE hypertable_name='price_snapshot'`).Scan(&n)
    require.Equal(t, 1, n)
}

func TestQueryRange_90d(t *testing.T) {
    r, pid := setupWithProduct(t)
    seedDailyPrices(t, r, pid, 120) // 120 ngày
    rows, _ := r.QueryRange(ctx, pid, time.Now().AddDate(0,0,-90), time.Now())
    require.LessOrEqual(t, len(rows), 90)
    require.True(t, sortedByTS(rows))
}

func TestCheck_PriceNonPositive(t *testing.T) {
    r, pid := setupWithProduct(t)
    _, err := r.InsertSnapshot(ctx, PriceSnapshot{ProductID: pid, TS: t0, Price: 0})
    require.Error(t, err) // CHECK price > 0
}

func TestConflict_Idempotent(t *testing.T) {
    r, pid := setupWithProduct(t)
    s := PriceSnapshot{ProductID: pid, TS: t0, Price: 100_000}
    r.InsertSnapshot(ctx, s)
    r.InsertSnapshot(ctx, s) // trùng PK → DO NOTHING
    require.Equal(t, 1, countRows(t, r, pid))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0002 (bảng + hypertable) -> 0003 (cagg) -> 0004 (compress + retention) -> repo.go + delta.go -> tests. Migration chạy qua `golang-migrate` hoặc `tern`. Continuous aggregate phải tạo `WITH NO DATA` rồi để policy backfill (tránh khóa dài lúc deploy trên bảng lớn).

---

## §7 - Phụ thuộc

- **TASK-PRICE-001** - `tracked_product` phải tồn tại trước (FK `product_id`).
- **TASK-SCRAPE-005 (downstream)** - scraper gọi `InsertSnapshot` với delta-only.
- **TASK-DEAL-001 (downstream)** - sale ảo đọc `QueryRange` 90 ngày.
- **TASK-WEB-003 / TASK-DEAL-003 (downstream)** - biểu đồ đọc `price_daily`.
- Extension/lib: TimescaleDB 2.x (compression, continuous aggregate, retention); driver `pgx`.

---

## §8 - Payload ví dụ

### Scraper ghi snapshot (nội bộ)

```go
written, err := priceRepo.InsertSnapshot(ctx, price.PriceSnapshot{
    ProductID: 90112,
    TS:        time.Now(),
    Price:     89_000,        // VND
    ListPrice: ptr(int64(149_000)),
    Stock:     ptr(int32(37)),
    Sold:      ptr(int32(1240)),
    FlashSale: true,
})
// written=true nếu giá đổi so với lần quét trước; false nếu bỏ qua (delta-only)
```

### Đọc lịch sử 90 ngày (cho biểu đồ / sale ảo)

```sql
SELECT day, min_p, max_p, close_p
FROM price_daily
WHERE product_id = 90112
  AND day >= now() - INTERVAL '90 days'
ORDER BY day;
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Phân vùng theo `platform_id` ngoài `product_id` cho nén - slice sau nếu skew.
- Downsampling raw (giữ phút trong 7 ngày, giờ trong 30 ngày) - tối ưu storage giai đoạn sau.
- Multi-currency (THB/IDR khi mở SEA) - gắn `currency` vào `tracked_product`, giữ `price` BIGINT theo minor unit của nước.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| FK product_id không tồn tại | lỗi pgx | 400/skip | Scraper tạo tracked_product trước (TASK-PRICE-001) |
| price <= 0 | DB CHECK | từ chối ghi | Scraper validate trước khi gửi |
| list_price < price | DB CHECK | từ chối ghi | Sửa nguồn parse (có thể đảo cột) |
| Trùng (product_id, ts) | ON CONFLICT | DO NOTHING | Theo thiết kế (idempotent retry) |
| Cagg refresh chậm/đọng | TS policy log | biểu đồ trễ vài giờ | Refresh thủ công cửa sổ; chỉnh schedule_interval |
| Nén làm chậm ghi vào chunk cũ | hiếm (chunk cũ ít ghi) | INSERT vào chunk nén -> giải nén ngầm | Chấp nhận; chunk nóng không bị nén (30d) |
| Storage time-series phình | dashboard Grafana | chi phí tăng | Kiểm tra retention policy chạy; delta-only hoạt động |
| Delta-only bỏ sót thay đổi (so sánh thiếu trường) | property test | mất tín hiệu giá | `changed()` so đủ price/list_price/stock/flash |
| Đọc range lớn (>1 năm raw) | p95 query metric | chậm | Chuyển sang price_daily cho khoảng dài |
| Migration cagg khóa bảng lớn | deploy treo | downtime | Tạo WITH NO DATA + backfill qua policy |

---

## §11 - Ghi chú

- `price_snapshot` là bảng nền tảng - mọi tính năng lõi (sale ảo, biểu đồ, dự đoán đáy, so sánh chéo, B2B data) đọc từ đây hoặc từ `price_daily`.
- Delta-only là đòn bẩy unit economics lớn nhất phía storage: phần lớn lần quét giá không đổi nên không ghi.
- BIGINT VND tránh hoàn toàn sai số float trên phép so sánh phần trăm của thuật toán sale ảo.
- Continuous aggregate tách "độ phân giải ghi" (giây/phút) khỏi "độ phân giải đọc biểu đồ" (ngày) - đọc nhanh mà vẫn giữ raw chi tiết.
- Retention 18 tháng raw + aggregate vô hạn cân bằng chi phí với nhu cầu lịch sử dài của ML dự đoán đáy.
- Khi mở SEA, gắn currency vào tracked_product; price vẫn BIGINT theo minor unit từng nước.

---

*Hết TASK-PRICE-002. Status: ready_to_implement (mục tiêu audit 10/10).*
