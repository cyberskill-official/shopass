---
id: FR-PRICE-005
title: "Thuật toán matching canonical_key - chuẩn hóa title đa tầng + key xác định + fuzzy pg_trgm + hàng đợi manual-review cho dedup sản phẩm chéo 3 sàn"
module: PRICE
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-PRICE-001, FR-PRICE-004, FR-PRICE-002, FR-SCRAPE-002]
depends_on: [FR-PRICE-001]
blocks: [FR-PRICE-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (canonical_key chuẩn hóa so sánh chéo sàn)"
  - "docs/... §5.6 (moat đa sàn, khoảng trống so sánh giá chéo 3 sàn), §3.7 (GET /v1/compare)"
source_decisions:
  - "DEC-PRICE-20: chuẩn hóa title đa tầng - fold dấu tiếng Việt + lowercase + bỏ nhiễu marketing/emoji/boilerplate seller, rồi tách brand + model + thuộc tính nổi bật"
  - "DEC-PRICE-21: canonical_key xác định (deterministic) - brand + model chuẩn hóa + hash thuộc tính nổi bật, cho nhóm khớp chính xác trước khi fuzzy"
  - "DEC-PRICE-22: fuzzy match qua pg_trgm similarity + token-set ratio, gộp listing gần-trùng vào cùng canonical_key theo ngưỡng tin cậy"
  - "DEC-PRICE-23: hàng đợi manual-review cho merge confidence thấp - merge sai = giá so sánh sai, KHÔNG bao giờ auto-merge dưới ngưỡng"
  - "DEC-PRICE-24: recompute idempotent khi title đổi - cùng input cho cùng canonical_key, không sinh nhóm rác khi scraper cập nhật title"

language: "Go 1.22 (price-svc); PostgreSQL 16 + pg_trgm; optional Python embedding helper"
service: shopass/services/price/
new_files:
  - services/price/internal/canon/normalize.go
  - services/price/internal/canon/key.go
  - services/price/internal/canon/match.go
  - services/price/internal/canon/review_queue.go
  - services/price/internal/canon/normalize_test.go
  - services/price/internal/canon/match_test.go
  - services/price/migrations/0005_canonical_review.sql
modified_files:
  - services/price/internal/price/product_repo.go   # thêm SetCanonicalKey(productID, key)
allowed_tools:
  - file_read: services/price/**
  - file_write: services/price/**
  - bash: cd services/price && go test ./...
disallowed_tools:
  - auto-merge hai listing khi confidence dưới ngưỡng (vi phạm DEC-PRICE-23, merge sai = giá so sánh sai)
  - so khớp trên title thô chưa fold dấu/chưa bỏ nhiễu (vi phạm DEC-PRICE-20, hỏng token brand/model)
  - sinh canonical_key ngẫu nhiên/không xác định (vi phạm DEC-PRICE-21, recompute đổi key vô cớ)

effort_hours: 8
sub_tasks:
  - "1.5h: normalize.go - fold dấu, lowercase, bỏ emoji/nhiễu marketing/boilerplate seller, gom khoảng trắng, tách brand/model/attr"
  - "1.0h: key.go - CanonicalKey(brand, model, attrs) xác định bằng hash thuộc tính nổi bật"
  - "1.5h: match.go - truy vấn pg_trgm similarity + token-set ratio, tính confidence, quyết định merge/review"
  - "0.5h: review_queue.go - enqueue merge confidence thấp, API duyệt/từ chối"
  - "0.5h: 0005_canonical_review.sql - CREATE EXTENSION pg_trgm + bảng canonical_review_queue + GIN trgm index trên tracked_product(title)"
  - "0.5h: product_repo.go - SetCanonicalKey(ctx, productID, key) ghi ngược idempotent"
  - "1.0h: normalize_test.go - fold dấu, bỏ nhiễu marketing, tách token brand/model"
  - "1.5h: match_test.go - cùng sản phẩm khác sàn gộp, khác sản phẩm không gộp, confidence thấp vào review"
risk_if_skipped: "canonical_key là moat đa sàn của SănDeal (§5.6). Không có thuật toán matching thì so sánh giá chéo 3 sàn (FR-PRICE-004, GET /v1/compare) không có hàng để JOIN - mỗi listing Shopee/TikTok/Lazada là một dòng rời, không ai biết chúng là cùng một sản phẩm vật lý. Khoảng trống so sánh giá chéo sàn mà BeeCost bỏ ngỏ sẽ vẫn bỏ ngỏ. Nếu làm ẩu (auto-merge dưới ngưỡng), hệ thống gộp nhầm hai SKU khác nhau và hiển thị giá so sánh sai - phá vỡ niềm tin, đúng cái bài học hậu-Honey mà SănDeal phải tránh. Thiếu fold dấu thì 'điện thoại' và 'dien thoai' thành hai nhóm, dedup vô dụng cho thị trường VN."
---

## §1 - Mô tả (BCP-14 normative)

Service PRICE **MUST** tính `canonical_key` cho mỗi `tracked_product` để gộp các listing cùng một sản phẩm vật lý trên Shopee/TikTok/Lazada thành một nhóm so sánh giá, ghi key ngược vào cột `tracked_product.canonical_key` (cột + index `idx_tp_canonical` do FR-PRICE-001 định nghĩa). Pipeline gồm chuẩn hóa title, dựng key xác định, fuzzy match và hàng đợi duyệt tay. Hợp đồng:

1. **MUST** chuẩn hóa title đa tầng qua `Normalize(title string) string` (DEC-PRICE-20): fold dấu tiếng Việt (đ->d, ă->a, ê->e...), lowercase, bỏ emoji và ký tự điều khiển, bỏ nhiễu marketing (`[chính hãng]`, `freeship`, `giảm sốc`, `mã giảm`, `hot`, `sale`), bỏ boilerplate seller (tên shop, hashtag, dấu phân tách `|`, `-`, `chính hãng bảo hành`), gom nhiều khoảng trắng thành một.
2. **MUST** từ title đã chuẩn hóa tách ra `brand`, `model` và các thuộc tính nổi bật (`capacity`, `color`, `size`) qua bộ token và từ điển brand; thuộc tính dạng `128gb`, `256 gb`, `xanh`, `size l` được chuẩn hóa về dạng canonical (`128gb`, `256gb`, `blue`, `l`).
3. **MUST** dựng `canonical_key` xác định qua `CanonicalKey(brand, model string, attrs map[string]string) string` (DEC-PRICE-21): ghép `brand` + `model` chuẩn hóa + hash (SHA-256, cắt 12 hex) của tập thuộc tính nổi bật đã sắp xếp. Cùng input **MUST** cho cùng output (không phụ thuộc thứ tự map, không random).
4. **MUST** dùng key xác định ở #3 cho bước nhóm chính xác (exact-ish) TRƯỚC, rồi mới fuzzy: hai listing cùng `canonical_key` thô là cùng nhóm, không cần fuzzy.
5. **MUST** fuzzy match các listing chưa khớp chính xác qua `Match(candidate Candidate) (MatchResult, error)` (DEC-PRICE-22): truy vấn `pg_trgm` similarity trên `tracked_product.title` đã chuẩn hóa kết hợp token-set ratio, lấy các ứng viên có `similarity >= sim_threshold` (mặc định 0,55).
6. **MUST** tính `confidence` (0..1) cho mỗi cặp ghép từ similarity trigram, độ trùng token brand/model và khớp thuộc tính; chỉ gộp tự động khi `confidence >= merge_threshold` (mặc định 0,82).
7. **MUST** đẩy mọi cặp có `low_threshold <= confidence < merge_threshold` (mặc định `low_threshold` = 0,60) vào `canonical_review_queue` để người duyệt quyết định; **MUST NEVER** auto-merge khi `confidence < merge_threshold` (DEC-PRICE-23). Merge sai làm `GET /v1/compare` hiển thị giá của sản phẩm khác.
8. **MUST** bỏ qua (không gộp, không enqueue) các cặp có `confidence < low_threshold` - coi là sản phẩm khác nhau.
9. **MUST** ghi `canonical_key` ngược qua `product_repo.SetCanonicalKey(ctx, productID int64, key string) error`, idempotent: gọi lại với cùng key không đổi gì, gọi với key mới cập nhật đúng một dòng.
10. **MUST** recompute idempotent khi title đổi (DEC-PRICE-24): scraper cập nhật `title` -> chạy lại pipeline; cùng title chuẩn hóa cho cùng `canonical_key`, không sinh nhóm rác hay tách nhóm vô cớ.
11. **MUST** bật extension `pg_trgm` (`CREATE EXTENSION IF NOT EXISTS pg_trgm`) và tạo `GIN` trigram index trên `tracked_product(title)` để truy vấn similarity không quét toàn bảng.
12. **SHOULD** phát OTel metric: `canon_key_computed_total{platform_id}` (counter), `canon_merge_auto_total` (counter), `canon_review_enqueued_total` (counter), `canon_match_duration_ms` (histogram).
13. **MAY** dùng helper Python tính embedding (cosine similarity) cho các cặp khó mà trigram không phân định được; lõi vẫn ở Go, embedding là tín hiệu phụ cộng vào `confidence`, không thay thế ngưỡng.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao chuẩn hóa title đa tầng (DEC-PRICE-20)?** Title sàn TMĐT VN là nhiễu nặng: `[CHÍNH HÃNG] Tai nghe Sony WH-1000XM5 Freeship Giảm Sốc Mã FMCG50 - Shop ABC`. So khớp trên chuỗi thô là vô vọng - cùng một tai nghe, ba sàn viết ba kiểu marketing khác nhau. Fold dấu là bắt buộc riêng cho VN: `điện thoại` và `dien thoai` phải về cùng một dạng, nếu không mỗi cách gõ thành một nhóm. Bỏ nhiễu marketing và boilerplate seller để chỉ còn `brand + model + thuộc tính` - phần thực sự định danh sản phẩm.

**Vì sao key xác định trước, fuzzy sau (DEC-PRICE-21)?** Phần lớn listing của cùng sản phẩm, sau khi chuẩn hóa sạch, cho ra `brand + model` y hệt. Nhóm chính xác bằng key xác định bắt trọn các trường hợp dễ với chi phí gần bằng không (so chuỗi). Chỉ phần đuôi khó - title thiếu model, viết tắt, lệch thuộc tính - mới cần fuzzy đắt hơn. Làm fuzzy cho mọi cặp là lãng phí và làm tăng rủi ro gộp nhầm.

**Vì sao hàng đợi manual-review (DEC-PRICE-23)?** Đây là quyết định bảo vệ niềm tin lớn nhất của FR này. Merge sai không phải lỗi vô hại: gộp nhầm `iPhone 15` với `iPhone 15 Pro` làm `GET /v1/compare` hiển thị giá Pro cho người đang xem bản thường - người dùng mất tiền, mất niềm tin, đúng vết xe Honey mà §5.6 cảnh báo. Vùng confidence "lưng chừng" (0,60-0,82) là vùng nguy hiểm nhất: đủ giống để cám dỗ gộp, đủ khác để có thể sai. Đẩy vùng này cho người duyệt, không bao giờ tự gộp, là cái giá rẻ để tránh hậu quả đắt.

**Vì sao pg_trgm thay vì so bằng đơn thuần?** So bằng (`title_a = title_b`) chỉ bắt được trùng tuyệt đối - gần như không xảy ra giữa ba sàn. `pg_trgm` đo độ tương đồng trigram trực tiếp trong Postgres, có GIN index nên truy vấn `WHERE title % $1` không quét toàn bảng. Giữ phép so độ tương đồng sát dữ liệu (trong DB) nhanh hơn kéo toàn bộ title ra Go rồi so từng cặp, và token-set ratio bù phần trigram yếu khi thứ tự từ đổi.

**Vì sao recompute idempotent (DEC-PRICE-24)?** Scraper cập nhật `title` thường xuyên (seller đổi tên, thêm khuyến mãi). Nếu mỗi lần title đổi sinh ra `canonical_key` khác, nhóm so sánh sẽ vỡ vụn liên tục, biểu đồ so sánh nhảy loạn. Hàm chuẩn hóa và `CanonicalKey` phải thuần (pure): cùng nội dung định danh cho cùng key, bất kể marketing bề mặt đổi thế nào.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/price/migrations/0005_canonical_review.sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- GIN trigram index để truy vấn similarity không quét toàn bảng
CREATE INDEX IF NOT EXISTS idx_tp_title_trgm
  ON tracked_product USING gin (title gin_trgm_ops);

-- Hàng đợi duyệt tay cho merge confidence thấp (DEC-PRICE-23)
CREATE TABLE canonical_review_queue (
  id            BIGSERIAL PRIMARY KEY,
  product_id    BIGINT      NOT NULL REFERENCES tracked_product(id),
  candidate_key TEXT        NOT NULL,              -- canonical_key đề xuất gộp vào
  confidence    REAL        NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
  status        TEXT        NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','approved','rejected')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  decided_at    TIMESTAMPTZ,
  UNIQUE (product_id, candidate_key)
);

CREATE INDEX idx_crq_pending ON canonical_review_queue (status, created_at)
  WHERE status = 'pending';
```

### Normalize (Go)

```go
// services/price/internal/canon/normalize.go
package canon

// Normalize đưa title thô về dạng chuẩn để so khớp (DEC-PRICE-20).
// Thuần: cùng input → cùng output, không phụ thuộc trạng thái ngoài.
func Normalize(title string) string {
    s := strings.ToLower(title)
    s = foldDiacritics(s)            // đ→d, ă→a, ê→e... (riêng cho VN)
    s = emojiRe.ReplaceAllString(s, " ")
    s = marketingNoiseRe.ReplaceAllString(s, " ") // chinh hang, freeship, giam soc, ma giam, hot, sale
    s = sellerBoilerplateRe.ReplaceAllString(s, " ")
    s = nonAlnumRe.ReplaceAllString(s, " ")       // giữ chữ-số, bỏ ký hiệu phân tách
    s = wsRe.ReplaceAllString(s, " ")             // gom khoảng trắng
    return strings.TrimSpace(s)
}

// Attrs tách brand + model + thuộc tính nổi bật từ title đã chuẩn hóa (§1 #2).
type Attrs struct {
    Brand  string
    Model  string
    Salient map[string]string // capacity/color/size đã chuẩn hóa
}

func Extract(normalized string) Attrs { /* token + từ điển brand + regex thuộc tính */ }
```

### CanonicalKey (Go)

```go
// services/price/internal/canon/key.go
package canon

