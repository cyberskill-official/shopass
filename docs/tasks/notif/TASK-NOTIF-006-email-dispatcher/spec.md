---
id: TASK-NOTIF-006
title: "Email dispatcher - interface EmailProvider pluggable (SES mặc định / SendGrid / Postmark), gửi qua SES API, xử lý bounce/complaint qua SNS đánh dấu verified=false, backoff cho 4xx/5xx tạm thời, cập nhật notification.status sent|failed; kênh rẻ thứ nhì dưới push trên sms"
module: NOTIF
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-NOTIF-001, TASK-NOTIF-003, TASK-NOTIF-002, TASK-NOTIF-005, TASK-INFRA-002]
depends_on: [TASK-NOTIF-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.6 (email dispatcher: SES/SendGrid/Postmark rất rẻ ~$0,10/1.000 email với SES, abstraction EmailProvider pluggable, xử lý bounce/complaint qua SNS, tôn trọng throttling/rate-limit, retry 4xx/5xx tạm thời)"
  - "docs/... §3.6 (mô hình chi phí thông báo: push > email > sms; email là kênh rẻ thứ nhì), §3.8 (NFR đỉnh tải 00:00)"
source_decisions:
  - "DEC-NOTIF-60: gửi email qua một interface EmailProvider pluggable; SES/SendGrid/Postmark là các implementation thay được sau cùng một hợp đồng, KHÔNG khóa cứng một SDK vào dispatcher"
  - "DEC-NOTIF-61: SES là provider mặc định vì rẻ nhất (~$0,10/1.000 email); email là kênh rẻ thứ nhì trong cost model push > email > sms (§3.6)"
  - "DEC-NOTIF-62: bounce cứng (hard bounce) và complaint từ SES SNS notification -> set user_channel_token.verified=false cho (user_id, 'email'); ngừng gửi tới địa chỉ chết, bảo vệ reputation/deliverability"
  - "DEC-NOTIF-63: lỗi tạm thời (throttling 429/Throttling, 5xx, timeout, soft bounce) -> exponential backoff có jitter, tôn trọng throttling/rate-limit của provider; không drop, trả việc về hàng đợi"
  - "DEC-NOTIF-64: dispatcher là consumer downstream của fan-out (TASK-NOTIF-003); chỉ nhặt dòng notification channel='email' status='queued'; nội dung Title/Body do template engine TASK-NOTIF-001 render sẵn, dispatcher chỉ truyền đi; gửi xong cập nhật status=sent|failed + sent_at"

language: "Go 1.22 (notif-svc)"
service: shopass/services/notif/
new_files:
  - services/notif/internal/email/provider.go
  - services/notif/internal/email/ses.go
  - services/notif/internal/email/backoff.go
  - services/notif/internal/email/dispatcher.go
  - services/notif/internal/email/bounce.go
  - services/notif/internal/email/provider_test.go
  - services/notif/internal/email/backoff_test.go
  - services/notif/internal/email/dispatcher_test.go
  - services/notif/internal/email/bounce_test.go
modified_files:
  - services/notif/internal/notif/repo.go            # thêm ClaimEmailBatch + MarkSent/MarkFailed (tái dùng) + InvalidateEmail
allowed_tools:
  - file_read: services/notif/**
  - file_write: services/notif/**
  - bash: cd services/notif && go test ./...
disallowed_tools:
  - khóa cứng một SDK nhà cung cấp (chỉ AWS SES) vào dispatcher thay vì sau interface EmailProvider (vi phạm DEC-NOTIF-60, không thay được provider)
  - tiếp tục gửi tới địa chỉ đã hard bounce / complaint (vi phạm DEC-NOTIF-62, giết reputation, vào blocklist SES)
  - drop thẳng message khi ăn throttling/5xx tạm thời thay vì backoff + trả về hàng đợi (vi phạm DEC-NOTIF-63, mất thông báo)
  - tự routing/chọn kênh hay render lại Title/Body (thuộc TASK-NOTIF-001; dispatcher chỉ truyền nội dung đã render)

effort_hours: 4
sub_tasks:
  - "0.5h: provider.go - interface EmailProvider (Send) + struct EmailMessage (To, Subject, HTMLBody, TextBody) + phân loại SendOutcome (sent/retry/dead/failed)"
  - "0.75h: ses.go - implementation SES qua SES v2 SendEmail, map lỗi (Throttling/5xx -> retry, MessageRejected địa chỉ chết -> dead), trả SendOutcome"
  - "0.5h: backoff.go - exponential backoff có jitter, tôn trọng Retry-After/throttling, phân biệt lỗi retry vs vĩnh viễn"
  - "0.75h: bounce.go - handler webhook SNS (hard bounce + complaint) -> InvalidateEmail(user) verified=false; soft bounce -> không vô hiệu"
  - "0.75h: dispatcher.go - ClaimEmailBatch -> Provider.Send -> MarkSent/MarkFailed; vòng RunOnce"
  - "0.25h: repo.go - ClaimEmailBatch (FOR UPDATE SKIP LOCKED status='queued' channel='email'), InvalidateEmail (verified=false cho channel='email')"
  - "0.5h: 4 test file - send thành công -> status=sent; hard bounce SNS -> verified=false; throttling -> backoff; provider swap sau interface; status update"

risk_if_skipped: "TASK-NOTIF-006 là cái miệng cuối của kênh email - kênh rẻ thứ nhì trong cost model §3.6 (push > email > sms) và là chỗ dựa cho mọi user không cài app (không có token push, không đáng gửi SMS đắt). Không có nó thì mọi dòng notification channel='email' mà fan-out (TASK-NOTIF-003) xếp vào hàng đợi nằm chết ở status='queued', cảnh báo giá (TASK-TRACK-004) và đáy giá (TASK-DEAL-006) không tới hộp thư người dùng, và phần lớn người chỉ-có-email của tệp web bị bỏ rơi đúng khúc đáng giá nhất. Nếu khóa cứng một SDK provider vào dispatcher thì khi SES bị giới hạn quota lúc đỉnh, hoặc giá SendGrid/Postmark đổi, ta phải mổ lại dispatcher thay vì cắm provider khác sau interface - mất hết sự linh hoạt mà §3.6 cố ý đòi. Nếu không xử lý hard bounce và complaint từ SES SNS bằng cách đánh dấu verified=false thì ta gửi mãi vào địa chỉ chết hoặc địa chỉ đã bấm spam: SES tính bounce rate và complaint rate, vượt ngưỡng (bounce > 5%, complaint > 0,1%) là tài khoản bị đưa vào review rồi tạm dừng gửi - một hộp thư hỏng kéo cả hệ thống email sập theo. Nếu không backoff cho throttling/5xx tạm thời mà drop thẳng thì lúc đỉnh 00:00 (giờ vàng flash sale, alert nổ hàng loạt) ta đụng rate-limit SES và người dùng mất đúng email họ chờ cả ngày."
---

## §1 - Mô tả (BCP-14 normative)

Dispatcher email **MUST** gửi email qua một interface `EmailProvider` thay-được (SES mặc định / SendGrid / Postmark), là consumer downstream của fan-out (TASK-NOTIF-003): nó nhặt các dòng `notification` có `channel='email'` đang `status='queued'`, truyền nội dung đã render đi, rồi cập nhật `status` -> `sent`/`failed` + `sent_at`. Dispatcher KHÔNG routing kênh và KHÔNG render lại nội dung (việc của TASK-NOTIF-001). Hợp đồng:

1. **MUST** định nghĩa một interface `EmailProvider` với phương thức `Send(ctx, EmailMessage) (SendOutcome, error)` làm hợp đồng chung (DEC-NOTIF-60). SES, SendGrid, Postmark **MUST** là các implementation thay được sau interface này; dispatcher gọi interface, **MUST NOT** khóa cứng một SDK nhà cung cấp.
2. **MUST** dùng SES làm provider mặc định (DEC-NOTIF-61) vì rẻ nhất (~`$0,10`/`1.000` email). Email là kênh rẻ thứ nhì trong cost model `push > email > sms` (§3.6) - dưới push gần-miễn-phí, trên SMS đắt.
3. **MUST** lấy địa chỉ đích (email) từ `user_channel_token` của TASK-NOTIF-001: bản ghi `channel='email'`, `verified=true`. Địa chỉ chưa verified hoặc đã bị vô hiệu (đã hard bounce/complaint) **MUST NOT** được gửi tới.
4. **MUST** truyền nguyên Title/Body do template engine TASK-NOTIF-001 render sẵn (DEC-NOTIF-64): `EmailMessage.Subject` lấy từ `Rendered.Title`, `EmailMessage.HTMLBody`/`TextBody` lấy từ `Rendered.Body`. Dispatcher **MUST NOT** chèn thêm hay render lại nội dung - escape an toàn đã làm ở TASK-NOTIF-001.
5. **MUST** xử lý hard bounce và complaint từ SES SNS notification (DEC-NOTIF-62): một handler webhook nhận SNS message; với `notificationType='Bounce'` mà `bounceType='Permanent'` (hard bounce) HOẶC `notificationType='Complaint'` -> set `user_channel_token.verified=false` cho `(user_id, 'email')` của địa chỉ đó. Soft bounce (`bounceType='Transient'`) **MUST NOT** vô hiệu địa chỉ.
6. **MUST** tôn trọng throttling/rate-limit của provider và retry lỗi tạm thời bằng exponential backoff có jitter (DEC-NOTIF-63): throttling (HTTP `429`/`Throttling`), `5xx`, timeout, soft bounce -> backoff rồi thử lại; **MUST** tôn trọng `Retry-After` khi provider trả về. Message lỗi tạm thời **MUST NOT** bị drop; nó giữ `status='queued'` để thử lại sau.
7. **MUST** phân loại phản hồi provider thành bốn nhóm và hành xử đúng:
- Thành công (gửi nhận, có `MessageId`) -> `MarkSent`: `status='sent'`, `sent_at=now()`.
- Lỗi tạm thời (throttling, `5xx`, timeout) -> retry theo backoff; vượt số lần thử tối đa thì `MarkFailed` để dead-letter (TASK-NOTIF-003).
- Địa chỉ chết (`MessageRejected` do email không tồn tại, hoặc địa chỉ đã trong suppression list) -> `MarkFailed` + đánh dấu địa chỉ verified=false.
- Lỗi vĩnh viễn khác (cấu hình sai, sender chưa verified domain) -> `MarkFailed`, log để vận hành xử lý.
8. **MUST** nhặt việc an toàn khi chạy nhiều worker song song: `ClaimEmailBatch` dùng `SELECT ... FOR UPDATE SKIP LOCKED` trên `notification` (`channel='email'`, `status='queued'`), để hai worker không gửi trùng một dòng (DEC-NOTIF-64).
9. **MUST** idempotent ở mức hợp lý: một dòng `notification` đã `status='sent'` **MUST NOT** bị gửi lại; `ClaimEmailBatch` chỉ lấy `status='queued'`, và `MarkSent`/`MarkFailed` là cập nhật có điều kiện trên trạng thái hiện tại.
10. **MUST** trả error (không panic) khi provider trả lỗi không phân loại được hoặc credential hết hạn/sai, để dispatcher loop áp retry/backoff và không kẹt worker.
11. **SHOULD** phát OTel metric: `email_send_total{provider, result}` (counter: sent|retry|failed|dead), `email_retry_total{reason}` (counter), `email_bounce_total{type}` (counter: hard|soft|complaint), `email_address_invalidated_total` (counter), `email_send_duration_ms` (histogram).
12. **MUST** giữ ranh giới trách nhiệm: dispatcher chỉ gửi email cho dòng đã được TASK-NOTIF-001 chọn kênh và render. Nó **MUST NOT** tự chọn kênh, **MUST NOT** đổi nội dung, và **MUST NOT** gọi push/SMS provider.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao interface EmailProvider pluggable (DEC-NOTIF-60)?** Thị trường gửi email có ba lựa chọn tương đương về API: SES, SendGrid, Postmark. Mỗi cái mạnh một kiểu - SES rẻ nhất và gắn sẵn AWS, SendGrid dễ template và analytics, Postmark deliverability cao cho transactional. Khóa cứng một SDK vào dispatcher nghĩa là đổi nhà cung cấp phải mổ lại lõi gửi. Đặt một interface `EmailProvider` với đúng một phương thức `Send` cho ta cắm provider khác sau cùng một hợp đồng: nếu SES bị siết quota lúc đỉnh hay giá đổi, ta đổi một dòng khởi tạo, không đụng dispatcher.

**Vì sao SES mặc định và vì sao email là kênh rẻ thứ nhì (DEC-NOTIF-61)?** §3.6 nêu rõ mô hình chi phí: push gần như miễn phí, email rất rẻ (~`$0,10`/`1.000` email với SES), SMS đắt. Email nằm đúng giữa - dưới push, trên SMS. Với mô hình free-tier tài trợ bằng affiliate, mỗi email là biến phí trực tiếp nên ta chọn provider rẻ nhất làm mặc định. Email còn là chỗ dựa cho cả tệp user chỉ-có-web (không cài app, không token push): không có kênh email thì routing TASK-NOTIF-001 chỉ còn SMS đắt hoặc bỏ luôn - cả hai đều tệ.

**Vì sao xử lý bounce/complaint set verified=false (DEC-NOTIF-62)?** SES không chỉ tính tiền theo lượng gửi, nó còn theo dõi bounce rate và complaint rate của cả tài khoản. Vượt ngưỡng (bounce > 5%, complaint > 0,1%) là bị đưa vào review rồi tạm dừng gửi - một hộp thư hỏng kéo cả hệ thống email sập. Hard bounce là địa chỉ không tồn tại; complaint là người dùng bấm "spam". Gửi tiếp vào đó vừa vô ích vừa giết reputation. SES bắn SNS notification cho mỗi bounce/complaint; ta nghe và set `verified=false` để routing TASK-NOTIF-001 tự thôi chọn email cho user đó. Soft bounce (hộp thư đầy, server tạm lỗi) thì khác - đó là tạm thời, không vô hiệu địa chỉ.

**Vì sao backoff + tôn trọng throttling, không drop (DEC-NOTIF-63)?** Đỉnh tải của SănDeal là 00:00 - flash sale mở, hàng loạt alert giá nổ cùng lúc (§3.8). Đúng lúc đó dễ đụng rate-limit SES nhất (SES có giới hạn gửi/giây theo tài khoản). Throttling là tín hiệu "chậm lại", không phải "bỏ đi". Drop thẳng nghĩa là người dùng mất đúng email họ chờ cả ngày. Backoff có jitter tránh cả đàn worker thử lại đồng pha (thundering herd); tôn trọng `Retry-After` là làm theo đúng nhịp provider yêu cầu thay vì đoán. Soft bounce cũng vào nhánh retry - hộp thư có thể hết đầy sau ít phút.

**Vì sao consumer của fan-out chứ không tự lấy thẳng từ DB (DEC-NOTIF-64)?** Fan-out (TASK-NOTIF-003) là nơi áp flatten-the-curve và rải tải theo thời gian; dispatcher chỉ là cái miệng cuối nhặt dòng `status='queued'` đã được fan-out chuẩn bị. Tách như vậy cho phép nhiều dispatcher email chạy song song mà không phải tự lo việc rải đỉnh, và giữ logic gửi (rate-limit, backoff, bounce) gọn trong một chỗ. `FOR UPDATE SKIP LOCKED` làm việc nhặt an toàn khi scale ngang worker. Cấu trúc claim/mark/status-update này gương đúng với dispatcher FCM (TASK-NOTIF-002) để hai kênh nhất quán.

**Vì sao dispatcher chỉ truyền, không render (DEC-NOTIF-64, §1 #4)?** Template engine TASK-NOTIF-001 đã render Title/Body và escape an toàn dữ liệu scraper ngoài (tên sản phẩm, giá) rồi. Nếu dispatcher lại đụng vào nội dung thì hai chỗ cùng sửa một thứ, khó test và dễ lệch, lại có nguy cơ làm hỏng escape đã làm. Dispatcher dừng đúng ở "gửi địa chỉ này nội dung này qua provider, ghi kết quả". Trách nhiệm đơn, test gọn.

---

## §3 - Hợp đồng API / DDL

### Repo: nhặt việc + cập nhật trạng thái (Go)

Không cần migration mới - tái dùng `notification` và `user_channel_token` của TASK-NOTIF-001. Chỉ bổ sung hàm repo (`MarkSent`/`MarkFailed` đã có từ TASK-NOTIF-002, tái dùng).

```go
// services/notif/internal/notif/repo.go (bổ sung)

// ClaimEmailBatch nhặt tối đa n dòng email đang queued, khóa hàng để worker khác bỏ qua.
// Dùng FOR UPDATE SKIP LOCKED để nhiều dispatcher chạy song song không gửi trùng.
func (r *Repo) ClaimEmailBatch(ctx context.Context, n int) ([]EmailJob, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT n.id, n.user_id, t.address AS email, n.template, n.payload
        FROM notification n
        JOIN user_channel_token t
          ON t.user_id = n.user_id AND t.channel = 'email' AND t.verified = true
        WHERE n.channel = 'email' AND n.status = 'queued'
        ORDER BY n.scheduled_at NULLS FIRST, n.id
        FOR UPDATE OF n SKIP LOCKED
        LIMIT $1`, n)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var jobs []EmailJob
    for rows.Next() {
        var j EmailJob
        if err := rows.Scan(&j.NotifID, &j.UserID, &j.Email, &j.Template, &j.Payload); err != nil {
            return nil, err
        }
        jobs = append(jobs, j)
    }
    return jobs, rows.Err()
}

// InvalidateEmail: tắt địa chỉ chết (hard bounce/complaint) -> routing TASK-NOTIF-001 thôi chọn email.
func (r *Repo) InvalidateEmail(ctx context.Context, userID int64) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE user_channel_token SET verified=false, updated_at=now()
         WHERE user_id=$1 AND channel='email'`, userID)
    return err
}

// InvalidateEmailByAddress: vô hiệu theo địa chỉ (SNS chỉ cho email, không cho user_id).
func (r *Repo) InvalidateEmailByAddress(ctx context.Context, address string) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE user_channel_token SET verified=false, updated_at=now()
         WHERE channel='email' AND address=$1`, address)
    return err
}
```

### Interface EmailProvider thay-được (Go)

```go
// services/notif/internal/email/provider.go

// EmailMessage: nội dung đã render bởi TASK-NOTIF-001, dispatcher chỉ truyền (§1 #4).
type EmailMessage struct {
    To       string // địa chỉ verified từ user_channel_token
    Subject  string // = Rendered.Title
    HTMLBody string // = Rendered.Body (đã escape ở TASK-NOTIF-001)
    TextBody string // bản plaintext fallback
}

// SendResult phân loại phản hồi provider để dispatcher hành xử (§1 #7).
type SendResult int

const (
    ResultSent    SendResult = iota // gửi nhận, có MessageId
    ResultRetry                     // throttling/5xx/timeout/soft bounce -> backoff
    ResultDead                      // địa chỉ chết / bị reject vĩnh viễn -> verified=false
    ResultFailed                    // lỗi vĩnh viễn khác (cấu hình/sender)
)

type SendOutcome struct {
    Result     SendResult
    MessageID  string
    RetryAfter time.Duration // > 0 khi provider trả Retry-After (§1 #6)
}

// EmailProvider: hợp đồng chung; SES/SendGrid/Postmark là implementation thay được (DEC-NOTIF-60).
type EmailProvider interface {
    Send(ctx context.Context, msg EmailMessage) (SendOutcome, error)
    Name() string // "ses" | "sendgrid" | "postmark" cho metric
}
```

### Implementation SES (provider mặc định, Go)

```go
// services/notif/internal/email/ses.go

// SESProvider là implementation mặc định (DEC-NOTIF-61) - rẻ nhất.
type SESProvider struct {
    client *sesv2.Client
    from   string // sender đã verified domain
}

func (p *SESProvider) Name() string { return "ses" }

// Send gửi một email qua SES v2 SendEmail; map lỗi sang SendOutcome (§1 #7).
func (p *SESProvider) Send(ctx context.Context, msg EmailMessage) (SendOutcome, error) {
    out, err := p.client.SendEmail(ctx, &sesv2.SendEmailInput{
        FromEmailAddress: &p.from,
        Destination:      &types.Destination{ToAddresses: []string{msg.To}},
        Content: &types.EmailContent{Simple: &types.Message{
            Subject: &types.Content{Data: &msg.Subject},
            Body: &types.Body{
                Html: &types.Content{Data: &msg.HTMLBody},
                Text: &types.Content{Data: &msg.TextBody},
            },
        }},
    })
    if err != nil {
        return classifySESError(err), err
    }
    return SendOutcome{Result: ResultSent, MessageID: aws.ToString(out.MessageId)}, nil
}

// classifySESError ánh xạ lỗi SES sang nhóm xử lý.
func classifySESError(err error) SendOutcome {
    var thr *types.TooManyRequestsException
    if errors.As(err, &thr) {
        return SendOutcome{Result: ResultRetry, RetryAfter: retryAfterFromErr(err)} // throttling (DEC-NOTIF-63)
    }
    var rej *types.MessageRejected
    if errors.As(err, &rej) {
        return SendOutcome{Result: ResultDead} // địa chỉ bị reject -> coi như chết
    }
    var le *types.LimitExceededException
    if errors.As(err, &le) {
        return SendOutcome{Result: ResultRetry} // vượt rate tạm thời -> backoff
    }
    if isServer5xxOrTimeout(err) {
        return SendOutcome{Result: ResultRetry}
    }
    return SendOutcome{Result: ResultFailed} // cấu hình/sender chưa verified -> vận hành xử lý
}
```

### Bounce/complaint handler qua SES SNS (Go)

```go
// services/notif/internal/email/bounce.go

// snsNotification: hình dạng tối giản của SES bounce/complaint qua SNS.
type snsNotification struct {
    NotificationType string `json:"notificationType"` // "Bounce" | "Complaint"
    Bounce           struct {
        BounceType        string `json:"bounceType"`        // "Permanent" | "Transient"
        BouncedRecipients []struct {
            EmailAddress string `json:"emailAddress"`
        } `json:"bouncedRecipients"`
    } `json:"bounce"`
    Complaint struct {
        ComplainedRecipients []struct {
            EmailAddress string `json:"emailAddress"`
        } `json:"complainedRecipients"`
    } `json:"complaint"`
}

// HandleSNS xử lý webhook SES SNS: hard bounce + complaint -> verified=false (DEC-NOTIF-62).
// Soft bounce (Transient) KHÔNG vô hiệu địa chỉ - đó là lỗi tạm thời.
func (h *BounceHandler) HandleSNS(ctx context.Context, raw []byte) error {
    var n snsNotification
    if err := json.Unmarshal(raw, &n); err != nil {
        return err
    }
    switch n.NotificationType {
    case "Bounce":
        if n.Bounce.BounceType != "Permanent" {
            metrics.Bounce("soft")
            return nil // soft bounce -> bỏ qua, không vô hiệu (§1 #5)
        }
        for _, r := range n.Bounce.BouncedRecipients {
            if err := h.repo.InvalidateEmailByAddress(ctx, r.EmailAddress); err != nil {
                return err
            }
            metrics.Bounce("hard")
            metrics.AddressInvalidated()
        }
    case "Complaint":
        for _, r := range n.Complaint.ComplainedRecipients {
            if err := h.repo.InvalidateEmailByAddress(ctx, r.EmailAddress); err != nil {
                return err
            }
            metrics.Bounce("complaint")
            metrics.AddressInvalidated()
        }
    }
    return nil
}
```

### Backoff (Go)

```go
// services/notif/internal/email/backoff.go

// nextDelay: exponential có jitter; ưu tiên Retry-After của provider khi có (§1 #6).
func nextDelay(attempt int, retryAfter time.Duration) time.Duration {
    if retryAfter > 0 {
        return retryAfter // tôn trọng throttling provider (DEC-NOTIF-63)
    }
    base := time.Duration(1<<attempt) * time.Second // 1s,2s,4s,8s...
    if base > 5*time.Minute {
        base = 5 * time.Minute
    }
    jitter := time.Duration(rand.Int63n(int64(base / 2)))
    return base/2 + jitter // +-50% tránh thundering herd lúc 00:00
}
```

---

## §4 - Acceptance criteria

1. `EmailProvider` là interface; `SESProvider` thỏa interface (`var _ EmailProvider = (*SESProvider)(nil)` biên dịch). Dispatcher gọi qua interface, KHÔNG tham chiếu thẳng `sesv2` trong `dispatcher.go`.
2. Provider mặc định khởi tạo là `SESProvider` (`Name()=="ses"`); thay bằng một fake `EmailProvider` khác trong test mà dispatcher chạy không đổi (chứng minh swap).
3. `SESProvider.Send` thành công (có `MessageId`) -> `MarkSent`: dòng `notification` chuyển `status='sent'` và `sent_at` được set.
4. SES trả throttling (`TooManyRequestsException`) -> `SendOutcome.Result=ResultRetry`; dòng giữ `status='queued'` (không bị drop), worker lập lịch thử lại theo backoff.
5. Có `Retry-After` từ provider -> `nextDelay` trả đúng giá trị đó; không có -> tăng theo cấp số nhân có jitter, có trần (<= 5 phút).
6. SNS `notificationType='Bounce'`, `bounceType='Permanent'` -> `InvalidateEmailByAddress` set `user_channel_token.verified=false` cho địa chỉ đó; metric `email_bounce_total{hard}` tăng.
7. SNS `notificationType='Bounce'`, `bounceType='Transient'` (soft bounce) -> KHÔNG vô hiệu địa chỉ; `email_bounce_total{soft}` tăng.
8. SNS `notificationType='Complaint'` -> `InvalidateEmailByAddress` set `verified=false`; `email_bounce_total{complaint}` tăng.
9. `ClaimEmailBatch` chỉ trả dòng `channel='email'`, `status='queued'`, và chỉ ghép với `user_channel_token` `verified=true`; địa chỉ chưa verified hoặc đã bị vô hiệu KHÔNG được nhặt.
10. Hai worker `ClaimEmailBatch` đồng thời (FOR UPDATE SKIP LOCKED) KHÔNG nhặt trùng cùng một dòng.
11. Dòng đã `status='sent'` -> `MarkSent`/`MarkFailed` lần nữa không đổi gì (cập nhật có điều kiện `status='queued'`); không gửi lại.
12. Metric `email_send_total{provider, result}` tăng đúng nhánh (sent/retry/failed/dead); `email_address_invalidated_total` tăng khi hard bounce/complaint.

---

## §5 - Kiểm thử (verification)

```go
// services/notif/internal/email/provider_test.go

// fakeProvider chứng minh dispatcher chạy sau interface, không khóa SES (DEC-NOTIF-60).
type fakeProvider struct {
    out  SendOutcome
    sent []EmailMessage
}

func (f *fakeProvider) Name() string { return "fake" }
func (f *fakeProvider) Send(_ context.Context, m EmailMessage) (SendOutcome, error) {
    f.sent = append(f.sent, m)
    return f.out, nil
}

func TestSESProvider_SatisfiesInterface(t *testing.T) {
    var _ EmailProvider = (*SESProvider)(nil) // SES là một implementation của interface
    var _ EmailProvider = (*fakeProvider)(nil)
}

func TestDispatch_Success_MarksSent(t *testing.T) {
    fp := &fakeProvider{out: SendOutcome{Result: ResultSent, MessageID: "0100-abc"}}
    d, repo, _, nid := setupQueuedEmail(t, fp)
    d.RunOnce(ctx)
    require.Equal(t, "sent", repo.status(t, nid))
    require.NotNil(t, repo.sentAt(t, nid))
    require.Len(t, fp.sent, 1)
    require.Equal(t, "Giá đã giảm về mức bạn chờ", fp.sent[0].Subject) // truyền Title đã render (§1 #4)
}

func TestDispatch_ProviderSwap_NoDispatcherChange(t *testing.T) {
    // Cùng dispatcher, đổi provider -> vẫn chạy (chứng minh interface pluggable).
    fp := &fakeProvider{out: SendOutcome{Result: ResultSent, MessageID: "x"}}
    d, repo, _, nid := setupQueuedEmail(t, fp)
    d.RunOnce(ctx)
    require.Equal(t, "sent", repo.status(t, nid)) // SES bị thay bằng fake, không đụng dispatcher
}

// services/notif/internal/email/dispatcher_test.go

func TestDispatch_Throttling_TriggersRetry_NotDropped(t *testing.T) {
    fp := &fakeProvider{out: SendOutcome{Result: ResultRetry, RetryAfter: 0}}
    d, repo, _, nid := setupQueuedEmail(t, fp)
    d.RunOnce(ctx)
    require.Equal(t, "queued", repo.status(t, nid)) // không drop, giữ queued để thử lại (DEC-NOTIF-63)
}

func TestDispatch_Rejected_MarksFailedAndDead(t *testing.T) {
    fp := &fakeProvider{out: SendOutcome{Result: ResultDead}}
    d, repo, addr, nid := setupQueuedEmail(t, fp)
    d.RunOnce(ctx)
    require.Equal(t, "failed", repo.status(t, nid))
    require.False(t, repo.emailVerified(t, addr)) // địa chỉ chết -> verified=false
}

// services/notif/internal/email/bounce_test.go

func TestBounce_HardBounce_InvalidatesAddress(t *testing.T) {
    h, repo, addr := setupBounceHandler(t, "user@dead.example")
    raw := []byte(`{"notificationType":"Bounce","bounce":{"bounceType":"Permanent",
        "bouncedRecipients":[{"emailAddress":"user@dead.example"}]}}`)
    require.NoError(t, h.HandleSNS(ctx, raw))
    require.False(t, repo.emailVerified(t, addr)) // hard bounce -> ngừng gửi (DEC-NOTIF-62)
}

func TestBounce_SoftBounce_KeepsAddress(t *testing.T) {
    h, repo, addr := setupBounceHandler(t, "user@full.example")
    raw := []byte(`{"notificationType":"Bounce","bounce":{"bounceType":"Transient",
        "bouncedRecipients":[{"emailAddress":"user@full.example"}]}}`)
    require.NoError(t, h.HandleSNS(ctx, raw))
    require.True(t, repo.emailVerified(t, addr)) // soft bounce -> KHÔNG vô hiệu (§1 #5)
}

func TestBounce_Complaint_InvalidatesAddress(t *testing.T) {
    h, repo, addr := setupBounceHandler(t, "user@spam.example")
    raw := []byte(`{"notificationType":"Complaint",
        "complaint":{"complainedRecipients":[{"emailAddress":"user@spam.example"}]}}`)
    require.NoError(t, h.HandleSNS(ctx, raw))
    require.False(t, repo.emailVerified(t, addr)) // bấm spam -> ngừng gửi, bảo vệ reputation
}

// services/notif/internal/email/backoff_test.go

func TestBackoff_RespectsRetryAfter(t *testing.T) {
    require.Equal(t, 30*time.Second, nextDelay(0, 30*time.Second)) // tôn trọng provider
}

func TestBackoff_ExponentialWithCap(t *testing.T) {
    d := nextDelay(20, 0) // attempt lớn
    require.LessOrEqual(t, d, 5*time.Minute) // có trần
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `repo.go` (ClaimEmailBatch + InvalidateEmailByAddress trên schema sẵn của TASK-NOTIF-001; MarkSent/MarkFailed tái dùng từ TASK-NOTIF-002) -> `provider.go` (interface EmailProvider + EmailMessage + SendOutcome) -> `ses.go` (SESProvider + classifySESError) -> `backoff.go` (retry) -> `bounce.go` (HandleSNS hard bounce + complaint) -> `dispatcher.go` (vòng RunOnce: claim -> Provider.Send -> mark) -> tests. Dispatcher chạy thành pool worker, mỗi worker `ClaimEmailBatch` rồi gửi qua `EmailProvider` đã tiêm; `FOR UPDATE SKIP LOCKED` giữ an toàn khi scale ngang. Provider tiêm qua constructor (`NewDispatcher(repo, provider)`) nên thay SES bằng SendGrid/Postmark chỉ là đổi tham số, không đụng dispatcher. Fan-out (TASK-NOTIF-003) là thứ đẩy dòng vào `status='queued'` và lo rải đỉnh; dispatcher chỉ nhặt và gửi. SNS endpoint là một HTTP handler riêng (đăng ký SES bắn vào), gọi `HandleSNS`.

---

## §7 - Phụ thuộc

- **TASK-NOTIF-001** - bảng `notification` (đọc `channel='email'`, cập nhật `status`/`sent_at`) và `user_channel_token` (đọc địa chỉ `verified=true`, set `verified=false` khi hard bounce/complaint). Title/Body do template engine TASK-NOTIF-001 render sẵn; dispatcher chỉ truyền, KHÔNG render lại.
- **TASK-NOTIF-003 (upstream)** - fan-out chuyển dòng `notification` channel='email' sang `status='queued'` và rải đỉnh (flatten-the-curve); dispatcher là consumer downstream nhặt các dòng đó. Dead-letter các dòng `status='failed'` cũng do fan-out lo.
- **TASK-NOTIF-002 (anh em)** - dispatcher FCM cùng cấu trúc claim/mark/status-update; task này gương theo để hai kênh nhất quán (push và email).
- **TASK-INFRA-002** - `app_user` cho FK `user_id` (gián tiếp qua schema TASK-NOTIF-001).
- Lib: `github.com/aws/aws-sdk-go-v2/service/sesv2` (SES v2 SendEmail), `encoding/json` (parse SNS), `net/http` (SNS webhook), `pgx` (FOR UPDATE SKIP LOCKED).

---

## §8 - Payload ví dụ

### Request gửi SES (dispatcher dựng từ dòng notification email + Title/Body đã render)

```go
out, err := provider.Send(ctx, email.EmailMessage{
    To:       "user@example.com",            // địa chỉ verified từ user_channel_token
    Subject:  "Giá đã giảm về mức bạn chờ",   // = Rendered.Title (TASK-NOTIF-001)
    HTMLBody: "<p>Sản phẩm bạn theo dõi còn <b>79.000 VND</b>.</p>", // = Rendered.Body (đã escape)
    TextBody: "San pham ban theo doi con 79.000 VND.",
})
// out.Result=ResultSent, out.MessageID="0100-..." nếu SES nhận; ResultRetry nếu throttling
```

### Bounce SNS notification từ SES (hard bounce -> vô hiệu địa chỉ)

```json
{
  "notificationType": "Bounce",
  "bounce": {
    "bounceType": "Permanent",
    "bounceSubType": "General",
    "bouncedRecipients": [
      { "emailAddress": "user@dead.example", "diagnosticCode": "smtp; 550 5.1.1 user unknown" }
    ],
    "timestamp": "2026-06-28T00:00:05.000Z"
  }
}
```

### Dòng notification: queued -> sent sau khi gửi thành công

```sql
-- Trước (fan-out đã đẩy vào hàng đợi):
-- 55020 | email | price_below | queued | sent_at = NULL
UPDATE notification SET status='sent', sent_at=now() WHERE id=55020 AND status='queued';
-- Sau: 55020 | email | price_below | sent | sent_at = 2026-06-28 00:00:05+07

-- Địa chỉ chết (hard bounce/complaint): tắt verified
UPDATE user_channel_token SET verified=false, updated_at=now()
  WHERE channel='email' AND address='user@dead.example';
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Dedicated IP + warmup (gửi tăng dần để xây sender reputation) - chỉ cần khi volume email đủ lớn để thoát shared IP pool của SES; quy trình vận hành, chưa cần code.
- DKIM/SPF/DMARC ký domain - cấu hình DNS phía hạ tầng (TASK-INFRA), không thuộc lõi dispatcher; bật trước khi gửi production để vào inbox thay vì spam.
- Link unsubscribe + List-Unsubscribe header (one-click) - bắt buộc cho bulk/marketing; alert giá là transactional nên ưu tiên thấp hơn, thêm khi có digest/marketing email.
- Đo open/click rate (tracking pixel + link wrap) - cần cho tối ưu nội dung; để task analytics sau, tránh đụng privacy sớm.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Khóa cứng SDK SES vào dispatcher | code review + test interface | không thay được provider | Gọi qua EmailProvider, tiêm provider (DEC-NOTIF-60) |
| Throttling/rate-limit SES lúc đỉnh 00:00 | `email_send_total{retry}` | nguy cơ mất email nếu drop | Backoff + Retry-After, giữ queued, thử lại (DEC-NOTIF-63) |
| Gửi tiếp tới địa chỉ hard bounce | bounce rate SES tăng | tài khoản vào review, dừng gửi | Hard bounce SNS -> verified=false (DEC-NOTIF-62) |
| Người dùng bấm spam (complaint) | complaint rate SES tăng | reputation tụt, vào blocklist | Complaint SNS -> verified=false, thôi gửi |
| Nhầm soft bounce thành hard | test SoftBounce_KeepsAddress | vô hiệu nhầm địa chỉ sống | Chỉ bounceType='Permanent' mới vô hiệu (§1 #5) |
| Sender domain chưa verified | SES MessageRejected/config | gửi hỏng hàng loạt | ResultFailed + log; verify domain trước (DKIM/SPF) |
| Hai worker gửi trùng một dòng | test concurrency | email đúp tới user | FOR UPDATE SKIP LOCKED (§1 #8) |
| Gửi lại dòng đã sent | cập nhật có điều kiện | spam người dùng | MarkSent/Claim chỉ chạm status='queued' (§1 #9) |
| Dispatcher render lại nội dung | code review | hỏng escape của TASK-NOTIF-001 | Chỉ truyền Title/Body đã render (§1 #4) |

---

## §11 - Ghi chú

- Dispatcher email là cái miệng cuối của kênh email - kênh rẻ thứ nhì trong cost model §3.6 (push > email > sms); nó nhặt dòng `queued`, truyền nội dung đã render, ghi `sent`/`failed`, không hơn.
- Interface EmailProvider là điểm linh hoạt cố ý: SES mặc định vì rẻ nhất (~$0,10/1.000 email), nhưng SendGrid/Postmark cắm vào sau cùng hợp đồng chỉ bằng đổi tham số constructor.
- Bounce/complaint từ SES SNS set `verified=false` thay vì xóa hàng: routing TASK-NOTIF-001 tự thôi chọn email, vẫn còn vết chẩn đoán; phân biệt hard bounce (vĩnh viễn, vô hiệu) với soft bounce (tạm thời, retry).
- Bảo vệ reputation là sống còn với email: SES theo dõi bounce rate (ngưỡng ~5%) và complaint rate (ngưỡng ~0,1%) của cả tài khoản; một hộp thư hỏng có thể kéo cả hệ thống email vào review.
- Ranh giới cứng với TASK-NOTIF-001: dispatcher không routing, không render lại Title/Body; chỉ gửi email cho dòng đã định kênh. Push/SMS có dispatcher riêng (TASK-NOTIF-002/007).
- Cấu trúc claim/mark/status-update gương đúng dispatcher FCM (TASK-NOTIF-002) để hai kênh push và email nhất quán: cùng `FOR UPDATE SKIP LOCKED`, cùng `MarkSent`/`MarkFailed`, cùng backoff có jitter chống thundering herd lúc 00:00.

---

*Hết TASK-NOTIF-006. Status: ready_to_review (MVP ready for human review).*
