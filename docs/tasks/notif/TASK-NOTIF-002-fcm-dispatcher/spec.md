---
id: TASK-NOTIF-002
title: "FCM Web/Android dispatcher - gửi push qua FCM HTTP v1, token bucket 600k/phút, xử lý 429 RESOURCE_EXHAUSTED bằng backoff + Retry-After, đánh dấu token chết UNREGISTERED, cập nhật notification.status sent|failed"
module: NOTIF
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-NOTIF-001, TASK-NOTIF-003, TASK-NOTIF-004, TASK-NOTIF-005, TASK-INFRA-002]
depends_on: [TASK-NOTIF-001]
blocks: [TASK-MOBILE-001]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.6 (FCM HTTP v1, quota 600k/phút/project, 429 RESOURCE_EXHAUSTED, token management, UNREGISTERED)"
  - "docs/... §3.6 (mô hình chi phí thông báo: push gần miễn phí là kênh ưu tiên), §3.8 (NFR đỉnh tải 00:00)"
source_decisions:
  - "DEC-NOTIF-10: dùng FCM HTTP v1 API (googleapis.com/v1/projects/{id}/messages:send) với OAuth2 service-account, KHÔNG dùng legacy server-key API (đã deprecated)"
  - "DEC-NOTIF-11: token bucket nội bộ ~600.000 message/phút/project (refill mỗi phút) để tự giới hạn dưới quota mặc định FCM trước khi gửi"
  - "DEC-NOTIF-12: HTTP 429 RESOURCE_EXHAUSTED -> exponential backoff có jitter, tôn trọng header Retry-After; không drop, trả việc về hàng đợi"
  - "DEC-NOTIF-13: phản hồi UNREGISTERED / INVALID_ARGUMENT(token) -> set user_channel_token.verified=false (ngừng gửi tới token chết), notification.status=failed"
  - "DEC-NOTIF-14: dispatcher là consumer downstream của fan-out (TASK-NOTIF-003); chỉ nhặt dòng notification channel='push' status='queued'; gửi xong cập nhật status=sent|failed + sent_at"

language: "Go 1.22 (notif-svc)"
service: shopass/services/notif/
new_files:
  - services/notif/internal/fcm/client.go
  - services/notif/internal/fcm/message.go
  - services/notif/internal/fcm/backoff.go
  - services/notif/internal/fcm/dispatcher.go
  - services/notif/internal/fcm/quota.go
  - services/notif/internal/fcm/client_test.go
  - services/notif/internal/fcm/backoff_test.go
  - services/notif/internal/fcm/dispatcher_test.go
modified_files:
  - services/notif/internal/notif/repo.go            # thêm ClaimPushBatch + MarkSent/MarkFailed + InvalidateToken
