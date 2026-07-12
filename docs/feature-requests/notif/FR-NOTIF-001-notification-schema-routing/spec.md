---
id: FR-NOTIF-001
title: "Schema notification + template engine + routing kênh theo cost model (push > email > sms) - bảng notification, render template an toàn, chọn kênh rẻ nhất khả dụng làm điểm vào của fan-out"
module: NOTIF
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-INFRA-002, FR-NOTIF-002, FR-NOTIF-003, FR-TRACK-004, FR-DEAL-006]
depends_on: [FR-INFRA-002]
blocks: [FR-NOTIF-002, FR-NOTIF-003, FR-TRACK-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model notification)"
  - "docs/... §3.6 (mô hình chi phí thông báo: push gần miễn phí > email rất rẻ > SMS đắt; ưu tiên push > email > sms)"
source_decisions:
  - "DEC-NOTIF-01: bảng notification (user_id, channel, template, payload JSONB, scheduled_at, sent_at, status) là sổ ghi mỗi thông báo, một dòng cho mỗi (user, kênh) sau khi routing fan-out"
  - "DEC-NOTIF-02: template engine render text/title từ template + payload, escape an toàn (không cho payload chèn markup phá nội dung); template thiếu biến trả lỗi, không gửi nửa vời"
  - "DEC-NOTIF-03: routing theo cost model push > email > sms (§3.6) - chọn kênh rẻ nhất mà user có khả năng nhận (có token push -> push; else email; sms chỉ cho high-value/OTP)"
  - "DEC-NOTIF-04: request từ producer (FR-TRACK-004 / FR-DEAL-006) gồm user_id + channel[] mong muốn + template + data; routing giao nhau channel[] với khả năng nhận của user rồi quyết kênh thực"
  - "DEC-NOTIF-05: status lifecycle pending -> queued -> sent | failed; FR-NOTIF-003 (fan-out) và các dispatcher cập nhật; FR này định nghĩa schema + routing + render, KHÔNG tự gọi nhà cung cấp"

language: "Go 1.22 (notif-svc); PostgreSQL 16"
service: shopass/services/notif/
new_files:
  - services/notif/migrations/0001_notification.sql
  - services/notif/migrations/0002_user_channel_token.sql
  - services/notif/internal/notif/template.go
  - services/notif/internal/notif/routing.go
  - services/notif/internal/notif/repo.go
  - services/notif/internal/notif/template_test.go
  - services/notif/internal/notif/routing_test.go
