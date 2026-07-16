---
id: TASK-SCRAPE-002
title: "Shopee internal-API adapter - /api/v4/pdp/get_pc + /api/v4/recommend, truy cập is_login:false, parse JSON -> PriceSnapshot"
module: SCRAPE
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-SCRAPE-001, TASK-SCRAPE-003, TASK-SCRAPE-005, TASK-SCRAPE-006, TASK-PRICE-001, TASK-PRICE-002, TASK-TRACK-001]
depends_on: [TASK-SCRAPE-001]
blocks: [TASK-SCRAPE-005, TASK-SCRAPE-006, TASK-TRACK-001]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (Shopee endpoints /api/v4/pdp/get_pc, /api/v4/recommend, is_login:false)"
  - "docs/... §3.3 (hybrid reverse-engineer internal JSON), §3.9 (anti-bot Shopee Medium-High)"
source_decisions:
  - "DEC-SCRAPE-06: ưu tiên internal JSON endpoint /api/v4/pdp/get_pc (rẻ, nhanh) khi truy cập được không cần login"
  - "DEC-SCRAPE-07: parse JSON Shopee về price.PriceSnapshot, quy đổi giá Shopee (micro-VND x100000) sang BIGINT VND"
  - "DEC-SCRAPE-08: rơi xuống Playwright farm (TASK-SCRAPE-003) khi endpoint trả lỗi/HTML challenge thay vì JSON"
  - "DEC-SCRAPE-09: không gửi cookie phiên người dùng; adapter backend dùng proxy residential + fingerprint riêng (tách hẳn extension)"

language: "Go 1.22 (scrape-svc); HTTP client + JSON parse; fallback gọi farm Playwright"
service: shopass/services/scrape/
new_files:
  - services/scrape/internal/adapters/shopee/adapter.go
  - services/scrape/internal/adapters/shopee/endpoints.go
  - services/scrape/internal/adapters/shopee/parse.go
  - services/scrape/internal/adapters/shopee/parse_test.go
  - services/scrape/internal/adapters/shopee/adapter_test.go
  - services/scrape/internal/adapters/shopee/testdata/pdp_get_pc.json
modified_files:
  - services/scrape/internal/orchestrator/registry.go     # đăng ký ShopeeAdapter theo platform_id
