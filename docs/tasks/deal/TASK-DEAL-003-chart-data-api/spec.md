---
id: TASK-DEAL-003
title: "GET /v1/products/{id}/chart - feed dữ liệu biểu đồ giá có chú giải tín hiệu deal: thân daily từ price_daily + overlay median90/trailing_min/verdict sale ảo + mốc ngày đôi, p95 <500ms"
module: DEAL
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-PRICE-002, TASK-PRICE-003, TASK-DEAL-001, TASK-DEAL-002, TASK-WEB-003]
depends_on: [TASK-PRICE-002]
blocks: [TASK-WEB-003]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.7 (API GET /v1/products/{id}/chart?range=90d)"
  - "docs/... §3.8 (NFR biểu đồ <500ms)"
  - "docs/... §3.5 (mốc ngày đôi 1.1...12.12, annotation sale ảo)"
source_decisions:
  - "DEC-DEAL-20: feed biểu đồ đọc continuous aggregate price_daily (bucket ngày) cho phần thân để giữ p95 <500ms - không quét raw price_snapshot cho cả khoảng"
  - "DEC-DEAL-21: overlay annotation phía server (median90, trailing_min, verdict sale ảo, mốc ngày đôi) để verdict đồng nhất với TASK-DEAL-001 và client nhẹ"
  - "DEC-DEAL-22: DEAL-003 là feed chart-shaped (downsample daily + chú giải) - phân biệt rõ với TASK-PRICE-003 là chuỗi giá thô (daily + raw tail, không chú giải)"
  - "DEC-DEAL-23: tôn trọng cổng độ chín TASK-DEAL-002 - WARMING gắn cờ đang tích lũy, NEW (<14 ngày) trả dữ liệu nhưng verdict=UNKNOWN"
  - "DEC-DEAL-24: range chỉ nhận allowlist 7d/30d/90d/180d/1y, default 90d, range lạ trả 400 - chặn quét raw không giới hạn vỡ NFR"

language: "Go 1.22 (deal-svc); PostgreSQL 16 + TimescaleDB 2.x (đọc price_daily)"
service: shopass/services/deal/
new_files:
  - services/deal/internal/api/chart.go
  - services/deal/internal/api/chart_test.go
  - services/deal/internal/chart/annotate.go
modified_files:
  - services/deal/internal/api/router.go               # đăng ký route GET /v1/products/{id}/chart
