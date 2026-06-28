---
id: FR-PRICE-004
title: "GET /v1/compare?canonical_key=... - so sánh giá hiện tại cùng một sản phẩm trên 3 sàn (Shopee/TikTok/Lazada) + đánh dấu rẻ nhất phía server cho moat đa sàn"
module: PRICE
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-PRICE-005, FR-PRICE-001, FR-PRICE-002, FR-PRICE-003]
depends_on: [FR-PRICE-005]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.7 (GET /v1/compare?canonical_key=...)"
  - "docs/... §5.6 (moat đa sàn, so sánh giá chéo 3 sàn), §6.1 (catalog tính năng so sánh chéo sàn)"
source_decisions:
  - "DEC-PRICE-40: compare khóa theo canonical_key (output của FR-PRICE-005); endpoint không tự so khớp, chỉ JOIN theo key đã gán"
  - "DEC-PRICE-41: giá hiện tại = snapshot mới nhất mỗi product qua DISTINCT ON (product_id) ... ORDER BY product_id, ts DESC, không quét price_daily (cần ts giây/phút cho độ tươi)"
  - "DEC-PRICE-42: tính rẻ nhất phía server và set cờ is_cheapest, không để client tự so (tránh lệch logic giữa web/app)"
  - "DEC-PRICE-43: trả ts mỗi sàn trong payload để UI hiển thị độ tươi của từng giá (minh bạch hậu-Honey §5.4)"
  - "DEC-PRICE-44: canonical_key chỉ có 1 sàn vẫn trả về dòng đó (không ép đủ 3 sàn); 1 dòng thì không cần badge rẻ nhất ngoài chính nó"

language: "Go 1.22 (price-svc); PostgreSQL 16 + TimescaleDB 2.x"
service: shopass/services/price/
new_files:
  - services/price/internal/api/compare.go
  - services/price/internal/api/compare_test.go
  - services/price/internal/price/compare_query.go
modified_files:
  - services/price/internal/api/router.go            # đăng ký route GET /v1/compare