// CanonicalKey dựng key xác định (DEC-PRICE-21): brand + model + hash thuộc tính.
// Cùng input → cùng output (sắp xếp thuộc tính trước khi hash).
func CanonicalKey(brand, model string, attrs map[string]string) string {
    keys := make([]string, 0, len(attrs))
    for k := range attrs {
        keys = append(keys, k)
    }
    sort.Strings(keys) // xác định, không phụ thuộc thứ tự map
    var b strings.Builder
    for _, k := range keys {
        b.WriteString(k)
        b.WriteByte('=')
        b.WriteString(attrs[k])
        b.WriteByte(';')
    }
    sum := sha256.Sum256([]byte(b.String()))
    return brand + ":" + model + ":" + hex.EncodeToString(sum[:])[:12]
}
```

### Match (Go + pg_trgm)

```go
// services/price/internal/canon/match.go
package canon

type Candidate struct {
    ProductID int64
    Title     string
    Attrs     Attrs
}

type MatchResult struct {
    CanonicalKey string
    Confidence   float64
    Action       string // "merge" | "review" | "skip"
}

// Match tìm nhóm cho candidate: key xác định trước, fuzzy pg_trgm sau (§1 #4-#8).
func (m *Matcher) Match(ctx context.Context, c Candidate) (MatchResult, error) {
    norm := Normalize(c.Title)
    rows, err := m.pool.Query(ctx,
        `SELECT id, canonical_key, similarity(title, $1) AS sim
           FROM tracked_product
          WHERE title % $1 AND canonical_key IS NOT NULL
          ORDER BY sim DESC
          LIMIT 10`, norm)
    if err != nil {
        return MatchResult{}, err
    }
    best := m.bestCandidate(rows, c) // gộp sim trigram + token-set + khớp attr
    switch {
    case best.Confidence >= m.mergeThreshold: // 0,82
        return MatchResult{best.Key, best.Confidence, "merge"}, nil
    case best.Confidence >= m.lowThreshold:    // 0,60
        return MatchResult{best.Key, best.Confidence, "review"}, nil
    default:
        return MatchResult{CanonicalKey(c.Attrs.Brand, c.Attrs.Model, c.Attrs.Salient),
            best.Confidence, "skip"}, nil // sản phẩm mới, key riêng
    }
}
```

---

## §4 - Acceptance criteria

1. `Normalize("[CHÍNH HÃNG] Điện Thoại iPhone 15 Freeship ")` trả `"dien thoai iphone 15"` (fold dấu, lowercase, bỏ marketing/emoji, gom khoảng trắng).
2. `Extract` của title điện thoại trả `Brand="apple"` (qua từ điển), `Model="iphone 15"`, `Salient` chứa color/capacity nếu có.
3. `CanonicalKey("apple","iphone 15", {"capacity":"128gb"})` gọi 2 lần (map xáo thứ tự) trả CÙNG một chuỗi.
4. Key xác định nhóm trước: hai listing chuẩn hóa ra cùng `brand+model+attrs` cùng `canonical_key`, không qua fuzzy.
5. `Match` của listing Lazada cùng sản phẩm với một listing Shopee đã có key -> `Action="merge"`, trả đúng `canonical_key` của nhóm.
6. `Match` của sản phẩm khác hẳn (iPhone 15 vs Galaxy S24) -> `Action="skip"`, không gộp.
7. `Match` cặp confidence trong [0,60; 0,82) -> `Action="review"`, thêm 1 dòng `canonical_review_queue` status `pending`.
8. `Match` cặp confidence < 0,60 -> `Action="skip"`, KHÔNG enqueue, KHÔNG gộp.
9. `SetCanonicalKey(pid, key)` gọi lại cùng key -> không đổi; gọi với key mới -> cập nhật đúng 1 dòng (idempotent).
10. Title đổi rồi recompute -> cùng nội dung định danh cho cùng `canonical_key` (không sinh nhóm rác).
11. Migration chạy sạch -> `pg_trgm` bật, `idx_tp_title_trgm` GIN tồn tại, `canonical_review_queue` tồn tại với CHECK `confidence` và CHECK `status`.
12. Truy vấn `WHERE title % $1` dùng GIN trgm index (EXPLAIN không Seq Scan trên bảng lớn).
13. Metric `canon_merge_auto_total` tăng khi `merge`; `canon_review_enqueued_total` tăng khi `review`.

---

## §5 - Kiểm thử (verification)

```go
// services/price/internal/canon/normalize_test.go
func TestNormalize_FoldsDiacritics(t *testing.T) {
    got := Normalize("Điện Thoại Thông Minh")
    require.Equal(t, "dien thoai thong minh", got)
}

