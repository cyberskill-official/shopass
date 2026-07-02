---
id: FR-TRACK-001
title: "POST /v1/track - parse item_url theo sàn -> upsert tracked_product -> enqueue scrape job: API người dùng bắt đầu theo dõi một SKU, idempotent một lần track không nhân bản"
module: TRACK
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-PRICE-001, FR-SCRAPE-002, FR-TRACK-002, FR-TRACK-003, FR-INFRA-001]
depends_on: [FR-PRICE-001, FR-SCRAPE-002]
blocks: [FR-TRACK-002, FR-TRACK-003]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.7 (POST /v1/track {platform, item_url} -> tạo tracked_product)"
  - "docs/... §3.4 (tracked_product), §5.1 (cold-start - track sớm để tích lũy lịch sử giá)"
source_decisions:
  - "DEC-TRACK-01: POST /v1/track nhận {platform, item_url}, parse item_url ra platform_item_id + shop_id theo từng sàn, KHÔNG tin client gửi sẵn id"
  - "DEC-TRACK-02: tạo tracked_product qua Upsert idempotent của FR-PRICE-001 - một SKU track nhiều lần (kể cả nhiều user) chỉ một dòng"
  - "DEC-TRACK-03: sau khi có tracked_product, enqueue một scrape job ưu tiên (priming) để có snapshot đầu tiên ngay, giải cold-start §5.1"
  - "DEC-TRACK-04: gắn user_id (từ JWT của gateway FR-INFRA-001) vào tracked_product qua bảng nối user_tracked_product - một user track cùng SKU chỉ một lần"
  - "DEC-TRACK-05: platform trong body phải khớp allowlist code sàn (shopee/tiktok/lazada); url không parse được trả 400, không enqueue gì"

language: "Go 1.22 (track-svc); PostgreSQL 16"
service: shopass/services/track/
new_files:
  - services/track/migrations/0001_user_tracked_product.sql
  - services/track/internal/api/track.go
  - services/track/internal/track/url_parser.go
  - services/track/internal/track/repo.go
  - services/track/internal/track/url_parser_test.go
  - services/track/internal/api/track_test.go
modified_files:
  - services/track/internal/api/router.go            # đăng ký route POST /v1/track