allowed_tools:
  - file_read: services/scrape/**
  - file_write: services/scrape/**
  - bash: cd services/scrape && go test ./...
disallowed_tools:
  - gửi cookie/token phiên của người dùng kèm request backend (vi phạm DEC-SCRAPE-09, §3.2)
  - lưu giá dạng float/numeric (vi phạm DEC-PRICE-05 của TASK-PRICE-002)
  - bỏ qua fallback Playwright khi nhận HTML challenge (làm mất dữ liệu khi Shopee siết)

effort_hours: 8
sub_tasks:
  - "1.0h: endpoints.go - dựng URL /api/v4/pdp/get_pc?item_id=&shop_id= + /api/v4/recommend"
  - "1.5h: parse.go - map JSON Shopee -> PriceSnapshot, quy đổi micro-VND, đọc flash_sale/stock/sold"
  - "1.5h: adapter.go - Fetch(ctx, job): gọi endpoint qua proxy, phát hiện challenge -> fallback farm"
  - "1.0h: testdata + parse_test.go - fixture JSON thật (ẩn danh) -> assert price/list_price/flash"
  - "1.0h: adapter_test.go - JSON hợp lệ -> snapshot; HTML challenge -> gọi fallback; is_login:false path"
  - "1.0h: registry.go - đăng ký adapter; integration với orchestrator pool"
  - "1.0h: OTel metric shopee_api_status_total{code} + shopee_fallback_total"

risk_if_skipped: "Shopee là sàn lớn nhất VN và là bề mặt MVP (§3 backlog). Không có adapter này thì không có nguồn giá Shopee - sale ảo (TASK-DEAL-001), biểu đồ (TASK-DEAL-003), so sánh chéo (TASK-PRICE-004) đều rỗng. Nếu gửi cookie người dùng kèm request backend là vi phạm nguyên tắc token-not-on-server (§3.8, §5.4) và biến SănDeal thành thứ bị nghi malware. Nếu không có fallback Playwright, ngày Shopee trả HTML challenge thay vì JSON là ngày dữ liệu Shopee đứng hình. Đây là adapter sàn đầu tiên và là khuôn mẫu cho TikTok/Lazada."
---

## §1 - Mô tả (BCP-14 normative)

Adapter Shopee **MUST** triển khai interface `PlatformAdapter` của TASK-SCRAPE-001, lấy giá qua internal JSON endpoint của Shopee khi truy cập được, parse về `price.PriceSnapshot`, và rơi xuống Playwright farm khi gặp challenge. Hợp đồng:

1. **MUST** triển khai `Fetch(ctx, job ScrapeJob) (price.PriceSnapshot, error)` và `PlatformID() int16` trả mã của Shopee.
2. **MUST** gọi endpoint `/api/v4/pdp/get_pc?item_id={itemid}&shop_id={shopid}` làm đường lấy giá chính (DEC-SCRAPE-06); `item_id`/`shop_id` lấy từ `tracked_product.platform_item_id` (TASK-PRICE-001).
3. **MUST** truy cập theo ngữ cảnh `is_login:false` (không phiên đăng nhập) - adapter backend KHÔNG dùng cookie phiên của bất kỳ người dùng nào (DEC-SCRAPE-09, §3.2). Yêu cầu đi qua proxy residential (TASK-SCRAPE-004) và fingerprint của farm, tách hẳn khỏi extension.
4. **MUST** parse JSON Shopee về `PriceSnapshot` (DEC-SCRAPE-07):
    - `price` <- trường giá hiện tại, quy đổi đơn vị Shopee (giá Shopee thường là micro-đơn-vị x100000) về BIGINT VND nguyên.
    - `list_price` <- giá gốc/niêm yết (before-discount), cùng quy đổi.
    - `stock`, `sold` <- tồn kho và đã bán nếu có.
    - `flash_sale` <- true khi payload báo đang flash sale (`flash_sale`/`is_flash_sale`/`upcoming_flash_sale` đang chạy).
5. **MUST** quy đổi giá KHÔNG dùng float trung gian: chia số nguyên `raw / 100000` (hoặc đọc trường đã là VND nếu Shopee trả vậy), tránh sai số float (đồng bộ DEC-PRICE-05 của TASK-PRICE-002).
6. **MUST** phát hiện challenge/anti-bot: nếu phản hồi không phải JSON hợp lệ (HTML, trang verify, HTTP 4xx/5xx của WAF), adapter **MUST** rơi xuống Playwright farm (TASK-SCRAPE-003) qua `DEC-SCRAPE-08` thay vì trả snapshot rỗng/sai.
7. **MUST** trả error (không panic) khi cả endpoint lẫn fallback đều thất bại, để orchestrator (TASK-SCRAPE-001 #5) áp retry/backoff.
8. **SHOULD** dùng `/api/v4/recommend/recommend` như nguồn phụ để khám phá SKU liên quan/đối thủ (phục vụ canonical_key của TASK-PRICE-005), nhưng đường giá chính là `pdp/get_pc`.
9. **MUST** xử lý đúng item không tồn tại / đã gỡ: Shopee trả `error != 0` hoặc thiếu `item` -> adapter trả lỗi phân loại `ErrItemGone`, để orchestrator hạ tier/ngừng quét thay vì retry vô ích.
10. **SHOULD** phát OTel metric: `shopee_api_status_total{http_code}` (counter), `shopee_fallback_total{reason}` (counter), `shopee_parse_fail_total` (counter).
11. **MUST** không lưu trữ phản hồi thô chứa dữ liệu cá nhân; chỉ trích `(price, list_price, stock, sold, flash_sale, ts)` rồi bỏ phần còn lại (tối thiểu hóa, §5.4).
12. **MUST** đặt `ts = time.Now()` (thời điểm quét) và `ProductID = job.ProductID` trên snapshot trước khi trả về orchestrator.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao ưu tiên internal JSON (DEC-SCRAPE-06)?** Gọi `/api/v4/pdp/get_pc` trả JSON cấu trúc là cách rẻ và nhanh nhất: không cần render trình duyệt, không tốn CPU farm, băng thông nhỏ. Playwright đắt hơn nhiều lần. Dùng JSON khi được, chỉ trả tiền cho farm khi buộc phải.

**Vì sao is_login:false và không cookie người dùng (DEC-SCRAPE-09)?** Đây là ranh giới đạo đức và pháp lý cốt lõi của SănDeal (§5.4). Scraping backend là hành vi của hệ thống, phải đứng bằng hạ tầng riêng (proxy + fingerprint farm), tuyệt đối không mượn phiên của người dùng. Extension đọc giỏ hàng client-side là chuyện khác (first-party, TASK-EXT-002); hai đường này không được trộn.

**Vì sao quy đổi giá bằng số nguyên (§1 #5)?** Giá Shopee trong payload thường là số rất lớn (micro-đơn-vị). Chia bằng float đưa sai số vào tiền tệ, làm hỏng phép so `current_price >= median90 * 0.97` của sale ảo (TASK-DEAL-001). Chia số nguyên giữ chính xác tuyệt đối, đồng bộ với BIGINT VND của TASK-PRICE-002.

**Vì sao bắt buộc fallback Playwright (DEC-SCRAPE-08)?** Shopee có anti-bot Medium-High (§3.9): chạy thư viện JS riêng + fingerprinting, đôi khi trả HTML challenge thay vì JSON. Nếu adapter coi HTML là "không có giá" và bỏ qua, dữ liệu Shopee sẽ thủng từng mảng mỗi khi sàn siết. Fallback giữ độ phủ.

**Vì sao phân loại ErrItemGone (§1 #9)?** Item bị gỡ sẽ luôn trả lỗi; retry mãi là lãng phí proxy và làm bẩn dead-letter. Phân biệt "lỗi tạm thời cần retry" với "item đã chết cần ngừng quét" giúp farm không tự đốt tài nguyên.

**Vì sao tối thiểu hóa phản hồi (§1 #11)?** Payload Shopee có thể chứa thông tin shop/seller/review. SănDeal chỉ cần các trường giá. Vứt phần còn lại ngay khi parse xong là cách rẻ nhất để tuân nguyên tắc tối thiểu hóa dữ liệu (§5.4) và giảm rủi ro lưu nhầm dữ liệu cá nhân.

---

## §3 - Hợp đồng API / DDL

### Endpoints (Go)

```go
// services/scrape/internal/adapters/shopee/endpoints.go
const (
    pdpPath       = "/api/v4/pdp/get_pc"
    recommendPath = "/api/v4/recommend/recommend"
)

// pdpURL dựng URL lấy giá chính cho một item Shopee.
func pdpURL(base string, itemID, shopID int64) string {
    return fmt.Sprintf("%s%s?item_id=%d&shop_id=%d&detail_level=0",
        base, pdpPath, itemID, shopID)
}
```

### Parse (Go)

```go
// services/scrape/internal/adapters/shopee/parse.go
const shopeePriceUnit = 100_000 // Shopee trả giá theo micro-đơn-vị

type pdpResponse struct {
    Error int `json:"error"`
    Data  struct {
        Item struct {
            Price          int64 `json:"price"`            // micro-VND
            PriceBeforeDisc int64 `json:"price_before_discount"`
            Stock          int32 `json:"stock"`
            HistoricalSold int32 `json:"historical_sold"`
            FlashSale      *struct{ Status int `json:"status"` } `json:"flash_sale"`
        } `json:"item"`
    } `json:"data"`
}

// toSnapshot quy đổi micro-VND -> BIGINT VND bằng phép chia số nguyên (không float).
func (r pdpResponse) toSnapshot(productID int64, ts time.Time) (price.PriceSnapshot, error) {
    if r.Error != 0 {
        return price.PriceSnapshot{}, ErrItemGone
    }
    it := r.Data.Item
    snap := price.PriceSnapshot{
        ProductID: productID,
        TS:        ts,
        Price:     it.Price / shopeePriceUnit,
        Stock:     &it.Stock,
        Sold:      &it.HistoricalSold,
        FlashSale: it.FlashSale != nil && it.FlashSale.Status == 1,
    }
    if it.PriceBeforeDisc > 0 {
        lp := it.PriceBeforeDisc / shopeePriceUnit
        snap.ListPrice = &lp
    }
    return snap, nil
}
```

### Adapter (Go)

```go
// services/scrape/internal/adapters/shopee/adapter.go
var ErrItemGone = errors.New("shopee: item removed or unavailable")

func (a *ShopeeAdapter) Fetch(ctx context.Context, job orchestrator.ScrapeJob) (price.PriceSnapshot, error) {
    itemID, shopID := splitRef(job)         // từ tracked_product.platform_item_id
    body, ct, err := a.http.GetViaProxy(ctx, pdpURL(a.base, itemID, shopID))
    if err == nil && isJSON(ct) {
        var resp pdpResponse
        if jsonErr := json.Unmarshal(body, &resp); jsonErr == nil {
            return resp.toSnapshot(job.ProductID, time.Now())
        }
        metrics.ShopeeParseFail()
    }
    // Challenge/HTML/lỗi WAF -> rơi xuống Playwright farm (TASK-SCRAPE-003)
    metrics.ShopeeFallback("challenge_or_parse")
    return a.farm.RenderPrice(ctx, job)
}
```

---

## §4 - Acceptance criteria

1. `ShopeeAdapter` thỏa interface `PlatformAdapter` (compile-time assertion `var _ orchestrator.PlatformAdapter = (*ShopeeAdapter)(nil)`).
2. `pdpURL` dựng đúng `/api/v4/pdp/get_pc?item_id=...&shop_id=...` từ ref của `tracked_product`.
3. Parse fixture JSON hợp lệ -> `Price` đúng VND (đã chia 100000), `ListPrice` đúng, `FlashSale` đúng theo `flash_sale.status`.
4. Quy đổi giá dùng phép chia số nguyên; với `price=89_000_00000` -> `Price=89_000` chính xác, không sai số.
5. `error != 0` trong payload -> `Fetch` trả `ErrItemGone` (không retry vô ích).
6. Phản hồi là HTML challenge (content-type không phải JSON) -> adapter gọi `farm.RenderPrice` (fallback Playwright).
7. Cả endpoint lẫn fallback fail -> `Fetch` trả error (orchestrator sẽ retry/backoff), không panic.
8. Adapter KHÔNG đính kèm cookie/token phiên người dùng vào request (kiểm qua header request trong test).
9. Snapshot trả về có `ProductID = job.ProductID` và `ts` ~ thời điểm quét.
10. Chỉ các trường giá được giữ; phản hồi thô không được lưu (kiểm không có field dư rò ra ngoài `PriceSnapshot`).
11. Metric `shopee_api_status_total`, `shopee_fallback_total` tăng đúng theo nhánh xử lý.
12. Tích hợp orchestrator: job Shopee qua pool -> adapter chạy -> `InsertSnapshot` delta-only được gọi (TASK-SCRAPE-001 #8).

---

## §5 - Kiểm thử (verification)

```go
// services/scrape/internal/adapters/shopee/parse_test.go
func TestParse_ValidPDP(t *testing.T) {
    raw := readFixture(t, "pdp_get_pc.json")
    var resp pdpResponse
    require.NoError(t, json.Unmarshal(raw, &resp))
    snap, err := resp.toSnapshot(90112, t0)
    require.NoError(t, err)
    require.Equal(t, int64(89_000), snap.Price)         // 89_000_00000 / 100000
    require.Equal(t, int64(149_000), *snap.ListPrice)
    require.True(t, snap.FlashSale)
}

func TestParse_IntegerDivision_NoFloatError(t *testing.T) {
    resp := pdpResponse{}
    resp.Data.Item.Price = 333_333_00000   // 333_333 VND chính xác
    snap, _ := resp.toSnapshot(1, t0)
    require.Equal(t, int64(333_333), snap.Price)
}

func TestParse_ItemGone(t *testing.T) {
    resp := pdpResponse{Error: 4}
    _, err := resp.toSnapshot(1, t0)
    require.ErrorIs(t, err, ErrItemGone)
}
```

```go
// services/scrape/internal/adapters/shopee/adapter_test.go
var _ orchestrator.PlatformAdapter = (*ShopeeAdapter)(nil)

func TestFetch_HTMLChallenge_FallsBackToFarm(t *testing.T) {
    a := newAdapter(t, stubHTTP("<html>verify</html>", "text/html"))
    called := false
    a.farm = farmFunc(func(ctx context.Context, j orchestrator.ScrapeJob) (price.PriceSnapshot, error) {
        called = true
        return price.PriceSnapshot{ProductID: j.ProductID, Price: 1}, nil
    })
    _, err := a.Fetch(ctx, job(90112))
    require.NoError(t, err)
    require.True(t, called) // đã rơi xuống Playwright
}

func TestFetch_NoUserCookieSent(t *testing.T) {
    var sentCookie string
    a := newAdapter(t, captureHeader("Cookie", &sentCookie))
    a.Fetch(ctx, job(90112))
    require.Empty(t, sentCookie) // backend không mượn phiên người dùng
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: endpoints.go (URL builder) -> parse.go (JSON -> snapshot, test trước với fixture) -> adapter.go (Fetch + fallback) -> registry.go (đăng ký vào orchestrator) -> tests. Fixture `pdp_get_pc.json` là JSON thật đã ẩn danh (xóa field shop/seller/review, chỉ giữ cấu trúc giá). Đường lấy giá chính là `pdp/get_pc`; `recommend` là tùy chọn phụ cho TASK-PRICE-005.

---

## §7 - Phụ thuộc

- **TASK-SCRAPE-001** - interface `PlatformAdapter`, orchestrator điều phối + gọi `InsertSnapshot`.
- **TASK-SCRAPE-003 (fallback)** - Playwright farm `RenderPrice` khi gặp challenge.
- **TASK-SCRAPE-004** - proxy residential cho `GetViaProxy`.
- **TASK-PRICE-001** - `tracked_product.platform_item_id` cung cấp `item_id`/`shop_id`.
- **TASK-PRICE-002 (downstream)** - snapshot ghi qua `InsertSnapshot` delta-only.
- **TASK-PRICE-005 (tùy chọn)** - `recommend` hỗ trợ canonical_key matching.

---

## §8 - Payload ví dụ

### Phản hồi /api/v4/pdp/get_pc (ẩn danh, chỉ phần giá)

```json
{
  "error": 0,
  "data": {
    "item": {
      "price": 8900000000,
      "price_before_discount": 14900000000,
      "stock": 37,
      "historical_sold": 1240,
      "flash_sale": { "status": 1 }
    }
  }
}
```

### Snapshot adapter trả về orchestrator

```go
price.PriceSnapshot{
    ProductID: 90112,
    TS:        time.Now(),
    Price:     89_000,        // VND (8900000000 / 100000)
    ListPrice: ptr(int64(149_000)),
    Stock:     ptr(int32(37)),
    Sold:      ptr(int32(1240)),
    FlashSale: true,
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Đơn vị giá chính xác của Shopee (x100000 hay đọc trường VND trực tiếp) - xác minh trên fixture thật lúc build; hằng `shopeePriceUnit` tập trung một chỗ để chỉnh.
- Tham số ký/anti-bot bổ sung cho `pdp/get_pc` nếu Shopee siết là-login-only - khi đó fallback farm gánh nhiều hơn.
- Khai thác `recommend` để dò biến thể (màu/size) cùng SKU - gắn khi TASK-PRICE-005.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Shopee trả HTML challenge | content-type không JSON | mất giá nếu bỏ qua | Fallback Playwright (§1 #6) |
| Đơn vị giá đổi (x100000 sai) | parse_test trên fixture | giá sai 10^5 lần | Hằng `shopeePriceUnit` một chỗ, test khóa |
| Item đã gỡ | `error != 0` | retry vô ích | `ErrItemGone` -> orchestrator ngừng quét |
| Gửi nhầm cookie người dùng | test `NoUserCookieSent` | vi phạm §5.4 | Adapter backend không bao giờ đính cookie |
| Float trong quy đổi | parse_test | sai số tiền tệ | Phép chia số nguyên (§1 #5) |
| JSON đổi schema (Shopee A/B) | `shopee_parse_fail_total` | parse hụt | TASK-SCRAPE-006 giám sát + cập nhật struct |
| WAF rate-limit 429 | `shopee_api_status_total` | bị siết | Fallback + orchestrator backoff + proxy rotation |
| Lưu nhầm phản hồi thô | review code | rủi ro dữ liệu cá nhân | Chỉ trích trường giá, vứt phần còn lại (§1 #11) |
| Fallback farm cũng fail | error trả về | quét hụt lần này | Orchestrator retry/backoff (TASK-SCRAPE-001 #5) |

---

## §11 - Ghi chú

- Adapter Shopee là khuôn mẫu cho TikTok (TASK-SCRAPE-007) và Lazada (TASK-SCRAPE-008): cùng interface, khác cách lấy dữ liệu.
- Ranh giới cứng: scraping backend (proxy + fingerprint farm, is_login:false) tách hoàn toàn khỏi extension đọc giỏ hàng (first-party). Không trộn phiên người dùng vào backend.
- Quy đổi giá bằng số nguyên đồng bộ BIGINT VND của TASK-PRICE-002 - không sai số float trên tiền tệ.
- Fallback Playwright là bảo hiểm độ phủ trước anti-bot Medium-High của Shopee (§3.9); JSON là đường nhanh-rẻ mặc định.
- Đường giá chính là `pdp/get_pc`; `recommend` để mở rộng độ phủ SKU và phục vụ so sánh chéo sàn về sau.

---

*Hết TASK-SCRAPE-002. Status: ready_to_implement (mục tiêu audit 10/10).*
