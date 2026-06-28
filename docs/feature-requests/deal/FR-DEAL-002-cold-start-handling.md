---
id: FR-DEAL-002
title: "Xử lý cold-start cho sale ảo + biểu đồ giá - 3 trạng thái trưởng thành NEW/WARMING/MATURE, category priors làm fallback, cổng baseline 90 ngày trước khi bật tính năng công khai cho 1 SKU"
module: DEAL
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-DEAL-001, FR-DEAL-004, FR-PRICE-002, FR-PRICE-001]
depends_on: [FR-DEAL-001]
blocks: [FR-DEAL-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.1 (cold-start chicken-and-egg, ra mắt tính năng khi đủ ~90 ngày dữ liệu cho top SKU)"
  - "docs/... §3.5 (category priors, sản phẩm <14 ngày trả UNKNOWN)"
source_decisions:
  - "DEC-DEAL-10: ba trạng thái trưởng thành theo số ngày lịch sử - NEW (<14 ngày), WARMING (14-90 ngày), MATURE (>=90 ngày)"
  - "DEC-DEAL-11: cổng baseline 90 ngày - chỉ bật sale ảo/biểu đồ công khai cho 1 SKU khi nó đạt MATURE (§5.1)"
  - "DEC-DEAL-12: category_prior là thống kê gộp theo category_id, dùng làm fallback khi SKU chưa đủ lịch sử"
  - "DEC-DEAL-13: SKU <14 ngày luôn trả verdict UNKNOWN (nhất quán với FR-DEAL-001), không đoán mò"
  - "DEC-DEAL-14: category_prior chỉ gộp từ SKU MATURE - loại SKU non khỏi mẫu để tránh thiên lệch phản hồi"

language: "Go 1.22 (deal-svc); PostgreSQL 16 + TimescaleDB 2.x (category_prior aggregate)"
service: shopass/services/deal/
new_files:
  - services/deal/migrations/0001_category_prior.sql
  - services/deal/internal/coldstart/maturity.go
  - services/deal/internal/coldstart/priors.go
  - services/deal/internal/coldstart/maturity_test.go
modified_files:
  - services/deal/internal/fakesale/detect.go            # hỏi cổng maturity trước khi phát verdict tự tin
allowed_tools:
  - file_read: services/deal/**
  - file_write: services/deal/**
  - bash: cd services/deal && go test ./...
disallowed_tools:
  - phát verdict tự tin cho SKU <14 ngày (vi phạm DEC-DEAL-13, phải trả UNKNOWN)
  - bật sale ảo/biểu đồ công khai cho SKU chưa đạt MATURE (vi phạm DEC-DEAL-11)
  - gộp SKU non (NEW/WARMING) vào category_prior (vi phạm DEC-DEAL-14, gây thiên lệch)

effort_hours: 6
sub_tasks:
  - "0.5h: 0001_category_prior.sql - materialized view gộp median/discount depth/sample count theo category_id, chỉ từ SKU MATURE"
  - "1.0h: maturity.go - Maturity(daysOfHistory) State + hằng số ngưỡng 14/90"
  - "1.0h: maturity.go - IsFeatureReady(product) gating sale ảo/biểu đồ công khai ở 90 ngày"
  - "1.0h: priors.go - PriorFor(categoryID) CategoryPrior + dùng làm fallback khi SKU non"
  - "0.5h: detect.go - chèn cổng maturity trước khi surface verdict tự tin (handoff FR-DEAL-001)"
  - "1.5h: maturity_test.go - 4 test bảng (biên 13/14/89/90, cổng 90d, prior loại SKU non, fallback khi NEW)"
  - "0.5h: OTel metric deal_maturity_state_total{state} + category_prior_fallback_total{category_id}"

risk_if_skipped: "SănDeal có bài toán con-gà-quả-trứng (§5.1): tính năng lõi (verdict sale ảo, dự đoán đáy, biểu đồ) cần lịch sử giá, nhưng lịch sử cần thời gian tích lũy. Không có chính sách cold-start thì hệ thống sẽ phát verdict bừa cho SKU mới chỉ có vài ngày dữ liệu, gọi nhầm sale xịn thành sale ảo (hoặc ngược lại), phá niềm tin người dùng (§5.4) ngay từ lần dùng đầu. Thiếu cổng baseline 90 ngày thì biểu đồ và verdict công khai ra mắt khi dữ liệu còn quá mỏng. Thiếu category_prior thì SKU mới không có điểm tựa nào để ước lượng tạm, và FR-DEAL-004 (Prophet baseline) mất nguồn prior cho cold-start. Thiếu việc loại SKU non khỏi prior thì thống kê category bị kéo lệch bởi chính những SKU chưa chín, tạo vòng phản hồi sai."
---

## §1 - Mô tả (BCP-14 normative)

Service DEAL **MUST** định nghĩa chính sách cold-start cho bộ phát hiện sale ảo (FR-DEAL-001) và biểu đồ giá: phân SKU thành 3 trạng thái trưởng thành theo số ngày lịch sử, dùng category priors làm fallback khi thiếu lịch sử, và đặt cổng baseline 90 ngày trước khi bật tính năng công khai cho từng SKU. Hợp đồng:

1. **MUST** tính `daysOfHistory` cho mỗi SKU = số ngày từ `first_seen` của `tracked_product` (hoặc snapshot `price_snapshot` sớm nhất, lấy mốc nào sớm hơn) tới hiện tại.
2. **MUST** ánh xạ `daysOfHistory` sang trạng thái: `NEW` khi `< 14` ngày, `WARMING` khi `14 <= d < 90` ngày, `MATURE` khi `>= 90` ngày (DEC-DEAL-10). Ngưỡng là hằng số có tên, không phải số ma thuật rải rác.
3. **MUST** trả verdict `UNKNOWN` cho mọi SKU ở trạng thái `NEW` (`< 14` ngày), nhất quán với FR-DEAL-001 (DEC-DEAL-13). Không đoán mò khi dữ liệu chưa đủ.
4. **MUST** ở trạng thái `WARMING` (14-90 ngày): cho phép verdict độ tin cậy thấp (gắn cờ `low_confidence`) và hiển thị biểu đồ kèm chú thích `"đang tích lũy"`; KHÔNG được phát verdict độ tin cậy đầy đủ.
5. **MUST** ở trạng thái `MATURE` (`>= 90` ngày): cho phép verdict độ tin cậy đầy đủ và biểu đồ công khai không kèm chú thích cảnh báo dữ liệu mỏng.
6. **MUST** định nghĩa materialized view `category_prior` gộp theo `category_id`: `median_price` tiêu biểu, `typical_discount_depth` (độ sâu giảm giá điển hình), `sample_count` (số SKU đóng góp). Đây là prior để mượn sức mạnh thống kê từ cả category.
7. **MUST** chỉ gộp `category_prior` từ SKU ở trạng thái `MATURE` (DEC-DEAL-14): loại mọi SKU `NEW`/`WARMING` khỏi mẫu để tránh thiên lệch phản hồi (SKU non kéo prior lệch, rồi prior lệch lại làm SKU non trông sai).
8. **MUST** expose hàm cổng `IsFeatureReady(product) bool` (DEC-DEAL-11): trả `true` chỉ khi SKU đạt `MATURE` (`>= 90` ngày). Sale ảo công khai và biểu đồ công khai cho 1 SKU **MUST** đi qua cổng này; cổng `false` thì chỉ được hiển thị trạng thái "đang tích lũy", không phát verdict công khai.
9. **MUST** dùng `category_prior` làm fallback CHỈ khi SKU chưa `MATURE`: với SKU `NEW`/`WARMING`, các ước lượng (median tham chiếu, độ sâu giảm điển hình) lấy từ prior của `category_id`; với SKU `MATURE`, dùng lịch sử riêng của SKU, không dùng prior.
10. **MUST** xử lý SKU không có `category_id` (NULL): trả `UNKNOWN` và không có fallback prior (không thể mượn sức mạnh từ category nào).
11. **SHOULD** tính lại `category_prior` theo lịch (refresh materialized view hằng ngày) để prior bám theo dữ liệu MATURE mới nhất.
12. **MUST** đặt sàn `min_sample_count` cho prior: nếu `sample_count` của một `category_id` dưới ngưỡng (ví dụ `< 30` SKU MATURE), prior của category đó coi như không đủ tin và fallback trả `UNKNOWN` thay vì prior mỏng.
13. **SHOULD** phát OTel metric: `deal_maturity_state_total{state}` (counter theo NEW/WARMING/MATURE), `category_prior_fallback_total{category_id}` (counter mỗi lần dùng prior), `category_prior_sample_count{category_id}` (gauge).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao cổng 90 ngày (DEC-DEAL-11, §5.1)?** §5.1 chốt nguyên tắc ra mắt: chỉ bật tính năng khi đủ khoảng 90 ngày dữ liệu cho top SKU. Sale ảo so giá hiện tại với median 90 ngày và đáy gần đây; nếu cửa sổ chưa đủ 90 ngày, "median 90 ngày" thực ra chỉ là median của vài tuần và rất dễ bị một đợt flash sale kéo lệch. Cổng `IsFeatureReady` ép tính năng công khai chờ tới khi cửa sổ đủ dài, đổi độ trễ ra mắt lấy độ chính xác.

**Vì sao UNKNOWN tốt hơn một phán đoán sai (§5.4)?** Tài sản lớn nhất của SănDeal là niềm tin: người dùng tin "sale ảo" nghĩa là thật sự ảo. Một verdict sai trên SKU mới (gọi sale xịn thành ảo, hoặc ngược lại) đắt hơn nhiều so với việc nói thẳng "chưa đủ dữ liệu". `NEW -> UNKNOWN` là sự khiêm tốn có chủ đích: thà im lặng còn hơn nói sai và mất uy tín ngay lần đầu người dùng kiểm chứng.

**Vì sao category priors (DEC-DEAL-12)?** Một SKU mới chưa có lịch sử riêng, nhưng category của nó thì có. Điện thoại trong cùng nhóm ngành có biên giảm giá điển hình giống nhau, có mùa double-date giống nhau. Mượn thống kê gộp của category cho ta một điểm tựa hợp lý để ước lượng tạm trong giai đoạn WARMING, thay vì khởi đầu từ con số không. Đây cũng chính là prior mà FR-DEAL-004 (Prophet baseline) cần để khởi tạo dự đoán đáy cho SKU cold-start.

**Vì sao loại SKU non khỏi prior aggregate (DEC-DEAL-14)?** Nếu category_prior gộp cả SKU NEW/WARMING, ta tạo một vòng phản hồi: SKU non có dữ liệu nhiễu kéo median và discount depth của category lệch đi, rồi prior lệch đó lại được áp ngược lên các SKU non khác làm chúng trông càng sai. Chỉ gộp từ SKU MATURE giữ prior "sạch" - nó phản ánh hành vi giá đã ổn định, không phản ánh nhiễu của giai đoạn cold-start.

**Vì sao sàn min_sample_count (§1 #12)?** Một category mới chỉ có 3-4 SKU MATURE thì median của nó không đáng tin hơn một SKU đơn lẻ là bao. Đặt sàn (ví dụ 30 SKU) buộc prior chỉ phát huy khi category đủ dày; dưới sàn thì thà trả UNKNOWN còn hơn dựa vào một prior mỏng và dễ lệch.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/deal/migrations/0001_category_prior.sql
-- Gộp thống kê theo category_id, CHỈ từ SKU đã MATURE (>= 90 ngày lịch sử).
-- mature_sku: 1 dòng/SKU MATURE kèm median giá riêng và độ sâu giảm gần nhất.
CREATE MATERIALIZED VIEW category_prior AS
  WITH mature_sku AS (
    SELECT tp.category_id,
           percentile_cont(0.5) WITHIN GROUP (ORDER BY pd.close_p) AS sku_median,
           max(pd.max_p) AS sku_list,    -- xấp xỉ giá niêm yết = đỉnh quan sát
           min(pd.min_p) AS sku_floor
    FROM tracked_product tp
    JOIN price_daily pd ON pd.product_id = tp.id
    WHERE tp.category_id IS NOT NULL
      AND now() - tp.first_seen >= INTERVAL '90 days'   -- chỉ SKU MATURE
    GROUP BY tp.id, tp.category_id
  )
  SELECT category_id,
         percentile_cont(0.5) WITHIN GROUP (ORDER BY sku_median)              AS median_price,
         percentile_cont(0.5) WITHIN GROUP (
           ORDER BY (sku_list - sku_floor)::float / NULLIF(sku_list, 0))      AS typical_discount_depth,
         count(*)                                                             AS sample_count
  FROM mature_sku
  GROUP BY category_id;

CREATE UNIQUE INDEX idx_category_prior_cat ON category_prior (category_id);
-- Refresh hằng ngày (REFRESH ... CONCURRENTLY, đăng ký ở scheduler deal-svc, §1 #11).
```

### Maturity (Go)

```go
// services/deal/internal/coldstart/maturity.go
package coldstart

type State int

const (
    StateNew     State = iota // < 14 ngày  -> UNKNOWN
    StateWarming              // 14-90 ngày -> low-confidence + "đang tích lũy"
    StateMature              // >= 90 ngày -> full confidence
)

const (
    warmingDays = 14
    matureDays  = 90
    minSamples  = 30 // sàn sample_count cho prior (§1 #12)
)

// Maturity ánh xạ số ngày lịch sử sang trạng thái trưởng thành (DEC-DEAL-10).
func Maturity(daysOfHistory int) State {
    switch {
    case daysOfHistory < warmingDays:
        return StateNew
    case daysOfHistory < matureDays:
        return StateWarming
    default:
        return StateMature
    }
}

// IsFeatureReady: cổng baseline 90 ngày cho sale ảo/biểu đồ công khai (DEC-DEAL-11).
func IsFeatureReady(p Product) bool {
    return Maturity(p.DaysOfHistory()) == StateMature
}
```

### Priors (Go)

```go
// services/deal/internal/coldstart/priors.go
package coldstart

type CategoryPrior struct {
    CategoryID    int64   `db:"category_id"`
    MedianPrice   int64   `db:"median_price"`
    DiscountDepth float64 `db:"typical_discount_depth"`
    SampleCount   int     `db:"sample_count"`
}

// PriorFor đọc prior của 1 category; ok=false khi không có category,
// hoặc sample_count dưới sàn min_sample_count (prior quá mỏng, §1 #12).
func (r *Repo) PriorFor(ctx context.Context, categoryID int64) (CategoryPrior, bool, error) {
    var cp CategoryPrior
    err := r.pool.QueryRow(ctx,
        `SELECT category_id, median_price, typical_discount_depth, sample_count
         FROM category_prior WHERE category_id = $1`, categoryID).
        Scan(&cp.CategoryID, &cp.MedianPrice, &cp.DiscountDepth, &cp.SampleCount)
    if errors.Is(err, pgx.ErrNoRows) {
        return CategoryPrior{}, false, nil
    }
    if err != nil {
        return CategoryPrior{}, false, err
    }
    if cp.SampleCount < minSamples {
        return cp, false, nil // prior mỏng -> coi như không đủ tin
    }
    return cp, true, nil
}
```

---

## §4 - Acceptance criteria

1. `daysOfHistory` của 1 SKU = số ngày từ `first_seen` (hoặc snapshot sớm nhất) tới nay, tính đúng kể cả khi snapshot sớm hơn `first_seen`.
2. `Maturity(d)` trả `NEW` với `d<14`, `WARMING` với `14<=d<90`, `MATURE` với `d>=90`.
3. SKU `NEW` (`d<14`) -> verdict `UNKNOWN` (handoff FR-DEAL-001), không có ngoại lệ.
4. SKU `WARMING` -> verdict gắn cờ `low_confidence` VÀ biểu đồ kèm chú thích `"đang tích lũy"`; không phát verdict đầy đủ.
5. SKU `MATURE` -> cho phép verdict đầy đủ và biểu đồ công khai không cảnh báo dữ liệu mỏng.
6. Materialized view `category_prior` tồn tại, có cột `median_price`, `typical_discount_depth`, `sample_count` theo `category_id`.
7. `category_prior` KHÔNG chứa đóng góp từ SKU `NEW`/`WARMING` (chỉ SKU `>= 90` ngày vào mẫu).
8. `IsFeatureReady(product)` trả `true` chỉ khi SKU `MATURE`; `false` cho `NEW` và `WARMING`.
9. SKU non (`NEW`/`WARMING`) dùng `PriorFor(category_id)` làm fallback; SKU `MATURE` dùng lịch sử riêng, không gọi prior.
10. SKU có `category_id` NULL -> `UNKNOWN`, `PriorFor` trả `ok=false` (không fallback).
11. Sau refresh theo lịch, `category_prior` phản ánh đúng tập SKU `MATURE` hiện tại.
12. `PriorFor` của category có `sample_count < 30` trả `ok=false` (prior mỏng, fallback về UNKNOWN).
13. Metric `deal_maturity_state_total{state}` tăng đúng nhãn; `category_prior_fallback_total` tăng khi dùng prior.

---

## §5 - Kiểm thử (verification)

```go
// services/deal/internal/coldstart/maturity_test.go
func TestMaturity_Boundaries(t *testing.T) {
    cases := []struct {
        days int
        want State
    }{
        {13, StateNew}, {14, StateWarming}, {89, StateWarming}, {90, StateMature},
    }
    for _, c := range cases {
        require.Equal(t, c.want, Maturity(c.days), "days=%d", c.days)
    }
}

func TestIsFeatureReady_Gate90d(t *testing.T) {
    require.False(t, IsFeatureReady(fakeProduct(89)))
    require.True(t, IsFeatureReady(fakeProduct(90)))
    require.True(t, IsFeatureReady(fakeProduct(200)))
}

func TestPrior_ExcludesImmature(t *testing.T) {
    r := setupRepo(t)
    seedProduct(t, r, catID, 120, 100_000) // MATURE, vào mẫu
    seedProduct(t, r, catID, 5, 10_000)    // NEW, PHẢI bị loại
    refreshCategoryPrior(t, r)

    cp, ok, _ := r.PriorFor(ctx, catID)
    require.True(t, ok)
    require.Equal(t, 1, cp.SampleCount)        // chỉ SKU MATURE
    require.Equal(t, int64(100_000), cp.MedianPrice) // SKU NEW không kéo lệch
}

func TestPriorFallback_WhenNew(t *testing.T) {
    r := setupRepo(t)
    seedMatureCategory(t, r, catID, 40) // 40 SKU MATURE -> qua sàn
    refreshCategoryPrior(t, r)

    newp := fakeProductInCategory(5, catID) // NEW
    require.Equal(t, StateNew, Maturity(newp.DaysOfHistory()))
    cp, ok, _ := r.PriorFor(ctx, catID)
    require.True(t, ok) // SKU non mượn prior của category làm fallback
    require.Greater(t, cp.MedianPrice, int64(0))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0001_category_prior.sql` (materialized view + unique index) -> `maturity.go` (State + Maturity + IsFeatureReady) -> `priors.go` (PriorFor + sàn min_sample_count) -> chèn cổng vào `detect.go` (handoff FR-DEAL-001) -> tests. Materialized view tính từ `price_daily` (FR-PRICE-002) và `tracked_product` (FR-PRICE-001); refresh hằng ngày qua scheduler của deal-svc (`REFRESH MATERIALIZED VIEW CONCURRENTLY category_prior`, cần unique index nên đã tạo `idx_category_prior_cat`). Driver dùng `pgx`. Cổng `IsFeatureReady` đặt ngay đầu nhánh surface verdict công khai trong `detect.go`: nếu `false`, trả nhánh "đang tích lũy" thay vì verdict.

---

## §7 - Phụ thuộc

- **FR-DEAL-001** - bộ phát hiện sale ảo cung cấp verdict enum (gồm `UNKNOWN`); FR này nối vào cổng maturity trước khi surface verdict tự tin.
- **FR-PRICE-001** - `tracked_product.first_seen` và `category_id` là input tính `daysOfHistory` và khóa gộp prior.
- **FR-PRICE-002** - `price_daily` (continuous aggregate) và `price_snapshot` là nguồn cho `category_prior`.
- **FR-DEAL-004 (downstream)** - Prophet baseline dùng category priors làm điểm khởi đầu cho SKU cold-start.
- Extension/lib: PostgreSQL `percentile_cont`, materialized view `REFRESH ... CONCURRENTLY`; driver `pgx`.

---

## §8 - Payload ví dụ

### SKU mới 5 ngày -> UNKNOWN + fallback prior

```go
p := loadProduct(ctx, 90112) // first_seen = 5 ngày trước, category_id = 4221
state := coldstart.Maturity(p.DaysOfHistory())   // StateNew
ready := coldstart.IsFeatureReady(p)             // false -> không bật tính năng công khai
verdict := "UNKNOWN"                              // §1 #3, không đoán mò

// fallback: mượn prior của category để ước lượng tạm trong UI "đang tích lũy"
cp, ok, _ := repo.PriorFor(ctx, 4221)
// ok=true nếu category 4221 có >= 30 SKU MATURE; dùng cp.MedianPrice / cp.DiscountDepth
```

### SKU chín 120 ngày -> IsFeatureReady true

```go
p := loadProduct(ctx, 77003) // first_seen = 120 ngày trước
state := coldstart.Maturity(p.DaysOfHistory())   // StateMature
ready := coldstart.IsFeatureReady(p)             // true -> sale ảo + biểu đồ công khai bật
// dùng lịch sử riêng của SKU (median90, đáy gần đây), KHÔNG dùng category_prior
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Ngưỡng `warmingDays`/`matureDays` có nên khác theo category (hàng thời trang đổi giá nhanh hơn điện máy) - cấu hình per-category ở slice sau.
- `min_sample_count` cố định 30 hay co giãn theo độ phân tán giá của category - tinh chỉnh khi có dữ liệu thực.
- Prior theo trục thời gian (mùa double-date, payday) ngoài median tĩnh - tách sang FR-DEAL-004 nơi Prophet xử lý mùa vụ.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| category_id NULL | check trong code | không có prior để mượn | Trả UNKNOWN, không fallback (§1 #10) |
| prior tính từ quá ít mẫu | sàn min_sample_count | prior lệch, ước lượng sai | sample_count < 30 -> ok=false, fallback UNKNOWN |
| SKU non lọt vào prior aggregate | test ExcludesImmature | vòng phản hồi thiên lệch | WHERE first_seen >= 90d trong view |
| Lỗ hổng lịch sử undercount số ngày | so first_seen với snapshot sớm nhất | SKU chín bị coi là non | Lấy mốc sớm hơn giữa first_seen và snapshot đầu |
| Materialized view stale | tuổi refresh / log scheduler | prior trễ vài giờ tới 1 ngày | Refresh hằng ngày; chạy thủ công nếu cần |
| Verdict tự tin lọt qua cho SKU WARMING | test cổng | mất niềm tin (§5.4) | IsFeatureReady chặn ở 90d trước khi surface |
| Ngưỡng 14/90 rải rác dạng số ma thuật | code review | lệch ngưỡng giữa các nơi | Hằng số có tên warmingDays/matureDays |
| REFRESH CONCURRENTLY thiếu unique index | lỗi Postgres | refresh khóa hoặc fail | idx_category_prior_cat tạo cùng migration |

---

## §11 - Ghi chú

- Cold-start là bài toán con-gà-quả-trứng cốt lõi của SănDeal (§5.1): tính năng cần lịch sử, lịch sử cần thời gian. FR này biến nó thành chính sách rõ ràng thay vì hành vi ngầm.
- `UNKNOWN` không phải lỗi mà là câu trả lời trung thực: thà im lặng còn hơn phát verdict sai và mất niềm tin (§5.4).
- Loại SKU non khỏi category_prior là chi tiết dễ sai nhất: quên điều kiện `first_seen >= 90d` thì prior tự nhiễm nhiễu của chính giai đoạn cold-start.
- Cổng `IsFeatureReady` tách "đã có dữ liệu" khỏi "đủ dữ liệu để công khai" - SKU WARMING vẫn tích lũy và hiển thị "đang tích lũy", chỉ không phát verdict công khai.
- Ưu tiên backfill sớm cho top SKU (FR-SCRAPE-002) để chúng qua cổng 90 ngày càng nhanh càng tốt - đây là đòn bẩy rút ngắn thời gian ra mắt tính năng.
- category_prior là cầu nối tới FR-DEAL-004: cùng một prior dùng cho fallback verdict cũng là điểm khởi đầu cho Prophet baseline trên SKU cold-start.

---

*Hết FR-DEAL-002. Status: ready_to_implement (mục tiêu audit 10/10).*