allowed_tools:
  - file_read: services/track/**
  - file_read: services/price/**
  - file_write: services/track/**
  - bash: cd services/track && go test ./...
disallowed_tools:
  - tin client gửi sẵn platform_item_id thay vì parse từ item_url (vi phạm DEC-TRACK-01, nhận id rác/giả mạo)
  - tạo tracked_product trực tiếp thay vì gọi Upsert của FR-PRICE-001 (vi phạm DEC-TRACK-02, nhân bản SKU)
  - bỏ qua bước enqueue scrape job priming (vi phạm DEC-TRACK-03, SKU mới không có snapshot, biểu đồ rỗng)

effort_hours: 5
sub_tasks:
  - "1.0h: url_parser.go - parse item_url Shopee/TikTok/Lazada ra (platform_item_id, shop_id); regex per-sàn + validate"
  - "0.5h: 0001_user_tracked_product.sql - bảng nối user_tracked_product + UNIQUE(user_id, product_id)"
  - "1.0h: repo.go - LinkUserProduct (ON CONFLICT DO NOTHING) + đọc lại tracked_product qua Upsert của price-svc"
  - "1.0h: track.go - handler: validate platform, parse url, Upsert tracked_product, link user, enqueue scrape job, 201/400"
  - "0.5h: router.go - đăng ký route sau JWT middleware (FR-INFRA-001)"
  - "1.0h: url_parser_test.go + track_test.go - 6 test (parse 3 sàn, url xấu 400, platform lạ 400, idempotent track 2 lần, enqueue gọi đúng)"

risk_if_skipped: "POST /v1/track là cửa duy nhất để người dùng đưa một sản phẩm vào hệ thống theo dõi - không có nó thì wishlist (FR-TRACK-002), alert (FR-TRACK-003) và biểu đồ giá không có thực thể sản phẩm nào của người dùng để gắn vào, toàn bộ vòng giá trị theo dõi giá đứng. Nếu tin client gửi sẵn platform_item_id thay vì parse từ item_url thì hệ thống nhận id rác hoặc id giả mạo trỏ tới SKU người khác, làm hỏng quan hệ user-sản phẩm. Nếu tạo tracked_product thẳng thay vì qua Upsert idempotent của FR-PRICE-001 thì mỗi lần track nhân bản một dòng, registry phình và price_snapshot trỏ id rác. Nếu bỏ bước enqueue scrape job priming thì SKU mới track không có snapshot đầu tiên, người dùng vào xem thấy biểu đồ rỗng đúng lúc họ vừa quan tâm - phá trải nghiệm cold-start (§5.1)."
---

## §1 - Mô tả (BCP-14 normative)

Service TRACK **MUST** expose endpoint REST `POST /v1/track` nhận `{platform, item_url}`, parse `item_url` theo từng sàn ra định danh SKU, upsert vào `tracked_product` (qua FR-PRICE-001), gắn người dùng vào SKU đó, rồi enqueue một scrape job priming để có snapshot giá đầu tiên. Hợp đồng:

1. **MUST** phục vụ route `POST /v1/track` với thân JSON `{platform: string, item_url: string}`. Thiếu trường bắt buộc trả `400`.
2. **MUST** validate `platform` thuộc allowlist code sàn `{shopee, tiktok, lazada}` (DEC-TRACK-05); ngoài allowlist trả `400` với thân `{"error":"unsupported platform"}`. KHÔNG enqueue gì khi validate thất bại.
3. **MUST** parse `item_url` theo từng sàn ra `(platform_item_id, shop_id)` (DEC-TRACK-01): handler tự bóc id từ url, KHÔNG nhận `platform_item_id` do client gửi sẵn. URL không khớp mẫu của `platform` trả `400` với thân `{"error":"invalid item_url"}`.
4. **MUST** map `platform` code sang `platform_id` (SMALLINT) qua bảng `platform` (FR-INFRA-002) trước khi upsert; code hợp lệ nhưng chưa seed trong `platform` trả `400`.
5. **MUST** tạo hoặc lấy `tracked_product` qua `Upsert` idempotent của FR-PRICE-001 (DEC-TRACK-02): cùng `(platform_id, platform_item_id)` chỉ một dòng dù nhiều user track. Handler **MUST NOT** ghi `tracked_product` bằng INSERT trần.
6. **MUST** gắn người dùng vào SKU qua bảng nối `user_tracked_product (user_id, product_id, tracked_at)` với `UNIQUE (user_id, product_id)` (DEC-TRACK-04); link idempotent qua `ON CONFLICT (user_id, product_id) DO NOTHING`. Một user track cùng SKU nhiều lần vẫn một dòng nối.
7. **MUST** lấy `user_id` từ JWT do API Gateway (FR-INFRA-001) gắn; handler KHÔNG tự parse token. Request thiếu auth bị gateway chặn trước.
8. **MUST** enqueue một scrape job priming sau khi có `tracked_product.id` (DEC-TRACK-03): job ưu tiên để scraper (FR-SCRAPE-002) lấy snapshot giá đầu tiên ngay, không chờ vòng quét định kỳ. Enqueue đặt sau khi link user thành công.
9. **MUST** trả `201 Created` với thân `{product_id, platform, already_tracked}`; `already_tracked=true` khi user đã track SKU này từ trước (link là no-op), `false` khi vừa tạo link mới.
10. **MUST** đảm bảo thứ tự bền vững: nếu enqueue scrape job lỗi sau khi đã upsert + link, handler **MUST** vẫn trả `201` (sản phẩm đã được theo dõi) và ghi log lỗi enqueue để retry nền - không lăn ngược việc link user.
11. **MUST** đặt `Content-Type: application/json; charset=utf-8`; giá (nếu trả kèm) là `BIGINT` VND (int64), đồng nhất với DEC-PRICE-05.
12. **SHOULD** phát OTel: `track_requests_total{platform, status}` (counter), `track_new_product_total{platform}` (counter đếm SKU lần đầu vào hệ thống), `track_scrape_enqueue_total{result}` (counter).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao parse item_url phía server thay vì nhận id sẵn (DEC-TRACK-01)?** Người dùng dán một đường link sản phẩm họ copy từ app sàn - đó là input tự nhiên. Nếu để client tự bóc `platform_item_id` rồi gửi, ta tin một giá trị client kiểm soát: client lỗi gửi id sai, hoặc kẻ xấu gửi id trỏ tới SKU bất kỳ để bơm dữ liệu rác. Parse url phía server cho ta một nguồn sự thật, đồng thời chuẩn hóa được các biến thể url (có/không query tracking, short link) về cùng một id.

**Vì sao upsert qua FR-PRICE-001 thay vì INSERT (DEC-TRACK-02)?** Nhiều user có thể cùng theo dõi một SKU hot. SKU là thực thể dùng chung, không thuộc riêng user nào. `Upsert` idempotent của FR-PRICE-001 đảm bảo đúng một dòng `tracked_product` cho mỗi SKU; quan hệ "ai theo dõi" nằm ở bảng nối riêng. Tách vậy giữ registry sản phẩm sạch và cho phép chia sẻ chi phí scraping (một SKU quét một lần phục vụ mọi người theo dõi).

**Vì sao enqueue scrape job priming (DEC-TRACK-03)?** Cold-start là rủi ro đã nêu (§5.1): SKU mới chưa có lịch sử giá. Nếu chỉ chờ vòng quét định kỳ, người dùng vừa track xong vào xem ngay sẽ thấy biểu đồ rỗng - đúng lúc họ quan tâm nhất. Một job priming ưu tiên lấy snapshot đầu tiên trong vài giây, cho biểu đồ ít nhất một điểm và khởi động chuỗi giá.

**Vì sao bảng nối user_tracked_product với UNIQUE (DEC-TRACK-04)?** Quan hệ user-SKU là nhiều-nhiều: một user theo dõi nhiều SKU, một SKU được nhiều user theo dõi. Bảng nối là cách chuẩn. `UNIQUE (user_id, product_id)` làm việc track idempotent - bấm track hai lần không tạo hai link, và là khóa `ON CONFLICT` tự nhiên.

**Vì sao vẫn trả 201 khi enqueue lỗi (§1 #10)?** Hành động người dùng yêu cầu là "theo dõi sản phẩm này", và điều đó đã hoàn thành khi upsert + link xong. Snapshot priming là tối ưu trải nghiệm, không phải điều kiện thành công. Lăn ngược cả việc track chỉ vì hàng đợi scrape tạm trục trặc là phản trực giác; thay vào đó ghi log để retry nền và để vòng quét định kỳ bù.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/track/migrations/0001_user_tracked_product.sql
CREATE TABLE user_tracked_product (
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  product_id BIGINT      NOT NULL REFERENCES tracked_product(id),
  tracked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, product_id)
);

-- Liệt kê nhanh mọi SKU một user đang theo dõi (cho wishlist/alert/dashboard)
CREATE INDEX idx_utp_user ON user_tracked_product (user_id);
-- Đếm nhanh số người theo dõi một SKU (chia sẻ chi phí scraping, ưu tiên tần suất quét)
CREATE INDEX idx_utp_product ON user_tracked_product (product_id);
```

### Parser url theo sàn (Go)

```go
// services/track/internal/track/url_parser.go

// ParsedItem là kết quả bóc từ item_url.
type ParsedItem struct {
    PlatformItemID string
    ShopID         string // có thể rỗng nếu sàn không nhúng shop trong url
}

// shopeeItemRe khớp .../<shop_id>.<item_id> hoặc ...-i.<shop_id>.<item_id>
var shopeeItemRe = regexp.MustCompile(`i\.(\d+)\.(\d+)`)
var lazadaItemRe = regexp.MustCompile(`/products/.*-i(\d+)`)
var tiktokItemRe = regexp.MustCompile(`/product/(\d+)`)

// ParseItemURL bóc (platform_item_id, shop_id) theo từng sàn (DEC-TRACK-01).
// Trả ok=false nếu url không khớp mẫu của platform -> handler trả 400.
func ParseItemURL(platform, rawURL string) (ParsedItem, bool) {
    u, err := url.Parse(rawURL)
    if err != nil || u.Host == "" {
        return ParsedItem{}, false
    }
    switch platform {
    case "shopee":
        m := shopeeItemRe.FindStringSubmatch(u.Path)
        if m == nil {
            return ParsedItem{}, false
        }
        return ParsedItem{ShopID: m[1], PlatformItemID: m[2]}, true
    case "lazada":
        m := lazadaItemRe.FindStringSubmatch(u.Path)
        if m == nil {
            return ParsedItem{}, false
        }
        return ParsedItem{PlatformItemID: m[1]}, true
    case "tiktok":
        m := tiktokItemRe.FindStringSubmatch(u.Path)
        if m == nil {
            return ParsedItem{}, false
        }
        return ParsedItem{PlatformItemID: m[1]}, true
    default:
        return ParsedItem{}, false
    }
}
```

### Handler (Go)

```go
// services/track/internal/api/track.go

type TrackRequest struct {
    Platform string `json:"platform"`
    ItemURL  string `json:"item_url"`
}
type TrackResponse struct {
    ProductID      int64  `json:"product_id"`
    Platform       string `json:"platform"`
    AlreadyTracked bool   `json:"already_tracked"`
}

func (h *Handler) HandleTrack(w http.ResponseWriter, req *http.Request) {
    userID := auth.UserID(req.Context()) // do gateway gắn (FR-INFRA-001)
    var body TrackRequest
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid body")
        return
    }
    platformID, ok := h.platforms.IDByCode(body.Platform) // map code -> platform_id
    if !ok {
        writeErr(w, http.StatusBadRequest, "unsupported platform")
        return
    }
    item, ok := track.ParseItemURL(body.Platform, body.ItemURL)
    if !ok {
        writeErr(w, http.StatusBadRequest, "invalid item_url")
        return
    }
    tp, err := h.price.Upsert(req.Context(), price.TrackedProduct{ // FR-PRICE-001
        PlatformID:     platformID,
        PlatformItemID: item.PlatformItemID,
        ShopID:         nilIfEmpty(item.ShopID),
    })
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    linkedNew, err := h.repo.LinkUserProduct(req.Context(), userID, tp.ID)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    if err := h.scrapeQueue.EnqueuePriming(req.Context(), tp.ID); err != nil {
        log.Warn("priming enqueue failed", "product_id", tp.ID, "err", err) // §1 #10: không lăn ngược
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(TrackResponse{
        ProductID: tp.ID, Platform: body.Platform, AlreadyTracked: !linkedNew,
    })
}
```

---

## §4 - Acceptance criteria

1. `POST /v1/track {platform:"shopee", item_url:"https://shopee.vn/...-i.88123.20114455667"}` trả `201` + `product_id > 0`.
2. URL Shopee parse đúng `shop_id=88123`, `platform_item_id=20114455667`; Lazada `/products/...-i7788` -> `7788`; TikTok `/product/990011` -> `990011`.
3. `platform` ngoài allowlist (vd `"amazon"`) trả `400` + `{"error":"unsupported platform"}`, KHÔNG enqueue.
4. `item_url` không khớp mẫu của `platform` trả `400` + `{"error":"invalid item_url"}`.
5. Track cùng SKU lần hai bởi cùng user trả `201` + `already_tracked=true`; số dòng `user_tracked_product` vẫn 1.
6. Hai user khác nhau track cùng SKU -> một dòng `tracked_product`, hai dòng `user_tracked_product`.
7. `tracked_product` được tạo qua `Upsert` (cùng `(platform_id, platform_item_id)` không nhân bản); xác nhận chỉ một dòng.
8. Sau track thành công, một scrape job priming được enqueue đúng một lần với `product_id` vừa có.
9. Enqueue scrape lỗi (mock queue fail) -> handler vẫn trả `201`, link user vẫn tồn tại, log cảnh báo được ghi.
10. Request thiếu JWT bị gateway chặn (handler không bao giờ thấy); handler không tự xác thực.
11. `platform` hợp lệ nhưng chưa seed trong bảng `platform` -> `400`.
12. Metric `track_new_product_total` tăng khi SKU lần đầu vào hệ thống; `track_requests_total{status}` tăng theo mỗi phản hồi.

---

## §5 - Kiểm thử (verification)

```go
// services/track/internal/track/url_parser_test.go
func TestParse_Shopee(t *testing.T) {
    it, ok := ParseItemURL("shopee", "https://shopee.vn/Tai-nghe-i.88123.20114455667?sp_atk=x")
    require.True(t, ok)
    require.Equal(t, "88123", it.ShopID)
    require.Equal(t, "20114455667", it.PlatformItemID)
}

func TestParse_Lazada(t *testing.T) {
    it, ok := ParseItemURL("lazada", "https://www.lazada.vn/products/abc-pro-i7788.html")
    require.True(t, ok)
    require.Equal(t, "7788", it.PlatformItemID)
}

func TestParse_TikTok(t *testing.T) {
    it, ok := ParseItemURL("tiktok", "https://www.tiktok.com/view/product/990011")
    require.True(t, ok)
    require.Equal(t, "990011", it.PlatformItemID)
}

func TestParse_BadURL(t *testing.T) {
    _, ok := ParseItemURL("shopee", "https://shopee.vn/no-item-id-here")
    require.False(t, ok) // -> handler trả 400
}

// services/track/internal/api/track_test.go
func TestTrack_NewProduct_201(t *testing.T) {
    h, q := setupHandler(t)
    rec := doPOST(t, h, "/v1/track",
        `{"platform":"shopee","item_url":"https://shopee.vn/x-i.88123.20114455667"}`)
    require.Equal(t, 201, rec.Code)
    require.Equal(t, 1, q.PrimingCount()) // đúng một job priming
}

func TestTrack_UnsupportedPlatform_400(t *testing.T) {
    h, q := setupHandler(t)
    rec := doPOST(t, h, "/v1/track", `{"platform":"amazon","item_url":"https://x"}`)
    require.Equal(t, 400, rec.Code)
    require.Equal(t, 0, q.PrimingCount()) // không enqueue khi validate fail
}

func TestTrack_Idempotent_SameUser(t *testing.T) {
    h, _ := setupHandler(t)
    body := `{"platform":"shopee","item_url":"https://shopee.vn/x-i.88123.20114455667"}`
    doPOST(t, h, "/v1/track", body)
    rec := doPOST(t, h, "/v1/track", body)
    var resp TrackResponse
    decode(t, rec, &resp)
    require.True(t, resp.AlreadyTracked)
    require.Equal(t, 1, countLinks(t, h)) // không nhân bản link
}

func TestTrack_EnqueueFails_Still201(t *testing.T) {
    h, q := setupHandler(t)
    q.FailNext() // queue từ chối job kế
    rec := doPOST(t, h, "/v1/track",
        `{"platform":"shopee","item_url":"https://shopee.vn/x-i.88123.20114455667"}`)
    require.Equal(t, 201, rec.Code)       // §1 #10: sản phẩm vẫn được theo dõi
    require.Equal(t, 1, countLinks(t, h)) // link user không bị lăn ngược
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `url_parser.go` (parse 3 sàn) -> migration `0001_user_tracked_product.sql` -> `repo.go` (`LinkUserProduct` ON CONFLICT DO NOTHING) -> `track.go` (handler ghép parse + Upsert price-svc + link + enqueue) -> đăng ký route trong `router.go` sau JWT middleware (FR-INFRA-001) -> tests. Handler dùng `http.ServeMux` Go 1.22. `EnqueuePriming` đẩy vào cùng hàng đợi scrape của FR-SCRAPE-001/002 với cờ ưu tiên. `Upsert` gọi sang package `price` (cùng monorepo) hoặc qua client nội bộ - không tự ghi `tracked_product`.

---

## §7 - Phụ thuộc

- **FR-PRICE-001** - cung cấp `Upsert(tracked_product)` idempotent và bảng `tracked_product`; là điều kiện cứng cho việc tạo/lấy SKU.
- **FR-SCRAPE-002** - tiêu thụ scrape job priming, lấy snapshot giá đầu tiên của SKU mới track.
- **FR-INFRA-001 (gateway)** - gắn JWT auth và `user_id` vào context trước handler.
- **FR-INFRA-002** - bảng `platform` để map code sàn -> `platform_id`; bảng `app_user` cho FK `user_id`.
- **FR-TRACK-002 / FR-TRACK-003 (downstream)** - wishlist và alert_rule gắn vào `tracked_product` đã track ở đây.
- Lib: `pgx`, `encoding/json`, `net/http`, `regexp`, `net/url`.

---

## §8 - Payload ví dụ

### Request

```
curl -s -X POST -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{"platform":"shopee","item_url":"https://shopee.vn/Tai-nghe-ABC-i.88123.20114455667"}' \
  "https://api.sandeal.vn/v1/track"
```

### Response (201)

```json
{
  "product_id": 90112,
  "platform": "shopee",
  "already_tracked": false
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Resolve short link (vd `s.shopee.vn/...`) bằng một HEAD request theo redirect trước khi parse - thêm khi gặp nhiều short link trong thực tế; không cản FR này.
- Giới hạn số SKU theo dõi theo tier (free vs Premium) - gắn vào FR-BILL khi có gating.
- Track theo `canonical_key` để tự theo dõi cùng món trên cả 3 sàn một lần - chờ FR-PRICE-005 so khớp; mở rộng sau.
- Bỏ theo dõi (`DELETE /v1/track/{product_id}`) - thêm khi làm UI quản lý wishlist (FR-WEB-004).

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| platform ngoài allowlist | validate code | 400 unsupported platform | UI chỉ cho chọn 3 sàn đã hỗ trợ |
| item_url không khớp mẫu | parser ok=false | 400 invalid item_url | Hướng dẫn người dùng dán đúng link sản phẩm |
| platform_item_id client gửi giả | parse server-side | bỏ qua, chỉ tin url | Không nhận id từ client (DEC-TRACK-01) |
| Trùng track cùng user | ON CONFLICT DO NOTHING | already_tracked=true | Theo thiết kế (idempotent) |
| Nhiều user cùng SKU | Upsert + bảng nối | một SKU, nhiều link | Theo thiết kế (chia sẻ SKU) |
| Enqueue scrape lỗi | log warn | vẫn 201, không có snapshot ngay | Retry nền + vòng quét định kỳ bù |
| platform chưa seed | IDByCode false | 400 | Seed platform (FR-INFRA-002) chạy trước |
| Short link chưa resolve | parser ok=false | 400 | Hoãn: resolve redirect (xem §9) |
| Race hai request track cùng SKU+user | ON CONFLICT | một link, không lỗi | Theo thiết kế (idempotent) |

---

## §11 - Ghi chú

- `POST /v1/track` là cửa vào duy nhất của vòng theo dõi giá: parse url -> SKU dùng chung -> link user -> priming snapshot.
- Parse url phía server là rào an toàn: client chỉ đưa link, server quyết định id, tránh id rác và giả mạo.
- Tách `tracked_product` (dùng chung) khỏi `user_tracked_product` (riêng user) cho phép chia sẻ chi phí scraping: một SKU hot quét một lần phục vụ mọi người theo dõi.
- Scrape job priming là liều thuốc cold-start ở mức từng SKU: người dùng vừa track có ngay điểm giá đầu tiên thay vì biểu đồ rỗng.
- Việc track không bị lăn ngược chỉ vì enqueue trục trặc - hành động người dùng (theo dõi) và tối ưu trải nghiệm (priming) tách rời nhau.
- Khi mở SEA, thêm mẫu parse url cho sàn nước khác vào `ParseItemURL`; phần còn lại của handler không đổi.

---

*Hết FR-TRACK-001. Status: ready_to_implement (mục tiêu audit 10/10).*
