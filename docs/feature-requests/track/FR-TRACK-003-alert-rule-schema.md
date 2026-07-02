---
id: FR-TRACK-003
title: "Schema + API alert_rule + bảng alert - bốn rule_type (price_below/drop_pct/real_sale/bottom_predicted), channel[] (push/email/sms), active flag; CRUD có phân quyền, validate threshold theo loại"
module: TRACK
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-TRACK-001, FR-TRACK-004, FR-DEAL-006, FR-WEB-004, FR-INFRA-002]
depends_on: [FR-TRACK-001]
blocks: [FR-DEAL-006, FR-TRACK-004, FR-WEB-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (alert_rule rule_type/threshold/channel[]/active + bảng alert)"
  - "docs/... §3.7 (POST /v1/alerts {product_id, rule_type, threshold, channel[]})"
source_decisions:
  - "DEC-TRACK-20: alert_rule.rule_type thuộc enum chốt {price_below, drop_pct, real_sale, bottom_predicted}; CHECK constraint chặn loại lạ"
  - "DEC-TRACK-21: channel là TEXT[] mỗi phần tử thuộc {push, email, sms}; rule có thể chọn nhiều kênh; validate từng phần tử"
  - "DEC-TRACK-22: ngữ nghĩa threshold phụ thuộc rule_type - price_below cần BIGINT VND, drop_pct cần phần trăm 1..99, real_sale/bottom_predicted KHÔNG cần threshold. threshold dùng BIGINT có chủ đích (khác NUMERIC ở §3.4): threshold luôn là số nguyên, giữ đồng nhất DEC-PRICE-05, KHÔNG đổi về NUMERIC"
  - "DEC-TRACK-23: bảng alert là sổ ghi mỗi lần luật bắn (alert_rule_id, fired_at, payload JSONB, status); FR-TRACK-004 ghi, FR này chỉ định nghĩa schema + đọc lịch sử"
  - "DEC-TRACK-24: active BOOLEAN bật/tắt luật không cần xóa; engine FR-TRACK-004 chỉ đánh giá luật active=true; mọi CRUD kiểm chủ sở hữu (chống IDOR)"

language: "Go 1.22 (track-svc); PostgreSQL 16"
service: shopass/services/track/
new_files:
  - services/track/migrations/0003_alert_rule.sql
  - services/track/internal/api/alert_rule.go
  - services/track/internal/track/alert_rule_repo.go
  - services/track/internal/track/alert_validate.go
  - services/track/internal/track/alert_validate_test.go
  - services/track/internal/api/alert_rule_test.go
modified_files:
  - services/track/internal/api/router.go            # đăng ký các route alert
allowed_tools:
  - file_read: services/track/**
  - file_write: services/track/**
  - bash: cd services/track && go test ./...
disallowed_tools:
  - chấp nhận rule_type ngoài enum hoặc channel ngoài {push,email,sms} (vi phạm DEC-TRACK-20/21, luật rác engine không hiểu)
  - bỏ kiểm chủ sở hữu trên route CRUD (vi phạm DEC-TRACK-24, lỗ IDOR sửa/xóa luật người khác)
  - lưu threshold của price_below dạng float (vi phạm đồng nhất DEC-PRICE-05, sai số so giá)

effort_hours: 6
sub_tasks:
  - "0.5h: 0003_alert_rule.sql - bảng alert_rule (CHECK rule_type) + bảng alert + index (product_id, rule_type, active)"
  - "1.0h: alert_validate.go - validate threshold/channel theo từng rule_type (DEC-TRACK-22)"
  - "1.0h: alert_rule_repo.go - CreateRule, ListRules(user), ToggleActive, DeleteRule, ListAlerts(rule) (scope user_id)"
  - "1.0h: alert_rule.go - 5 handler CRUD + kiểm chủ sở hữu + validate + 201/200/404/400"
  - "0.5h: router.go - đăng ký route sau JWT middleware (FR-INFRA-001)"
  - "2.0h: alert_validate_test.go + alert_rule_test.go - 8 test (4 rule_type validate, channel xấu, cross-user 404, toggle, list alerts scope)"

risk_if_skipped: "alert_rule là cách người dùng khai báo 'báo tôi khi nào' - đây là trái tim của lời hứa SănDeal (theo dõi giá rồi nhắc đúng lúc). Không có nó thì engine kích hoạt (FR-TRACK-004) không có luật nào để đánh giá và push (FR-NOTIF-002) không có việc gì để gửi - toàn bộ chuỗi cảnh báo chết. Nếu chấp nhận rule_type ngoài enum hoặc channel ngoài {push,email,sms} thì engine nhận luật rác nó không biết xử lý, hoặc dispatcher nhận kênh không tồn tại - lỗi runtime lan xuống. Nếu bỏ kiểm chủ sở hữu thì user sửa/tắt/xóa được luật cảnh báo của người khác (lỗ IDOR), làm người khác mất cảnh báo họ đã đặt - vừa phá trải nghiệm vừa đụng PDPL. Sai ngữ nghĩa threshold (vd nhận price_below âm hoặc drop_pct 500%) tạo luật không bao giờ bắn hoặc bắn loạn."
---

## §1 - Mô tả (BCP-14 normative)

Service TRACK **MUST** cung cấp schema và API CRUD cho `alert_rule` (luật cảnh báo của user trên một SKU) và bảng `alert` (sổ ghi mỗi lần luật bắn), với validate `threshold`/`channel` theo từng `rule_type` và kiểm chủ sở hữu mọi thao tác. Hợp đồng:

1. **MUST** định nghĩa `alert_rule (id BIGSERIAL PK, user_id BIGINT REFERENCES app_user(id), product_id BIGINT REFERENCES tracked_product(id), rule_type TEXT, threshold BIGINT, channel TEXT[] NOT NULL, active BOOLEAN DEFAULT true, created_at TIMESTAMPTZ DEFAULT now())`.
2. **MUST** ràng buộc `rule_type` - `{price_below, drop_pct, real_sale, bottom_predicted}` qua CHECK (DEC-TRACK-20). `rule_type` ngoài enum bị DB từ chối.
3. **MUST** lưu `channel` là `TEXT[]` với mỗi phần tử - `{push, email, sms}` (DEC-TRACK-21); validate từng phần tử ở tầng API trước khi ghi (mảng rỗng hoặc kênh lạ trả `400`). `push` là kênh mặc định nếu client không gửi.
4. **MUST** lưu `threshold` dạng `BIGINT` và diễn giải theo `rule_type` (DEC-TRACK-22):
    - `price_below`: `threshold` là giá VND (int64), **MUST** > 0; cảnh báo khi giá hiện tại <= threshold.
    - `drop_pct`: `threshold` là phần trăm nguyên, **MUST** trong `[1, 99]`; cảnh báo khi giá giảm >= threshold% so với mốc tham chiếu.
    - `real_sale` và `bottom_predicted`: `threshold` **MUST** là NULL (tín hiệu từ engine DEAL, không có ngưỡng người dùng đặt).
5. **MUST** validate quan hệ `rule_type` <-> `threshold` ở tầng API (DEC-TRACK-22): `price_below`/`drop_pct` thiếu `threshold` hợp lệ trả `400`; `real_sale`/`bottom_predicted` kèm `threshold` khác NULL trả `400`.
6. **MUST** định nghĩa bảng `alert (id BIGSERIAL PK, alert_rule_id BIGINT REFERENCES alert_rule(id) ON DELETE CASCADE, fired_at TIMESTAMPTZ NOT NULL, payload JSONB, status TEXT NOT NULL DEFAULT 'pending')` (DEC-TRACK-23). FR này chỉ định nghĩa schema và đọc lịch sử; FR-TRACK-004 ghi dòng `alert`.
7. **MUST** phục vụ `POST /v1/alerts {product_id, rule_type, threshold?, channel[]?}` tạo luật gắn `user_id` từ JWT; `product_id` phải có trong `tracked_product` (FK), nếu không trả `400`. Trả `201` + luật đã tạo.
8. **MUST** phục vụ `GET /v1/alerts` trả mọi luật của caller; `PATCH /v1/alerts/{id} {active}` bật/tắt; `DELETE /v1/alerts/{id}` xóa luật (và `alert` con qua CASCADE).
9. **MUST** kiểm chủ sở hữu mọi route theo `{id}` (DEC-TRACK-24): `alert_rule.user_id != caller` trả `404` (không `403`, không lộ tồn tại). KHÔNG route nào bỏ qua.
10. **MUST** phục vụ `GET /v1/alerts/{id}/history` trả các dòng `alert` của luật đó (sắp `fired_at` giảm dần), chỉ khi caller sở hữu luật.
11. **MUST** tạo index hỗ trợ engine FR-TRACK-004 truy vấn luật cần đánh giá: `idx_ar_eval ON alert_rule (product_id, rule_type) WHERE active = true` (chỉ index luật đang bật, partial index nhẹ).
12. **SHOULD** phát OTel `alert_rule_ops_total{op, rule_type, status}` và `alert_rule_active_total{rule_type}` (gauge số luật đang bật theo loại) để theo dõi phân bố luật.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao rule_type là enum CHECK (DEC-TRACK-20)?** Bốn loại cảnh báo phủ đúng nhu cầu: "về giá X" (price_below), "giảm Y%" (drop_pct), "sale thật chứ không ảo" (real_sale - dùng FR-DEAL-001), "sắp chạm đáy" (bottom_predicted - dùng FR-DEAL-006). Engine đánh giá (FR-TRACK-004) viết một nhánh xử lý cho mỗi loại. Cho `rule_type` tùy ý nghĩa là engine gặp loại nó không có nhánh -> hoặc bỏ sót cảnh báo, hoặc panic. CHECK ở DB là rào cuối cùng đảm bảo chỉ có bốn loại đã biết.

**Vì sao ngữ nghĩa threshold đổi theo rule_type (DEC-TRACK-22)?** Một cột `threshold` phục vụ nhiều loại luật, nhưng đơn vị khác nhau: VND tuyệt đối với price_below, phần trăm với drop_pct, không dùng với real_sale/bottom_predicted (tín hiệu đến từ engine DEAL). Nếu không validate quan hệ này, một luật `drop_pct` với threshold = 89000 nghĩa là "giảm 89000 phần trăm" - vô nghĩa, không bao giờ bắn. Validate theo loại ở tầng API bắt lỗi ngay lúc tạo thay vì để luật chết âm thầm.

**Vì sao channel là TEXT[] nhiều kênh (DEC-TRACK-21)?** Người dùng muốn "push cho tôi, và email nếu là deal lớn". Một luật chọn được nhiều kênh là linh hoạt tự nhiên. Validate từng phần tử thuộc {push, email, sms} chặn kênh không có dispatcher. Mặc định `push` (kênh rẻ nhất, §3.6) khi client không nói rõ, khớp mô hình chi phí push > email > sms.

**Vì sao tách bảng alert khỏi alert_rule (DEC-TRACK-23)?** `alert_rule` là cấu hình (ý định, sống lâu); `alert` là sự kiện (một lần bắn, bất biến). Trộn lại làm rối. Tách ra cho ta sổ lịch sử "luật này đã bắn khi nào, gửi gì" để hiển thị cho user và để FR-TRACK-004 dedup (không bắn lại cùng cảnh báo). FR này dựng schema + đọc; engine FR-TRACK-004 ghi.

**Vì sao partial index chỉ trên active=true (§1 #11)?** Engine FR-TRACK-004 quét "mọi luật đang bật của một SKU vừa đổi giá". Luật `active=false` không bao giờ được đánh giá, nên không cần nằm trong index. Partial index `WHERE active = true` nhỏ hơn và nhanh hơn full index, đúng hình dạng truy vấn của engine.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/track/migrations/0003_alert_rule.sql
CREATE TABLE alert_rule (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  product_id BIGINT      NOT NULL REFERENCES tracked_product(id),
  rule_type  TEXT        NOT NULL
               CHECK (rule_type IN ('price_below','drop_pct','real_sale','bottom_predicted')),
  threshold  BIGINT,                                  -- VND (price_below) hoặc % (drop_pct); NULL cho real_sale/bottom_predicted
  channel    TEXT[]      NOT NULL DEFAULT ARRAY['push'],
  active      BOOLEAN     NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Engine FR-TRACK-004 chỉ quét luật đang bật của một SKU -> partial index
CREATE INDEX idx_ar_eval ON alert_rule (product_id, rule_type) WHERE active = true;
CREATE INDEX idx_ar_user ON alert_rule (user_id);

CREATE TABLE alert (
  id            BIGSERIAL   PRIMARY KEY,
  alert_rule_id BIGINT      NOT NULL REFERENCES alert_rule(id) ON DELETE CASCADE,
  fired_at      TIMESTAMPTZ NOT NULL,
  payload       JSONB,
  status        TEXT        NOT NULL DEFAULT 'pending' -- pending|sent|failed (FR-TRACK-004/NOTIF cập nhật)
);
CREATE INDEX idx_alert_rule_time ON alert (alert_rule_id, fired_at DESC);
```

### Validate threshold theo rule_type (Go)

```go
// services/track/internal/track/alert_validate.go

var validChannels = map[string]bool{"push": true, "email": true, "sms": true}

// ValidateRule kiểm quan hệ rule_type <-> threshold <-> channel (DEC-TRACK-21/22).
// Trả lỗi mô tả để handler map 400.
func ValidateRule(ruleType string, threshold *int64, channel []string) error {
    if len(channel) == 0 {
        return errors.New("channel rỗng")
    }
    for _, c := range channel {
        if !validChannels[c] {
            return fmt.Errorf("channel không hợp lệ: %s", c)
        }
    }
    switch ruleType {
    case "price_below":
        if threshold == nil || *threshold <= 0 {
            return errors.New("price_below cần threshold (VND) > 0")
        }
    case "drop_pct":
        if threshold == nil || *threshold < 1 || *threshold > 99 {
            return errors.New("drop_pct cần threshold trong [1,99]")
        }
    case "real_sale", "bottom_predicted":
        if threshold != nil {
            return fmt.Errorf("%s không nhận threshold", ruleType)
        }
    default:
        return fmt.Errorf("rule_type không hợp lệ: %s", ruleType)
    }
    return nil
}
```

### Handler (Go) - tạo luật

```go
// services/track/internal/api/alert_rule.go
func (h *Handler) HandleCreateRule(w http.ResponseWriter, req *http.Request) {
    userID := auth.UserID(req.Context())
    var body struct {
        ProductID int64    `json:"product_id"`
        RuleType  string   `json:"rule_type"`
        Threshold *int64   `json:"threshold"`
        Channel   []string `json:"channel"`
    }
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid body")
        return
    }
    if len(body.Channel) == 0 {
        body.Channel = []string{"push"} // mặc định kênh rẻ nhất
    }
    if err := track.ValidateRule(body.RuleType, body.Threshold, body.Channel); err != nil {
        writeErr(w, http.StatusBadRequest, err.Error())
        return
    }
    rule, err := h.repo.CreateRule(req.Context(), userID, body.ProductID,
        body.RuleType, body.Threshold, body.Channel)
    if err != nil {
        if isFKViolation(err) { // product_id chưa track
            writeErr(w, http.StatusBadRequest, "product not tracked")
            return
        }
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(rule)
}
```

---

## §4 - Acceptance criteria

1. `POST /v1/alerts {product_id, rule_type:"price_below", threshold:89000, channel:["push","email"]}` trả `201`; luật lưu `threshold=89000` (int64), `channel={push,email}`.
2. `rule_type` ngoài enum (vd `"price_above"`) trả `400` (validate API) và bị CHECK chặn nếu chạm DB.
3. `channel` chứa phần tử lạ (vd `["telegram"]`) trả `400`; `channel` rỗng -> mặc định `["push"]`.
4. `price_below` thiếu `threshold` hoặc `threshold <= 0` trả `400`.
5. `drop_pct` với `threshold` ngoài `[1,99]` (vd `500`) trả `400`.
6. `real_sale`/`bottom_predicted` kèm `threshold` khác NULL trả `400`; không kèm `threshold` -> `201`.
7. `product_id` chưa có trong `tracked_product` -> `400` (FK).
8. `GET /v1/alerts` chỉ trả luật của caller; luật user khác không xuất hiện.
9. Route `{id}` với luật user khác -> `404` (không `403`).
10. `PATCH /v1/alerts/{id} {active:false}` tắt luật; engine FR-TRACK-004 sau đó không đánh giá luật này (không còn trong partial index).
11. `DELETE /v1/alerts/{id}` xóa luật và mọi dòng `alert` con (CASCADE); `GET .../history` sau đó trả rỗng/404.
12. Index `idx_ar_eval` tồn tại dạng partial (`WHERE active = true`); metric `alert_rule_ops_total` tăng theo thao tác.

---

## §5 - Kiểm thử (verification)

```go
// services/track/internal/track/alert_validate_test.go
func TestValidate_PriceBelow(t *testing.T) {
    require.NoError(t, ValidateRule("price_below", ptr(int64(89_000)), []string{"push"}))
    require.Error(t, ValidateRule("price_below", nil, []string{"push"}))          // thiếu threshold
    require.Error(t, ValidateRule("price_below", ptr(int64(0)), []string{"push"})) // <= 0
}

func TestValidate_DropPct(t *testing.T) {
    require.NoError(t, ValidateRule("drop_pct", ptr(int64(20)), []string{"push"}))
    require.Error(t, ValidateRule("drop_pct", ptr(int64(500)), []string{"push"})) // ngoài [1,99]
}

func TestValidate_SignalRules_NoThreshold(t *testing.T) {
    require.NoError(t, ValidateRule("real_sale", nil, []string{"push"}))
    require.Error(t, ValidateRule("real_sale", ptr(int64(10)), []string{"push"})) // không nhận threshold
    require.NoError(t, ValidateRule("bottom_predicted", nil, []string{"email"}))
}

func TestValidate_Channel(t *testing.T) {
    require.Error(t, ValidateRule("real_sale", nil, []string{"telegram"})) // kênh lạ
    require.Error(t, ValidateRule("real_sale", nil, []string{}))           // rỗng
}

// services/track/internal/api/alert_rule_test.go
func TestCreate_UnknownProduct_400(t *testing.T) {
    h := setupHandler(t)
    rec := doPOSTAs(t, h, userA, "/v1/alerts",
        `{"product_id":999999,"rule_type":"real_sale","channel":["push"]}`)
    require.Equal(t, 400, rec.Code) // FK product not tracked
}

func TestCrossUser_404(t *testing.T) {
    h := setupHandler(t)
    rid := createRule(t, h, userB) // luật của userB
    rec := doPATCHAs(t, h, userA, "/v1/alerts/"+itoa(rid), `{"active":false}`)
    require.Equal(t, 404, rec.Code) // không 403 (DEC-TRACK-24)
}

func TestToggleActive_RemovesFromEvalIndex(t *testing.T) {
    h := setupHandler(t)
    rid := createRule(t, h, userA)
    doPATCHAs(t, h, userA, "/v1/alerts/"+itoa(rid), `{"active":false}`)
    require.False(t, ruleActive(t, h, rid)) // engine FR-TRACK-004 sẽ bỏ qua
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0003_alert_rule.sql` (2 bảng + CHECK rule_type + partial index) -> `alert_validate.go` (ValidateRule theo loại) -> `alert_rule_repo.go` (5 hàm scope user, mỗi route {id} qua kiểm chủ sở hữu) -> `alert_rule.go` (handler) -> đăng ký route trong `router.go` sau JWT middleware (FR-INFRA-001) -> tests. Validate ở tầng API là tuyến đầu (thông báo lỗi rõ); CHECK ở DB là tuyến cuối (an toàn dữ liệu). Lưu ý alias: một số FR DEAL gọi cờ này là `enabled`; schema chuẩn dùng `active` theo §3.4 tài liệu nguồn - engine FR-TRACK-004 đọc đúng tên `active`.

---

## §7 - Phụ thuộc

- **FR-TRACK-001** - SKU phải đã track (FK `product_id` tới `tracked_product`).
- **FR-INFRA-002** - bảng `app_user` cho FK `user_id`.
- **FR-INFRA-001 (gateway)** - gắn JWT auth và `user_id`.
- **FR-TRACK-004 (downstream)** - engine đọc `alert_rule` (qua `idx_ar_eval`) và ghi dòng `alert`; tên cột `active`/`channel`/`threshold` là hợp đồng với nó.
- **FR-DEAL-006 (downstream)** - batch đáy giá khớp `alert_rule` type `bottom_predicted`.
- **FR-WEB-004 (downstream)** - UI quản lý alert tiêu thụ các route này.
- Lib: `pgx` (hỗ trợ `TEXT[]`), `encoding/json`, `net/http`.

---

## §8 - Payload ví dụ

### Tạo luật price_below

```
curl -s -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
  -d '{"product_id":90112,"rule_type":"price_below","threshold":89000,"channel":["push","email"]}' \
  "https://api.sandeal.vn/v1/alerts"
```

### Response (201)

```json
{
  "id": 5012,
  "product_id": 90112,
  "rule_type": "price_below",
  "threshold": 89000,
  "channel": ["push", "email"],
  "active": true
}
```

### Lịch sử bắn (GET /v1/alerts/5012/history)

```json
[
  { "fired_at": "2026-06-27T12:00:05Z", "status": "sent",
    "payload": { "price": 79000, "old_price": 99000 } }
]
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Mốc tham chiếu cho `drop_pct` (so với giá hôm qua, median 7 ngày, hay giá lúc tạo luật) - chốt ở FR-TRACK-004 khi định nghĩa cách đánh giá; FR này chỉ lưu phần trăm.
- Lịch quiet-hours (không gửi cảnh báo ban đêm trừ khi user đồng ý) - gắn vào FR-NOTIF khi có tùy chọn người dùng.
- Giới hạn số luật theo tier (free vs Premium) - gắn vào FR-BILL.
- Luật trên `canonical_key` (cảnh báo khi bất kỳ sàn nào về giá X) - chờ FR-PRICE-005; mở rộng sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| rule_type ngoài enum | validate API + CHECK | 400 / từ chối ghi | Hai tầng chặn (DEC-TRACK-20) |
| channel lạ | ValidateRule | 400 | Chỉ push/email/sms có dispatcher |
| price_below threshold <= 0 | ValidateRule | 400 | Buộc VND > 0 |
| drop_pct ngoài [1,99] | ValidateRule | 400 | Phần trăm hợp lệ |
| real_sale kèm threshold | ValidateRule | 400 | Tín hiệu engine, không ngưỡng |
| product_id chưa track | FK 23503 | 400 product not tracked | Buộc qua FR-TRACK-001 |
| Truy cập luật user khác | kiểm chủ sở hữu | 404 | DEC-TRACK-24 (chống IDOR) |
| Xóa luật còn alert con | ON DELETE CASCADE | alert tự xóa | CASCADE ở FK |
| Engine đọc luật active=false | partial index | bỏ qua | WHERE active=true (§1 #11) |
| Nhầm tên cột active/enabled | code review | engine không khớp | Chuẩn hóa `active` theo §3.4 |

---

## §11 - Ghi chú

- `alert_rule` là nơi người dùng nói "báo tôi khi nào"; bốn `rule_type` phủ giá tuyệt đối, phần trăm giảm, sale thật, và đáy dự đoán.
- Ngữ nghĩa `threshold` đổi theo loại là điểm dễ sai nhất - validate quan hệ rule_type <-> threshold ở tầng API bắt luật chết âm thầm ngay lúc tạo.
- Tách `alert` (sự kiện bất biến) khỏi `alert_rule` (cấu hình sống lâu) cho ta sổ lịch sử và nền dedup cho FR-TRACK-004.
- Partial index `WHERE active=true` khớp đúng truy vấn của engine (chỉ luật đang bật), nhẹ hơn full index.
- Kiểm chủ sở hữu trả 404 (không 403) đóng lỗ IDOR trên khóa BIGSERIAL tuần tự, đồng nhất với FR-TRACK-002.
- Khi mở SEA, `threshold` của price_below vẫn int64 theo minor unit từng nước; enum rule_type và validate không đổi.

---

*Hết FR-TRACK-003. Status: ready_to_implement (mục tiêu audit 10/10).*