func TestNormalize_StripsMarketingNoise(t *testing.T) {
    // \U0001F525 là một emoji thật (lửa) nhúng vào chuỗi để chứng minh Normalize bóc emoji;
    // nguồn giữ ASCII thuần, runtime vẫn nhận đúng rune emoji.
    got := Normalize("[CHÍNH HÃNG] Tai nghe Sony WH-1000XM5 Freeship Giảm Sốc \U0001F525 - Shop ABC")
    require.Equal(t, "tai nghe sony wh 1000xm5", got)
}

func TestCanonicalKey_Deterministic(t *testing.T) {
    a := CanonicalKey("apple", "iphone 15", map[string]string{"capacity": "128gb", "color": "blue"})
    b := CanonicalKey("apple", "iphone 15", map[string]string{"color": "blue", "capacity": "128gb"})
    require.Equal(t, a, b) // thứ tự map không đổi key
}
```

```go
// services/price/internal/canon/match_test.go
func TestMatch_SameProductDifferentPlatform_Merges(t *testing.T) {
    m, key := setupWithProduct(t, "shopee", "Tai nghe Sony WH-1000XM5 Chính Hãng")
    res, _ := m.Match(ctx, Candidate{
        ProductID: 0,
        Title:     "[LAZMALL] Sony WH 1000XM5 Freeship - Headphone",
    })
    require.Equal(t, "merge", res.Action)
    require.Equal(t, key, res.CanonicalKey)
}

