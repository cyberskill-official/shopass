---
id: FR-AFFIL-002
title: "POST /v1/affiliate/link - deep link affiliate CHỈ tạo khi user chủ động bấm 'Mua qua SănDeal'; hiển thị link đích rõ + disclosure; KHÔNG auto-cookie nền (né Honey)"
module: AFFIL
priority: MUST
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-AFFIL-001, FR-AFFIL-003, FR-AFFIL-004, FR-INFRA-001, FR-EXT-003, NFR-AFFIL-001]
depends_on: [FR-AFFIL-001]
blocks: [FR-AFFIL-003, FR-AFFIL-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.7 (POST /v1/affiliate/link {product_id} -> deep link affiliate, CHỈ khi user bấm)"
  - "docs/... §4.2 (mô hình affiliate hợp lệ duy nhất: user chủ động bấm -> deep link disclosure; KHÔNG cookie-stuffing Honey-style)"
source_decisions:
  - "DEC-AFFIL-06: endpoint CHỈ trả deep link khi request mang cờ chủ ý user (user-initiated); không có chế độ tạo link nền/tự động"
  - "DEC-AFFIL-07: mỗi lần tạo link sinh một sub_id mới + ghi affiliate_click (FR-AFFIL-001) - link và sổ ghi click là một hành động nguyên tử"
  - "DEC-AFFIL-08: response BẮT BUỘC kèm disclosure (SănDeal có thể nhận hoa hồng) + URL đích hiển thị rõ; client phải cho user thấy đích trước khi điều hướng"
  - "DEC-AFFIL-09: server KHÔNG tự set/sửa cookie affiliate trên domain sàn; chỉ trả URL deep link để client mở trong tab user bấm (last-click hợp lệ)"
  - "DEC-AFFIL-10: deep link build theo network template (FR-AFFIL-003) với sub_id nhúng làm tham số tracking; thiếu mapping network -> 503, không bịa link"

language: "Go 1.22 (affil-svc); PostgreSQL 16; REST qua API Gateway (FR-INFRA-001)"
service: shopass/services/affil/
new_files:
  - services/affil/internal/api/link.go
  - services/affil/internal/affil/deeplink.go
  - services/affil/internal/affil/disclosure.go
  - services/affil/internal/api/link_test.go
  - services/affil/internal/affil/deeplink_test.go
modified_files:
  - services/affil/internal/api/router.go            # đăng ký POST /v1/affiliate/link sau JWT middleware
allowed_tools:
  - file_read: services/affil/**
  - file_write: services/affil/**
  - bash: cd services/affil && go test ./...
disallowed_tools:
  - tạo link/ghi click khi không có cờ user-initiated (vi phạm DEC-AFFIL-06, NFR-AFFIL-001 - Honey-style)
  - trả response thiếu disclosure hoặc giấu URL đích (vi phạm DEC-AFFIL-08)
  - server tự set/sửa cookie affiliate trên domain sàn (vi phạm DEC-AFFIL-09)
  - bịa deep link khi thiếu mapping network (vi phạm DEC-AFFIL-10)

effort_hours: 6
sub_tasks:
  - "1.0h: deeplink.go - build URL deep link theo network template + nhúng sub_id; trả lỗi nếu thiếu template"
  - "0.5h: disclosure.go - văn bản disclosure tiếng Việt + cấu trúc response (deep_link, target_url, disclosure)"
  - "1.5h: link.go - handler: bắt buộc cờ user-initiated, lấy user_id từ JWT, sinh sub_id, RecordClick, build link, trả disclosure"
  - "0.5h: router.go - đăng ký route sau JWT middleware (FR-INFRA-001)"
  - "1.0h: link_test.go - thiếu cờ user-initiated -> 400 và KHÔNG ghi click; happy path 200 có disclosure + target_url; product lạ -> 404"
  - "1.0h: deeplink_test.go - sub_id nhúng đúng tham số; thiếu network template -> 503; URL đích là domain sàn hợp lệ"
  - "0.5h: OTel metric affiliate_link_created_total{platform,network} + affiliate_link_rejected_total{reason}"

risk_if_skipped: "Đây là điểm thực thi kỹ thuật của lằn ranh đạo đức trung tâm SănDeal (§4.2). Mô hình affiliate hợp lệ DUY NHẤT là user chủ động bấm 'Mua qua SănDeal' -> deep link hiển thị rõ + disclosure. Nếu endpoint cho phép tạo link/ghi click mà không có chủ ý user, SănDeal rơi đúng vào mô hình cookie-stuffing kiểu Honey: bị Chrome Web Store gỡ (chính sách 3/2025 thực thi 10/06/2025), bị network đình chỉ (chuỗi Rakuten/Impact/Awin tháng 01/2026), và mất moat niềm tin hậu-Honey. Nếu response giấu URL đích hoặc thiếu disclosure, người dùng không biết mình đi đâu và SănDeal hưởng lợi - vi phạm yêu cầu minh bạch. Nếu server tự set cookie affiliate trên domain sàn thì đó là chèn cookie nền - bất hợp pháp theo policy mới. Đây là FR mà nếu làm sai sẽ phá hủy toàn bộ định vị sản phẩm."
---

## §1 - Mô tả (BCP-14 normative)

Service AFFIL **MUST** expose `POST /v1/affiliate/link` chỉ trả một deep link affiliate khi request là user-initiated (user chủ động bấm "Mua qua SănDeal"), kèm disclosure và URL đích hiển thị rõ, và ghi một `affiliate_click` nguyên tử với hành động đó. Server **MUST NOT** tự chèn/sửa cookie affiliate nền. Hợp đồng:

1. **MUST** phục vụ `POST /v1/affiliate/link` với thân `{product_id: int64, user_initiated: true}`. Trường `user_initiated` **MUST** là `true`; nếu thiếu hoặc `false` -> trả `400` với `{"error":"link requires explicit user action"}` và **MUST NOT** ghi `affiliate_click` (DEC-AFFIL-06, NFR-AFFIL-001).
2. **MUST** lấy `user_id` từ JWT do API Gateway (FR-INFRA-001) gắn; handler KHÔNG tự parse token. Request thiếu auth bị gateway chặn trước.
3. **MUST** tra `product_id` trong `tracked_product` để lấy `platform_id` + định danh sản phẩm sàn; `product_id` không tồn tại -> `404` `{"error":"product not found"}`, không ghi click.
4. **MUST** sinh một `sub_id` mới (FR-AFFIL-001 `subid.go`) cho mỗi lần tạo link, và ghi `affiliate_click` qua `RecordClick` (FR-AFFIL-001) như một hành động nguyên tử với việc trả link (DEC-AFFIL-07). Nếu `RecordClick` lỗi -> `500`, không trả link (không trả link mà không ghi, và không ghi mà không có chủ ý user).
5. **MUST** build deep link theo template của `network` cho sàn đó (FR-AFFIL-003), nhúng `sub_id` làm tham số tracking (DEC-AFFIL-10). Thiếu mapping network cho sàn -> `503` `{"error":"affiliate network unavailable"}`, KHÔNG bịa link tự ghép tay.
6. Response thành công (`200`) **MUST** chứa: `deep_link` (URL affiliate để client mở), `target_url` (URL sản phẩm đích trên sàn, hiển thị cho user thấy đi đâu), và `disclosure` (văn bản tiếng Việt nói rõ SănDeal có thể nhận hoa hồng) (DEC-AFFIL-08).
7. `target_url` **MUST** trỏ tới domain sàn hợp lệ của `platform_id` (shopee.vn / tiktok.com / lazada.vn); client **MUST** được kỳ vọng hiển thị đích này trước khi điều hướng (disclosure-then-redirect, không che giấu).
8. Server **MUST NOT** set, sửa, hay drop cookie trên bất kỳ domain sàn nào (DEC-AFFIL-09). Cookie affiliate (nếu có) chỉ được tạo bởi chính trang sàn khi client mở `deep_link` trong tab mà user bấm - đó là last-click hợp lệ do hành động người dùng.
9. Endpoint **MUST NOT** có chế độ batch/prefetch/background tạo link cho nhiều sản phẩm "phòng khi user mua" (NFR-AFFIL-001): mỗi link tương ứng một hành động bấm thật của user tại thời điểm đó.
10. **MUST** trả `disclosure` không rỗng trong mọi response `200`; thiếu disclosure là lỗi tạo response (fail-closed) - KHÔNG trả `deep_link` mà không kèm disclosure.
11. **MUST** đặt `Content-Type: application/json; charset=utf-8`. Mọi giá trị tiền (nếu trả kèm) là `BIGINT` VND, đồng nhất DEC-PRICE-05/DEC-AFFIL-04.
12. **SHOULD** phát OTel: `affiliate_link_created_total{platform_id, network}` (counter), `affiliate_link_rejected_total{reason}` (counter; `reason - {not_user_initiated, product_not_found, network_unavailable}`).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao bắt buộc cờ user_initiated (DEC-AFFIL-06)?** Toàn bộ định vị của SănDeal hậu-Honey nằm ở một câu: affiliate chỉ hợp lệ khi người dùng chủ động bấm. Honey bị gỡ vì làm ngược lại - thay cookie nền không cần user. Bằng cách yêu cầu một cờ chủ ý tường minh và từ chối (kèm không ghi click) khi thiếu, endpoint không thể bị dùng để tạo link nền. Đây không phải kiểm tra hình thức: nó là hợp đồng kỹ thuật của lời hứa đạo đức.

**Vì sao link và ghi click là nguyên tử (DEC-AFFIL-07)?** Một deep link affiliate luôn ứng với một hành động bấm. Sinh `sub_id` mới và ghi `affiliate_click` cùng lúc với việc trả link giữ sổ cái khớp thực tế: mỗi link đã trả đều có một click được ghi, và mỗi click đều do user bấm. Không có link "ghost" không ghi, cũng không có click ghi mà không phải do bấm.

**Vì sao bắt buộc disclosure + target_url hiển thị (DEC-AFFIL-08)?** Minh bạch là yêu cầu của chính sách Chrome Web Store mới (§4.2) và là moat niềm tin. Người dùng phải biết hai điều trước khi đi: họ sẽ tới đâu (`target_url`) và SănDeal có thể hưởng hoa hồng (`disclosure`). Giấu một trong hai là kiểu hành vi khiến Honey mất niềm tin. Trả cả hai trong response và kỳ vọng client hiển thị chúng là cách biến minh bạch thành mặc định.

**Vì sao server không chạm cookie sàn (DEC-AFFIL-09)?** Chèn/sửa cookie affiliate trên domain sàn từ phía ta chính là cookie-stuffing - thứ bị cấm. Mô hình hợp lệ là: ta chỉ trả một URL; khi user (đã bấm) mở URL đó, chính trang sàn set cookie affiliate của nó qua điều hướng bình thường. Last-click khi đó do hành động người dùng tạo ra, không phải do ta nhét nền.

**Vì sao thiếu network template thì 503 chứ không tự ghép (DEC-AFFIL-10)?** Mỗi network (Involve Asia, Accesstrade) có định dạng deep link riêng với tham số tracking riêng. Tự ghép tay một URL "trông giống" dễ tạo link sai không quy được conversion (mất tiền) hoặc vi phạm định dạng network. Thiếu mapping là lỗi cấu hình thật; trả `503` để lộ ra và sửa, an toàn hơn là bịa một link hỏng.

**Vì sao cấm batch/prefetch (§1 #9)?** Tạo sẵn link cho nhiều sản phẩm "phòng khi user mua" là một dạng tự động hóa nền trá hình - tinh thần của nó là gắn affiliate không cần chủ ý từng lần. Một bấm, một link. Ràng buộc này đóng cửa hậu cho hành vi kiểu Honey len vào qua tối ưu hiệu năng.

---

## §3 - Hợp đồng API / DDL

### Handler (Go)

```go
// services/affil/internal/api/link.go

type LinkRequest struct {
    ProductID     int64 `json:"product_id"`
    UserInitiated bool  `json:"user_initiated"`   // BẮT BUỘC true (§1 #1)
}
type LinkResponse struct {
    DeepLink   string `json:"deep_link"`
    TargetURL  string `json:"target_url"`         // hiển thị đích cho user (§1 #6,#7)
    Disclosure string `json:"disclosure"`         // không rỗng (§1 #10)
}

func (h *Handler) HandleCreateLink(w http.ResponseWriter, req *http.Request) {
    userID := auth.UserID(req.Context()) // do gateway gắn (FR-INFRA-001)
    var body LinkRequest
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        writeErr(w, 400, "invalid body"); return
    }
    if !body.UserInitiated { // §1 #1 - không ghi click, từ chối
        metrics.LinkRejected("not_user_initiated")
        writeErr(w, 400, "link requires explicit user action"); return
    }
    tp, ok := h.products.Get(req.Context(), body.ProductID)
    if !ok {
        metrics.LinkRejected("product_not_found")
        writeErr(w, 404, "product not found"); return
    }
    network, tmpl, ok := h.networks.TemplateFor(tp.PlatformID) // FR-AFFIL-003
    if !ok {
        metrics.LinkRejected("network_unavailable")
        writeErr(w, 503, "affiliate network unavailable"); return // không bịa link (§1 #5)
    }
    subID := affil.NewSubID()
    if _, err := h.repo.RecordClick(req.Context(), affil.AffiliateClick{ // nguyên tử với link (§1 #4)
        UserID: userID, PlatformID: tp.PlatformID, ProductID: &tp.ID,
        SubID: subID, Network: network,
    }); err != nil {
        writeErr(w, 500, "internal error"); return // không trả link nếu không ghi được
    }
    deep := affil.BuildDeepLink(tmpl, tp, subID) // nhúng sub_id (§1 #5)
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    _ = json.NewEncoder(w).Encode(LinkResponse{
        DeepLink:   deep,
        TargetURL:  tp.TargetURL(),              // domain sàn hợp lệ (§1 #7)
        Disclosure: affil.Disclosure(),          // không rỗng (§1 #10)
    })
    metrics.LinkCreated(tp.PlatformID, network)
}
```

### Deep link builder (Go)

```go
// services/affil/internal/affil/deeplink.go

// BuildDeepLink ghép URL affiliate theo template network, nhúng sub_id làm tham số tracking.
// KHÔNG set cookie, KHÔNG chạm domain sàn - chỉ trả chuỗi URL (§1 #8).
func BuildDeepLink(tmpl NetworkTemplate, p Product, subID string) string {
    target := p.TargetURL() // URL sản phẩm trên sàn
    u, _ := url.Parse(tmpl.BaseURL)
    q := u.Query()
    q.Set(tmpl.TargetParam, target) // network bọc URL đích
    q.Set(tmpl.SubIDParam, subID)   // sub_id để postback đối soát (FR-AFFIL-001)
    u.RawQuery = q.Encode()
    return u.String()
}
```

### Disclosure (Go)

```go
// services/affil/internal/affil/disclosure.go

// Disclosure trả văn bản minh bạch tiếng Việt, không rỗng (§1 #10).
func Disclosure() string {
    return "Khi ban mua qua lien ket nay, SanDeal co the nhan mot khoan hoa hong tu san. " +
        "Gia ban tra khong thay doi. Ban dang duoc chuyen toi trang san chinh thuc."
}
```

---

## §4 - Acceptance criteria

1. `POST /v1/affiliate/link {product_id:90112, user_initiated:true}` (hợp lệ) -> `200` với `deep_link`, `target_url`, `disclosure` đều không rỗng.
2. Request thiếu `user_initiated` (hoặc `false`) -> `400` `{"error":"link requires explicit user action"}` VÀ không có dòng `affiliate_click` nào được ghi (kiểm DB sau request).
3. `product_id` không tồn tại -> `404`, không ghi click.
4. Mỗi request hợp lệ sinh một `sub_id` mới và một dòng `affiliate_click` đúng `user_id/platform_id/network`.
5. `deep_link` chứa `sub_id` vừa sinh làm tham số tracking (kiểm query string).
6. `target_url` trỏ tới domain sàn của `platform_id` (ví dụ `shopee.vn` cho platform 1).
7. Thiếu mapping network cho sàn -> `503`, KHÔNG trả `deep_link`, không ghi click.
8. `disclosure` không rỗng trong mọi response `200`; chứa cụm "hoa hong" (minh bạch).
9. `RecordClick` lỗi (mock repo fail) -> `500`, KHÔNG trả `deep_link` (không link mà không ghi).
10. Không tồn tại endpoint/đường batch tạo nhiều link một lần (grep router: chỉ một route, nhận một `product_id`).
11. Server không gọi bất kỳ API set-cookie nào trên domain sàn (review + test: handler chỉ trả JSON, không HTTP client tới sàn).
12. Metric `affiliate_link_created_total` tăng khi `200`; `affiliate_link_rejected_total{reason}` tăng đúng nhãn khi từ chối.

---

## §5 - Kiểm thử (verification)

```go
// services/affil/internal/api/link_test.go
func TestLink_NotUserInitiated_Rejected_NoClick(t *testing.T) {
    h, repo := setupHandler(t)
    rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112,"user_initiated":false}`)
    require.Equal(t, 400, rec.Code)
    require.Equal(t, 0, repo.ClickCount()) // KHÔNG ghi click khi thiếu chủ ý (§1 #1)
}

func TestLink_MissingFlag_Rejected(t *testing.T) {
    h, repo := setupHandler(t)
    rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112}`) // thiếu user_initiated
    require.Equal(t, 400, rec.Code)
    require.Equal(t, 0, repo.ClickCount())
}

func TestLink_HappyPath_HasDisclosureAndTarget(t *testing.T) {
    h, repo := setupHandler(t)
    rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112,"user_initiated":true}`)
    require.Equal(t, 200, rec.Code)
    var resp LinkResponse
    decode(t, rec, &resp)
    require.NotEmpty(t, resp.DeepLink)
    require.Contains(t, resp.TargetURL, "shopee.vn")
    require.Contains(t, resp.Disclosure, "hoa hong")
    require.Equal(t, 1, repo.ClickCount()) // ghi đúng một click
}

func TestLink_ProductNotFound_404(t *testing.T) {
    h, repo := setupHandler(t)
    rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":999999,"user_initiated":true}`)
    require.Equal(t, 404, rec.Code)
    require.Equal(t, 0, repo.ClickCount())
}

func TestLink_NoNetwork_503_NoClick(t *testing.T) {
    h, repo := setupHandlerNoNetwork(t) // không cấu hình template
    rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112,"user_initiated":true}`)
    require.Equal(t, 503, rec.Code)
    require.Equal(t, 0, repo.ClickCount()) // không bịa link, không ghi
}

func TestLink_RecordClickFails_NoLink(t *testing.T) {
    h, repo := setupHandler(t)
    repo.FailNextClick()
    rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112,"user_initiated":true}`)
    require.Equal(t, 500, rec.Code)
    require.NotContains(t, rec.Body.String(), "deep_link") // không link mà không ghi (§1 #4)
}
```

```go
// services/affil/internal/affil/deeplink_test.go
func TestDeepLink_EmbedsSubID(t *testing.T) {
    tmpl := NetworkTemplate{BaseURL: "https://go.involve.asia/aff", TargetParam: "url", SubIDParam: "sub_id"}
    link := BuildDeepLink(tmpl, sampleProduct(), "sd_abc123")
    require.Contains(t, link, "sub_id=sd_abc123")
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `deeplink.go` (build URL từ template + nhúng sub_id) -> `disclosure.go` (văn bản minh bạch) -> `link.go` (handler: chặn nếu thiếu cờ user_initiated, tra product, lấy network template, sinh sub_id, RecordClick nguyên tử, trả disclosure + target_url) -> đăng ký route trong `router.go` sau JWT middleware (FR-INFRA-001) -> tests. Handler dùng `http.ServeMux` Go 1.22. Network template lấy từ FR-AFFIL-003 (`networks.TemplateFor`). Không tạo HTTP client tới domain sàn - handler chỉ trả JSON, mọi điều hướng/cookie xảy ra phía client khi user mở `deep_link`.

---

## §7 - Phụ thuộc

- **FR-AFFIL-001** - cung cấp `RecordClick`, `NewSubID`, bảng `affiliate_click`; là điều kiện cứng để ghi click khi tạo link.
- **FR-AFFIL-003 (downstream + cấu hình)** - cung cấp network template (`TemplateFor`) để build deep link; postback dùng `sub_id` sinh ở đây.
- **FR-AFFIL-004 (downstream)** - guardrails né Honey kiểm endpoint này không có đường tự động + có disclosure.
- **FR-INFRA-001 (gateway)** - gắn JWT auth và `user_id` vào context trước handler.
- **FR-EXT-003** - extension chỉ gửi dữ liệu sạch (productId); nút "Mua qua SănDeal" trong extension/web gọi endpoint này khi user bấm.
- **NFR-AFFIL-001** - ràng buộc compliance: endpoint này là điểm thực thi "chỉ user-initiated + disclosure".
- Lib: `net/http`, `net/url`, `encoding/json`.

---

## §8 - Payload ví dụ

### Request (user vừa bấm "Mua qua SănDeal")

```http
POST /v1/affiliate/link HTTP/1.1
Authorization: Bearer <JWT-SanDeal>
Content-Type: application/json

{"product_id": 90112, "user_initiated": true}
```

### Response (200)

```json
{
  "deep_link": "https://go.involve.asia/aff?url=https%3A%2F%2Fshopee.vn%2Fproduct%2F88123%2F20114455667&sub_id=sd_ab12cd34ef56",
  "target_url": "https://shopee.vn/product/88123/20114455667",
  "disclosure": "Khi ban mua qua lien ket nay, SanDeal co the nhan mot khoan hoa hong tu san. Gia ban tra khong thay doi. Ban dang duoc chuyen toi trang san chinh thuc."
}
```

### Response (400 - thiếu chủ ý user, KHÔNG ghi click)

```json
{"error": "link requires explicit user action"}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Trang trung gian "interstitial" của SănDeal hiển thị disclosure + đếm ngược rồi mới redirect (thay vì client tự hiển thị) - cân nhắc cho web; client hiện chịu trách nhiệm hiển thị `target_url` + `disclosure`.
- Rút gọn deep link qua dịch vụ link riêng để đẹp URL - không cần ở P2; giữ URL network nguyên bản để đối soát rõ.
- Đa network cho cùng sàn (chọn network trả hoa hồng cao nhất tại thời điểm) - thêm logic chọn ở FR-AFFIL-003; FR này dùng template mặc định của sàn.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Tạo link không có chủ ý user | link test (§4 #2) | cookie-stuffing kiểu Honey, bị gỡ | Bắt buộc user_initiated, không ghi nếu thiếu (§1 #1) |
| Response giấu đích / thiếu disclosure | happy-path test | mất niềm tin, vi phạm policy | Bắt buộc target_url + disclosure không rỗng (§1 #6,#10) |
| Server tự set cookie sàn | review (§4 #11) | chèn cookie nền bất hợp pháp | Chỉ trả URL, client mở (DEC-AFFIL-09) |
| Bịa deep link khi thiếu network | 503 test | link sai không quy conversion | Trả 503, không tự ghép (§1 #5) |
| Trả link mà không ghi click | RecordClick-fail test | sổ cái lệch | Nguyên tử: lỗi ghi -> không trả link (§1 #4) |
| Batch/prefetch nhiều link | grep router (§4 #10) | tự động hóa nền trá hình | Một bấm một link (§1 #9) |
| product_id không tồn tại | 404 test | link trỏ rác | Tra tracked_product trước (§1 #3) |
| Thiếu JWT | gateway chặn | - | Auth ở gateway (FR-INFRA-001) |
| sub_id trùng (hi hữu) | UNIQUE ở FR-AFFIL-001 | RecordClick lỗi -> 500 | Sinh lại sub_id; entropy đủ để cực hiếm |

---

## §11 - Ghi chú

- Endpoint này là điểm thực thi kỹ thuật của lằn ranh đạo đức trung tâm: affiliate chỉ hợp lệ khi user chủ động bấm (§4.2).
- Cờ `user_initiated` + không-ghi-khi-thiếu đóng cửa với mô hình cookie-stuffing kiểu Honey ngay tại API.
- Response luôn kèm `target_url` (đi đâu) + `disclosure` (có thể hưởng hoa hồng): minh bạch là mặc định, không phải tùy chọn.
- Server không bao giờ chạm cookie domain sàn; cookie last-click do chính trang sàn set khi user (đã bấm) mở deep link - đó là last-click hợp lệ.
- Link và ghi click là nguyên tử: mỗi link đã trả đều có một click ghi do user bấm, không có link ghost cũng không có click nền.
- Cấm batch/prefetch đóng cửa hậu cho tự động hóa nền len vào qua tối ưu hiệu năng.
- Đây là nền cho FR-AFFIL-003 (postback dùng sub_id), FR-AFFIL-004 (guardrails kiểm chính endpoint này) và FR-AFFIL-005 (cashback).

---

*Hết FR-AFFIL-002. Status: ready_to_implement (mục tiêu audit 10/10).*