modified_files: []
allowed_tools:
  - file_read: services/notif/**
  - file_write: services/notif/**
  - bash: cd services/notif && go test ./...
disallowed_tools:
  - tự gọi FCM/APNs/SES/SMS trong FR này (vi phạm DEC-NOTIF-05; gửi thuộc FR-NOTIF-002..007 qua fan-out)
  - chọn SMS khi push/email khả dụng cho thông báo thường (vi phạm DEC-NOTIF-03, đốt chi phí - SMS đắt nhất)
  - render template chèn payload chưa escape (vi phạm DEC-NOTIF-02, nội dung vỡ/inject)

effort_hours: 6
sub_tasks:
  - "0.5h: 0001_notification.sql - bảng notification + index (status, scheduled_at) + 0002 user_channel_token (push token/email/phone verified)"
  - "1.5h: template.go - registry template (price_below, drop_pct, real_sale, bottom_predicted) + render escape an toàn + lỗi khi thiếu biến"
  - "1.5h: routing.go - ResolveChannel: giao channel[] mong muốn với khả năng nhận, áp cost model push>email>sms (DEC-NOTIF-03)"
  - "1.0h: repo.go - InsertNotification (status pending) + GetUserChannels (token/email/phone)"
  - "1.5h: template_test.go + routing_test.go - 8 test (render 4 template, escape, thiếu biến lỗi, routing 4 trường hợp khả dụng, sms chỉ high-value)"

risk_if_skipped: "FR-NOTIF-001 là nền của toàn bộ module thông báo - bảng notification, template engine và routing kênh mà mọi dispatcher (FR-NOTIF-002..007) và fan-out (FR-NOTIF-003) phụ thuộc. Không có nó thì engine alert (FR-TRACK-004) và batch đáy giá (FR-DEAL-006) có cảnh báo cần gửi nhưng không có nơi ghi và không có logic chọn kênh - chuỗi thông báo đứt ở khúc cuối. Nếu routing không theo cost model (push > email > sms §3.6) mà chọn SMS bừa thì chi phí thông báo bùng nổ: SMS tới VN ~200-500 VND/tin nội địa và ~$0,1552 qua Twilio, đắt gấp hàng nghìn lần push gần-miễn-phí - một sai lầm routing đủ phá unit economics. Nếu template engine không escape payload thì giá/tên sản phẩm do scraper lấy về có thể chứa ký tự phá nội dung hoặc chèn markup, làm thông báo vỡ hoặc thành kênh injection. Tự gọi nhà cung cấp ở đây thay vì qua fan-out làm mất rate-limit và flatten-the-curve, ăn 429 lúc đỉnh 00:00."
---

## §1 - Mô tả (BCP-14 normative)

Service NOTIF **MUST** định nghĩa bảng `notification`, một template engine render an toàn, và một bộ routing chọn kênh theo cost model (push > email > sms). FR này là điểm vào của fan-out; nó KHÔNG tự gọi nhà cung cấp. Hợp đồng:

1. **MUST** định nghĩa bảng `notification (id BIGSERIAL PK, user_id BIGINT REFERENCES app_user(id), channel TEXT NOT NULL, template TEXT NOT NULL, payload JSONB, scheduled_at TIMESTAMPTZ, sent_at TIMESTAMPTZ, status TEXT NOT NULL DEFAULT 'pending', created_at TIMESTAMPTZ DEFAULT now())`.
2. **MUST** ràng buộc `channel` - `{push, email, sms}` và `status` - `{pending, queued, sent, failed}` qua CHECK (DEC-NOTIF-05). Một dòng `notification` đại diện đúng một (user, kênh) đã được routing chọn.
3. **MUST** định nghĩa bảng `user_channel_token (user_id BIGINT, channel TEXT, platform TEXT, address TEXT, verified BOOLEAN DEFAULT false, updated_at TIMESTAMPTZ, PRIMARY KEY (user_id, channel, platform))` lưu khả năng nhận của user: token FCM/APNs cho `push`, email cho `email`, số điện thoại cho `sms`. Cột `platform` thuộc `{ios, android, web, email, sms}` tách kênh `push` thành hai nhánh gửi: FR-NOTIF-002 (FCM) nhặt `platform IN ('android','web')`, FR-NOTIF-005 (APNs) nhặt `platform='ios'`; nhờ `platform` trong khóa chính, một user có thể có đồng thời token iOS + Android + Web và nhận push trên từng thiết bị. Chỉ bản ghi `verified=true` (push token còn hạn, email/phone đã xác minh) mới được routing coi là khả dụng.
4. **MUST** cung cấp template engine với registry các template chốt: `price_below`, `drop_pct`, `real_sale`, `bottom_predicted` (khớp `rule_type` của FR-TRACK-003). Mỗi template có `title` + `body` theo từng kênh.
5. **MUST** render template từ `payload` an toàn (DEC-NOTIF-02): mọi giá trị từ `payload` (giá, tên sản phẩm) **MUST** được escape phù hợp kênh trước khi chèn (HTML-escape cho email body, không cho ký tự điều khiển phá push). Payload KHÔNG được chèn markup/script vào nội dung.
6. **MUST** trả lỗi khi template thiếu biến bắt buộc (DEC-NOTIF-02): nếu `payload` thiếu khóa template cần (vd `price`), render trả lỗi và dòng `notification` KHÔNG chuyển sang `queued` - không gửi nội dung nửa vời như "Giá còn {price}".
7. **MUST** cung cấp `ResolveChannel(desired []string, caps UserChannels) (string, bool)` áp cost model (DEC-NOTIF-03): từ giao của `desired` (channel[] producer yêu cầu) và khả năng nhận khả dụng của user, chọn kênh **rẻ nhất** theo thứ tự ưu tiên `push > email > sms`. Trả `ok=false` nếu không kênh nào khả dụng.
8. **MUST** chỉ chọn `sms` khi (a) `sms` nằm trong `desired`, (b) push và email đều KHÔNG khả dụng hoặc producer đánh dấu thông báo high-value/OTP, và (c) user có số điện thoại verified (DEC-NOTIF-03). SMS KHÔNG được chọn cho thông báo thường khi push/email khả dụng.
9. **MUST** expose `InsertNotification(ctx, n Notification) (int64, error)` ghi một dòng `status='pending'` sau khi routing đã quyết kênh và render đã thành công; trả `id`. Đồng thời giữ ranh giới với tầng gửi (DEC-NOTIF-05): FR này KHÔNG gọi FCM/APNs/SES/SendGrid/SMS. Sau khi có dòng `notification` pending, fan-out (FR-NOTIF-003) và các dispatcher (FR-NOTIF-002/005/006/007) lo việc gửi và cập nhật `status` -> `sent`/`failed` + `sent_at`.
10. **MUST** tạo index hỗ trợ fan-out lấy việc: `idx_notif_dispatch ON notification (status, scheduled_at) WHERE status IN ('pending','queued')` (partial, nhẹ).
11. **MUST** lưu mọi giá trong `payload`/render dạng `BIGINT` VND khi là tiền (đồng nhất DEC-PRICE-05); format hiển thị (vd "79.000 VND") làm ở bước render, dữ liệu gốc giữ int64.
12. **SHOULD** phát OTel: `notification_created_total{channel, template}` (counter), `notification_routing_total{chosen_channel, downgraded}` (counter - đếm khi phải hạ kênh do không khả dụng), `template_render_error_total{template}` (counter).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao routing theo cost model push > email > sms (DEC-NOTIF-03)?** §3.6 nêu rõ mô hình chi phí: push (FCM/APNs) gần như miễn phí, email rất rẻ (~$0,10/1.000 với SES), SMS đắt (200-500 VND/tin nội địa, ~$0,1552 qua Twilio - gấp hàng nghìn lần push). Với mô hình free-tier tài trợ bằng affiliate, kênh gửi là biến phí trực tiếp. Routing luôn chọn kênh rẻ nhất mà user thực sự nhận được giữ chi phí thông báo gần sàn, bảo vệ unit economics. SMS để dành cho việc thực sự đáng (OTP, alert giá trị cao).

**Vì sao tách user_channel_token với verified (DEC-NOTIF-03, §1 #3)?** Routing cần biết user "có thể nhận qua kênh nào". Một user cài app có token push; một user khác chỉ có email; một số có cả. `verified` quan trọng: gửi push tới token hết hạn lãng phí và ăn lỗi; gửi SMS tới số chưa xác minh có thể sai người. Chỉ coi bản ghi verified là khả dụng làm routing quyết đúng và tránh gửi vào hư không.

**Vì sao template engine escape an toàn và lỗi khi thiếu biến (DEC-NOTIF-02)?** Nội dung thông báo nhúng dữ liệu từ scraper (tên sản phẩm, giá) - dữ liệu ngoài, không tin được. Nếu chèn thẳng, một tên sản phẩm chứa ký tự đặc biệt làm vỡ email HTML hoặc trông như injection. Escape theo kênh chặn việc đó. Lỗi-khi-thiếu-biến tránh gửi "Giá còn {price}" (placeholder lộ) - thà không gửi còn hơn gửi nội dung hỏng làm mất niềm tin.

**Vì sao một dòng notification cho mỗi (user, kênh) sau routing (DEC-NOTIF-01)?** Producer gửi một "ý định cảnh báo" với `channel[]` mong muốn. Routing quy về đúng một kênh thực (kênh rẻ nhất khả dụng). Ghi một dòng `notification` cho kết quả đó cho ta đơn vị theo dõi rõ: trạng thái gửi, thời điểm, kênh thật. Fan-out lấy các dòng pending này làm việc.

**Vì sao FR này không tự gửi (DEC-NOTIF-05, §1 #9)?** Gửi ở quy mô là module riêng (FR-NOTIF-002..007) với fan-out, rate-limit FCM, flatten-the-curve, dead-letter. Trộn routing/render với gọi nhà cung cấp làm hai thứ khó test chung và dễ vượt rate-limit. FR này dừng ở "quyết kênh + dựng nội dung + ghi pending"; phần gửi là việc của các FR sau, tách trách nhiệm sạch.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/notif/migrations/0001_notification.sql
CREATE TABLE notification (
  id           BIGSERIAL   PRIMARY KEY,
  user_id      BIGINT      NOT NULL REFERENCES app_user(id),
  channel      TEXT        NOT NULL CHECK (channel IN ('push','email','sms')),
  template     TEXT        NOT NULL,
  payload      JSONB,
  scheduled_at TIMESTAMPTZ,
  sent_at      TIMESTAMPTZ,
  status       TEXT        NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending','queued','sent','failed')),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Fan-out (FR-NOTIF-003) lấy việc theo trạng thái + lịch -> partial index
CREATE INDEX idx_notif_dispatch ON notification (status, scheduled_at)
  WHERE status IN ('pending','queued');

-- services/notif/migrations/0002_user_channel_token.sql
CREATE TABLE user_channel_token (
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  channel    TEXT        NOT NULL CHECK (channel IN ('push','email','sms')),
  platform   TEXT        NOT NULL CHECK (platform IN ('ios','android','web','email','sms')),
  address    TEXT        NOT NULL,                 -- FCM/APNs token | email | phone
  verified   BOOLEAN     NOT NULL DEFAULT false,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, channel, platform)         -- mot user co the co token ios + android + web
);
-- platform tach push thanh FCM (android/web) vs APNs (ios): FR-NOTIF-002 nhat platform IN ('android','web'),
-- FR-NOTIF-005 nhat platform='ios'. Email/sms dat platform='email'/'sms' cho dong nhat khoa.
```

### Routing theo cost model (Go)

```go
// services/notif/internal/notif/routing.go

// channelRank: số nhỏ = rẻ hơn = ưu tiên cao hơn (push > email > sms, §3.6).
var channelRank = map[string]int{"push": 0, "email": 1, "sms": 2}

type UserChannels struct {
    Push  bool // có token push verified
    Email bool // có email verified
    SMS   bool // có phone verified
}

// ResolveChannel chọn kênh rẻ nhất trong giao(desired, khả dụng) (DEC-NOTIF-03).
// highValue cho phép SMS ngay cả khi push/email khả dụng (OTP/alert giá trị cao).
func ResolveChannel(desired []string, caps UserChannels, highValue bool) (string, bool) {
    avail := func(c string) bool {
        switch c {
        case "push":
            return caps.Push
        case "email":
            return caps.Email
        case "sms":
            return caps.SMS && (highValue || (!caps.Push && !caps.Email)) // §1 #8: SMS chỉ khi đáng
        }
        return false
    }
    best, found := "", false
    for _, c := range desired {
        if !avail(c) {
            continue
        }
        if !found || channelRank[c] < channelRank[best] {
            best, found = c, true
        }
    }
    return best, found
}
```

### Template engine an toàn (Go)

```go
// services/notif/internal/notif/template.go

type Rendered struct{ Title, Body string }

// templates: registry chốt khớp rule_type của FR-TRACK-003.
var templates = map[string]func(map[string]any) (Rendered, error){
    "price_below": func(d map[string]any) (Rendered, error) {
        price, ok := intField(d, "price")
        if !ok {
            return Rendered{}, fmt.Errorf("template price_below thiếu price")
        }
        return Rendered{
            Title: "Giá đã giảm về mức bạn chờ",
            Body:  fmt.Sprintf("Sản phẩm bạn theo dõi còn %s.", formatVND(price)),
        }, nil
    },
    // drop_pct, real_sale, bottom_predicted tương tự, mỗi cái kiểm đủ biến.
}

// Render dựng nội dung; thiếu biến -> lỗi, KHÔNG trả nội dung nửa vời (DEC-NOTIF-02).
func Render(template string, data map[string]any) (Rendered, error) {
    fn, ok := templates[template]
    if !ok {
        return Rendered{}, fmt.Errorf("template không tồn tại: %s", template)
    }
    return fn(data) // mỗi fn tự escape giá trị ngoài trước khi chèn
}

// formatVND format int64 -> "79.000 VND"; dữ liệu gốc luôn là int64 (DEC-PRICE-05).
func formatVND(v int64) string { /* nhóm hàng nghìn bằng dấu chấm */ return "" }
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> bảng `notification` + `user_channel_token` tồn tại; CHECK `channel`/`status` hoạt động.
2. `INSERT notification` với `channel='telegram'` hoặc `status='done'` -> lỗi CHECK.
3. `Render("price_below", {price:79000})` trả `Body` chứa "79.000 VND"; số gốc là int64.
4. `Render("price_below", {})` (thiếu `price`) -> lỗi; KHÔNG trả nội dung có placeholder.
5. Render với tên sản phẩm chứa ký tự đặc biệt (vd `<b>`) -> được escape, không chèn markup thô vào email body.
6. `ResolveChannel(["push","email"], {Push:true, Email:true}, false)` -> `"push"` (rẻ nhất khả dụng).
7. `ResolveChannel(["push","email"], {Push:false, Email:true}, false)` -> `"email"` (push không khả dụng, hạ kênh).
8. `ResolveChannel(["push","sms"], {Push:false, SMS:true}, false)` -> `"sms"` (push không khả dụng, sms là lựa chọn còn lại).
9. `ResolveChannel(["push","sms"], {Push:true, SMS:true}, false)` -> `"push"`; SMS KHÔNG được chọn cho thông báo thường khi push khả dụng.
10. `ResolveChannel(["sms"], {SMS:true}, true)` (high-value) -> `"sms"` được phép.
11. `ResolveChannel(["push"], {Push:false}, false)` -> `ok=false` (không kênh khả dụng); producer xử lý.
12. `idx_notif_dispatch` tồn tại dạng partial; metric `notification_routing_total{downgraded}` tăng khi phải hạ kênh.

