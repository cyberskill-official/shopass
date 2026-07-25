---
id: TASK-NOTIF-007
title: "SMS dispatcher cho Việt Nam - SMSProvider abstraction (SpeedSMS/eSMS/VietGuys/Mobifone) + Twilio fallback, brandname đăng ký Cục An toàn thông tin, guard chỉ gửi SMS cho high-value/OTP vì SMS là kênh đắt nhất (push > email > sms), cập nhật notification.status sent|failed"
module: NOTIF
priority: SHOULD
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-NOTIF-001, TASK-NOTIF-002, TASK-NOTIF-003, TASK-INFRA-002]
depends_on: [TASK-NOTIF-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.6 (SMS Việt Nam: nhà cung cấp SpeedSMS/eSMS-ViHAT/VietGuys/Mobifone, Twilio fallback, brandname đăng ký Cục An toàn thông tin, giá nội địa ~200-500 VND/tin, Twilio ~$0,1552/SMS, SMS là kênh đắt nhất nên chỉ dùng high-value/OTP)"
  - "docs/... §3.6 (mô hình chi phí thông báo: push gần miễn phí > email rất rẻ > SMS đắt nhất), §3.8 (NFR đỉnh tải 00:00)"
source_decisions:
  - "DEC-NOTIF-70: SMSProvider là interface pluggable - SpeedSMS, eSMS (ViHAT), VietGuys, Mobifone là provider VN sơ cấp; Twilio là fallback khi provider VN lỗi"
  - "DEC-NOTIF-71: ưu tiên nhà cung cấp nội địa VN (giá ~200-500 VND/tin) làm sơ cấp; Twilio (~$0,1552/SMS, đắt gấp ~15-40 lần) chỉ là fallback cho OTP/high-value khi provider VN fail"
  - "DEC-NOTIF-72: brandname (sender ID) MUST được đăng ký/khai báo với Cục An toàn thông tin trước khi gửi; phí duy trì brandname ~50.000 VND/tháng/nhà mạng"
  - "DEC-NOTIF-73: guard chi phí - dispatcher MUST chỉ gửi SMS cho dòng notification đã được routing TASK-NOTIF-001 đánh dấu high-value/OTP; nếu không phải high-value/OTP thì assert-từ chối (defense-in-depth, không tin mù routing)"
  - "DEC-NOTIF-74: dispatcher là consumer downstream của fan-out (TASK-NOTIF-003); chỉ nhặt dòng notification channel='sms' status='queued'; gửi xong cập nhật status=sent|failed + sent_at"
  - "DEC-NOTIF-75: SMS Long Code rẻ hơn brandname ~50%; brandname dùng cho alert (uy tín thương hiệu), long code là tùy chọn hạ chi phí - chốt brandname mặc định, long code hoãn"

language: "Go 1.22 (notif-svc)"
service: shopass/services/notif/
new_files:
  - services/notif/internal/sms/provider.go
  - services/notif/internal/sms/speedsms.go
  - services/notif/internal/sms/twilio.go
  - services/notif/internal/sms/dispatcher.go
  - services/notif/internal/sms/guard.go
  - services/notif/internal/sms/provider_test.go
  - services/notif/internal/sms/dispatcher_test.go
  - services/notif/internal/sms/guard_test.go
modified_files:
  - services/notif/internal/notif/repo.go            # thêm ClaimSMSBatch + (tái dùng MarkSent/MarkFailed của TASK-NOTIF-002)
allowed_tools:
  - file_read: services/notif/**
  - file_write: services/notif/**
  - bash: cd services/notif && go test ./...
disallowed_tools:
  - gửi SMS cho dòng notification thường (không high-value/OTP) khi push/email khả dụng (vi phạm DEC-NOTIF-73, đốt chi phí - SMS là kênh đắt nhất)
  - dùng Twilio làm provider sơ cấp cho SMS nội địa VN (vi phạm DEC-NOTIF-71, ~$0,1552/SMS đắt gấp hàng chục lần nhà cung cấp VN)
  - gửi SMS với sender ID chưa đăng ký brandname với Cục An toàn thông tin (vi phạm DEC-NOTIF-72, nhà mạng chặn / phạt)
  - tự routing/chọn kênh hay render lại nội dung (thuộc TASK-NOTIF-001; dispatcher chỉ gửi SMS đã định kênh)

effort_hours: 6
sub_tasks:
  - "1.0h: provider.go - interface SMSProvider (Send -> SMSResult) + phân loại kết quả (sent/retry/permanent_fail), kiểu Brandname, đăng ký provider"
  - "1.0h: speedsms.go - impl provider VN (SpeedSMS) gọi HTTP API, brandname sender, phân loại phản hồi"
  - "0.75h: twilio.go - wrapper Twilio làm fallback (chỉ kích hoạt khi provider VN fail + dòng là OTP/high-value)"
  - "1.0h: guard.go - assert high-value/OTP trước khi gửi (DEC-NOTIF-73); từ chối + MarkFailed nếu dòng thường lọt xuống đây"
  - "1.5h: dispatcher.go - ClaimSMSBatch -> guard -> provider VN -> (lỗi) Twilio fallback -> MarkSent/MarkFailed"
  - "0.25h: repo.go - ClaimSMSBatch (SELECT ... FOR UPDATE SKIP LOCKED channel='sms' status='queued')"
  - "0.5h: 3 test file - send sent; provider VN fail -> Twilio fallback; non-high-value bị guard từ chối; brandname dùng đúng; status update; cost guard"

risk_if_skipped: "TASK-NOTIF-007 là cái miệng cuối của kênh SMS - kênh ĐẮT NHẤT trong cost model §3.6 (push > email > sms), và cũng là kênh duy nhất đến được người dùng không cài app, không mở email, đúng cho OTP đăng nhập và alert giá trị cao. Không có nó thì mọi dòng notification channel='sms' mà fan-out (TASK-NOTIF-003) xếp hàng nằm chết ở status='queued', OTP không tới, alert high-value rớt. Nhưng rủi ro lớn hơn là làm SAI: nếu dispatcher không có guard chi phí và gửi SMS cho thông báo giá thường (lẽ ra đi push gần-miễn-phí), chi phí bùng nổ - SMS nội địa VN ~200-500 VND/tin, qua Twilio ~$0,1552/SMS (khoảng 4.000 VND, gấp hàng chục lần), trong khi push gần như 0 đồng; một flash sale 00:00 với hàng chục nghìn alert đi nhầm SMS đủ đốt sạch ngân sách thông báo cả tháng và phá unit economics của mô hình free-tier tài trợ bằng affiliate. Nếu dùng Twilio làm sơ cấp thay vì nhà cung cấp VN, mỗi tin đắt gấp hàng chục lần vô ích. Nếu gửi với sender ID chưa đăng ký brandname với Cục An toàn thông tin, nhà mạng VN chặn tin hoặc đánh spam, tin không tới và có thể bị phạt. Guard high-value/OTP và ưu tiên provider VN là hai cái van giữ kênh đắt nhất này chỉ chạy khi thực sự đáng và với giá rẻ nhất có thể."
---

## §1 - Mô tả (BCP-14 normative)

Dispatcher SMS **MUST** gửi tin nhắn SMS tới người dùng VN qua một nhà cung cấp nội địa (SpeedSMS / eSMS-ViHAT / VietGuys / Mobifone) với Twilio làm fallback, là consumer downstream của fan-out (TASK-NOTIF-003): nó nhặt các dòng `notification` có `channel='sms'` đang `status='queued'`, gửi, rồi cập nhật `status` -> `sent`/`failed` + `sent_at`. Vì SMS là kênh đắt nhất trong cost model (push > email > sms), dispatcher **MUST** guard để chỉ gửi cho thông báo high-value/OTP. Dispatcher **MUST NOT** routing kênh và **MUST NOT** render lại nội dung (việc của TASK-NOTIF-001). Hợp đồng:

1. **MUST** định nghĩa interface `SMSProvider` pluggable (DEC-NOTIF-70) với phương thức `Send(ctx, msg SMSMessage) (SMSResult, error)`, để SpeedSMS / eSMS / VietGuys / Mobifone / Twilio đều cài đặt được và hoán đổi qua cấu hình mà không sửa dispatcher.
2. **MUST** dùng nhà cung cấp nội địa VN làm **sơ cấp** (DEC-NOTIF-71): SpeedSMS / eSMS (ViHAT) / VietGuys / Mobifone, giá nội địa ~`200-500 VND/tin` tùy nhà mạng. Twilio (~`$0,1552/SMS` tới VN, đắt gấp hàng chục lần) **MUST** chỉ đóng vai fallback.
3. **MUST** kích hoạt Twilio fallback khi provider VN sơ cấp trả lỗi tạm thời hoặc không gửi được (DEC-NOTIF-71), VÀ chỉ khi dòng `notification` là OTP/high-value (Twilio đắt, không dùng cho khối lượng lớn). Nếu provider VN thành công thì **MUST NOT** gọi Twilio.
4. **MUST** gửi với sender ID là một **brandname** đã đăng ký/khai báo với "Cục An toàn thông tin" (DEC-NOTIF-72). Dispatcher **MUST NOT** gửi với brandname chưa đăng ký - nhà mạng VN chặn hoặc đánh spam tin chưa khai báo. Brandname duy trì phí ~`50.000 VND/tháng/nhà mạng`.
5. **MUST** guard chi phí (DEC-NOTIF-73): trước khi gửi, dispatcher **MUST** assert dòng `notification` là high-value/OTP. Routing của TASK-NOTIF-001 đã đảm bảo `channel='sms'` chỉ xuất hiện cho high-value/OTP hoặc khi push+email không khả dụng, nhưng dispatcher **MUST** kiểm lại (defense-in-depth) và **MUST** từ chối + `MarkFailed` nếu một dòng thường lọt xuống đây. SMS **MUST NOT** được gửi cho thông báo giá thông thường khi push/email là lựa chọn rẻ hơn.
6. **MUST** dựng `SMSMessage` đúng cấu trúc: `to` (số điện thoại VN dạng E.164 hoặc nội địa theo yêu cầu provider), `body` (đã render bởi TASK-NOTIF-001, plaintext, không markup), `brandname` (sender ID đã đăng ký), và cờ `highValue` mang từ dòng `notification`.
7. **MUST** lấy số điện thoại đích từ `user_channel_token` của TASK-NOTIF-001: bản ghi `channel='sms'`, `verified=true`. Số chưa verified **MUST NOT** được gửi tới (gửi SMS tới số chưa xác minh có thể sai người + tốn tiền vô ích).
8. **MUST** phân loại phản hồi provider thành ba nhóm và hành xử đúng (mirror TASK-NOTIF-002 §1 #6):
- Thành công (provider chấp nhận, có message id) -> `MarkSent`: `status='sent'`, `sent_at=now()`.
- Lỗi tạm thời (5xx, timeout, hết số dư tạm, rate-limit provider) -> thử Twilio fallback (nếu high-value/OTP) hoặc retry; vượt số lần thử thì `MarkFailed` để dead-letter (TASK-NOTIF-003).
- Lỗi vĩnh viễn (số sai định dạng, brandname bị từ chối, nội dung vi phạm) -> `MarkFailed`, không retry.
9. **MUST** nhặt việc an toàn khi chạy nhiều worker song song: `ClaimSMSBatch` dùng `SELECT ... FOR UPDATE SKIP LOCKED` trên `notification` (`channel='sms'`, `status='queued'`) để hai worker không gửi trùng (DEC-NOTIF-74), tái dùng pattern của TASK-NOTIF-002.
10. **MUST** idempotent ở mức hợp lý: một dòng `notification` đã `status='sent'` **MUST NOT** bị gửi lại; `ClaimSMSBatch` chỉ lấy `status='queued'`, `MarkSent`/`MarkFailed` là cập nhật có điều kiện trên trạng thái hiện tại. SMS gửi trùng vừa tốn tiền vừa phiền người dùng.
11. **SHOULD** phát OTel metric: `sms_send_total{provider, result}` (counter: sent|retry|failed), `sms_fallback_twilio_total` (counter - đếm khi phải dùng Twilio), `sms_guard_rejected_total` (counter - đếm khi guard từ chối dòng không high-value), `sms_cost_estimate_vnd_total{provider}` (counter - ước lượng chi phí để giám sát), `sms_send_duration_ms` (histogram).
12. **MUST** giữ ranh giới trách nhiệm: dispatcher chỉ gửi SMS cho dòng đã được TASK-NOTIF-001 chọn kênh và render. Nó **MUST NOT** tự chọn kênh, **MUST NOT** đổi nội dung, và **MUST NOT** gọi FCM/APNs/email provider.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao SMS chỉ cho high-value/OTP và phải guard (DEC-NOTIF-73)?** §3.6 xếp SMS là kênh đắt nhất: push (FCM/APNs) gần như miễn phí, email rất rẻ, còn SMS nội địa VN ~200-500 VND/tin và qua Twilio ~$0,1552/SMS (khoảng 4.000 VND). Với mô hình free-tier tài trợ bằng affiliate, mỗi tin SMS là biến phí trực tiếp đắt gấp hàng nghìn lần push. Routing TASK-NOTIF-001 đã chỉ chọn SMS cho high-value/OTP, nhưng dispatcher vẫn kiểm lại vì một lỗi routing hoặc một dòng dựng sai đi nhầm SMS lúc đỉnh 00:00 (hàng chục nghìn alert) đủ đốt sạch ngân sách thông báo cả tháng. Guard là cái van cuối cùng: thà từ chối một dòng đáng ngờ còn hơn để chi phí trượt.

**Vì sao nhà cung cấp VN sơ cấp + Twilio fallback (DEC-NOTIF-71)?** Cùng một tin SMS tới số VN, nhà cung cấp nội địa (SpeedSMS, eSMS, VietGuys, Mobifone) tính ~200-500 VND, còn Twilio tính ~$0,1552 - đắt gấp hàng chục lần. Với khối lượng OTP/alert hằng ngày, gửi sơ cấp qua provider VN giữ chi phí ở sàn. Nhưng provider VN đôi khi lỗi (hết số dư, rate-limit, sự cố), và OTP đăng nhập thì không được rớt. Twilio làm fallback cho đúng các tin đáng (OTP/high-value) khi provider VN fail: đắt nhưng độ tin cậy cao, và chỉ cho phần nhỏ tràn nên tổng chi phí vẫn thấp. Hai lớp bổ trợ: rẻ làm chính, đắt làm lưới an toàn.

**Vì sao brandname phải đăng ký với Cục An toàn thông tin (DEC-NOTIF-72)?** Ở VN, gửi SMS quảng cáo/thông báo với một sender ID (brandname) phải khai báo trước với Cục An toàn thông tin; tin từ brandname chưa đăng ký bị nhà mạng chặn hoặc đánh spam, nghĩa là tin không tới và có thể bị phạt. Brandname còn là uy tín thương hiệu: người dùng thấy "SănDeal" thay vì một đầu số lạ thì tin tưởng hơn, chống lừa đảo giả mạo. Phí duy trì ~50.000 VND/tháng/nhà mạng là chi phí cố định nhỏ đổi lấy khả năng gửi hợp lệ và uy tín.

**Vì sao SMSProvider là interface pluggable (DEC-NOTIF-70)?** Thị trường SMS VN có nhiều nhà cung cấp với giá, độ tin cậy và API khác nhau (SpeedSMS, eSMS, VietGuys, Mobifone), và ta có thể cần đổi nhà cung cấp theo giá hợp đồng hoặc cộng thêm nhà cung cấp để phân tải. Đặt `SMSProvider` là interface cho phép hoán đổi provider qua cấu hình mà không sửa dispatcher; thêm provider mới chỉ là một file cài đặt interface. Twilio cũng cài đúng interface đó nên fallback gọn trong cùng một abstraction. Điều này cũng làm test dễ: stub một `SMSProvider` giả để kiểm dispatcher mà không gọi mạng thật.

**Vì sao consumer của fan-out chứ không tự lấy thẳng từ DB (DEC-NOTIF-74)?** Giống TASK-NOTIF-002, fan-out (TASK-NOTIF-003) là nơi áp flatten-the-curve và rải tải; dispatcher SMS chỉ là cái miệng cuối nhặt dòng `channel='sms'` `status='queued'`. Tách như vậy cho phép nhiều worker SMS chạy song song mà không tự lo việc rải đỉnh, và giữ logic gửi (chọn provider, fallback, guard, brandname) gọn một chỗ. `FOR UPDATE SKIP LOCKED` làm việc nhặt an toàn khi scale ngang.

**Vì sao ranh giới cứng với TASK-NOTIF-001 (§1 #12)?** TASK-NOTIF-001 đã quyết kênh (SMS chỉ khi đáng) và render nội dung an toàn. Nếu dispatcher lại đụng vào chọn kênh hay nội dung thì hai chỗ cùng sửa một thứ, khó test và dễ lệch - đúng như lập luận ở TASK-NOTIF-002. Dispatcher SMS dừng đúng ở "gửi số này nội dung này qua provider VN (hoặc Twilio fallback), ghi kết quả". Trách nhiệm đơn, test gọn.

---

## §3 - Hợp đồng API / DDL

### Repo: nhặt việc SMS (Go)

Không cần migration mới - tái dùng `notification` và `user_channel_token` của TASK-NOTIF-001, và `MarkSent`/`MarkFailed` của TASK-NOTIF-002. Chỉ bổ sung `ClaimSMSBatch`.

```go
// services/notif/internal/notif/repo.go (bổ sung)

// ClaimSMSBatch nhặt tối đa n dòng SMS đang queued, khóa hàng để worker khác bỏ qua.
// Dùng FOR UPDATE SKIP LOCKED để nhiều dispatcher SMS chạy song song không gửi trùng.
// high_value lấy từ payload (do TASK-NOTIF-001 đánh dấu) để guard ở §1 #5.
func (r *Repo) ClaimSMSBatch(ctx context.Context, n int) ([]SMSJob, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT n.id, n.user_id, t.address AS phone, n.payload,
               COALESCE((n.payload->>'high_value')::bool, false) AS high_value
        FROM notification n
        JOIN user_channel_token t
          ON t.user_id = n.user_id AND t.channel = 'sms' AND t.verified = true
        WHERE n.channel = 'sms' AND n.status = 'queued'
        ORDER BY n.scheduled_at NULLS FIRST, n.id
        FOR UPDATE OF n SKIP LOCKED
        LIMIT $1`, n)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var jobs []SMSJob
    for rows.Next() {
        var j SMSJob
        if err := rows.Scan(&j.NotifID, &j.UserID, &j.Phone, &j.Payload, &j.HighValue); err != nil {
            return nil, err
        }
        jobs = append(jobs, j)
    }
    return jobs, rows.Err()
}
```

### SMSProvider interface + kiểu kết quả (Go)

```go
// services/notif/internal/sms/provider.go

// SMSResult phân loại phản hồi provider để dispatcher hành xử (§1 #8).
type SMSResult int

const (
    SMSSent          SMSResult = iota // provider chấp nhận, có message id
    SMSRetry                          // 5xx/timeout/hết số dư tạm/rate-limit -> fallback hoặc retry
    SMSPermanentFail                  // số sai, brandname bị từ chối, nội dung vi phạm
)

// SMSMessage là nội dung đã render bởi TASK-NOTIF-001; dispatcher chỉ gửi, không render lại.
type SMSMessage struct {
    To        string // số VN E.164 (vd +84906878091) hoặc nội địa theo provider
    Body      string // plaintext đã render, không markup
    Brandname string // sender ID đã đăng ký Cục An toàn thông tin (§1 #4)
    HighValue bool   // OTP/alert giá trị cao - mang từ dòng notification
}

// SMSProvider pluggable: SpeedSMS/eSMS/VietGuys/Mobifone/Twilio đều cài đặt (DEC-NOTIF-70).
type SMSProvider interface {
    Name() string
    Send(ctx context.Context, msg SMSMessage) (SMSResult, error)
}
```

### Provider VN sơ cấp - SpeedSMS (Go)

```go
// services/notif/internal/sms/speedsms.go

// SpeedSMS là provider nội địa VN sơ cấp (~200-500 VND/tin, DEC-NOTIF-71).
type SpeedSMS struct {
    http      *http.Client
    apiURL    string
    token     string
    brandname string // brandname đã đăng ký; gửi đúng sender đã khai báo (§1 #4)
}

func (s *SpeedSMS) Name() string { return "speedsms" }

func (s *SpeedSMS) Send(ctx context.Context, msg SMSMessage) (SMSResult, error) {
    // type=2: gửi bằng brandname đã đăng ký (không phải đầu số ngẫu nhiên).
    body, _ := json.Marshal(map[string]any{
        "to":      []string{msg.To},
        "content": msg.Body,
        "sms_type": 2,
        "brandname": msg.Brandname,
    })
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(body))
    req.SetBasicAuth(s.token, "x") // SpeedSMS dùng token làm basic-auth user
    req.Header.Set("Content-Type", "application/json")
    resp, err := s.http.Do(req)
    if err != nil {
        return SMSRetry, err // timeout/mạng -> để dispatcher thử Twilio fallback
    }
    defer resp.Body.Close()
    return classifySpeedSMS(resp), nil
}

// classifySpeedSMS ánh xạ HTTP status + status code provider sang SMSResult.
func classifySpeedSMS(resp *http.Response) SMSResult {
    switch {
    case resp.StatusCode == 200:
        return SMSSent
    case resp.StatusCode == 429 || resp.StatusCode >= 500:
        return SMSRetry // rate-limit / sự cố provider -> fallback/retry
    case resp.StatusCode == 400 || resp.StatusCode == 402:
        // 400 số/nội dung sai, 402 hết số dư có thể là tạm; coi 400 là vĩnh viễn
        return SMSPermanentFail
    default:
        return SMSPermanentFail
    }
}
```

### Twilio fallback wrapper (Go)

```go
// services/notif/internal/sms/twilio.go

// Twilio là fallback (~$0,1552/SMS, DEC-NOTIF-71) - CHỈ dùng khi provider VN fail
// VÀ dòng là OTP/high-value. Không bao giờ là provider sơ cấp cho SMS nội địa.
type Twilio struct {
    http       *http.Client
    apiURL     string // .../2010-04-01/Accounts/{sid}/Messages.json
    sid, token string
    from       string // số/Alphanumeric Sender ID Twilio
}

func (t *Twilio) Name() string { return "twilio" }

func (t *Twilio) Send(ctx context.Context, msg SMSMessage) (SMSResult, error) {
    form := url.Values{}
    form.Set("To", msg.To)
    form.Set("From", t.from)
    form.Set("Body", msg.Body)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, t.apiURL,
        strings.NewReader(form.Encode()))
    req.SetBasicAuth(t.sid, t.token)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    resp, err := t.http.Do(req)
    if err != nil {
        return SMSRetry, err
    }
    defer resp.Body.Close()
    switch {
    case resp.StatusCode >= 200 && resp.StatusCode < 300:
        return SMSSent
    case resp.StatusCode == 429 || resp.StatusCode >= 500:
        return SMSRetry
    default:
        return SMSPermanentFail
    }
}
```

### Guard chi phí + dispatcher với fallback (Go)

```go
// services/notif/internal/sms/guard.go

// ErrNotHighValue: dòng không high-value/OTP lọt xuống SMS dispatcher (DEC-NOTIF-73).
var ErrNotHighValue = errors.New("sms: dòng không high-value/OTP - từ chối gửi SMS (kênh đắt nhất)")

// assertHighValue là van chi phí cuối: SMS chỉ cho high-value/OTP.
// Routing TASK-NOTIF-001 đã đảm bảo, nhưng kiểm lại (defense-in-depth) - không tin mù.
func assertHighValue(j notif.SMSJob) error {
    if !j.HighValue {
        return ErrNotHighValue
    }
    return nil
}
```

```go
// services/notif/internal/sms/dispatcher.go

// sendWithFallback: provider VN sơ cấp trước; nếu SMSRetry và dòng high-value -> Twilio fallback.
func (d *Dispatcher) sendWithFallback(ctx context.Context, msg sms.SMSMessage) (sms.SMSResult, string, error) {
    res, err := d.primary.Send(ctx, msg) // provider VN (SpeedSMS/eSMS/VietGuys/Mobifone)
    if res == sms.SMSSent {
        return res, d.primary.Name(), err
    }
    if res == sms.SMSRetry && msg.HighValue && d.fallback != nil {
        metrics.SMSFallbackTwilio()
        fres, ferr := d.fallback.Send(ctx, msg) // Twilio (chỉ OTP/high-value)
        return fres, d.fallback.Name(), ferr
    }
    return res, d.primary.Name(), err
}

// RunOnce: claim -> guard -> gửi (VN -> Twilio fallback) -> mark.
func (d *Dispatcher) RunOnce(ctx context.Context) error {
    jobs, err := d.repo.ClaimSMSBatch(ctx, d.batch)
    if err != nil {
        return err
    }
    for _, j := range jobs {
        if err := assertHighValue(j); err != nil {
            metrics.SMSGuardRejected()
            _ = d.repo.MarkFailed(ctx, j.NotifID) // không gửi dòng không đáng (§1 #5)
            continue
        }
        msg := sms.SMSMessage{
            To: j.Phone, Body: bodyOf(j.Payload),
            Brandname: d.brandname, HighValue: j.HighValue,
        }
        res, provider, _ := d.sendWithFallback(ctx, msg)
        switch res {
        case sms.SMSSent:
            _ = d.repo.MarkSent(ctx, j.NotifID)
        default:
            _ = d.repo.MarkFailed(ctx, j.NotifID)
        }
        metrics.SMSSend(provider, res)
    }
    return nil
}
```

---

## §4 - Acceptance criteria

1. `SMSProvider` là interface; SpeedSMS, Twilio (và stub test) đều cài đặt `Send(ctx, SMSMessage) (SMSResult, error)` + `Name()`.
2. Provider VN sơ cấp gửi thành công (HTTP 200, có message id) -> `SMSSent` -> `MarkSent`: dòng `notification` `status='sent'`, `sent_at` được set.
3. Provider VN trả `SMSRetry` (429/5xx/timeout) VÀ dòng `high_value=true` -> Twilio fallback được gọi (`sms_fallback_twilio_total` tăng).
4. Provider VN trả `SMSSent` -> Twilio **KHÔNG** được gọi (không lãng phí ~$0,1552/SMS khi sơ cấp đã xong).
5. Provider VN trả `SMSRetry` nhưng dòng `high_value=false` -> KHÔNG Twilio fallback (Twilio chỉ cho OTP/high-value); dòng `MarkFailed`/retry theo cấu hình.
6. Dòng `notification` không high-value/OTP lọt vào `ClaimSMSBatch` -> guard `assertHighValue` từ chối (`ErrNotHighValue`), `sms_guard_rejected_total` tăng, dòng `MarkFailed`, KHÔNG gửi SMS.
7. `SMSMessage` gửi đi mang `Brandname` = brandname đã đăng ký; provider gửi với sender đó (không phải đầu số ngẫu nhiên).
8. `ClaimSMSBatch` chỉ trả dòng `channel='sms'`, `status='queued'`, ghép với `user_channel_token` `channel='sms'` `verified=true`; số chưa verified KHÔNG được nhặt.
9. Hai worker `ClaimSMSBatch` đồng thời (FOR UPDATE SKIP LOCKED) KHÔNG nhặt trùng cùng một dòng.
10. Provider VN trả `SMSPermanentFail` (số sai / brandname từ chối) -> `MarkFailed`, KHÔNG retry, KHÔNG Twilio.
11. Dòng đã `status='sent'` -> `MarkSent`/`MarkFailed` lần nữa không đổi gì (cập nhật có điều kiện `status='queued'`); không gửi lại (không tốn tiền đúp).
12. Metric `sms_send_total{provider,result}` tăng đúng nhánh; `sms_cost_estimate_vnd_total{provider}` tăng theo provider thực gửi (provider VN ~200-500 VND, Twilio ~4.000 VND).

---

## §5 - Kiểm thử (verification)

```go
// services/notif/internal/sms/provider_test.go

// stubProvider: SMSProvider giả để test dispatcher không gọi mạng thật.
type stubProvider struct {
    name string
    res  SMSResult
    got  *SMSMessage
}

func (s *stubProvider) Name() string { return s.name }
func (s *stubProvider) Send(ctx context.Context, m SMSMessage) (SMSResult, error) {
    *s.got = m
    return s.res, nil
}

func TestSpeedSMS_Success_ReturnsSent(t *testing.T) {
    srv := stubHTTP(t, 200, `{"status":"success","code":"00"}`)
    p := newSpeedSMS(t, srv.URL, "SANDEAL")
    res, err := p.Send(ctx, SMSMessage{To: "+84906878091", Body: "Ma OTP: 123456", Brandname: "SANDEAL", HighValue: true})
    require.NoError(t, err)
    require.Equal(t, SMSSent, res)
}

func TestSpeedSMS_5xx_ReturnsRetry(t *testing.T) {
    srv := stubHTTP(t, 503, `{}`)
    p := newSpeedSMS(t, srv.URL, "SANDEAL")
    res, _ := p.Send(ctx, SMSMessage{To: "+84906878091", Body: "x", Brandname: "SANDEAL", HighValue: true})
    require.Equal(t, SMSRetry, res) // sự cố provider VN -> để dispatcher fallback Twilio
}

func TestSpeedSMS_UsesRegisteredBrandname(t *testing.T) {
    var sentBrand string
    srv := captureBrandname(t, &sentBrand, 200, `{"status":"success"}`)
    p := newSpeedSMS(t, srv.URL, "SANDEAL")
    p.Send(ctx, SMSMessage{To: "+84906878091", Body: "x", Brandname: "SANDEAL", HighValue: true})
    require.Equal(t, "SANDEAL", sentBrand) // gửi đúng brandname đã đăng ký (§1 #4)
}

// services/notif/internal/sms/guard_test.go
func TestGuard_NonHighValue_Rejected(t *testing.T) {
    err := assertHighValue(notif.SMSJob{NotifID: 1, HighValue: false})
    require.ErrorIs(t, err, ErrNotHighValue) // SMS kênh đắt nhất -> chặn dòng thường (DEC-NOTIF-73)
}

func TestGuard_HighValue_Passes(t *testing.T) {
    require.NoError(t, assertHighValue(notif.SMSJob{NotifID: 1, HighValue: true}))
}

// services/notif/internal/sms/dispatcher_test.go
func TestDispatch_Success_MarksSent(t *testing.T) {
    d, repo, nid := setupQueuedSMS(t, &stubProvider{name: "speedsms", res: SMSSent, got: &SMSMessage{}}, nil, true)
    d.RunOnce(ctx)
    require.Equal(t, "sent", repo.status(t, nid))
    require.NotNil(t, repo.sentAt(t, nid))
}

func TestDispatch_PrimaryFails_FallsBackToTwilio(t *testing.T) {
    primary := &stubProvider{name: "speedsms", res: SMSRetry, got: &SMSMessage{}}
    fallback := &stubProvider{name: "twilio", res: SMSSent, got: &SMSMessage{}}
    d, repo, nid := setupQueuedSMS(t, primary, fallback, true) // high_value=true
    d.RunOnce(ctx)
    require.NotEmpty(t, fallback.got.To)           // Twilio fallback đã được gọi (nhận message)
    require.Equal(t, "sent", repo.status(t, nid))  // gửi thành công qua Twilio
}

func TestDispatch_PrimarySent_NoTwilioCall(t *testing.T) {
    primary := &stubProvider{name: "speedsms", res: SMSSent, got: &SMSMessage{}}
    fallback := &stubProvider{name: "twilio", res: SMSSent, got: &SMSMessage{}}
    d, repo, nid := setupQueuedSMS(t, primary, fallback, true)
    d.RunOnce(ctx)
    require.Empty(t, fallback.got.To)              // Twilio KHÔNG được gọi (sơ cấp đã xong)
    require.Equal(t, "sent", repo.status(t, nid))
}

func TestDispatch_NonHighValue_GuardRejects_NoSend(t *testing.T) {
    primary := &stubProvider{name: "speedsms", res: SMSSent, got: &SMSMessage{}}
    d, repo, nid := setupQueuedSMS(t, primary, nil, false) // high_value=false lọt xuống
    d.RunOnce(ctx)
    require.Empty(t, primary.got.To)               // KHÔNG gửi SMS (guard chặn, DEC-NOTIF-73)
    require.Equal(t, "failed", repo.status(t, nid))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `repo.go` (ClaimSMSBatch trên schema sẵn của TASK-NOTIF-001, tái dùng MarkSent/MarkFailed của TASK-NOTIF-002) -> `provider.go` (interface SMSProvider + SMSResult) -> `speedsms.go` (provider VN sơ cấp + classify) -> `twilio.go` (fallback wrapper) -> `guard.go` (assertHighValue) -> `dispatcher.go` (vòng RunOnce: claim -> guard -> sendWithFallback -> mark) -> tests. Dispatcher chạy thành pool worker, mỗi worker `ClaimSMSBatch` rồi gửi; `FOR UPDATE SKIP LOCKED` giữ an toàn khi scale ngang. Fan-out (TASK-NOTIF-003) là thứ đẩy dòng `channel='sms'` vào `status='queued'`; dispatcher chỉ nhặt và gửi. Provider VN sơ cấp và brandname cấu hình qua env/secret; thêm provider VN khác (eSMS/VietGuys/Mobifone) chỉ là một file cài đặt `SMSProvider`.

---

## §7 - Phụ thuộc

- **TASK-NOTIF-001** - bảng `notification` (đọc `channel='sms'`, cập nhật `status`/`sent_at`) và `user_channel_token` (đọc số điện thoại `verified=true`). Routing của TASK-NOTIF-001 đã đảm bảo `channel='sms'` chỉ xuất hiện cho high-value/OTP hoặc khi push+email không khả dụng - dispatcher kiểm lại bằng guard (defense-in-depth). Dispatcher KHÔNG routing/render - hai việc đó thuộc TASK-NOTIF-001.
- **TASK-NOTIF-003 (upstream)** - fan-out chuyển dòng `notification` SMS sang `status='queued'` và rải đỉnh (flatten-the-curve); dispatcher là consumer downstream nhặt các dòng đó. Dead-letter các dòng `status='failed'` cũng do fan-out lo.
- **TASK-NOTIF-002 (anh em)** - mirror cấu trúc claim/mark/status-update; tái dùng `MarkSent`/`MarkFailed`/idempotency pattern. FCM (push) và SMS là hai dispatcher song song trên cùng vòng đời `notification`.
- **TASK-INFRA-002** - `app_user` cho FK `user_id` (gián tiếp qua schema TASK-NOTIF-001).
- Lib/dịch vụ: provider VN (SpeedSMS / eSMS-ViHAT / VietGuys / Mobifone) qua HTTP API; Twilio (`Messages.json`) làm fallback; `net/http`, `encoding/json`, `net/url`, `pgx` (FOR UPDATE SKIP LOCKED). Brandname đăng ký với Cục An toàn thông tin (quy trình vận hành, ngoài code).

---

## §8 - Payload ví dụ

### Provider VN sơ cấp (SpeedSMS) gửi bằng brandname đã đăng ký

```json
{
  "to": ["+84906878091"],
  "content": "SanDeal: Ma OTP dang nhap cua ban la 123456 (het han sau 5 phut).",
  "sms_type": 2,
  "brandname": "SANDEAL"
}
```

### Twilio fallback (chỉ khi provider VN fail + dòng high-value/OTP)

```json
{
  "To": "+84906878091",
  "From": "SANDEAL",
  "Body": "SanDeal: Ma OTP dang nhap cua ban la 123456 (het han sau 5 phut)."
}
```

### Dòng notification: queued -> sent sau khi gửi thành công

```sql
-- Trước (fan-out đã đẩy vào hàng đợi, payload đánh dấu high_value):
-- 55020 | sms | otp_login | queued | sent_at = NULL | payload.high_value = true
UPDATE notification SET status='sent', sent_at=now() WHERE id=55020 AND status='queued';
-- Sau: 55020 | sms | otp_login | sent | sent_at = 2026-06-28 00:00:02+07
```

### Dòng thường lọt xuống SMS: guard từ chối, không gửi

```sql
-- payload.high_value = false (lẽ ra phải đi push) -> guard chặn, không tốn tiền SMS:
UPDATE notification SET status='failed' WHERE id=55021 AND status='queued';
-- sms_guard_rejected_total += 1
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- SMS Long Code (rẻ hơn brandname ~50%) thay cho brandname để hạ chi phí - đổi lại mất uy tín thương hiệu và có giới hạn throughput; chốt brandname mặc định, cân nhắc long code cho phân khúc không cần thương hiệu (DEC-NOTIF-75).
- Provider SMS cho SEA (Thái/Indonesia/Philippines khi mở rộng) - cùng interface `SMSProvider`, thêm nhà cung cấp khu vực; chọn theo nước người dùng giai đoạn sau.
- DLR (delivery receipt) - lấy báo cáo "đã gửi tới máy" từ provider qua webhook để biết tin thực sự tới (khác với "provider chấp nhận"); cần endpoint nhận callback, để task analytics sau.
- Tự động chuyển provider VN theo giá/độ tin cậy (multi-provider load-balance VN) - tối ưu chi phí khi volume đủ lớn; hiện một provider VN sơ cấp + Twilio fallback là đủ.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| SMS gửi cho dòng thường (không high-value) | guard `assertHighValue` | đốt chi phí (SMS đắt nhất) | Guard từ chối + MarkFailed (DEC-NOTIF-73), `sms_guard_rejected_total` |
| Dùng Twilio làm sơ cấp cho SMS nội địa | code review / cost metric | mỗi tin đắt gấp hàng chục lần | Provider VN sơ cấp, Twilio chỉ fallback (DEC-NOTIF-71) |
| Brandname chưa đăng ký Cục An toàn thông tin | nhà mạng chặn / test brandname | tin không tới, có thể bị phạt | Chỉ gửi với brandname đã đăng ký (DEC-NOTIF-72) |
| Provider VN lỗi/hết số dư lúc gửi OTP | `SMSRetry` + metric | OTP đăng nhập rớt | Twilio fallback cho high-value/OTP (DEC-NOTIF-71) |
| Twilio bị gọi cho khối lượng lớn (không cần) | `sms_fallback_twilio_total` bất thường | chi phí Twilio tăng | Fallback chỉ kích hoạt khi high-value + VN fail (§1 #3) |
| Số điện thoại chưa verified | join `verified=true` | gửi sai người / tốn tiền | ClaimSMSBatch chỉ ghép token verified (§1 #7) |
| Hai worker gửi trùng một dòng | test concurrency | SMS đúp tới user, tốn tiền | FOR UPDATE SKIP LOCKED (§1 #9) |
| Gửi lại dòng đã sent | cập nhật có điều kiện | spam + tốn tiền đúp | MarkSent/Claim chỉ chạm status='queued' (§1 #10) |
| Số sai định dạng / brandname bị từ chối | `SMSPermanentFail` | gửi hỏng lặp lại | MarkFailed, không retry, không Twilio (§1 #8) |

---

## §11 - Ghi chú

- Dispatcher SMS là cái miệng cuối của kênh ĐẮT NHẤT trong cost model §3.6 (push > email > sms); nó nhặt dòng `queued`, guard high-value/OTP, gửi qua provider VN (hoặc Twilio fallback), ghi `sent`/`failed`, không hơn.
- Giá tham chiếu: SMS nội địa VN ~200-500 VND/tin tùy nhà mạng; Twilio ~$0,1552/SMS tới VN (khoảng 4.000 VND, đắt gấp hàng chục lần) nên chỉ làm fallback cho OTP/high-value; push gần như 0 đồng - đó là lý do guard tồn tại.
- Brandname (sender ID) phải đăng ký/khai báo với Cục An toàn thông tin; phí duy trì ~50.000 VND/tháng/nhà mạng - chi phí cố định nhỏ đổi lấy khả năng gửi hợp lệ và uy tín thương hiệu (chống giả mạo).
- SMS Long Code rẻ hơn brandname ~50% nhưng mất uy tín thương hiệu và giới hạn throughput; chốt brandname mặc định, long code hoãn (DEC-NOTIF-75).
- `SMSProvider` là interface pluggable: SpeedSMS/eSMS/VietGuys/Mobifone/Twilio hoán đổi qua cấu hình; thêm provider mới chỉ là một file cài đặt interface, không sửa dispatcher.
- Ranh giới cứng với TASK-NOTIF-001: dispatcher không routing, không render lại nội dung; chỉ gửi SMS cho dòng đã định kênh và đã được đánh dấu high-value/OTP. Push có dispatcher riêng (TASK-NOTIF-002), email có dispatcher riêng (TASK-NOTIF-006).

---

*Hết TASK-NOTIF-007. Status: ready_to_implement (mục tiêu audit 10/10).*