allowed_tools:
  - file_read: services/price/**
  - file_write: services/price/**
  - bash: cd services/price && go test ./...
disallowed_tools:
  - tính rẻ nhất phía client thay vì server (vi phạm DEC-PRICE-42, lệch logic web/app)
  - lấy giá từ price_daily thay vì snapshot mới nhất (vi phạm DEC-PRICE-41, mất ts độ tươi)
  - ép phải đủ 3 sàn mới trả về (vi phạm DEC-PRICE-44, bỏ rơi sản phẩm 1-2 sàn)

effort_hours: 6
sub_tasks:
  - "0.5h: router.go - đăng ký GET /v1/compare sau JWT middleware của gateway"
  - "1.5h: compare_query.go - DISTINCT ON (product_id) latest snapshot, JOIN tracked_product theo canonical_key + JOIN platform"
  - "1.0h: compare.go - parse + validate param canonical_key (400 nếu thiếu), tính is_cheapest, dựng response"
  - "0.5h: compare.go - xử lý unknown/empty key -> 404, single-platform -> trả 1 dòng"
  - "2.0h: compare_test.go - 5 test (3 sàn highlight rẻ nhất, thiếu key 400, key lạ rỗng, 1 sàn, hình dạng payload + ts)"
  - "0.5h: OTel metric compare_request_total{result} + compare_query_duration_ms (p95 < 500ms)"

risk_if_skipped: "GET /v1/compare là tính năng lõi của moat đa sàn SănDeal (§5.6): cho một canonical_key, trả giá hiện tại cùng một sản phẩm vật lý trên cả 3 sàn và chỉ ra sàn rẻ nhất. Thiếu nó thì toàn bộ công sức so khớp chéo sàn của FR-PRICE-005 không có đầu ra cho người dùng - dữ liệu đã gom nhóm nhưng không ai đọc được. Khoảng trống so sánh giá chéo 3 sàn mà §5.6 chỉ ra vẫn bỏ ngỏ. Nếu tính rẻ nhất phía client, web và app dễ lệch logic và hiển thị badge khác nhau cho cùng dữ liệu. Nếu lấy giá từ price_daily thay vì snapshot mới nhất, mất ts giây/phút nên không thể hiển thị độ tươi - đúng cái minh bạch hậu-Honey mà SănDeal cam kết. Endpoint này cũng là nền cho UI so sánh và gói B2B sau này."
---

## §1 - Mô tả (BCP-14 normative)

Service PRICE **MUST** expose REST endpoint `GET /v1/compare?canonical_key=...`: cho một `canonical_key` (do FR-PRICE-005 sinh và lưu trên `tracked_product`), trả giá hiện tại của cùng một sản phẩm vật lý trên từng sàn (Shopee/TikTok/Lazada), kèm cờ rẻ nhất. Hợp đồng:

1. **MUST** đăng ký route `GET /v1/compare` sau JWT middleware của API gateway; chỉ request có token hợp lệ mới qua (auth tập trung ở gateway, handler không tự xác thực).
2. **MUST** yêu cầu query param `canonical_key`; nếu thiếu hoặc rỗng -> trả `400 Bad Request` với thân lỗi rõ, KHÔNG truy vấn DB.
3. **MUST** JOIN `tracked_product` lọc theo `canonical_key` -> snapshot giá mới nhất mỗi `product_id` -> `platform`, dựng đúng một dòng cho mỗi sàn có listing thuộc key đó (DEC-PRICE-40).
4. **MUST** lấy giá hiện tại của mỗi `product_id` qua `DISTINCT ON (product_id) ... ORDER BY product_id, ts DESC` trên `price_snapshot` (DEC-PRICE-41); KHÔNG dùng `price_daily` (cần `ts` giây/phút cho độ tươi).
5. **MUST** trả mỗi dòng gồm `{platform_code, platform_name, product_id, price, ts, item_url}`. `platform_code` lấy từ `platform.code`; `platform_name` suy ra từ `code` phía server (`displayName`: shopee->"Shopee", tiktok->"TikTok Shop", lazada->"Lazada") vì bảng `platform` (FR-INFRA-002) KHÔNG có cột `name`. `item_url` lấy từ `platform_item_id` của `tracked_product`, hoặc tham chiếu để client dựng link.
6. **MUST** tính sàn rẻ nhất phía server và set `is_cheapest = true` đúng cho dòng có `price` nhỏ nhất (DEC-PRICE-42); các dòng còn lại `false`. Nếu có nhiều dòng cùng giá nhỏ nhất, set `is_cheapest` cho mọi dòng bằng giá đó.
7. **MUST** lưu và trả `price` dạng `BIGINT` (VND, `int64`, không thập phân) - thống nhất với FR-PRICE-002, tránh sai số float.
8. **MUST** trả `ts` (TIMESTAMPTZ) của từng dòng để UI hiển thị độ tươi của giá mỗi sàn (DEC-PRICE-43); giá có thể cũ khác nhau giữa các sàn.
9. **MUST** xử lý `canonical_key` chỉ có một sàn: vẫn trả về dòng đó với `is_cheapest = true` (chỉ một lựa chọn), KHÔNG ép phải đủ 3 sàn (DEC-PRICE-44).
10. **MUST** xử lý `canonical_key` lạ hoặc không có listing nào -> trả `404 Not Found` (chọn 404, không trả mảng rỗng `200`, để client phân biệt rõ "không có sản phẩm" với "có nhưng rỗng").
11. **MUST** đạt `p95 < 500ms` cho truy vấn compare (dựa trên index `idx_tp_canonical` của FR-PRICE-001 + index `ts` của hypertable).
12. **SHOULD** phát OTel metric `compare_request_total{result}` (`result - {ok, not_found, bad_request}`) và `compare_query_duration_ms` (histogram) để theo dõi p95.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao khóa theo canonical_key (DEC-PRICE-40, depends_on FR-PRICE-005)?** So khớp "cùng một sản phẩm vật lý trên 3 sàn" là bài toán khó và đã được giải ở FR-PRICE-005 (chuẩn hóa title, key xác định, fuzzy, duyệt tay). Endpoint compare không lặp lại việc đó - nó chỉ tin `canonical_key` đã gán và JOIN theo. Tách rõ "so khớp" khỏi "đọc so sánh" giữ handler mỏng, nhanh, và mọi rủi ro gộp nhầm nằm gọn ở một nơi (FR-PRICE-005).

**Vì sao DISTINCT ON latest-per-product thay vì price_daily (DEC-PRICE-41)?** So sánh giá cần giá hiện tại đúng tại thời điểm hỏi, không phải giá đóng cửa của ngày. `price_daily` chỉ có `close_p` theo ngày, mất độ phân giải giây/phút và quan trọng hơn là mất `ts` chính xác. `DISTINCT ON (product_id) ORDER BY product_id, ts DESC` lấy đúng snapshot mới nhất mỗi sản phẩm kèm `ts` thật, nên UI biết giá này tươi đến đâu. Index `(product_id, ts DESC)` của hypertable làm phép này nhanh.

**Vì sao tính rẻ nhất phía server (DEC-PRICE-42)?** Nếu để web và app mỗi nơi tự so giá, hai client dễ lệch nhau khi xử lý giá bằng nhau, giá thiếu, hay làm tròn. Tính một lần ở server và trả cờ `is_cheapest` cho mọi client đọc giống nhau - một nguồn sự thật, không lệch badge giữa nền tảng.

**Vì sao trả ts mỗi sàn (DEC-PRICE-43, minh bạch §5.4)?** SănDeal định vị quanh niềm tin hậu-Honey: không giấu, không thổi phồng. Một giá Lazada quét cách đây 2 giờ và một giá Shopee quét 5 phút trước không nên trông ngang nhau. Trả `ts` để UI nói thẳng "giá cập nhật lúc...", người dùng tự đánh giá độ tin. Giấu độ tươi là mầm mất niềm tin; phơi nó ra là rẻ và đúng.

**Vì sao 1 sàn vẫn trả, không ép đủ 3 (DEC-PRICE-44)?** Nhiều sản phẩm chỉ bán trên một hoặc hai sàn. Ép phải đủ 3 sàn mới trả về sẽ bỏ rơi phần lớn long-tail. Trả những gì có (dù chỉ một dòng) vẫn hữu ích: người dùng thấy giá hiện có và biết sàn đó là rẻ nhất theo nghĩa "lựa chọn duy nhất".

**Vì sao key lạ trả 404 chứ không phải mảng rỗng (§1 #10)?** Mảng rỗng `200` mơ hồ: client không phân biệt được "canonical_key này không tồn tại" với "tồn tại nhưng tạm không có giá". `404` nói rõ không có gì để so sánh, để UI hiện trạng thái đúng (ví dụ "chưa theo dõi sản phẩm này") thay vì khung so sánh trống.

---

## §3 - Hợp đồng API / DDL

### Truy vấn (Go + SQL)

```go
// services/price/internal/price/compare_query.go
package price

// CompareRow là một dòng giá hiện tại của một sàn cho một canonical_key.
// platform chỉ có cột `code` (FR-INFRA-002 sở hữu: id/code/country/base_url) - KHÔNG có cột name;
// tên hiển thị suy ra từ code phía server (displayName), không SELECT cột không tồn tại.
type CompareRow struct {
    PlatformCode string    `db:"platform_code"`
    ProductID    int64     `db:"product_id"`
    Price        int64     `db:"price"`        // VND, không thập phân
    TS           time.Time `db:"ts"`           // độ tươi của giá sàn này
    PlatformItem string    `db:"platform_item_id"`
}

// CompareByCanonicalKey: snapshot mới nhất mỗi product (DISTINCT ON),
// JOIN tracked_product lọc theo canonical_key + JOIN platform (chỉ lấy code).
func (r *Repo) CompareByCanonicalKey(ctx context.Context, key string) ([]CompareRow, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT pf.code AS platform_code,
               tp.id   AS product_id,
               ls.price,
               ls.ts,
               tp.platform_item_id
        FROM tracked_product tp
        JOIN platform pf ON pf.id = tp.platform_id
        JOIN LATERAL (
            SELECT DISTINCT ON (ps.product_id) ps.price, ps.ts
            FROM price_snapshot ps
            WHERE ps.product_id = tp.id
            ORDER BY ps.product_id, ps.ts DESC
        ) ls ON true
        WHERE tp.canonical_key = $1
        ORDER BY ls.price ASC`, key)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    return scanCompareRows(rows)
}

// displayName suy tên hiển thị từ platform.code (không có cột name trong schema §3.4).
func displayName(code string) string {
    switch code {
    case "shopee":
        return "Shopee"
    case "tiktok":
        return "TikTok Shop"
    case "lazada":
        return "Lazada"
    default:
        return code
    }
}
```

### Response structs (Go)

```go
// services/price/internal/api/compare.go
package api

type CompareItem struct {
    PlatformCode string `json:"platform_code"`
    PlatformName string `json:"platform_name"`
    ProductID    int64  `json:"product_id"`
    Price        int64  `json:"price"`        // VND
    Currency     string `json:"currency"`     // "VND"
    TS           string `json:"ts"`           // RFC3339, độ tươi
    ItemURL      string `json:"item_url"`
    IsCheapest   bool   `json:"is_cheapest"`
}

type CompareResponse struct {
    CanonicalKey string        `json:"canonical_key"`
    Items        []CompareItem `json:"items"`
}
```

### Handler (Go)

```go
// services/price/internal/api/compare.go
func (h *Handler) Compare(w http.ResponseWriter, r *http.Request) {
    key := strings.TrimSpace(r.URL.Query().Get("canonical_key"))
    if key == "" {
        writeErr(w, http.StatusBadRequest, "canonical_key là bắt buộc")
        metrics.Compare("bad_request")
        return
    }
    rows, err := h.repo.CompareByCanonicalKey(r.Context(), key)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "lỗi truy vấn")
        return
    }
    if len(rows) == 0 {
        writeErr(w, http.StatusNotFound, "không có sản phẩm cho canonical_key này")
        metrics.Compare("not_found")
        return
    }
    items := toItems(rows)
    markCheapest(items) // set is_cheapest phía server (DEC-PRICE-42)
    writeJSON(w, http.StatusOK, CompareResponse{CanonicalKey: key, Items: items})
    metrics.Compare("ok")
}

// markCheapest đánh dấu mọi dòng có price = min (xử lý cả trường hợp bằng giá).
func markCheapest(items []CompareItem) {
    if len(items) == 0 {
        return
    }
    min := items[0].Price
    for _, it := range items {
        if it.Price < min {
            min = it.Price
        }
    }
    for i := range items {
        items[i].IsCheapest = items[i].Price == min
    }
}
```

---

## §4 - Acceptance criteria

1. `GET /v1/compare` đăng ký sau JWT middleware của gateway; request không token bị gateway chặn (handler không thấy).
2. `GET /v1/compare` thiếu `canonical_key` (hoặc rỗng) -> `400`, thân lỗi rõ, KHÔNG truy vấn DB.
3. `canonical_key` có listing 3 sàn -> trả 3 dòng, mỗi sàn một dòng (JOIN tracked_product theo key + platform).
4. Mỗi dòng lấy giá từ snapshot mới nhất của product đó (DISTINCT ON product_id, ts DESC), KHÔNG từ price_daily.
5. Mỗi dòng có đủ `{platform_code, platform_name, product_id, price, ts, item_url}`.
6. Dòng giá nhỏ nhất có `is_cheapest=true`; các dòng khác `false`; nhiều dòng bằng giá nhỏ nhất -> tất cả `true`.
7. `price` là số nguyên VND (int64); JSON không có phần thập phân.
8. Mỗi dòng trả `ts` (RFC3339) phản ánh thời điểm snapshot mới nhất của sàn đó.
9. `canonical_key` chỉ có 1 sàn -> trả 1 dòng, `is_cheapest=true`, không lỗi.
10. `canonical_key` lạ / không có listing -> `404` (không phải mảng rỗng `200`).
11. p95 truy vấn compare `< 500ms` trên dữ liệu thực (đo qua `compare_query_duration_ms`).
12. Metric `compare_request_total{result}` tăng đúng nhánh (`ok` / `not_found` / `bad_request`).

---

## §5 - Kiểm thử (verification)

```go
// services/price/internal/api/compare_test.go
func TestCompare_ThreePlatforms_HighlightsCheapest(t *testing.T) {
    h, key := setupThreePlatforms(t) // shopee 89k, tiktok 92k, lazada 85k cùng canonical_key
    rec := doGET(t, h, "/v1/compare?canonical_key="+key)
    require.Equal(t, http.StatusOK, rec.Code)
    var resp CompareResponse
    json.Unmarshal(rec.Body.Bytes(), &resp)
    require.Len(t, resp.Items, 3)
    cheapest := pickCheapest(resp.Items)
    require.Equal(t, "lazada", cheapest.PlatformCode) // 85k rẻ nhất
    require.True(t, cheapest.IsCheapest)
    require.Equal(t, 1, countCheapest(resp.Items)) // đúng một dòng được đánh dấu
}

func TestCompare_MissingKey_400(t *testing.T) {
    h, _ := setupThreePlatforms(t)
    rec := doGET(t, h, "/v1/compare") // thiếu canonical_key
    require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCompare_UnknownKey_Empty(t *testing.T) {
    h, _ := setupThreePlatforms(t)
    rec := doGET(t, h, "/v1/compare?canonical_key=khong-ton-tai")
    require.Equal(t, http.StatusNotFound, rec.Code) // 404, không phải mảng rỗng
}

func TestCompare_SinglePlatform(t *testing.T) {
    h, key := setupSinglePlatform(t, "shopee", 120_000) // chỉ 1 sàn
    rec := doGET(t, h, "/v1/compare?canonical_key="+key)
    require.Equal(t, http.StatusOK, rec.Code)
    var resp CompareResponse
    json.Unmarshal(rec.Body.Bytes(), &resp)
    require.Len(t, resp.Items, 1)
    require.True(t, resp.Items[0].IsCheapest) // lựa chọn duy nhất
}

func TestCompare_PayloadShape(t *testing.T) {
    h, key := setupThreePlatforms(t)
    rec := doGET(t, h, "/v1/compare?canonical_key="+key)
    var resp CompareResponse
    json.Unmarshal(rec.Body.Bytes(), &resp)
    it := resp.Items[0]
    require.NotEmpty(t, it.PlatformCode)
    require.NotEmpty(t, it.PlatformName)
    require.Greater(t, it.Price, int64(0))     // int64 VND
    require.NotEmpty(t, it.TS)                  // ts độ tươi luôn có
    require.Equal(t, "VND", it.Currency)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `compare_query.go` (DISTINCT ON latest + JOIN platform lấy `code`) -> `compare.go` (parse + validate + `toItems` + `markCheapest` + dựng response) -> `router.go` (đăng ký `GET /v1/compare` sau JWT middleware) -> tests. Handler dùng driver `pgx`, không tự xác thực (gateway lo JWT). `toItems` đặt `PlatformName = displayName(row.PlatformCode)` (bảng `platform` không có cột `name`). Truy vấn xếp `ORDER BY ls.price ASC` để dòng đầu là rẻ nhất; `markCheapest` vẫn quét lại để xử lý đúng trường hợp bằng giá. Item URL dựng từ `platform_item_id` + `platform_code` theo bảng quy ước link sàn, hoặc trả thẳng tham chiếu cho client ghép.

---

## §7 - Phụ thuộc

- **FR-PRICE-005** - sinh và gán `canonical_key` cho `tracked_product`; không có key thì compare không có hàng để JOIN.
- **FR-PRICE-001** - `tracked_product(canonical_key, platform_id, platform_item_id, title)` + index `idx_tp_canonical`.
- **FR-INFRA-002** - bảng `platform(id, code, country, base_url)` (KHÔNG có cột `name`); compare JOIN lấy `code`, suy `platform_name` từ `code` phía server.
- **FR-PRICE-002** - `price_snapshot(product_id, ts, price)` là nguồn giá hiện tại (snapshot mới nhất).
- **FR-PRICE-003 (liên quan)** - chuẩn hóa truy vấn lịch sử/giá hiện tại dùng chung tầng repo.
- Lib: driver `pgx`; TimescaleDB 2.x (index `ts` của hypertable cho latest-per-product nhanh).

---

## §8 - Payload ví dụ

### Request

```bash
curl -s -H "Authorization: Bearer $JWT" \
  "https://api.sandeal.vn/v1/compare?canonical_key=sony:wh%201000xm5:9f2a3c1d0b7e"
```

### Response (200, 3 sàn, Lazada rẻ nhất)

```json
{
  "canonical_key": "sony:wh 1000xm5:9f2a3c1d0b7e",
  "items": [
    {
      "platform_code": "lazada",
      "platform_name": "Lazada",
      "product_id": 90233,
      "price": 5990000,
      "currency": "VND",
      "ts": "2026-06-27T08:55:12Z",
      "item_url": "https://www.lazada.vn/products/i90233.html",
      "is_cheapest": true
    },
    {
      "platform_code": "shopee",
      "platform_name": "Shopee",
      "product_id": 90112,
      "price": 6190000,
      "currency": "VND",
      "ts": "2026-06-27T08:59:40Z",
      "item_url": "https://shopee.vn/product/i90112",
      "is_cheapest": false
    },
    {
      "platform_code": "tiktok",
      "platform_name": "TikTok Shop",
      "product_id": 90150,
      "price": 6290000,
      "currency": "VND",
      "ts": "2026-06-27T07:12:03Z",
      "item_url": "https://shop.tiktok.com/view/i90150",
      "is_cheapest": false
    }
  ]
}
```

`ts` của TikTok cũ hơn (07:12) so với Shopee/Lazada (08:55-08:59), UI nên nói rõ độ tươi từng giá.

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Ngưỡng "giá quá cũ" để gắn cờ stale (ví dụ `ts` quá 6 giờ) - thêm trường `is_stale` ở slice sau, dữ liệu `ts` đã đủ để UI tự quyết tạm thời.
- Trả thêm `list_price`/`flash_sale` mỗi dòng để hiện mức giảm - mở rộng response khi UI so sánh cần, không đổi khóa truy vấn.
- Cache kết quả compare theo `canonical_key` (TTL ngắn) - bật khi tải tăng; ban đầu đọc thẳng để luôn tươi.
- Trả thêm phí ship ước tính mỗi sàn để "rẻ nhất" tính cả ship - cần nguồn phí ship, để giai đoạn sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| canonical_key sai do over-merge (FR-PRICE-005) | report giá lệch + mẫu duyệt | so sánh nhầm sản phẩm khác (giá Pro cho bản thường) | Ngưỡng merge cao + review queue ở FR-PRICE-005 (DEC-PRICE-23) |
| Thiếu param canonical_key | guard handler | 400 trước khi chạm DB | Validate sớm, thân lỗi rõ (§1 #2) |
| canonical_key lạ / chưa gán | rows rỗng | 404 | Trả 404 phân biệt với rỗng (§1 #10) |
| Giá 1 sàn cũ/stale | so ts mỗi dòng | "rẻ nhất" dựa giá cũ | Trả ts để UI cảnh báo độ tươi (DEC-PRICE-43); cờ stale ở slice sau |
| Thiếu 1-2 sàn (sản phẩm long-tail) | đếm items < 3 | so sánh không đủ 3 dòng | Vẫn trả phần có, không ép đủ 3 (DEC-PRICE-44) |
| Lấy nhầm price_daily thay snapshot | review + mất ts | giá đóng cửa ngày, không tươi | Dùng DISTINCT ON price_snapshot (DEC-PRICE-41); AC #4 |
| Hai dòng bằng giá nhỏ nhất | test bằng giá | chỉ một dòng được badge | markCheapest set mọi dòng = min (§1 #6) |
| Truy vấn chậm khi key nhiều listing | p95 compare_query_duration_ms | vỡ NFR <500ms | Index idx_tp_canonical + ts hypertable; LIMIT/giám sát |
| Tính rẻ nhất phía client lệch web/app | so badge giữa nền tảng | hiển thị khác nhau cùng dữ liệu | Tính server, trả is_cheapest (DEC-PRICE-42) |

---

## §11 - Ghi chú

- `GET /v1/compare` là đầu ra người dùng của moat đa sàn: FR-PRICE-005 gom nhóm, FR-PRICE-004 đọc nhóm đó thành so sánh giá 3 sàn.
- Khóa theo `canonical_key` giữ handler mỏng: mọi rủi ro gộp nhầm nằm ở FR-PRICE-005, compare chỉ tin và JOIN.
- `DISTINCT ON (product_id) ... ts DESC` trả giá mới nhất kèm `ts` thật - khác `price_daily` chỉ có giá đóng cửa ngày, mất độ tươi.
- Tính `is_cheapest` phía server là một nguồn sự thật cho mọi client; trả `ts` mỗi sàn là cam kết minh bạch hậu-Honey (§5.4).
- Trả về cả khi chỉ một sàn (không ép đủ 3) phục vụ long-tail; key lạ trả 404 để client phân biệt rõ trạng thái.
- Endpoint này underpin UI so sánh và là viên gạch cho gói dữ liệu B2B sau (xuất giá chéo sàn theo canonical_key).

---

*Hết FR-PRICE-004. Status: ready_to_implement (mục tiêu audit 10/10).*
