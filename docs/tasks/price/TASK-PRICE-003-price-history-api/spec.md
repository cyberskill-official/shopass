---
id: TASK-PRICE-003
title: "GET /v1/products/{id}/price-history - REST trả time-series giá một SKU: đọc price_daily cho phần thân + stitch raw tail từ price_snapshot, p95 <500ms cho biểu đồ"
module: PRICE
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-PRICE-002, TASK-WEB-003, TASK-DEAL-003, TASK-INFRA-001]
depends_on: [TASK-PRICE-002]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.7 (GET /v1/products/{id}/price-history?range=90d)"
  - "docs/... §3.8 (NFR biểu đồ <500ms)"
source_decisions:
  - "DEC-PRICE-30: đọc continuous aggregate price_daily cho phần thân của khoảng, stitch raw price_snapshot cho cái đuôi kể từ bucket ngày gần nhất - biểu đồ tươi mà không chờ cagg refresh hằng giờ"
  - "DEC-PRICE-31: range chỉ nhận allowlist 7d/30d/90d/180d/1y, default 90d, range lạ trả 400 - chặn quét raw không giới hạn vỡ NFR"
  - "DEC-PRICE-32: đảm bảo p95 <500ms bằng cách chỉ chạm cagg cho phần thân và tối đa một chunk raw cho cái đuôi (đuôi luôn nằm trong chunk nóng 7 ngày)"
  - "DEC-PRICE-33: close_p (last(price) trong ngày) là giá hiển thị của điểm ngày; min_p/max_p chỉ để vẽ dải biên"
  - "DEC-PRICE-34: product_id không tồn tại trong tracked_product trả 404, phân biệt rõ với 200 + chuỗi rỗng (SKU có thật nhưng chưa có snapshot)"

language: "Go 1.22 (price-svc); PostgreSQL 16 + TimescaleDB 2.x"
service: shopass/services/price/
new_files:
  - services/price/internal/api/price_history.go
  - services/price/internal/api/price_history_test.go
  - services/price/internal/price/history_query.go
modified_files:
  - services/price/internal/api/router.go            # đăng ký route GET /v1/products/{id}/price-history