---

## §5 - Kiểm thử (verification)

```go
// services/notif/internal/notif/routing_test.go
func TestRouting_PicksCheapestAvailable(t *testing.T) {
    c, ok := ResolveChannel([]string{"push", "email"},
        UserChannels{Push: true, Email: true}, false)
    require.True(t, ok)
    require.Equal(t, "push", c) // rẻ nhất (§3.6)
}

func TestRouting_DowngradesWhenPushUnavailable(t *testing.T) {
    c, _ := ResolveChannel([]string{"push", "email"},
        UserChannels{Push: false, Email: true}, false)
    require.Equal(t, "email", c)
}

func TestRouting_SMSNotChosenWhenPushAvailable(t *testing.T) {
    c, _ := ResolveChannel([]string{"push", "sms"},
        UserChannels{Push: true, SMS: true}, false)
    require.Equal(t, "push", c) // không đốt SMS khi push khả dụng (DEC-NOTIF-03)
}

func TestRouting_SMSAllowedForHighValue(t *testing.T) {
    c, ok := ResolveChannel([]string{"sms"}, UserChannels{SMS: true}, true)
    require.True(t, ok)
    require.Equal(t, "sms", c)
}

func TestRouting_NoChannelAvailable(t *testing.T) {
    _, ok := ResolveChannel([]string{"push"}, UserChannels{Push: false}, false)
    require.False(t, ok)
}

// services/notif/internal/notif/template_test.go
func TestRender_PriceBelow(t *testing.T) {
    r, err := Render("price_below", map[string]any{"price": int64(79_000)})
    require.NoError(t, err)
    require.Contains(t, r.Body, "79.000 VND")
}

func TestRender_MissingVar_Errors(t *testing.T) {
    _, err := Render("price_below", map[string]any{})
    require.Error(t, err) // không gửi nội dung nửa vời (DEC-NOTIF-02)
}

func TestRender_EscapesPayload(t *testing.T) {
    r, _ := Render("price_below", map[string]any{
        "price": int64(79_000), "product_name": "<script>x</script>",
    })
    require.NotContains(t, r.Body, "<script>") // escape, không inject
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0001_notification.sql` + `0002_user_channel_token.sql` -> `template.go` (registry + render escape) -> `routing.go` (ResolveChannel cost model) -> `repo.go` (InsertNotification + GetUserChannels) -> tests. Producer (FR-TRACK-004 / FR-DEAL-006) gọi `GetUserChannels` -> `ResolveChannel` -> `Render` -> `InsertNotification` (pending). Không có lời gọi nhà cung cấp nào trong package `notif` ở FR này. Format VND nhóm hàng nghìn bằng dấu chấm ở bước render; dữ liệu gốc giữ int64.