func TestMatch_DifferentProduct_DoesNotMerge(t *testing.T) {
    m, _ := setupWithProduct(t, "shopee", "iPhone 15 128GB")
    res, _ := m.Match(ctx, Candidate{Title: "Samsung Galaxy S24 Ultra 256GB"})
    require.Equal(t, "skip", res.Action)
}

func TestMatch_LowConfidence_GoesToReviewQueue(t *testing.T) {
    m, _ := setupWithProduct(t, "shopee", "iPhone 15 128GB")
    res, _ := m.Match(ctx, Candidate{Title: "iPhone 15 Pro 128GB"}) // gần nhưng khác bản
    require.Equal(t, "review", res.Action)
    require.Equal(t, 1, countPending(t, m)) // 1 dòng review, KHÔNG auto-merge
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0005 (pg_trgm + GIN index + review queue) -> `normalize.go` (chuẩn hóa + Extract) -> `key.go` (CanonicalKey xác định) -> `match.go` (pg_trgm + confidence + quyết định) -> `review_queue.go` (enqueue + duyệt) -> `product_repo.go::SetCanonicalKey` -> tests. Từ điển brand và regex nhiễu marketing để ở file dữ liệu cùng package, mở rộng theo dữ liệu thực mà không sửa logic. Embedding helper Python (nếu bật) chạy ngoài tiến trình, trả điểm cosine qua một endpoint nội bộ, cộng vào `confidence` như tín hiệu phụ.

---

## §7 - Phụ thuộc

- **FR-PRICE-001** - `tracked_product(title, canonical_key)` và `idx_tp_canonical` phải tồn tại trước (ghi key ngược).
- **FR-PRICE-004 (downstream)** - `GET /v1/compare?canonical_key=...` JOIN theo `canonical_key` mà FR này sinh ra.
- **FR-SCRAPE-002 (tùy chọn)** - `recommend` của Shopee cấp SKU liên quan, làm giàu ứng viên matching.
- **FR-PRICE-002 (liên quan)** - sau khi gộp, mỗi `canonical_key` đọc `price_snapshot` của các `product_id` thành viên để dựng dòng so sánh.
- Extension/lib: PostgreSQL `pg_trgm` (similarity, GIN index); driver `pgx`; tùy chọn helper embedding Python.

---

## §8 - Payload ví dụ

### Hai title cùng sản phẩm, chuẩn hóa về cùng canonical_key

```text
Shopee : "[CHÍNH HÃNG] Tai nghe Sony WH-1000XM5 Freeship (emoji) Mã FMCG50 - Shop ABC"
Lazada : "Sony WH 1000XM5 Headphone Chống Ồn | Bảo Hành 12 Tháng [LAZMALL]"

Normalize → "tai nghe sony wh 1000xm5"   (cả hai)
Extract   → brand="sony", model="wh 1000xm5", salient={}
CanonicalKey("sony","wh 1000xm5",{}) → "sony:wh 1000xm5:9f2a3c1d0b7e"  (cả hai)
```

### JOIN nuôi GET /v1/compare

```sql
-- FR-PRICE-004 dùng canonical_key để gom giá 3 sàn
SELECT p.code AS platform, tp.title, d.close_p AS price
FROM tracked_product tp
JOIN platform p          ON p.id = tp.platform_id
JOIN LATERAL (
  SELECT close_p FROM price_daily
  WHERE product_id = tp.id ORDER BY day DESC LIMIT 1
) d ON true
WHERE tp.canonical_key = 'sony:wh 1000xm5:9f2a3c1d0b7e';
-- → 3 dòng (shopee / tiktok / lazada) cho cùng một sản phẩm vật lý
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Embedding cosine cho cặp khó (helper Python) - bật khi tỷ lệ review queue cao hơn ngưỡng vận hành.
- Học ngưỡng `merge_threshold`/`low_threshold` từ nhãn duyệt tay tích lũy - tinh chỉnh giai đoạn sau, ban đầu hằng số.
- Gộp theo biến thể (variant) trong cùng listing (màu/dung lượng) - slice sau khi mô hình variant của FR-PRICE-001 ổn định.
- Từ điển brand đa ngôn ngữ khi mở SEA (TH/ID) - mở rộng theo nước, fold dấu theo ngôn ngữ đích.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Over-merge: gộp hai SKU khác (iPhone 15 vs 15 Pro) | mẫu review + report giá lệch | `GET /v1/compare` hiển thị giá sai, mất niềm tin | Ngưỡng `merge_threshold` cao + review queue cho vùng lưng chừng (DEC-PRICE-23) |
| Under-merge: cùng SKU không gộp | tỷ lệ key trùng thấp | so sánh chéo sàn thiếu dòng, moat yếu | Hạ nhẹ `sim_threshold`, làm giàu từ điển brand/đồng nghĩa |
| Fold dấu phá token brand (`Samsung`->`samsung` ok; nhãn có dấu hiếm) | test fold + spot-check | brand sai -> key sai | Từ điển brand chuẩn hóa sau fold; giữ token số model nguyên |
| pg_trgm chậm trên bảng triệu dòng | p95 `canon_match_duration_ms` | matching đọng, recompute trễ | GIN trgm index (§3); giới hạn LIMIT ứng viên; chạy theo lô |
| Recompute không idempotent (key đổi vô cớ) | test recompute | nhóm so sánh vỡ vụn, biểu đồ nhảy | `Normalize`/`CanonicalKey` thuần; hash thuộc tính đã sắp xếp (DEC-PRICE-24) |
| Nhiễu marketing mới (seller dùng từ lạ) | tỷ lệ token rác tăng | chuẩn hóa sót, similarity giảm | Bổ sung regex nhiễu trong file dữ liệu, không sửa logic |
| Review queue tồn đọng | `canon_review_enqueued_total` tăng, pending lớn | merge chậm, dòng so sánh thiếu | Bổ sung người duyệt; cân nhắc bật embedding để giảm vùng lưng chừng |
| Title rỗng/chỉ emoji | `Normalize` trả chuỗi rỗng | không tách được brand/model | Bỏ qua matching, giữ `canonical_key` NULL, gắn cờ cần scraper lấy lại title |
| Trùng `(product_id, candidate_key)` khi enqueue lặp | UNIQUE constraint | từ chối ghi trùng | `ON CONFLICT DO NOTHING` khi enqueue (idempotent) |

---

## §11 - Ghi chú

- `canonical_key` là moat đa sàn của SănDeal (§5.6): không có nó, so sánh giá chéo 3 sàn - khoảng trống mà BeeCost bỏ ngỏ - không thực hiện được.
- Triết lý thiết kế: thà under-merge (thiếu một dòng so sánh) còn hơn over-merge (hiển thị giá sai). Niềm tin hậu-Honey là tài sản dễ mất, khó lấy lại.
- Key xác định trước, fuzzy sau: bắt trọn phần dễ với chi phí gần không, chỉ tốn fuzzy cho phần đuôi khó.
- Vùng confidence lưng chừng (0,60-0,82) luôn vào người duyệt, không bao giờ tự gộp - đây là van an toàn chính của FR.
- Fold dấu là yêu cầu riêng cho thị trường VN: bỏ qua thì dedup vô dụng vì người Việt gõ có dấu và không dấu lẫn lộn.
- Hàm chuẩn hóa và `CanonicalKey` thuần (pure) bảo đảm recompute idempotent khi scraper cập nhật title - nhóm so sánh ổn định theo thời gian.

---

*Hết FR-PRICE-005. Status: ready_to_implement (mục tiêu audit 10/10).*