allowed_tools:
  - file_read: services/price/**
  - file_write: services/price/**
  - bash: cd services/price && go test ./...
disallowed_tools:
  - quét raw price_snapshot cho toàn khoảng (vi phạm DEC-PRICE-30/32, vỡ p95 <500ms)
  - nhận range tùy ý không qua allowlist (vi phạm DEC-PRICE-31, mở cửa quét raw không giới hạn)
  - trả price dạng float/string trong JSON (price là BIGINT VND, phải là int64)

effort_hours: 5
sub_tasks:
  - "0.5h: history_query.go - parse + validate range qua allowlist, map sang khoảng thời gian + mốc cắt đuôi"
  - "1.0h: history_query.go - QueryDailyBody (đọc price_daily) + QueryRawTail (đọc price_snapshot từ bucket ngày gần nhất)"
  - "1.0h: price_history.go - HTTP handler: parse id, gọi 2 truy vấn, ghép DTO, marshal JSON, 400/404 đúng chỗ"
  - "0.5h: router.go - đăng ký route sau JWT middleware của gateway (TASK-INFRA-001)"
  - "1.5h: price_history_test.go - 5 test (default 90d, bad range 400, unknown product 404, stitch raw tail, payload shape)"
  - "0.5h: OTel histogram price_history_duration_ms + span ghi range để soi p95"

risk_if_skipped: "Đây là API duy nhất nuôi biểu đồ giá (TASK-WEB-003) - không có nó thì màn hình lịch sử giá, lõi giá trị mà người dùng SănDeal vào xem mỗi ngày, không có nguồn dữ liệu. Nếu làm sai bằng cách quét raw price_snapshot cho cả khoảng thì p95 vỡ NFR <500ms (§3.8), biểu đồ giật trên SKU lịch sử dài. Nếu chỉ đọc cagg mà không stitch raw tail thì điểm giá gần nhất trễ tới một giờ (chờ continuous aggregate refresh), người dùng thấy giá cũ ngay lúc flash sale đang chạy - đúng thời điểm họ cần số tươi nhất. Thiếu allowlist range thì một request range=99999d quét toàn bộ raw, một câu query đủ kéo sập price-svc."
---

## §1 - Mô tả (BCP-14 normative)

Service PRICE **MUST** expose endpoint REST `GET /v1/products/{id}/price-history?range=90d` trả chuỗi thời gian giá của đúng một `tracked_product`, ghép phần thân từ continuous aggregate `price_daily` với cái đuôi raw từ `price_snapshot` để biểu đồ tươi tới phút mà vẫn đạt p95 <500ms. Hợp đồng:

1. **MUST** phục vụ route `GET /v1/products/{id}/price-history`; `{id}` là `tracked_product.id` (int64). Parse lỗi (id không phải số) trả `400`.
2. **MUST** chấp nhận query `range` chỉ trong allowlist `{7d, 30d, 90d, 180d, 1y}` (DEC-PRICE-31); thiếu `range` mặc định `90d`; giá trị ngoài allowlist trả `400` với thân lỗi `{"error":"invalid range"}`. KHÔNG nhận khoảng tùy ý.
3. **MUST** đọc phần thân của khoảng từ `price_daily` (DEC-PRICE-30): các điểm ngày `{day, min_p, max_p, close_p}` từ `now() - range` tới mốc cắt đuôi.
4. **MUST** stitch cái đuôi raw: đọc `price_snapshot` kể từ bucket ngày gần nhất đã đóng (mốc `date_trunc('day', now())`) tới hiện tại, trả các điểm `{ts, price, flash_sale}` (DEC-PRICE-30). Đuôi này luôn nằm trong một chunk nóng.
5. **MUST** trả JSON đúng hình dạng: `{product_id, range, daily:[{day, min_p, max_p, close_p}], tail:[{ts, price, flash_sale}]}`.
6. **MUST** trả `404` khi `product_id` không tồn tại trong `tracked_product` (DEC-PRICE-34), phân biệt với `200` + `daily`/`tail` rỗng khi SKU có thật nhưng chưa có snapshot nào.
7. **MUST** đảm bảo p95 <500ms (§3.8): chỉ chạm `price_daily` cho phần thân và tối đa một chunk raw cho đuôi (DEC-PRICE-32). KHÔNG quét raw cho toàn khoảng.
8. **MUST** dùng `close_p` (last(price) trong ngày) làm giá hiển thị của điểm ngày (DEC-PRICE-33); `min_p`/`max_p` chỉ để vẽ dải biên.
9. **MUST** trả mọi giá dưới dạng `BIGINT` VND (int64) trong payload - KHÔNG dùng float hay string (đồng nhất với DEC-PRICE-05 của TASK-PRICE-002).
10. **MUST** xác thực qua JWT do API Gateway (TASK-INFRA-001) gắn; handler giả định đã qua middleware, không tự parse token. Request thiếu auth bị gateway chặn trước khi tới handler.
11. **SHOULD** phát OTel: `price_history_duration_ms` (histogram, nhãn `range`), `price_history_requests_total{range, status}` (counter) để soi p95 theo từng range.
12. **MUST** đặt `Content-Type: application/json; charset=utf-8` và trả `daily` sắp tăng dần theo `day`, `tail` sắp tăng dần theo `ts`.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao đọc cagg cho thân nhưng stitch raw cho đuôi (DEC-PRICE-30)?** Continuous aggregate `price_daily` refresh mỗi giờ (TASK-PRICE-002 §1 #6), nên bucket của ngày hôm nay chưa "đóng" và chưa phản ánh giá vừa đổi. Nếu chỉ đọc cagg, người dùng thấy giá trễ tới một giờ - tệ nhất đúng lúc flash sale đang chạy. Đọc `price_daily` cho phần thân (nhanh, bảng nhỏ đã tổng hợp) rồi nối thêm raw `price_snapshot` kể từ đầu ngày hôm nay cho ta đường giá tươi tới phút mà không phải quét raw cả khoảng.

**Vì sao range allowlist có cap (DEC-PRICE-31)?** Phần đuôi raw chỉ rẻ vì nó nằm gọn trong một chunk nóng. Nếu cho range tùy ý, một request `range=99999d` ép đọc raw qua hàng trăm chunk (kể cả chunk đã nén), một câu query đủ vỡ p95 và ngốn bộ nhớ. Allowlist `{7d,30d,90d,180d,1y}` khớp đúng các nút bấm trên UI biểu đồ và chặn cứng bề mặt tấn công.

**Vì sao close_p là giá hiển thị (DEC-PRICE-33)?** Một ngày có nhiều snapshot (đặc biệt khi flash sale). Đường giá biểu đồ cần một giá trị đại diện cho mỗi ngày, và giá đóng cửa (last trong ngày) là cái người dùng nhớ nhất - "hôm đó nó còn bao nhiêu". `min_p`/`max_p` vẫn trả kèm để vẽ dải biên độ dao động, nhưng đường chính bám `close_p`.

**Vì sao 404 tách khỏi 200-rỗng (DEC-PRICE-34)?** "SKU không tồn tại" và "SKU mới chưa có snapshot" là hai trạng thái khác nhau với UI: cái đầu là lỗi định tuyến (đường dẫn sai), cái sau là màn hình "đang thu thập dữ liệu giá". Trộn cả hai thành 200-rỗng làm frontend không phân biệt được nên hiển thị gì.

---

## §3 - Hợp đồng API / DDL

### Truy vấn (Go)

```go
// services/price/internal/price/history_query.go

// rangeWindows là allowlist range -> khoảng thời gian (DEC-PRICE-31).
var rangeWindows = map[string]time.Duration{
    "7d":   7 * 24 * time.Hour,
    "30d":  30 * 24 * time.Hour,
    "90d":  90 * 24 * time.Hour,
    "180d": 180 * 24 * time.Hour,
    "1y":   365 * 24 * time.Hour,
}

// ParseRange validate range, trả khoảng + ok=false nếu ngoài allowlist.
func ParseRange(raw string) (time.Duration, bool) {
    if raw == "" {
        return rangeWindows["90d"], true // default 90d
    }
    d, ok := rangeWindows[raw]
    return d, ok
}

// QueryDailyBody đọc phần thân từ price_daily (DEC-PRICE-30).
func (r *Repo) QueryDailyBody(ctx context.Context, productID int64, from time.Time) ([]DailyPoint, error) {
    rows, err := r.pool.Query(ctx,
        `SELECT day, min_p, max_p, close_p
         FROM price_daily
         WHERE product_id = $1 AND day >= $2
         ORDER BY day`, productID, from)
    if err != nil {
        return nil, err
    }
    return scanDaily(rows)
}

// QueryRawTail đọc raw kể từ đầu ngày hôm nay - tối đa một chunk nóng (DEC-PRICE-32).
func (r *Repo) QueryRawTail(ctx context.Context, productID int64) ([]TailPoint, error) {
    rows, err := r.pool.Query(ctx,
        `SELECT ts, price, flash_sale
         FROM price_snapshot
         WHERE product_id = $1 AND ts >= date_trunc('day', now())
         ORDER BY ts`, productID)
    if err != nil {
        return nil, err
    }
    return scanTail(rows)
}

// ProductExists phân biệt 404 với 200-rỗng (DEC-PRICE-34).
func (r *Repo) ProductExists(ctx context.Context, productID int64) (bool, error) {
    var ok bool
    err := r.pool.QueryRow(ctx,
        `SELECT EXISTS(SELECT 1 FROM tracked_product WHERE id = $1)`, productID).Scan(&ok)
    return ok, err
}
```

### DTO + handler (Go)

```go
// services/price/internal/api/price_history.go

type DailyPoint struct {
    Day    time.Time `json:"day"`
    MinP   int64     `json:"min_p"` // VND
    MaxP   int64     `json:"max_p"` // VND
    CloseP int64     `json:"close_p"` // VND, giá hiển thị (DEC-PRICE-33)
}

type TailPoint struct {
    TS        time.Time `json:"ts"`
    Price     int64     `json:"price"` // VND
    FlashSale bool      `json:"flash_sale"`
}

type HistoryResponse struct {
    ProductID int64        `json:"product_id"`
    Range     string       `json:"range"`
    Daily     []DailyPoint `json:"daily"`
    Tail      []TailPoint  `json:"tail"`
}

// HandlePriceHistory phục vụ GET /v1/products/{id}/price-history?range=90d.
func (h *Handler) HandlePriceHistory(w http.ResponseWriter, req *http.Request) {
    id, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
    if err != nil {
        writeErr(w, http.StatusBadRequest, "invalid product id")
        return
    }
    rangeRaw := req.URL.Query().Get("range")
    window, ok := price.ParseRange(rangeRaw)
    if !ok {
        writeErr(w, http.StatusBadRequest, "invalid range")
        return
    }
    if rangeRaw == "" {
        rangeRaw = "90d"
    }
    exists, err := h.repo.ProductExists(req.Context(), id)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    if !exists {
        writeErr(w, http.StatusNotFound, "product not found") // DEC-PRICE-34
        return
    }
    from := time.Now().Add(-window)
    daily, err := h.repo.QueryDailyBody(req.Context(), id, from)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    tail, err := h.repo.QueryRawTail(req.Context(), id)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    _ = json.NewEncoder(w).Encode(HistoryResponse{
        ProductID: id, Range: rangeRaw, Daily: daily, Tail: tail,
    })
}
```

---

## §4 - Acceptance criteria

1. `GET /v1/products/abc/price-history` (id không phải số) trả `400`.
2. `range` ngoài allowlist (vd `range=5d`) trả `400` + `{"error":"invalid range"}`; thiếu `range` xử lý như `90d`.
3. Phần `daily` đến từ `price_daily` cho khoảng `now() - range` (không quét raw cho phần thân).
4. Phần `tail` đến từ `price_snapshot` kể từ `date_trunc('day', now())`, không sớm hơn.
5. Response JSON đúng khóa `{product_id, range, daily[], tail[]}` với `daily[*]` có `{day, min_p, max_p, close_p}` và `tail[*]` có `{ts, price, flash_sale}`.
6. `product_id` không có trong `tracked_product` trả `404`; SKU có thật nhưng chưa có snapshot trả `200` + `daily`/`tail` rỗng.
7. p95 của endpoint <500ms trên SKU có lịch sử 1 năm (đo qua `price_history_duration_ms`).
8. Điểm `daily` dùng `close_p` làm giá hiển thị; `min_p`/`max_p` vẫn có mặt.
9. Mọi giá trong JSON là số nguyên VND (int64), không float, không string.
10. Request đã qua JWT gateway tới được handler; handler không tự xác thực token.
11. Metric `price_history_requests_total{range, status}` tăng đúng theo từng phản hồi.
12. `daily` sắp tăng theo `day`, `tail` sắp tăng theo `ts`; `Content-Type` là `application/json; charset=utf-8`.

---

## §5 - Kiểm thử (verification)

```go
// services/price/internal/api/price_history_test.go

func TestPriceHistory_Default90d(t *testing.T) {
    h, pid := setupWithHistory(t, 120) // 120 ngày dữ liệu
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/price-history") // không range
    require.Equal(t, 200, rec.Code)
    var body HistoryResponse
    decode(t, rec, &body)
    require.Equal(t, "90d", body.Range)
    require.LessOrEqual(t, len(body.Daily), 91) // ~90 ngày, không phải 120
}

func TestPriceHistory_BadRange_400(t *testing.T) {
    h, pid := setupWithHistory(t, 30)
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/price-history?range=5d")
    require.Equal(t, 400, rec.Code)
    require.Contains(t, rec.Body.String(), "invalid range")
}

func TestPriceHistory_UnknownProduct_404(t *testing.T) {
    h, _ := setupWithHistory(t, 30)
    rec := doGET(t, h, "/v1/products/999999/price-history?range=30d")
    require.Equal(t, 404, rec.Code) // DEC-PRICE-34
}

func TestPriceHistory_StitchesRawTail(t *testing.T) {
    h, pid := setupWithHistory(t, 30)
    // ghi một snapshot raw lúc trưa nay, sau bucket cagg gần nhất
    insertRaw(t, h, pid, todayAt(12, 0), 79_000, true)
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/price-history?range=30d")
    var body HistoryResponse
    decode(t, rec, &body)
    require.NotEmpty(t, body.Tail) // đuôi raw có mặt, không chờ cagg refresh
    require.Equal(t, int64(79_000), body.Tail[len(body.Tail)-1].Price)
    require.True(t, body.Tail[len(body.Tail)-1].FlashSale)
}

func TestPriceHistory_PayloadShape(t *testing.T) {
    h, pid := setupWithHistory(t, 90)
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/price-history?range=90d")
    var raw map[string]json.RawMessage
    require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw))
    for _, k := range []string{"product_id", "range", "daily", "tail"} {
        _, ok := raw[k]
        require.True(t, ok, "thiếu khóa %s", k)
    }
    require.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `history_query.go` (ParseRange + 2 truy vấn + ProductExists) -> `price_history.go` (DTO + handler) -> đăng ký route trong `router.go` sau JWT middleware của gateway (TASK-INFRA-001) -> tests. Handler dùng `http.ServeMux` của Go 1.22 với pattern `GET /v1/products/{id}/price-history` và `req.PathValue("id")`. Không thêm caching ở slice này; cagg + đuôi một chunk đã đủ cho p95 <500ms.

---

## §7 - Phụ thuộc

- **TASK-PRICE-002** - cung cấp `price_daily` (continuous aggregate) cho phần thân và `price_snapshot` (hypertable) cho đuôi raw. Đây là điều kiện cứng.
- **TASK-INFRA-001 (gateway)** - gắn JWT auth trước handler; route nằm sau middleware này.
- **TASK-WEB-003 (downstream)** - biểu đồ giá tiêu thụ response này; hình dạng JSON ở §3 là hợp đồng với nó.
- **TASK-DEAL-003 (sibling)** - API dữ liệu biểu đồ song song (sale ảo); dùng chung quy ước range allowlist + giá int64 VND.
- Lib: `pgx` (driver), `encoding/json`, `net/http` (ServeMux Go 1.22).

---

## §8 - Payload ví dụ

### Request

```
curl -s -H "Authorization: Bearer $JWT" \
  "https://api.sandeal.vn/v1/products/90112/price-history?range=90d"
```

### Response (200)

```json
{
  "product_id": 90112,
  "range": "90d",
  "daily": [
    { "day": "2026-03-29T00:00:00Z", "min_p": 149000, "max_p": 159000, "close_p": 149000 },
    { "day": "2026-03-30T00:00:00Z", "min_p": 129000, "max_p": 149000, "close_p": 129000 },
    { "day": "2026-06-26T00:00:00Z", "min_p": 99000,  "max_p": 119000, "close_p": 99000 }
  ],
  "tail": [
    { "ts": "2026-06-27T08:15:00Z", "price": 99000, "flash_sale": false },
    { "ts": "2026-06-27T12:00:00Z", "price": 79000, "flash_sale": true }
  ]
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Cache lớp HTTP (ETag theo `product_id` + ngày) nếu QPS biểu đồ cao - slice sau, đo trước.
- Tham số `?points=` để downsample điểm ngày cho khoảng `1y` trên màn hình nhỏ - tối ưu UI giai đoạn sau.
- Trả kèm `currency` khi mở SEA (THB/IDR) - bám theo `tracked_product`, giữ giá int64 theo minor unit từng nước.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| id không phải số | `ParseInt` lỗi | 400 | Frontend chỉ gửi id số (theo route biểu đồ) |
| range ngoài allowlist | `ParseRange` ok=false | 400 invalid range | UI chỉ render các nút 7d/30d/90d/180d/1y |
| product_id không tồn tại | `ProductExists` false | 404 | Phân biệt với 200-rỗng (DEC-PRICE-34) |
| SKU có thật, chưa có snapshot | query trả 0 dòng | 200 + daily/tail rỗng | UI hiện "đang thu thập dữ liệu giá" |
| cagg chưa refresh đuôi hôm nay | thiết kế | đuôi đến từ raw, không trễ | Stitch raw tail luôn bù phần cagg thiếu |
| quét raw toàn khoảng (bug) | p95 query metric vọt | vỡ <500ms | Giữ đuôi raw trong `date_trunc('day', now())` |
| chunk raw đuôi đã nén | hiếm (đuôi là chunk nóng) | đọc chậm hơn | Đuôi luôn <1 ngày nên nằm trong chunk chưa nén (30d) |
| range=1y trên SKU lịch sử dài | p95 metric | phần thân nhiều điểm | price_daily nhẹ; cân nhắc `?points=` sau |

---

## §11 - Ghi chú

- API này là cầu nối duy nhất giữa kho time-series giá (TASK-PRICE-002) và biểu đồ người dùng (TASK-WEB-003).
- Mẹo "cagg cho thân + raw cho đuôi" tách tốc độ (đọc bảng tổng hợp nhỏ) khỏi độ tươi (raw tới phút) mà không phải quét raw cả khoảng.
- Allowlist range vừa khớp các nút UI vừa là rào chắn cứng chống quét raw không giới hạn.
- `close_p` là giá hiển thị; `min_p`/`max_p` chỉ vẽ dải biên - đồng nhất với cách TASK-PRICE-002 dựng continuous aggregate.
- Giá luôn là int64 VND suốt đường đi từ DB tới JSON, không có bước chuyển float nào để tránh sai số.
- Khi mở SEA, đính `currency` vào payload theo `tracked_product`; giá vẫn là int64 theo minor unit từng nước.

---

*Hết TASK-PRICE-003. Status: ready_to_implement (mục tiêu audit 10/10).*