---

## §7 - Phụ thuộc

- **FR-INFRA-002** - bảng `app_user` cho FK `user_id`.
- **FR-TRACK-004 (upstream)** - engine alert enqueue request (user_id + channel[] + template + data); FR này routing + render + ghi pending.
- **FR-DEAL-006 (upstream)** - batch đáy giá cũng đẩy request qua điểm vào này.
- **FR-NOTIF-003 (downstream)** - fan-out lấy dòng `notification` pending (qua `idx_notif_dispatch`) để gửi.
- **FR-NOTIF-002/005/006/007 (downstream)** - dispatcher từng kênh cập nhật `status` + `sent_at`.
- Lib: `pgx` (JSONB), `encoding/json`, `html` (escape), `text/template` hoặc render thủ công.

---

## §8 - Payload ví dụ

### Producer đẩy request (nội bộ, từ FR-TRACK-004)

```json
{ "user_id": 4021, "channel": ["push", "email"], "template": "price_below",
  "data": { "price": 79000, "threshold": 89000, "product_id": 90112 } }
```

### Dòng notification sau routing (push được chọn)

```sql
SELECT id, channel, template, status FROM notification WHERE user_id = 4021 ORDER BY id DESC LIMIT 1;
-- 55012 | push | price_below | pending   (fan-out sẽ nhặt và gửi)
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Đa ngôn ngữ template (vi-VN trước; thêm locale khi mở SEA) - bám `app_user.locale`; mở rộng registry sau.
- Tùy chọn người dùng tắt từng kênh (preference center) - thêm bảng `user_channel_pref`; gắn vào routing sau.
- Gộp nhiều alert thành digest theo user trong khoảng ngắn - tối ưu trải nghiệm + chi phí giai đoạn sau.
- Quiet-hours (không push ban đêm trừ khi user đồng ý) - phối hợp với FR-NOTIF-004 scheduler.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| channel/status ngoài enum | CHECK | từ chối ghi | Hai giá trị chốt (DEC-NOTIF-02/05) |
| Chọn SMS khi push khả dụng | routing test | đốt chi phí | Cost model push>email>sms (DEC-NOTIF-03) |
| Template thiếu biến | Render lỗi | không gửi nửa vời | Lỗi-khi-thiếu (DEC-NOTIF-02) |
| Payload chèn markup | escape | nội dung an toàn | Escape theo kênh (§1 #5) |
| Push tới token hết hạn | verified=false | không chọn push | Chỉ coi bản ghi verified khả dụng |
| Không kênh khả dụng | ResolveChannel ok=false | producer biết | Trả false, không ghi rác |
| FR này tự gọi nhà cung cấp | code review | vượt rate-limit | Chỉ ghi pending (DEC-NOTIF-05) |
| Giá format sai (float) | kiểu int64 + test | sai số/hiển thị xấu | int64 gốc, format ở render |
| Fan-out không tìm được việc | idx_notif_dispatch | trễ gửi | Partial index theo status |

---

## §11 - Ghi chú

- FR-NOTIF-001 là nền module thông báo: bảng `notification`, template engine, routing - mọi dispatcher và fan-out neo vào.
- Routing cost model push > email > sms là đòn bẩy chi phí trực tiếp: SMS đắt gấp hàng nghìn lần push nên chỉ dùng khi thực sự đáng (OTP/high-value).
- `user_channel_token.verified` đảm bảo routing chỉ chọn kênh user thực sự nhận được, tránh gửi vào hư không.
- Template engine escape + lỗi-khi-thiếu-biến giữ nội dung an toàn với dữ liệu scraper ngoài và không gửi placeholder lộ.
- Ranh giới cứng: FR này dừng ở "quyết kênh + dựng nội dung + ghi pending"; gửi là việc của FR-NOTIF-002..007 qua fan-out.
- Tiền luôn int64 VND trong payload; format hiển thị làm ở render, không có bước float.

---

*Hết FR-NOTIF-001. Status: ready_to_implement (mục tiêu audit 10/10).*