allowed_tools:
  - file_read: services/notif/**
  - file_write: services/notif/**
  - bash: cd services/notif && go test ./...
disallowed_tools:
  - dùng legacy FCM server-key API (fcm/send) thay vì HTTP v1 (vi phạm DEC-NOTIF-10, đường đã deprecated)
  - drop thẳng message khi ăn 429 thay vì backoff + trả về hàng đợi (vi phạm DEC-NOTIF-12, mất thông báo)
  - tiếp tục gửi tới token trả UNREGISTERED (vi phạm DEC-NOTIF-13, đốt quota vào token chết)
  - tự routing/chọn kênh hay render lại nội dung (thuộc TASK-NOTIF-001; dispatcher chỉ gửi push đã định kênh)

effort_hours: 8
sub_tasks:
  - "1.0h: client.go - FCM HTTP v1 client, OAuth2 token từ service-account, gọi messages:send, phân loại phản hồi"
  - "0.75h: message.go - dựng JSON message HTTP v1 (token + notification + data + webpush/android override)"
  - "1.25h: backoff.go - exponential backoff có jitter, đọc Retry-After, phân biệt lỗi retry vs lỗi vĩnh viễn"
  - "1.0h: quota.go - token bucket 600k/phút/project, refill theo phút, chặn trước khi vượt quota"
  - "1.5h: dispatcher.go - ClaimPushBatch -> gửi -> MarkSent/MarkFailed + InvalidateToken khi UNREGISTERED"
  - "0.5h: repo.go - ClaimPushBatch (SELECT ... FOR UPDATE SKIP LOCKED status='queued'), MarkSent/MarkFailed, InvalidateToken"
  - "1.5h: 3 test file - send thành công -> status=sent; 429 -> backoff + tôn trọng Retry-After; UNREGISTERED -> verified=false + status=failed; quota; retry"
  - "0.5h: OTel metric fcm_send_total{result} + fcm_retry_total + fcm_token_invalidated_total + fcm_quota_throttled_total"

risk_if_skipped: "TASK-NOTIF-002 là cái miệng cuối cùng của kênh push - kênh rẻ nhất và là ưu tiên số một trong cost model §3.6. Không có nó thì mọi dòng notification channel='push' mà fan-out (TASK-NOTIF-003) xếp vào hàng đợi nằm chết ở status='queued', cảnh báo giá (TASK-TRACK-004) và đáy giá (TASK-DEAL-006) coi như không tới tay người dùng, toàn bộ vòng đời thông báo đứt đúng khúc đáng giá nhất. Nếu vẫn còn dùng legacy server-key API thì đường này đã bị Google ngắt, gửi là lỗi cứng. Nếu không xử lý 429 RESOURCE_EXHAUSTED bằng backoff và tôn trọng Retry-After thì lúc đỉnh 00:00 (giờ vàng flash sale, hàng loạt alert nổ cùng lúc) ta vượt quota 600.000 message/phút/project, FCM trả 429 hàng loạt, và nếu drop thẳng thì người dùng mất đúng cảnh báo họ chờ cả ngày. Nếu không đánh dấu token UNREGISTERED là verified=false thì ta gửi mãi vào token chết của những máy đã gỡ app, đốt quota và rate-limit vào hư không, lại càng dễ chạm trần lúc cao điểm. Token bucket nội bộ là cái van giữ ta luôn dưới ngưỡng trước khi FCM phải từ chối."
---

## §1 - Mô tả (BCP-14 normative)

Dispatcher FCM **MUST** gửi push tới Web và Android qua FCM HTTP v1 API, là consumer downstream của fan-out (TASK-NOTIF-003): nó nhặt các dòng `notification` có `channel='push'` đang `status='queued'`, gửi, rồi cập nhật `status` -> `sent`/`failed` + `sent_at`. Dispatcher KHÔNG routing kênh và KHÔNG render lại nội dung (việc của TASK-NOTIF-001). Hợp đồng:

1. **MUST** gửi qua FCM HTTP v1 endpoint `POST https://fcm.googleapis.com/v1/projects/{project_id}/messages:send` với Bearer OAuth2 access-token lấy từ service-account (DEC-NOTIF-10). **MUST NOT** dùng legacy server-key API (`/fcm/send`) đã deprecated.
2. **MUST** lấy địa chỉ đích (FCM registration token) từ `user_channel_token` của TASK-NOTIF-001: bản ghi `channel='push'`, `verified=true`, và `platform IN ('android','web')` (FCM chỉ phụ trách Android + Web; token `platform='ios'` thuộc TASK-NOTIF-005/APNs và **MUST NOT** bị FCM nhặt). Token chưa verified hoặc đã bị vô hiệu **MUST NOT** được gửi tới.
3. **MUST** dựng JSON message HTTP v1 đúng cấu trúc: `message.token`, `message.notification.{title,body}` (đã render bởi TASK-NOTIF-001), tùy chọn `message.data` (key-value string) cho deep-link, và override theo nền tảng (`webpush`, `android`) khi cần.
4. **MUST** áp một token bucket nội bộ ~`600.000` message/phút/project (refill mỗi phút) trước khi gọi FCM (DEC-NOTIF-11). Quota mặc định của FCM là 600.000 message/phút/project và phủ trên 99% nhà phát triển FCM; van nội bộ giữ ta dưới ngưỡng để hạn chế 429 từ phía server.
5. **MUST** xử lý HTTP `429` `RESOURCE_EXHAUSTED` bằng exponential backoff có jitter và **MUST** tôn trọng header `Retry-After` khi FCM trả về (DEC-NOTIF-12). Message bị 429 **MUST NOT** bị drop; nó được trả lại hàng đợi (giữ `status='queued'`) để thử lại sau, không mất thông báo.
6. **MUST** phân loại phản hồi FCM thành ba nhóm và hành xử đúng:
    - Thành công (HTTP 200, có `name`) -> `MarkSent`: `status='sent'`, `sent_at=now()`.
    - Lỗi tạm thời (429 `RESOURCE_EXHAUSTED`, 500/503 `UNAVAILABLE`, timeout) -> retry theo backoff; vượt số lần thử tối đa thì `MarkFailed` để dead-letter (TASK-NOTIF-003).
    - Lỗi token vĩnh viễn (404 `UNREGISTERED`, 400 `INVALID_ARGUMENT` cho token sai định dạng) -> token chết.
7. **MUST** xử lý `UNREGISTERED` / token không hợp lệ bằng cách set `user_channel_token.verified=false` cho `(user_id, 'push')` tương ứng (DEC-NOTIF-13) và `MarkFailed` dòng `notification`. Lần sau routing (TASK-NOTIF-001) sẽ không coi token đó là khả dụng nữa - ngừng gửi vào token chết.
8. **MUST** nhặt việc an toàn khi chạy nhiều worker song song: `ClaimPushBatch` dùng `SELECT ... FOR UPDATE SKIP LOCKED` trên `notification` (`channel='push'`, `status='queued'`), cập nhật sang trạng thái đang xử lý, để hai worker không gửi trùng một dòng (DEC-NOTIF-14).
9. **MUST** idempotent ở mức hợp lý: một dòng `notification` đã `status='sent'` **MUST NOT** bị gửi lại; `ClaimPushBatch` chỉ lấy `status='queued'`, và `MarkSent`/`MarkFailed` là cập nhật có điều kiện trên trạng thái hiện tại.
10. **MUST** trả error (không panic) khi OAuth2 token hết hạn/làm mới thất bại hoặc FCM trả lỗi không phân loại được, để dispatcher loop áp retry/backoff và không kẹt worker.
11. **SHOULD** phát OTel metric: `fcm_send_total{result}` (counter: sent|retry|failed|token_dead), `fcm_retry_total{reason}` (counter), `fcm_token_invalidated_total` (counter), `fcm_quota_throttled_total` (counter - đếm khi token bucket chặn), `fcm_send_duration_ms` (histogram).
12. **MUST** giữ ranh giới trách nhiệm: dispatcher chỉ gửi push cho dòng đã được TASK-NOTIF-001 chọn kênh và render. Nó **MUST NOT** tự chọn kênh, **MUST NOT** đổi nội dung, và **MUST NOT** gọi email/SMS provider.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao FCM HTTP v1 chứ không legacy server-key (DEC-NOTIF-10)?** Google đã deprecate legacy FCM API (`/fcm/send` với server key tĩnh) và tắt dần. HTTP v1 dùng OAuth2 service-account ngắn hạn, an toàn hơn (không phơi key tĩnh dài hạn), và là đường duy nhất còn được hỗ trợ lâu dài cùng các tính năng override theo nền tảng. Xây thẳng trên v1 để khỏi phải làm lại khi legacy ngừng hẳn.

**Vì sao token bucket 600k/phút nội bộ (DEC-NOTIF-11)?** Quota mặc định FCM là 600.000 message/phút/project và "phủ trên 99% nhà phát triển FCM". Ta đặt một cái van nội bộ ngay trước lời gọi FCM để chủ động ở dưới ngưỡng. Tự throttle rẻ hơn nhiều so với để FCM trả 429 rồi backoff: van nội bộ làm mượt đỉnh trước khi nó thành lỗi server, còn backoff là lưới an toàn cho phần tràn. Hai lớp bổ trợ nhau.

**Vì sao 429 phải backoff + tôn trọng Retry-After, không drop (DEC-NOTIF-12)?** Đỉnh tải của SănDeal là 00:00 - flash sale mở, hàng loạt alert giá nổ cùng lúc (§3.8). Đúng lúc đó là lúc dễ vượt quota nhất. 429 `RESOURCE_EXHAUSTED` là tín hiệu "chậm lại", không phải "bỏ đi". Drop thẳng nghĩa là người dùng mất đúng cảnh báo họ chờ cả ngày. Backoff có jitter tránh cả đàn worker thử lại đồng pha (thundering herd); tôn trọng `Retry-After` là làm theo đúng nhịp FCM yêu cầu thay vì đoán.

**Vì sao UNREGISTERED set verified=false thay vì xóa (DEC-NOTIF-13)?** Token trả `UNREGISTERED` là máy đã gỡ app hoặc token đã xoay vòng. Gửi tiếp vào đó là đốt quota và rate-limit vào hư không, lại càng dễ chạm trần lúc cao điểm. Đặt `verified=false` (thay vì xóa hàng) giữ ranh giới sạch với TASK-NOTIF-001: routing chỉ coi bản ghi `verified=true` là khả dụng, nên token chết tự rớt khỏi vòng chọn kênh mà vẫn còn vết để chẩn đoán. Khi thiết bị đăng ký lại, token mới ghi đè và `verified` bật lại.

**Vì sao consumer của fan-out chứ không tự lấy thẳng từ DB (DEC-NOTIF-14)?** Fan-out (TASK-NOTIF-003) là nơi áp flatten-the-curve và rải tải theo thời gian; dispatcher chỉ là cái miệng cuối nhặt dòng `status='queued'` đã được fan-out chuẩn bị. Tách như vậy cho phép nhiều dispatcher push chạy song song mà không phải tự lo việc rải đỉnh, và giữ logic gửi (rate-limit, backoff, token) gọn trong một chỗ. `FOR UPDATE SKIP LOCKED` làm việc nhặt an toàn khi scale ngang worker.

**Vì sao ranh giới cứng với TASK-NOTIF-001 (§1 #12)?** TASK-NOTIF-001 đã quyết kênh (push là rẻ nhất khả dụng) và render nội dung an toàn rồi. Nếu dispatcher lại đụng vào chọn kênh hay nội dung thì hai chỗ cùng sửa một thứ, khó test và dễ lệch. Dispatcher dừng đúng ở "gửi token này nội dung này qua FCM, ghi kết quả". Trách nhiệm đơn, test gọn.

---

## §3 - Hợp đồng API / DDL

### Repo: nhặt việc + cập nhật trạng thái (Go)

Không cần migration mới - tái dùng `notification` và `user_channel_token` của TASK-NOTIF-001. Chỉ bổ sung hàm repo.

```go
// services/notif/internal/notif/repo.go (bổ sung)

// ClaimPushBatch nhặt tối đa n dòng push đang queued, khóa hàng để worker khác bỏ qua.
// Dùng FOR UPDATE SKIP LOCKED để nhiều dispatcher chạy song song không gửi trùng.
func (r *Repo) ClaimPushBatch(ctx context.Context, n int) ([]PushJob, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT n.id, n.user_id, t.address AS token, n.payload
        FROM notification n
        JOIN user_channel_token t
          ON t.user_id = n.user_id AND t.channel = 'push' AND t.verified = true
        WHERE n.channel = 'push' AND n.status = 'queued'
        ORDER BY n.scheduled_at NULLS FIRST, n.id
        FOR UPDATE OF n SKIP LOCKED
        LIMIT $1`, n)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var jobs []PushJob
    for rows.Next() {
        var j PushJob
        if err := rows.Scan(&j.NotifID, &j.UserID, &j.Token, &j.Payload); err != nil {
            return nil, err
        }
        jobs = append(jobs, j)
    }
    return jobs, rows.Err()
}

// MarkSent: queued -> sent + sent_at (chỉ khi còn queued, idempotent).
func (r *Repo) MarkSent(ctx context.Context, id int64) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE notification SET status='sent', sent_at=now()
         WHERE id=$1 AND status='queued'`, id)
    return err
}

// MarkFailed: queued -> failed (vượt retry hoặc token chết); fan-out lo dead-letter.
func (r *Repo) MarkFailed(ctx context.Context, id int64) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE notification SET status='failed'
         WHERE id=$1 AND status='queued'`, id)
    return err
}

// InvalidateToken: tắt token chết (UNREGISTERED) -> routing TASK-NOTIF-001 thôi chọn push.
func (r *Repo) InvalidateToken(ctx context.Context, userID int64) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE user_channel_token SET verified=false, updated_at=now()
         WHERE user_id=$1 AND channel='push'`, userID)
    return err
}
```

### FCM HTTP v1 client (Go)

```go
// services/notif/internal/fcm/client.go

// SendResult phân loại phản hồi FCM để dispatcher hành xử (§1 #6).
type SendResult int

const (
    ResultSent      SendResult = iota // 200 OK, có name
    ResultRetry                       // 429/500/503/timeout -> backoff
    ResultTokenDead                   // UNREGISTERED / INVALID_ARGUMENT token
    ResultFailed                      // lỗi vĩnh viễn khác
)

type SendOutcome struct {
    Result     SendResult
    RetryAfter time.Duration // > 0 khi FCM trả header Retry-After (§1 #5)
}

// Send gửi một message qua FCM HTTP v1 (DEC-NOTIF-10).
func (c *Client) Send(ctx context.Context, msg Message) (SendOutcome, error) {
    body, _ := json.Marshal(map[string]any{"message": msg})
    tok, err := c.oauth.Token(ctx) // OAuth2 access-token từ service-account
    if err != nil {
        return SendOutcome{Result: ResultRetry}, err
    }
    url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.projectID)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+tok)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.http.Do(req)
    if err != nil {
        return SendOutcome{Result: ResultRetry}, err // timeout/mạng -> retry
    }
    defer resp.Body.Close()
    return classify(resp), nil
}

// classify ánh xạ HTTP status + error code FCM sang SendOutcome.
func classify(resp *http.Response) SendOutcome {
    switch resp.StatusCode {
    case 200:
        return SendOutcome{Result: ResultSent}
    case 429:
        return SendOutcome{Result: ResultRetry, RetryAfter: parseRetryAfter(resp.Header)}
    case 500, 503:
        return SendOutcome{Result: ResultRetry, RetryAfter: parseRetryAfter(resp.Header)}
    case 404:
        return SendOutcome{Result: ResultTokenDead} // UNREGISTERED
    case 400:
        // INVALID_ARGUMENT: token sai định dạng -> coi như token chết
        return SendOutcome{Result: ResultTokenDead}
    default:
        return SendOutcome{Result: ResultFailed}
    }
}
```

### Backoff + token bucket (Go)

```go
// services/notif/internal/fcm/backoff.go

// nextDelay: exponential có jitter; ưu tiên Retry-After của FCM khi có (§1 #5).
func nextDelay(attempt int, retryAfter time.Duration) time.Duration {
    if retryAfter > 0 {
        return retryAfter // tôn trọng FCM (DEC-NOTIF-12)
    }
    base := time.Duration(1<<attempt) * time.Second // 1s,2s,4s,8s...
    if base > 5*time.Minute {
        base = 5 * time.Minute
    }
    jitter := time.Duration(rand.Int63n(int64(base / 2)))
    return base/2 + jitter // ±50% tránh thundering herd lúc 00:00
}

// services/notif/internal/fcm/quota.go

// Bucket: token bucket ~600k/phút/project (DEC-NOTIF-11), refill mỗi phút.
type Bucket struct {
    mu       sync.Mutex
    tokens   int
    capacity int
    last     time.Time
}

func NewBucket() *Bucket { return &Bucket{tokens: 600_000, capacity: 600_000, last: time.Now()} }

// Allow trả false khi đã chạm trần phút này -> dispatcher chờ/throttle (fcm_quota_throttled_total).
func (b *Bucket) Allow() bool {
    b.mu.Lock()
    defer b.mu.Unlock()
    if time.Since(b.last) >= time.Minute {
        b.tokens, b.last = b.capacity, time.Now() // refill theo phút
    }
    if b.tokens <= 0 {
        return false
    }
    b.tokens--
    return true
}
```

---

## §4 - Acceptance criteria

1. `Client.Send` gọi đúng endpoint HTTP v1 `.../v1/projects/{project_id}/messages:send` với header `Authorization: Bearer <oauth-token>`; KHÔNG gọi legacy `/fcm/send`.
2. Phản hồi 200 (có `name`) -> `MarkSent`: dòng `notification` chuyển `status='sent'` và `sent_at` được set.
3. Phản hồi 429 `RESOURCE_EXHAUSTED` -> `SendOutcome.Result=ResultRetry`; dòng giữ `status='queued'` (không bị drop), worker lập lịch thử lại theo backoff.
4. Khi 429 kèm header `Retry-After: 30` -> `nextDelay` trả đúng 30s (tôn trọng FCM thay vì công thức exponential).
5. Không có `Retry-After` -> `nextDelay(attempt)` tăng theo cấp số nhân có jitter, có trần (<= 5 phút).
6. Phản hồi 404 `UNREGISTERED` -> `ResultTokenDead`; `InvalidateToken` set `user_channel_token.verified=false`, dòng `notification` -> `status='failed'`.
7. Phản hồi 400 `INVALID_ARGUMENT` (token sai định dạng) -> `ResultTokenDead`, xử lý như token chết.
8. `ClaimPushBatch` chỉ trả dòng `channel='push'`, `status='queued'`, và chỉ ghép với `user_channel_token` `verified=true`; token chưa verified KHÔNG được nhặt.
9. Hai worker `ClaimPushBatch` đồng thời (FOR UPDATE SKIP LOCKED) KHÔNG nhặt trùng cùng một dòng.
10. Token bucket: gọi thứ 600.001 trong cùng một phút -> `Allow()` trả false (`fcm_quota_throttled_total` tăng); sau khi sang phút mới -> refill, `Allow()` trả true.
11. Dòng đã `status='sent'` -> `MarkSent`/`MarkFailed` lần nữa không đổi gì (cập nhật có điều kiện `status='queued'`); không gửi lại.
12. Metric `fcm_send_total{result}` tăng đúng nhánh (sent/retry/failed/token_dead); `fcm_token_invalidated_total` tăng khi gặp UNREGISTERED.

---

## §5 - Kiểm thử (verification)

```go
// services/notif/internal/fcm/client_test.go
func TestSend_Success_MarksSent(t *testing.T) {
    srv := stubFCM(t, 200, `{"name":"projects/p/messages/123"}`, nil)
    c := newClient(t, srv.URL)
    out, err := c.Send(ctx, Message{Token: "tok-ok"})
    require.NoError(t, err)
    require.Equal(t, ResultSent, out.Result)
}

func TestSend_UsesHTTPv1Endpoint_NotLegacy(t *testing.T) {
    var gotPath string
    srv := captureRequest(t, &gotPath, 200, `{"name":"n"}`)
    c := newClient(t, srv.URL)
    c.Send(ctx, Message{Token: "tok"})
    require.Contains(t, gotPath, "/v1/projects/")
    require.Contains(t, gotPath, "messages:send")
    require.NotContains(t, gotPath, "/fcm/send") // legacy không được dùng (DEC-NOTIF-10)
}

func TestSend_429_TriggersRetry_NotDropped(t *testing.T) {
    srv := stubFCM(t, 429, `{"error":{"status":"RESOURCE_EXHAUSTED"}}`, nil)
    c := newClient(t, srv.URL)
    out, _ := c.Send(ctx, Message{Token: "tok"})
    require.Equal(t, ResultRetry, out.Result) // không drop, sẽ thử lại (DEC-NOTIF-12)
}

func TestSend_429_RespectsRetryAfter(t *testing.T) {
    srv := stubFCM(t, 429, `{}`, http.Header{"Retry-After": []string{"30"}})
    c := newClient(t, srv.URL)
    out, _ := c.Send(ctx, Message{Token: "tok"})
    require.Equal(t, 30*time.Second, nextDelay(0, out.RetryAfter)) // tôn trọng Retry-After
}

func TestSend_Unregistered_MarksTokenDead(t *testing.T) {
    srv := stubFCM(t, 404, `{"error":{"status":"UNREGISTERED"}}`, nil)
    c := newClient(t, srv.URL)
    out, _ := c.Send(ctx, Message{Token: "tok-dead"})
    require.Equal(t, ResultTokenDead, out.Result) // -> InvalidateToken + status=failed (DEC-NOTIF-13)
}

// services/notif/internal/fcm/dispatcher_test.go
func TestDispatch_Unregistered_InvalidatesTokenAndFails(t *testing.T) {
    d, repo, uid, nid := setupQueuedPush(t, stubFCM(t, 404, `{"error":{"status":"UNREGISTERED"}}`, nil))
    d.RunOnce(ctx)
    require.False(t, repo.tokenVerified(t, uid))           // verified=false
    require.Equal(t, "failed", repo.status(t, nid))        // notification.status=failed
}

func TestDispatch_Success_SetsSentAt(t *testing.T) {
    d, repo, _, nid := setupQueuedPush(t, stubFCM(t, 200, `{"name":"n"}`, nil))
    d.RunOnce(ctx)
    require.Equal(t, "sent", repo.status(t, nid))
    require.NotNil(t, repo.sentAt(t, nid))
}

// services/notif/internal/fcm/backoff_test.go
func TestQuota_BlocksOverLimitThenRefills(t *testing.T) {
    b := NewBucket()
    for i := 0; i < 600_000; i++ { require.True(t, b.Allow()) }
    require.False(t, b.Allow())     // vượt 600k/phút -> chặn (DEC-NOTIF-11)
    b.last = time.Now().Add(-2 * time.Minute)
    require.True(t, b.Allow())      // sang phút mới -> refill
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `repo.go` (ClaimPushBatch + MarkSent/MarkFailed + InvalidateToken trên schema sẵn của TASK-NOTIF-001) -> `message.go` (dựng JSON HTTP v1) -> `client.go` (Send + classify) -> `backoff.go` + `quota.go` (retry + van quota) -> `dispatcher.go` (vòng RunOnce: claim -> Allow -> Send -> mark) -> tests. Dispatcher chạy thành pool worker, mỗi worker `ClaimPushBatch` rồi gửi; `FOR UPDATE SKIP LOCKED` giữ an toàn khi scale ngang. Fan-out (TASK-NOTIF-003) là thứ đẩy dòng vào `status='queued'` và lo rải đỉnh; dispatcher chỉ nhặt và gửi.

---

## §7 - Phụ thuộc

- **TASK-NOTIF-001** - bảng `notification` (đọc `channel='push'`, cập nhật `status`/`sent_at`) và `user_channel_token` (đọc token `verified=true`, set `verified=false` khi token chết). Dispatcher KHÔNG routing/render - hai việc đó thuộc TASK-NOTIF-001.
- **TASK-NOTIF-003 (upstream)** - fan-out chuyển dòng `notification` push sang `status='queued'` và rải đỉnh (flatten-the-curve); dispatcher là consumer downstream nhặt các dòng đó. Dead-letter các dòng `status='failed'` cũng do fan-out lo.
- **TASK-INFRA-002** - `app_user` cho FK `user_id` (gián tiếp qua schema TASK-NOTIF-001).
- Lib: `golang.org/x/oauth2/google` (OAuth2 service-account cho HTTP v1), `net/http`, `encoding/json`, `pgx` (FOR UPDATE SKIP LOCKED).

---

## §8 - Payload ví dụ

### Request gửi FCM HTTP v1 (dispatcher dựng từ dòng notification push)

```json
{
  "message": {
    "token": "fcm-registration-token-cua-thiet-bi",
    "notification": {
      "title": "Giá đã giảm về mức bạn chờ",
      "body": "Sản phẩm bạn theo dõi còn 79.000 VND."
    },
    "data": { "product_id": "90112", "deeplink": "sandeal://product/90112" },
    "webpush": { "fcm_options": { "link": "https://sandeal.vn/p/90112" } },
    "android": { "priority": "high" }
  }
}
```

### Dòng notification: queued -> sent sau khi gửi thành công

```sql
-- Trước (fan-out đã đẩy vào hàng đợi):
-- 55012 | push | price_below | queued | sent_at = NULL
UPDATE notification SET status='sent', sent_at=now() WHERE id=55012 AND status='queued';
-- Sau: 55012 | push | price_below | sent | sent_at = 2026-06-28 00:00:03+07
```

### Token chết (UNREGISTERED): tắt verified, dòng -> failed

```sql
UPDATE user_channel_token SET verified=false, updated_at=now() WHERE user_id=4021 AND channel='push';
UPDATE notification        SET status='failed'                  WHERE id=55013 AND status='queued';
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- APNs (iOS) là dispatcher riêng (TASK-NOTIF-005) - cùng vòng đời `notification` nhưng provider khác; không gộp vào task này.
- Multicast / `messages:send` theo batch nhiều token một lần - tối ưu throughput giai đoạn sau nếu volume push đủ lớn để cần.
- Xin tăng quota tạm thời trước sự kiện đỉnh (vd siêu sale 11.11 vượt 600k/phút) - quy trình vận hành ở §11, chưa cần code.
- Đo deliverability thật (push tới mà người dùng có thấy không) - cần tín hiệu client-side, để task analytics sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Dùng legacy server-key API | test `UsesHTTPv1Endpoint` | gửi lỗi cứng (đã deprecated) | Chỉ gọi HTTP v1 messages:send (DEC-NOTIF-10) |
| 429 RESOURCE_EXHAUSTED lúc đỉnh 00:00 | `fcm_send_total{retry}` | nguy cơ mất push nếu drop | Backoff + Retry-After, giữ queued, thử lại (DEC-NOTIF-12) |
| Bỏ qua Retry-After, thử lại quá sớm | test `RespectsRetryAfter` | bị siết tiếp, vòng lặp 429 | `nextDelay` ưu tiên Retry-After của FCM |
| Token UNREGISTERED (app đã gỡ) | HTTP 404 + classify | đốt quota vào token chết | `InvalidateToken` verified=false + status=failed (DEC-NOTIF-13) |
| Token sai định dạng | HTTP 400 INVALID_ARGUMENT | gửi hỏng lặp lại | Coi như token chết, ngừng gửi |
| Vượt 600k/phút từ phía ta | token bucket `Allow()=false` | FCM sẽ trả 429 | Van nội bộ chặn trước (DEC-NOTIF-11), `fcm_quota_throttled_total` |
| Hai worker gửi trùng một dòng | test concurrency | push đúp tới user | FOR UPDATE SKIP LOCKED (§1 #8) |
| OAuth2 token hết hạn/làm mới lỗi | Send trả ResultRetry+err | không gửi được | Retry; cache + refresh token service-account |
| Gửi lại dòng đã sent | cập nhật có điều kiện | spam người dùng | MarkSent/Claim chỉ chạm status='queued' (§1 #9) |

---

## §11 - Ghi chú

- Dispatcher FCM là cái miệng cuối của kênh push - kênh ưu tiên số một trong cost model §3.6 (push gần miễn phí); nó nhặt dòng `queued`, gửi, ghi `sent`/`failed`, không hơn.
- Quota mặc định FCM là 600.000 message/phút/project và phủ trên 99% nhà phát triển; token bucket nội bộ giữ ta dưới ngưỡng, backoff là lưới an toàn cho phần tràn.
- Xin tăng quota phải đặt trước: gửi yêu cầu trước >= 15 ngày cho mức thường; nếu cần > 18.000.000 message/phút thì trước >= 30 ngày; mỗi project chỉ được tối đa 2 sự kiện temp-quota mỗi năm. Lên lịch trước các đợt siêu sale.
- Ranh giới cứng với TASK-NOTIF-001: dispatcher không routing, không render lại nội dung; chỉ gửi push cho dòng đã định kênh. Email/SMS có dispatcher riêng (TASK-NOTIF-006/007).
- UNREGISTERED set `verified=false` thay vì xóa hàng: routing tự thôi chọn push, vẫn còn vết chẩn đoán; thiết bị đăng ký lại sẽ ghi đè token và bật lại verified.
- Backoff có jitter cố ý +/-50% để cả đàn worker không thử lại đồng pha sau một đợt 429 hàng loạt lúc 00:00 (chống thundering herd).

---

*Hết TASK-NOTIF-002. Status: ready_to_implement (mục tiêu audit 10/10).*