allowed_tools:
  - file_read: services/deal/**
  - file_write: services/deal/**
  - bash: cd services/deal && go test ./...
disallowed_tools:
  - quét raw price_snapshot cho toàn khoảng để dựng biểu đồ (vi phạm DEC-DEAL-20, vỡ p95 <500ms)
  - nhận range tùy ý không qua allowlist (vi phạm DEC-DEAL-24, mở cửa quét raw không giới hạn)
  - tính verdict sale ảo ở client thay vì gọi TASK-DEAL-001 (vi phạm DEC-DEAL-21, verdict lệch giữa thẻ và biểu đồ)
  - trả price dạng float/string trong JSON (price là BIGINT VND, phải là int64)

effort_hours: 5
sub_tasks:
  - "0.5h: chart.go - parse + validate range qua allowlist, map sang khoảng thời gian, default 90d"
  - "1.0h: chart.go - đọc thân daily từ price_daily, kiểm tra tồn tại product (404), lấy độ chín TASK-DEAL-002"
  - "1.5h: annotate.go - tính median90 + trailing_min từ chuỗi daily, sinh mốc ngày đôi trong range, gắn verdict từ TASK-DEAL-001"
  - "0.5h: chart.go - ghép DTO {product_id, range, maturity, daily, annotations}, marshal JSON, đặt Content-Type"
  - "0.5h: router.go - đăng ký route sau JWT middleware của gateway (TASK-INFRA-001)"
  - "1.0h: chart_test.go - 5 test (default 90d, annotation median/trailing, mốc ngày đôi, cờ WARMING, unknown product 404)"

risk_if_skipped: "Đây là feed duy nhất nuôi biểu đồ giá có chú giải (TASK-WEB-003) - màn hình mà người dùng SănDeal mở để trả lời câu hỏi cốt lõi đây có phải sale thật không. Thiếu nó thì biểu đồ chỉ là đường giá trần trụi, không có dải verdict sale ảo, không có đường trailing_min, không có mốc ngày đôi, người dùng tự đoán. Nếu dựng sai bằng cách quét raw price_snapshot cho cả khoảng thì p95 vỡ NFR <500ms (§3.8), biểu đồ giật trên SKU lịch sử dài. Nếu tính verdict ở client thay vì gọi TASK-DEAL-001 thì dải verdict trên biểu đồ lệch với nhãn trên thẻ sản phẩm, người dùng mất tin. Nếu bỏ qua cổng độ chín TASK-DEAL-002 thì SKU mới <14 ngày bị dán nhãn SALE_XIN dựa trên 3 điểm dữ liệu - đúng kiểu kết luận ẩu mà SănDeal hứa loại bỏ."
---

## §1 - Mô tả (BCP-14 normative)

Service DEAL **MUST** expose endpoint REST `GET /v1/products/{id}/chart?range=90d` trả feed dữ liệu đã định hình cho biểu đồ của đúng một `tracked_product`: phần thân là chuỗi giá theo ngày đọc từ continuous aggregate `price_daily`, phủ thêm các chú giải tín hiệu deal (đường median90, đường trailing_min, verdict sale ảo, mốc ngày đôi) để biểu đồ tự giải thích đây có phải sale thật không, đạt p95 <500ms. Hợp đồng:

1. **MUST** phục vụ route `GET /v1/products/{id}/chart`; `{id}` là `tracked_product.id` (int64). Parse lỗi (id không phải số) trả `400`.
2. **MUST** chấp nhận query `range` chỉ trong allowlist `{7d, 30d, 90d, 180d, 1y}` (DEC-DEAL-24); thiếu `range` mặc định `90d`; giá trị ngoài allowlist trả `400` với thân lỗi `{"error":"invalid range"}`. KHÔNG nhận khoảng tùy ý.
3. **MUST** đọc phần thân biểu đồ từ `price_daily` (DEC-DEAL-20): các điểm ngày `{day, min_p, max_p, close_p}` cho khoảng `now() - range` tới hiện tại, sắp tăng dần theo `day`. KHÔNG quét raw `price_snapshot` cho phần thân.
4. **MUST** tính `median90` (trung vị `close_p` trong cửa sổ 90 ngày gần nhất, đơn vị VND) và `trailing_min` (giá thấp nhất `min_p` trong toàn khoảng được trả) từ chính chuỗi daily (DEC-DEAL-21). Hai số này là đường tham chiếu để vẽ trên biểu đồ.
5. **MUST** gắn verdict sale ảo bằng cách gọi đánh giá của **TASK-DEAL-001** (`SALE_AO`, `SALE_XIN`, `TAM_DUOC`, hoặc `UNKNOWN`) - KHÔNG tự tính lại verdict ở DEAL-003 hay ở client (DEC-DEAL-21). Verdict trên biểu đồ phải khớp nhãn trên thẻ sản phẩm.
6. **MUST** lấy độ chín dữ liệu từ **TASK-DEAL-002** và trả trong trường `maturity` (`MATURE`, `WARMING`, hoặc `NEW`) (DEC-DEAL-23). SKU `WARMING` vẫn trả biểu đồ kèm cờ `accumulating=true` (đang tích lũy); SKU `NEW` (<14 ngày) trả dữ liệu nhưng `verdict=UNKNOWN`.
7. **MUST** tính tập mốc ngày đôi `double_dates` nằm trong khoảng được trả (DEC-DEAL-21): các ngày `1.1, 2.2, 3.3 ... 12.12` (ngày dd bằng tháng mm) rơi vào `[now() - range, now()]`, trả dạng `YYYY-MM-DD`. Đây là các mốc lịch sale lớn của sàn VN để biểu đồ đánh dấu.
8. **MUST** trả JSON đúng hình dạng: `{product_id, range, maturity, daily:[{day, min_p, max_p, close_p}], annotations:{median90, trailing_min, verdict, accumulating, double_dates:[...]}}`.
9. **MUST** đảm bảo p95 <500ms (§3.8): chỉ chạm `price_daily` (bảng tổng hợp nhỏ) cho phần thân, chú giải tính trong bộ nhớ trên chính chuỗi daily đã đọc - không thêm vòng quét raw (DEC-DEAL-20).
10. **MUST** trả `404` khi `product_id` không tồn tại trong `tracked_product`, phân biệt với `200` + `daily` rỗng khi SKU có thật nhưng chưa có snapshot nào.
11. **MUST** trả mọi giá (`min_p`, `max_p`, `close_p`, `median90`, `trailing_min`) dưới dạng `BIGINT` VND (int64) trong payload - KHÔNG dùng float hay string (đồng nhất với DEC-PRICE-05 của TASK-PRICE-002).
12. **MUST** xác thực qua JWT do API Gateway (TASK-INFRA-001) gắn; handler giả định đã qua middleware, không tự parse token. Request thiếu auth bị gateway chặn trước khi tới handler.
13. **SHOULD** phát OTel: `chart_duration_ms` (histogram, nhãn `range`), `chart_requests_total{range, status, maturity}` (counter) để soi p95 theo từng range.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao tách feed biểu đồ riêng thay vì dùng TASK-PRICE-003?** TASK-PRICE-003 trả chuỗi giá thô `{daily, tail}` - đường giá trần, ghép raw tail cho tươi tới phút, không chú giải. DEAL-003 trả `{daily, annotations}` - đã downsample về bucket ngày để vẽ và phủ thêm dải verdict, đường median90/trailing_min, mốc ngày đôi. Hai endpoint phục vụ hai nhu cầu khác nhau: PRICE-003 cho ai cần số giá nguyên bản (API tích hợp, B2B), DEAL-003 cho biểu đồ deal mà người dùng cuối nhìn để quyết "mua hay đợi". Trộn cả hai vào một endpoint làm hợp đồng rối và buộc client tự suy ra annotation.

**Vì sao tính chú giải phía server (DEC-DEAL-21)?** verdict sale ảo phải đồng nhất giữa thẻ sản phẩm và biểu đồ - nếu thẻ ghi SALE_XIN mà dải trên biểu đồ vẽ SALE_AO thì người dùng mất tin ngay. Gọi chung một nguồn (TASK-DEAL-001) đảm bảo một verdict. median90 và trailing_min cũng nên tính một lần ở server: client chỉ vẽ, không phải tải cả chuỗi raw rồi tự tính lại - vừa lệch, vừa nặng.

**Vì sao mốc ngày đôi (DEC-DEAL-21)?** Sàn VN chạy sale lớn vào các ngày trùng số (1.1, 2.2 ... 12.12). Đánh dấu các mốc này trên biểu đồ giúp người dùng đọc đúng bối cảnh: một cú giảm giá đúng ngày 12.12 là sale lịch chứ không phải tín hiệu hiếm, còn giá thấp giữa tháng thường thì đáng chú ý hơn. Đây là kiến thức lịch bản địa mà biểu đồ giá thuần không thể hiện được.

**Vì sao bucket ngày đạt <500ms (DEC-DEAL-20)?** Biểu đồ không cần độ phân giải giây - cần min/max/close theo ngày để vẽ. Đọc `price_daily` (continuous aggregate đã tổng hợp, bảng nhỏ) cho phần thân thay vì quét raw `price_snapshot` (bảng tỷ dòng) là khác biệt giữa vài mili-giây và vài trăm mili-giây. Chú giải tính trên chính chuỗi daily đã nằm trong bộ nhớ nên gần như miễn phí.

**Vì sao tôn trọng cổng độ chín TASK-DEAL-002 (DEC-DEAL-23)?** verdict trên 3 điểm dữ liệu là rác. SKU mới <14 ngày chưa đủ lịch sử để biết giá nền thật, nên trả `verdict=UNKNOWN` thay vì đoán liều. SKU WARMING (đủ vài tuần nhưng chưa đủ 90 ngày) vẫn vẽ được biểu đồ nhưng gắn cờ đang tích lũy để frontend hiển thị lưu ý. Đây là lời hứa cốt lõi của SănDeal: không kết luận ẩu.

---

## §3 - Hợp đồng API / DDL

### Chú giải (Go)

```go
// services/deal/internal/chart/annotate.go

type DailyPoint struct {
    Day    time.Time `json:"day"`
    MinP   int64     `json:"min_p"`   // VND
    MaxP   int64     `json:"max_p"`   // VND
    CloseP int64     `json:"close_p"` // VND, giá hiển thị của điểm ngày
}

type Annotations struct {
    Median90     int64    `json:"median90"`     // VND, trung vị close_p 90 ngày
    TrailingMin  int64    `json:"trailing_min"` // VND, đáy min_p trong khoảng
    Verdict      string   `json:"verdict"`      // từ TASK-DEAL-001
    Accumulating bool     `json:"accumulating"` // true khi maturity=WARMING
    DoubleDates  []string `json:"double_dates"` // YYYY-MM-DD trong khoảng
}

// Build tính median90 + trailing_min từ chuỗi daily và sinh mốc ngày đôi (DEC-DEAL-21).
// verdict + accumulating được nhồi bởi caller (chart.go) từ TASK-DEAL-001/002.
func Build(daily []DailyPoint, from, to time.Time) Annotations {
    return Annotations{
        Median90:    median90(daily),
        TrailingMin: trailingMin(daily),
        DoubleDates: doubleDates(from, to),
    }
}

// median90 lấy trung vị close_p của các điểm trong 90 ngày gần nhất.
func median90(d []DailyPoint) int64 {
    cut := time.Now().AddDate(0, 0, -90)
    var v []int64
    for _, p := range d {
        if !p.Day.Before(cut) {
            v = append(v, p.CloseP)
        }
    }
    if len(v) == 0 {
        return 0
    }
    sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
    return v[len(v)/2]
}

// trailingMin lấy đáy min_p trong toàn chuỗi được trả.
func trailingMin(d []DailyPoint) int64 {
    if len(d) == 0 {
        return 0
    }
    m := d[0].MinP
    for _, p := range d[1:] {
        if p.MinP < m {
            m = p.MinP
        }
    }
    return m
}

// doubleDates liệt kê các ngày dd==mm (1.1...12.12) rơi trong [from, to] (DEC-DEAL-21).
func doubleDates(from, to time.Time) []string {
    var out []string
    for y := from.Year(); y <= to.Year(); y++ {
        for m := 1; m <= 12; m++ {
            d := time.Date(y, time.Month(m), m, 0, 0, 0, 0, time.UTC)
            if !d.Before(from) && !d.After(to) {
                out = append(out, d.Format("2006-01-02"))
            }
        }
    }
    return out
}
```

### DTO + handler (Go)

```go
// services/deal/internal/api/chart.go

type ChartResponse struct {
    ProductID   int64              `json:"product_id"`
    Range       string             `json:"range"`
    Maturity    string             `json:"maturity"` // MATURE | WARMING | NEW (TASK-DEAL-002)
    Daily       []chart.DailyPoint `json:"daily"`
    Annotations chart.Annotations  `json:"annotations"`
}

var rangeWindows = map[string]time.Duration{
    "7d": 7 * 24 * time.Hour, "30d": 30 * 24 * time.Hour,
    "90d": 90 * 24 * time.Hour, "180d": 180 * 24 * time.Hour,
    "1y": 365 * 24 * time.Hour,
}

// HandleChart phục vụ GET /v1/products/{id}/chart?range=90d.
func (h *Handler) HandleChart(w http.ResponseWriter, req *http.Request) {
    id, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
    if err != nil {
        writeErr(w, http.StatusBadRequest, "invalid product id")
        return
    }
    rng := req.URL.Query().Get("range")
    if rng == "" {
        rng = "90d"
    }
    window, ok := rangeWindows[rng]
    if !ok {
        writeErr(w, http.StatusBadRequest, "invalid range")
        return
    }
    exists, err := h.repo.ProductExists(req.Context(), id)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    if !exists {
        writeErr(w, http.StatusNotFound, "product not found")
        return
    }
    from, to := time.Now().Add(-window), time.Now()
    daily, err := h.repo.QueryDaily(req.Context(), id, from) // đọc price_daily
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    ann := chart.Build(daily, from, to)
    mat := h.deal.Maturity(req.Context(), id)        // TASK-DEAL-002
    ann.Accumulating = mat == "WARMING"
    ann.Verdict = h.deal.Verdict(req.Context(), id)  // TASK-DEAL-001
    if mat == "NEW" {
        ann.Verdict = "UNKNOWN" // <14 ngày: không kết luận (DEC-DEAL-23)
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    _ = json.NewEncoder(w).Encode(ChartResponse{
        ProductID: id, Range: rng, Maturity: mat, Daily: daily, Annotations: ann,
    })
}
```

---

## §4 - Acceptance criteria

1. `GET /v1/products/abc/chart` (id không phải số) trả `400`.
2. `range` ngoài allowlist (vd `range=5d`) trả `400` + `{"error":"invalid range"}`; thiếu `range` xử lý như `90d`.
3. Phần `daily` đến từ `price_daily` cho khoảng `now() - range`, sắp tăng theo `day`; không quét raw cho phần thân.
4. `annotations.median90` = trung vị `close_p` 90 ngày; `annotations.trailing_min` = đáy `min_p` trong khoảng, cả hai là int64 VND.
5. `annotations.verdict` lấy từ TASK-DEAL-001 (không tính lại); khớp nhãn thẻ sản phẩm.
6. `maturity` lấy từ TASK-DEAL-002; `WARMING` kèm `annotations.accumulating=true`; `NEW` (<14 ngày) trả `verdict=UNKNOWN`.
7. `annotations.double_dates` chứa đúng các ngày dd==mm (1.1...12.12) rơi trong khoảng, dạng `YYYY-MM-DD`.
8. Response JSON đúng khóa `{product_id, range, maturity, daily[], annotations{median90, trailing_min, verdict, accumulating, double_dates[]}}`.
9. p95 của endpoint <500ms trên SKU có lịch sử 1 năm (đo qua `chart_duration_ms`); chú giải tính trong bộ nhớ, không thêm vòng quét raw.
10. `product_id` không có trong `tracked_product` trả `404`; SKU có thật chưa có snapshot trả `200` + `daily` rỗng.
11. Mọi giá trong JSON là số nguyên VND (int64), không float, không string.
12. Request đã qua JWT gateway tới được handler; handler không tự xác thực token; metric `chart_requests_total{range, status, maturity}` tăng đúng.

---

## §5 - Kiểm thử (verification)

```go
// services/deal/internal/api/chart_test.go

func TestChart_Default90d(t *testing.T) {
    h, pid := setupWithHistory(t, 120) // 120 ngày dữ liệu
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/chart") // không range
    require.Equal(t, 200, rec.Code)
    var body ChartResponse
    decode(t, rec, &body)
    require.Equal(t, "90d", body.Range)
    require.LessOrEqual(t, len(body.Daily), 91) // ~90 ngày, không phải 120
}

func TestChart_Annotations_MedianTrailingMin(t *testing.T) {
    h, pid := setupWithHistory(t, 90)
    seedDailyCloses(t, h, pid, []int64{120_000, 100_000, 80_000, 100_000}) // đáy 80k
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/chart?range=90d")
    var body ChartResponse
    decode(t, rec, &body)
    require.Equal(t, int64(100_000), body.Annotations.Median90)    // trung vị
    require.Equal(t, int64(80_000), body.Annotations.TrailingMin)  // đáy
}

func TestChart_DoubleDateMarkers(t *testing.T) {
    h, pid := setupWithRange(t, "2026-04-01", "2026-06-27") // chứa 4.4, 5.5, 6.6
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/chart?range=90d")
    var body ChartResponse
    decode(t, rec, &body)
    require.Contains(t, body.Annotations.DoubleDates, "2026-04-04")
    require.Contains(t, body.Annotations.DoubleDates, "2026-05-05")
    require.NotContains(t, body.Annotations.DoubleDates, "2026-03-03") // ngoài khoảng
}

func TestChart_MaturityFlag_Warming(t *testing.T) {
    h, pid := setupWithMaturity(t, "WARMING", 40) // 40 ngày: đủ vẽ, chưa đủ 90
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/chart?range=90d")
    var body ChartResponse
    decode(t, rec, &body)
    require.Equal(t, "WARMING", body.Maturity)
    require.True(t, body.Annotations.Accumulating) // cờ đang tích lũy
}

func TestChart_UnknownProduct_404(t *testing.T) {
    h, _ := setupWithHistory(t, 30)
    rec := doGET(t, h, "/v1/products/999999/chart?range=30d")
    require.Equal(t, 404, rec.Code)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `annotate.go` (median90 + trailing_min + double_dates thuần, dễ unit-test) -> `chart.go` (DTO + handler: parse range, ProductExists, đọc price_daily, ghép verdict TASK-DEAL-001 + maturity TASK-DEAL-002) -> đăng ký route trong `router.go` sau JWT middleware của gateway (TASK-INFRA-001) -> tests. Handler dùng `http.ServeMux` của Go 1.22 với pattern `GET /v1/products/{id}/chart` và `req.PathValue("id")`. Không thêm caching ở slice này; đọc price_daily + tính chú giải trong bộ nhớ đã đủ cho p95 <500ms.

---

## §7 - Phụ thuộc

- **TASK-PRICE-002** - cung cấp continuous aggregate `price_daily` cho phần thân biểu đồ. Đây là điều kiện cứng (depends_on).
- **TASK-DEAL-001 (sibling)** - nguồn verdict sale ảo (`SALE_AO`/`SALE_XIN`/`TAM_DUOC`); DEAL-003 gọi để gắn vào `annotations.verdict`, không tự tính.
- **TASK-DEAL-002 (sibling)** - cổng độ chín dữ liệu; DEAL-003 lấy `maturity` và áp quy tắc WARMING/NEW.
- **TASK-INFRA-001 (gateway)** - gắn JWT auth trước handler; route nằm sau middleware này.
- **TASK-WEB-003 (downstream)** - biểu đồ giá có chú giải tiêu thụ response này; hình dạng JSON ở §3 là hợp đồng với nó.
- Lib: `pgx` (driver), `encoding/json`, `net/http` (ServeMux Go 1.22), `sort`.

---

## §8 - Payload ví dụ

### Request

```
curl -s -H "Authorization: Bearer $JWT" \
  "https://api.sandeal.vn/v1/products/90112/chart?range=90d"
```

### Response (200)

```json
{
  "product_id": 90112,
  "range": "90d",
  "maturity": "MATURE",
  "daily": [
    { "day": "2026-03-29T00:00:00Z", "min_p": 149000, "max_p": 159000, "close_p": 149000 },
    { "day": "2026-04-04T00:00:00Z", "min_p": 119000, "max_p": 149000, "close_p": 119000 },
    { "day": "2026-05-05T00:00:00Z", "min_p": 109000, "max_p": 139000, "close_p": 109000 },
    { "day": "2026-06-26T00:00:00Z", "min_p": 99000,  "max_p": 119000, "close_p": 99000 }
  ],
  "annotations": {
    "median90": 129000,
    "trailing_min": 99000,
    "verdict": "TAM_DUOC",
    "accumulating": false,
    "double_dates": ["2026-04-04", "2026-05-05", "2026-06-06"]
  }
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Cache lớp HTTP (ETag theo `product_id` + ngày + range) nếu QPS biểu đồ cao - slice sau, đo trước.
- Tham số `?points=` để downsample điểm ngày cho khoảng `1y` trên màn hình nhỏ - tối ưu UI giai đoạn sau.
- Dải verdict theo từng đoạn thời gian (band per khoảng) thay vì một verdict tổng - cân nhắc khi UI biểu đồ cần tô màu nhiều vùng.
- Trả kèm `currency` khi mở SEA (THB/IDR) - bám theo `tracked_product`, giữ giá int64 theo minor unit từng nước.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| id không phải số | `ParseInt` lỗi | 400 | Frontend chỉ gửi id số (theo route biểu đồ) |
| range ngoài allowlist | tra map ok=false | 400 invalid range | UI chỉ render các nút 7d/30d/90d/180d/1y |
| product_id không tồn tại | `ProductExists` false | 404 | Phân biệt với 200-rỗng (SKU mới chưa có snapshot) |
| SKU có thật, chưa có snapshot | query trả 0 dòng | 200 + daily rỗng | UI hiện "đang thu thập dữ liệu giá" |
| quét raw toàn khoảng (bug) | p95 query metric vọt | vỡ <500ms | Chỉ đọc price_daily, chú giải tính trong bộ nhớ |
| verdict client lệch verdict thẻ | so chéo nhãn | mất tin người dùng | Một nguồn verdict duy nhất TASK-DEAL-001 (DEC-DEAL-21) |
| SKU NEW bị gắn verdict ẩu | maturity=NEW | kết luận sai trên ít dữ liệu | Ép `verdict=UNKNOWN` khi <14 ngày (DEC-DEAL-23) |
| TASK-DEAL-002 chưa sẵn lúc gọi | maturity rỗng | thiếu cờ tích lũy | Mặc định an toàn: coi như WARMING tới khi đủ 90 ngày |
| median90 trên chuỗi rỗng | guard len==0 | trả 0 thay vì panic | `median90`/`trailingMin` trả 0 khi không có điểm |

---

## §11 - Ghi chú

- **Phân biệt với TASK-PRICE-003:** DEAL-003 và PRICE-003 đều phục vụ biểu đồ nhưng khác hợp đồng. PRICE-003 trả `{daily, tail}` - chuỗi giá thô, ghép raw tail cho tươi tới phút, KHÔNG chú giải; dành cho ai cần số giá nguyên bản. DEAL-003 trả `{daily, annotations}` - đã downsample về bucket ngày để vẽ và phủ thêm verdict sale ảo, đường median90/trailing_min, mốc ngày đôi; dành cho biểu đồ deal mà người dùng cuối đọc để quyết mua hay đợi. Một bên là feed thô, một bên là feed đã trang trí cho rendering.
- Chú giải tính phía server cho một verdict đồng nhất giữa thẻ và biểu đồ, và client chỉ vẽ chứ không tải cả chuỗi raw rồi tự tính.
- Mốc ngày đôi là kiến thức lịch sale bản địa VN - biểu đồ giá thuần không thể hiện được bối cảnh này.
- Đọc `price_daily` cho phần thân tách tốc độ (bảng tổng hợp nhỏ) khỏi nhu cầu chú giải, giữ p95 <500ms mà không quét raw cả khoảng.
- Cổng độ chín TASK-DEAL-002 là rào chắn chống kết luận ẩu: NEW trả UNKNOWN, WARMING gắn cờ tích lũy - đúng lời hứa cốt lõi của SănDeal.
- Giá luôn là int64 VND suốt đường đi từ DB tới JSON, không có bước chuyển float nào để tránh sai số.

---

*Hết TASK-DEAL-003. Status: ready_to_implement (mục tiêu audit 10/10).*
