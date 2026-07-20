---
id: TASK-NOTIF-005
title: "APNs iOS dispatcher - gửi push tới iPhone/iPad qua APNs HTTP/2 (api.push.apple.com), nhiều kết nối song song multiplexing, token-based auth JWT provider token (.p8), xử lý 410 token hết hạn -> verified=false, backoff cho 500/503, cập nhật notification.status sent|failed"
module: NOTIF
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-NOTIF-001, TASK-NOTIF-002, TASK-NOTIF-003, TASK-NOTIF-004, TASK-INFRA-002]
depends_on: [TASK-NOTIF-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.6 (APNs iOS: nhiều kết nối HTTP/2 song song tới api.push.apple.com, status 410 token hết hạn -> gỡ token, backoff 500/503, token-based auth JWT provider token .p8)"
  - "docs/... §3.6 (mô hình chi phí thông báo: push gần miễn phí là kênh ưu tiên), §3.8 (NFR đỉnh tải 00:00)"
source_decisions:
  - "DEC-NOTIF-50: dùng APNs HTTP/2 provider API (POST api.push.apple.com/3/device/{token}) với token-based auth (JWT provider token ký bằng .p8 ES256), KHÔNG dùng certificate-based legacy/binary; provider token cache <60 phút rồi ký lại"
  - "DEC-NOTIF-51: mở nhiều kết nối HTTP/2 song song tới api.push.apple.com và multiplex nhiều request trên mỗi kết nối để đạt throughput (một kết nối HTTP/2 không đủ cho đỉnh 00:00)"
  - "DEC-NOTIF-52: HTTP 410 (Unregistered, apns-id token không còn active) -> set user_channel_token.verified=false cho (user_id,'push') (gỡ token chết), notification.status=failed; không gửi lại token đó"
  - "DEC-NOTIF-53: HTTP 500/503 (InternalServerError / ServiceUnavailable, lỗi server APNs) -> exponential backoff có jitter, tôn trọng header Retry-After khi có; không drop, trả việc về hàng đợi (giữ status='queued')"
  - "DEC-NOTIF-54: dispatcher là consumer downstream của fan-out (TASK-NOTIF-003); chỉ nhặt dòng notification channel='push' status='queued' cho thiết bị iOS (platform='ios'); gửi xong cập nhật status=sent|failed + sent_at. KHÔNG routing/render (việc của TASK-NOTIF-001)"

language: "Go 1.22 (notif-svc)"
service: shopass/services/notif/
new_files:
  - services/notif/internal/apns/client.go
  - services/notif/internal/apns/payload.go
  - services/notif/internal/apns/jwt.go
  - services/notif/internal/apns/pool.go
  - services/notif/internal/apns/backoff.go
  - services/notif/internal/apns/dispatcher.go
  - services/notif/internal/apns/client_test.go
  - services/notif/internal/apns/pool_test.go
  - services/notif/internal/apns/dispatcher_test.go
modified_files:
  - services/notif/internal/notif/repo.go            # thêm ClaimIOSPushBatch (lọc platform='ios') tái dùng MarkSent/MarkFailed/InvalidateToken
allowed_tools:
  - file_read: services/notif/**
  - file_write: services/notif/**
  - bash: cd services/notif && go test ./...
disallowed_tools:
  - dùng certificate-based / binary legacy APNs API thay vì HTTP/2 token-based (vi phạm DEC-NOTIF-50, đường cũ khó vận hành và đang bị thu hẹp)
  - mở đúng một kết nối HTTP/2 rồi gửi tuần tự (vi phạm DEC-NOTIF-51, không đủ throughput lúc đỉnh)
  - drop thẳng message khi ăn 500/503 thay vì backoff + trả về hàng đợi (vi phạm DEC-NOTIF-53, mất thông báo)
  - tiếp tục gửi tới token trả 410 Unregistered (vi phạm DEC-NOTIF-52, đốt kết nối vào token chết)
  - tự routing/chọn kênh hay render lại nội dung (thuộc TASK-NOTIF-001; dispatcher chỉ gửi push iOS đã định kênh)

effort_hours: 5
sub_tasks:
  - "0.75h: jwt.go - ký JWT provider token ES256 từ .p8 (iss=team_id, kid=key_id), cache <60 phút rồi ký lại, header authorization: bearer"
  - "1.0h: client.go - APNs HTTP/2 client, gọi POST /3/device/{token} với apns-topic/apns-push-type/apns-priority, phân loại phản hồi theo status (200/410/429/500/503)"
  - "0.5h: payload.go - dựng aps payload JSON (alert title/body đã render bởi TASK-NOTIF-001, sound, badge, custom keys deep-link)"
  - "0.75h: pool.go - connection pool nhiều kết nối HTTP/2 song song tới api.push.apple.com, multiplex request, round-robin/least-busy"
  - "0.5h: backoff.go - exponential backoff có jitter cho 500/503, đọc Retry-After, phân biệt lỗi retry vs token chết"
  - "0.75h: dispatcher.go - ClaimIOSPushBatch -> gửi song song qua pool -> MarkSent/MarkFailed + InvalidateToken khi 410"
  - "0.25h: repo.go - ClaimIOSPushBatch (SELECT ... FOR UPDATE SKIP LOCKED channel='push' status='queued' platform='ios'); tái dùng MarkSent/MarkFailed/InvalidateToken của TASK-NOTIF-002"
  - "0.5h: 3 test file - send 200 -> status=sent; 410 -> verified=false + status=failed; 500/503 -> backoff + giữ queued; pool song song; status update"

risk_if_skipped: "TASK-NOTIF-005 là cái miệng cuối của kênh push cho iOS - nửa thiết bị của thị trường ứng dụng tiêu dùng VN/SEA chạy iPhone/iPad, và push là kênh rẻ nhất, ưu tiên số một trong cost model §3.6. Không có nó thì mọi dòng notification channel='push' của người dùng iOS mà fan-out (TASK-NOTIF-003) xếp vào hàng đợi nằm chết ở status='queued': cảnh báo giá (TASK-TRACK-004) và đáy giá (TASK-DEAL-006) không bao giờ tới iPhone, một nửa tập người dùng đáng giá nhất bị bỏ rơi đúng khúc cuối, dù TASK-NOTIF-002 đã lo xong Android/Web. Nếu chỉ mở một kết nối HTTP/2 rồi gửi tuần tự thì lúc đỉnh 00:00 (flash sale mở, hàng loạt alert nổ cùng lúc, §3.8) throughput nghẽn cổ chai, push iOS tới trễ hàng phút - đúng lúc cần nhanh nhất thì chậm nhất. Nếu không xử lý 410 (token không còn active vì người dùng gỡ app hoặc token xoay vòng) bằng cách set verified=false thì ta gửi mãi vào token chết, đốt kết nối và quota vào hư không, lại càng dễ nghẽn lúc cao điểm. Nếu không backoff cho 500/503 mà drop thẳng thì khi APNs có sự cố server thoáng qua, người dùng iOS mất đúng cảnh báo họ chờ cả ngày. Nếu vẫn ôm certificate-based legacy API thì vận hành nặng (chứng chỉ hết hạn, xoay vòng thủ công) và lệch khỏi đường token-based mà phần còn lại của hệ thống đã chuẩn hóa."
---

## §1 - Mô tả (BCP-14 normative)

Dispatcher APNs **MUST** gửi push tới iPhone/iPad qua APNs HTTP/2 provider API (`api.push.apple.com`), là consumer downstream của fan-out (TASK-NOTIF-003): nó nhặt các dòng `notification` có `channel='push'` đang `status='queued'` cho thiết bị iOS, gửi qua nhiều kết nối HTTP/2 song song, rồi cập nhật `status` -> `sent`/`failed` + `sent_at`. Dispatcher KHÔNG routing kênh và KHÔNG render lại nội dung (việc của TASK-NOTIF-001). Đây là anh em iOS của TASK-NOTIF-002 (FCM cho Android/Web); cùng vòng đời `notification`, khác nhà cung cấp. Hợp đồng:

1. **MUST** gửi qua APNs HTTP/2 endpoint `POST https://api.push.apple.com/3/device/{device_token}` với token-based auth (DEC-NOTIF-50). **MUST NOT** dùng certificate-based / binary legacy API.
2. **MUST** xác thực bằng JWT provider token (DEC-NOTIF-50): JWT ký thuật toán `ES256` bằng khóa `.p8` (APNs Auth Key), header `alg=ES256` + `kid=<key_id>`, claims `iss=<team_id>` + `iat=<now>`; đặt vào header `authorization: bearer <jwt>`. Provider token **MUST** được cache và ký lại trước khi quá 60 phút (APNs từ chối token cũ hơn 1 giờ với 403 `ExpiredProviderToken`).
3. **MUST** đặt các header APNs bắt buộc cho mỗi request: `apns-topic` (bundle id của app iOS), `apns-push-type` (`alert` cho cảnh báo giá), `apns-priority` (10 cho alert hiển thị ngay), và `apns-id` (UUID idempotency, log để truy vết). Header `apns-expiration` **SHOULD** đặt để APNs không giữ message quá lâu nếu thiết bị offline.
4. **MUST** lấy `device_token` đích từ `user_channel_token` của TASK-NOTIF-001: bản ghi `channel='push'`, `verified=true`, và thuộc thiết bị iOS. Token chưa verified hoặc đã bị vô hiệu **MUST NOT** được gửi tới.
5. **MUST** mở **nhiều kết nối HTTP/2 song song** tới `api.push.apple.com` và **multiplex** nhiều request trên mỗi kết nối (DEC-NOTIF-51). Một kết nối HTTP/2 đơn không đủ throughput cho đỉnh 00:00; pool nhiều kết nối + multiplexing là cách APNs khuyến nghị để đạt thông lượng cao.
6. **MUST** dựng aps payload JSON đúng cấu trúc: `aps.alert.{title,body}` (đã render bởi TASK-NOTIF-001), tùy chọn `aps.sound`, `aps.badge`, và custom key ngoài `aps` (vd `product_id`, `deeplink`) cho điều hướng. Payload **MUST** giữ dưới giới hạn 4KB của APNs cho alert push.
7. **MUST** phân loại phản hồi APNs theo HTTP status + `reason` và hành xử đúng:
- Thành công (HTTP `200`) -> `MarkSent`: `status='sent'`, `sent_at=now()`.
- Token chết (HTTP `410` `Unregistered`, hoặc `400` `BadDeviceToken`) -> token không còn active.
- Lỗi server tạm thời (HTTP `500` `InternalServerError`, `503` `ServiceUnavailable`, timeout) -> retry theo backoff.
- Bị siết (HTTP `429` `TooManyRequests`) -> backoff, tôn trọng `Retry-After`.
8. **MUST** xử lý HTTP `410` (`Unregistered` - apns-id/device token không còn active) bằng cách set `user_channel_token.verified=false` cho `(user_id, 'push')` tương ứng (DEC-NOTIF-52) và `MarkFailed` dòng `notification`. Lần sau routing (TASK-NOTIF-001) sẽ không coi token đó là khả dụng nữa - ngừng gửi vào token chết. HTTP `400 BadDeviceToken` xử lý cùng nhánh token chết.
9. **MUST** xử lý HTTP `500`/`503` (lỗi server APNs) bằng exponential backoff có jitter và **MUST** tôn trọng header `Retry-After` khi APNs trả về (DEC-NOTIF-53). Message bị 500/503 **MUST NOT** bị drop; nó được trả lại hàng đợi (giữ `status='queued'`) để thử lại sau, không mất thông báo. Vượt số lần thử tối đa thì `MarkFailed` để dead-letter (TASK-NOTIF-003).
10. **MUST** nhặt việc an toàn khi chạy nhiều worker song song: `ClaimIOSPushBatch` dùng `SELECT ... FOR UPDATE SKIP LOCKED` trên `notification` (`channel='push'`, `status='queued'`, thiết bị iOS), để hai worker không gửi trùng một dòng (DEC-NOTIF-54). Một dòng `notification` đã `status='sent'` **MUST NOT** bị gửi lại; `MarkSent`/`MarkFailed` là cập nhật có điều kiện trên trạng thái hiện tại.
11. **MUST** trả error (không panic) khi ký JWT provider token thất bại (`.p8` lỗi), kết nối HTTP/2 đứt, hoặc APNs trả lỗi không phân loại được, để dispatcher loop áp retry/backoff và không kẹt worker. Header `apns-id` trả về **SHOULD** được log để đối soát với APNs khi cần.
12. **MUST** giữ ranh giới trách nhiệm: dispatcher chỉ gửi push cho dòng đã được TASK-NOTIF-001 chọn kênh và render. Nó **MUST NOT** tự chọn kênh, **MUST NOT** đổi nội dung, và **MUST NOT** gọi FCM/email/SMS provider. **SHOULD** phát OTel metric: `apns_send_total{result}` (sent|retry|failed|token_dead), `apns_retry_total{reason}`, `apns_token_invalidated_total`, `apns_connections_active` (gauge), `apns_send_duration_ms` (histogram).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao APNs HTTP/2 token-based chứ không certificate-based legacy (DEC-NOTIF-50)?** APNs hiện đại là HTTP/2 với hai cách xác thực: certificate (.p12 gắn từng app, hết hạn hàng năm, xoay vòng thủ công nặng) và token-based (JWT provider token ký bằng một khóa .p8 dùng cho nhiều app/topic, không hết hạn cứng). Token-based nhẹ vận hành hơn hẳn: một khóa .p8, ký JWT ngắn hạn theo nhu cầu, không phải canh ngày hết hạn chứng chỉ. Phần còn lại của hệ thống (FCM cũng dùng OAuth2 ngắn hạn) đã chuẩn hóa theo hướng token ngắn hạn; làm APNs theo cùng kiểu giữ vận hành đồng nhất. Provider token phải ký lại trước 60 phút vì APNs từ chối token cũ hơn 1 giờ.

**Vì sao nhiều kết nối HTTP/2 song song + multiplexing (DEC-NOTIF-51)?** Đỉnh tải SănDeal là 00:00 - flash sale mở, hàng loạt alert giá nổ cùng lúc (§3.8), và một nửa số đó là người dùng iOS. HTTP/2 cho phép multiplex nhiều request trên một kết nối, nhưng một kết nối đơn vẫn có trần (số stream đồng thời, độ trễ round-trip). APNs khuyến nghị mở nhiều kết nối song song và rải request qua chúng để đạt throughput cao. Pool nhiều kết nối + multiplexing trên mỗi kết nối là hai lớp song song bổ trợ: nhiều kết nối phá trần stream của một kết nối, multiplexing tận dụng tối đa mỗi kết nối. Gửi tuần tự trên một kết nối là tự bóp cổ chai đúng lúc cần nhanh nhất.

**Vì sao 410 set verified=false thay vì xóa (DEC-NOTIF-52)?** APNs trả `410 Unregistered` khi device token không còn active - máy đã gỡ app, hoặc iOS xoay vòng token. Gửi tiếp vào đó là đốt kết nối và công sức vào hư không, lại càng dễ nghẽn lúc cao điểm. Đặt `verified=false` (thay vì xóa hàng) giữ ranh giới sạch với TASK-NOTIF-001 giống hệt cách TASK-NOTIF-002 xử lý `UNREGISTERED` của FCM: routing chỉ coi bản ghi `verified=true` là khả dụng, nên token chết tự rớt khỏi vòng chọn kênh mà vẫn còn vết để chẩn đoán. Khi thiết bị đăng ký lại, token mới ghi đè và `verified` bật lại. Một cơ chế, hai nhà cung cấp - dễ suy luận và đối xứng với anh em FCM.

**Vì sao 500/503 phải backoff + tôn trọng Retry-After, không drop (DEC-NOTIF-53)?** `500 InternalServerError` và `503 ServiceUnavailable` là lỗi phía server APNs, thường thoáng qua. Đúng lúc đỉnh 00:00 là lúc mọi hệ thống chịu tải nặng nhất, kể cả APNs. Đây là tín hiệu "thử lại sau", không phải "bỏ đi". Drop thẳng nghĩa là người dùng iOS mất đúng cảnh báo họ chờ cả ngày chỉ vì một trục trặc server tạm thời. Backoff có jitter tránh cả đàn worker thử lại đồng pha (thundering herd) sau một đợt lỗi hàng loạt; tôn trọng `Retry-After` là làm theo đúng nhịp APNs yêu cầu thay vì đoán. Cùng triết lý retry như TASK-NOTIF-002 áp cho 429/500/503 của FCM.

**Vì sao consumer của fan-out chứ không tự lấy thẳng từ DB (DEC-NOTIF-54)?** Fan-out (TASK-NOTIF-003) là nơi áp flatten-the-curve và rải tải theo thời gian; dispatcher APNs chỉ là cái miệng cuối nhặt dòng `status='queued'` đã được fan-out chuẩn bị, lọc đúng thiết bị iOS. Tách như vậy cho phép nhiều dispatcher APNs chạy song song mà không phải tự lo việc rải đỉnh, và giữ logic gửi (pool kết nối, backoff, token JWT) gọn trong một chỗ. `FOR UPDATE SKIP LOCKED` làm việc nhặt an toàn khi scale ngang worker. TASK-NOTIF-003 route các dòng `channel='push'` của thiết bị iOS sang dispatcher này, còn dòng push của Android/Web sang TASK-NOTIF-002.

**Vì sao ranh giới cứng với TASK-NOTIF-001 (§1 #12)?** TASK-NOTIF-001 đã quyết kênh (push là rẻ nhất khả dụng) và render nội dung an toàn rồi. Nếu dispatcher lại đụng vào chọn kênh hay nội dung thì hai chỗ cùng sửa một thứ, khó test và dễ lệch. Dispatcher APNs dừng đúng ở "gửi token iOS này nội dung này qua APNs HTTP/2, ghi kết quả". Trách nhiệm đơn, test gọn, và đối xứng hoàn toàn với TASK-NOTIF-002 ở phía FCM.

---

## §3 - Hợp đồng API / DDL

### Repo: nhặt việc iOS (Go)

Không cần migration mới - tái dùng `notification` và `user_channel_token` của TASK-NOTIF-001, và tái dùng `MarkSent`/`MarkFailed`/`InvalidateToken` đã thêm ở TASK-NOTIF-002. Chỉ bổ sung hàm nhặt việc lọc thiết bị iOS.

```go
// services/notif/internal/notif/repo.go (bổ sung)

// ClaimIOSPushBatch nhặt tối đa n dòng push iOS đang queued, khóa hàng để worker khác bỏ qua.
// Lọc platform='ios' qua user_channel_token; FOR UPDATE SKIP LOCKED để nhiều dispatcher APNs
// chạy song song không gửi trùng. MarkSent/MarkFailed/InvalidateToken tái dùng từ TASK-NOTIF-002.
func (r *Repo) ClaimIOSPushBatch(ctx context.Context, n int) ([]PushJob, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT n.id, n.user_id, t.address AS token, n.payload
        FROM notification n
        JOIN user_channel_token t
          ON t.user_id = n.user_id AND t.channel = 'push' AND t.verified = true
        WHERE n.channel = 'push' AND n.status = 'queued' AND t.platform = 'ios'
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
```

### JWT provider token ES256 từ .p8 (Go)

```go
// services/notif/internal/apns/jwt.go

// ProviderToken ký JWT ES256 bằng .p8 (DEC-NOTIF-50), cache <60 phút rồi ký lại.
type ProviderToken struct {
    mu      sync.Mutex
    keyID   string            // kid (header)
    teamID  string            // iss (claim)
    key     *ecdsa.PrivateKey // nạp từ .p8
    cached  string
    issued  time.Time
}

// Bearer trả JWT còn hạn; ký lại nếu cũ hơn ~55 phút (APNs từ chối token > 1 giờ).
func (p *ProviderToken) Bearer() (string, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    if p.cached != "" && time.Since(p.issued) < 55*time.Minute {
        return p.cached, nil
    }
    now := time.Now()
    header := map[string]string{"alg": "ES256", "kid": p.keyID}
    claims := map[string]any{"iss": p.teamID, "iat": now.Unix()}
    signed, err := signES256(header, claims, p.key) // base64url(header).base64url(claims).sig
    if err != nil {
        return "", err
    }
    p.cached, p.issued = signed, now
    return signed, nil
}
```

### APNs HTTP/2 client (Go)

```go
// services/notif/internal/apns/client.go

// SendResult phân loại phản hồi APNs để dispatcher hành xử (§1 #7).
type SendResult int

const (
    ResultSent      SendResult = iota // 200 OK
    ResultRetry                       // 429/500/503/timeout -> backoff
    ResultTokenDead                   // 410 Unregistered / 400 BadDeviceToken
    ResultFailed                      // lỗi vĩnh viễn khác
)

type SendOutcome struct {
    Result     SendResult
    RetryAfter time.Duration // > 0 khi APNs trả header Retry-After (§1 #9)
    APNsID     string        // header apns-id để đối soát/log (§1 #11)
}

// Send gửi một message qua APNs HTTP/2 (DEC-NOTIF-50). conn là một kết nối từ pool (§1 #5).
func (c *Client) Send(ctx context.Context, conn *http2Conn, token string, p Payload) (SendOutcome, error) {
    body, _ := json.Marshal(p)
    bearer, err := c.provider.Bearer() // JWT provider token, cache <60 phút
    if err != nil {
        return SendOutcome{Result: ResultRetry}, err
    }
    url := fmt.Sprintf("https://api.push.apple.com/3/device/%s", token)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    req.Header.Set("authorization", "bearer "+bearer)
    req.Header.Set("apns-topic", c.bundleID)        // bundle id app iOS
    req.Header.Set("apns-push-type", "alert")
    req.Header.Set("apns-priority", "10")           // alert hiển thị ngay
    req.Header.Set("apns-id", newUUID())            // idempotency + truy vết

    resp, err := conn.Do(req) // multiplex trên kết nối HTTP/2 sẵn (§1 #5)
    if err != nil {
        return SendOutcome{Result: ResultRetry}, err // timeout/mạng -> retry
    }
    defer resp.Body.Close()
    return classify(resp), nil
}

// classify ánh xạ HTTP status + reason APNs sang SendOutcome.
func classify(resp *http.Response) SendOutcome {
    out := SendOutcome{APNsID: resp.Header.Get("apns-id")}
    switch resp.StatusCode {
    case 200:
        out.Result = ResultSent
    case 410:
        out.Result = ResultTokenDead // Unregistered: token không còn active (DEC-NOTIF-52)
    case 400:
        out.Result = ResultTokenDead // BadDeviceToken -> coi như token chết
    case 429:
        out.Result, out.RetryAfter = ResultRetry, parseRetryAfter(resp.Header) // TooManyRequests
    case 500, 503:
        out.Result, out.RetryAfter = ResultRetry, parseRetryAfter(resp.Header) // lỗi server (DEC-NOTIF-53)
    default:
        out.Result = ResultFailed
    }
    return out
}
```

### Connection pool song song + backoff (Go)

```go
// services/notif/internal/apns/pool.go

// Pool giữ nhiều kết nối HTTP/2 song song tới api.push.apple.com (DEC-NOTIF-51).
// Multiplex nhiều request trên mỗi kết nối; chọn kết nối theo round-robin.
type Pool struct {
    conns []*http2Conn
    next  uint64 // round-robin counter (atomic)
}

func NewPool(size int, dial func() (*http2Conn, error)) (*Pool, error) {
    p := &Pool{conns: make([]*http2Conn, 0, size)}
    for i := 0; i < size; i++ {
        c, err := dial() // mỗi kết nối là một HTTP/2 tới api.push.apple.com
        if err != nil {
            return nil, err
        }
        p.conns = append(p.conns, c)
    }
    return p, nil
}

// Pick chọn kết nối kế tiếp (round-robin) để rải request qua nhiều kết nối song song.
func (p *Pool) Pick() *http2Conn {
    i := atomic.AddUint64(&p.next, 1)
    return p.conns[i%uint64(len(p.conns))]
}

// services/notif/internal/apns/backoff.go

// nextDelay: exponential có jitter; ưu tiên Retry-After của APNs khi có (§1 #9).
func nextDelay(attempt int, retryAfter time.Duration) time.Duration {
    if retryAfter > 0 {
        return retryAfter // tôn trọng APNs (DEC-NOTIF-53)
    }
    base := time.Duration(1<<attempt) * time.Second // 1s,2s,4s,8s...
    if base > 5*time.Minute {
        base = 5 * time.Minute
    }
    jitter := time.Duration(rand.Int63n(int64(base / 2)))
    return base/2 + jitter // ±50% tránh thundering herd lúc 00:00
}
```

---

## §4 - Acceptance criteria

1. `Client.Send` gọi đúng endpoint HTTP/2 `https://api.push.apple.com/3/device/{token}` với header `authorization: bearer <jwt>`, `apns-topic`, `apns-push-type`, `apns-priority`; KHÔNG dùng certificate/binary legacy.
2. JWT provider token ký `ES256` với `kid`/`iss` đúng; cache được tái dùng trong < 60 phút và ký lại sau ngưỡng (~55 phút), không ký mới mỗi request.
3. Phản hồi `200` -> `MarkSent`: dòng `notification` chuyển `status='sent'` và `sent_at` được set.
4. Phản hồi `410` `Unregistered` -> `ResultTokenDead`; `InvalidateToken` set `user_channel_token.verified=false`, dòng `notification` -> `status='failed'`.
5. Phản hồi `400` `BadDeviceToken` -> `ResultTokenDead`, xử lý như token chết (cùng nhánh với 410).
6. Phản hồi `500`/`503` -> `SendOutcome.Result=ResultRetry`; dòng giữ `status='queued'` (không bị drop), worker lập lịch thử lại theo backoff.
7. Khi `500`/`503`/`429` kèm header `Retry-After: 30` -> `nextDelay` trả đúng 30s (tôn trọng APNs thay vì công thức exponential).
8. Không có `Retry-After` -> `nextDelay(attempt)` tăng theo cấp số nhân có jitter, có trần (<= 5 phút).
9. Pool mở đúng `size` kết nối HTTP/2 tới `api.push.apple.com`; `Pick()` rải request qua các kết nối theo round-robin (nhiều kết nối song song, §1 #5).
10. `ClaimIOSPushBatch` chỉ trả dòng `channel='push'`, `status='queued'`, `platform='ios'`, và chỉ ghép với `user_channel_token` `verified=true`; token chưa verified hoặc thiết bị non-iOS KHÔNG được nhặt.
11. Hai worker `ClaimIOSPushBatch` đồng thời (FOR UPDATE SKIP LOCKED) KHÔNG nhặt trùng cùng một dòng; dòng đã `status='sent'` -> `MarkSent`/`MarkFailed` lần nữa không đổi gì.
12. Metric `apns_send_total{result}` tăng đúng nhánh (sent/retry/failed/token_dead); `apns_token_invalidated_total` tăng khi gặp 410; `apns_connections_active` phản ánh số kết nối pool.

---

## §5 - Kiểm thử (verification)

```go
// services/notif/internal/apns/client_test.go
func TestSend_Success_MarksSent(t *testing.T) {
    srv := stubAPNS(t, 200, "", nil) // 200 OK, body rỗng
    c := newClient(t, srv.URL)
    out, err := c.Send(ctx, dial(t, srv.URL), "tok-ok", Payload{})
    require.NoError(t, err)
    require.Equal(t, ResultSent, out.Result)
}

func TestSend_UsesHTTP2DeviceEndpoint_TokenAuth(t *testing.T) {
    var gotPath, gotAuth, gotTopic string
    srv := captureRequest(t, &gotPath, &gotAuth, &gotTopic, 200, "")
    c := newClient(t, srv.URL)
    c.Send(ctx, dial(t, srv.URL), "tok", Payload{})
    require.Contains(t, gotPath, "/3/device/")            // endpoint HTTP/2
    require.True(t, strings.HasPrefix(gotAuth, "bearer ")) // token-based JWT (DEC-NOTIF-50)
    require.NotEmpty(t, gotTopic)                          // apns-topic bắt buộc
}

func TestSend_410_MarksTokenDead(t *testing.T) {
    srv := stubAPNS(t, 410, `{"reason":"Unregistered"}`, nil)
    c := newClient(t, srv.URL)
    out, _ := c.Send(ctx, dial(t, srv.URL), "tok-dead", Payload{})
    require.Equal(t, ResultTokenDead, out.Result) // -> InvalidateToken + status=failed (DEC-NOTIF-52)
}

func TestSend_500_TriggersRetry_NotDropped(t *testing.T) {
    srv := stubAPNS(t, 500, `{"reason":"InternalServerError"}`, nil)
    c := newClient(t, srv.URL)
    out, _ := c.Send(ctx, dial(t, srv.URL), "tok", Payload{})
    require.Equal(t, ResultRetry, out.Result) // không drop, sẽ thử lại (DEC-NOTIF-53)
}

func TestSend_503_RespectsRetryAfter(t *testing.T) {
    srv := stubAPNS(t, 503, `{"reason":"ServiceUnavailable"}`,
        http.Header{"Retry-After": []string{"30"}})
    c := newClient(t, srv.URL)
    out, _ := c.Send(ctx, dial(t, srv.URL), "tok", Payload{})
    require.Equal(t, 30*time.Second, nextDelay(0, out.RetryAfter)) // tôn trọng Retry-After
}

// services/notif/internal/apns/pool_test.go
func TestPool_OpensParallelConnsAndRoundRobins(t *testing.T) {
    var dialed int32
    p, err := NewPool(4, func() (*http2Conn, error) {
        atomic.AddInt32(&dialed, 1)
        return fakeConn(), nil
    })
    require.NoError(t, err)
    require.Equal(t, int32(4), dialed)              // 4 kết nối song song (DEC-NOTIF-51)
    a, b := p.Pick(), p.Pick()
    require.NotSame(t, a, b)                         // rải qua các kết nối khác nhau
}

// services/notif/internal/apns/dispatcher_test.go
func TestDispatch_410_InvalidatesTokenAndFails(t *testing.T) {
    d, repo, uid, nid := setupQueuedIOSPush(t, stubAPNS(t, 410, `{"reason":"Unregistered"}`, nil))
    d.RunOnce(ctx)
    require.False(t, repo.tokenVerified(t, uid))    // verified=false
    require.Equal(t, "failed", repo.status(t, nid)) // notification.status=failed
}

func TestDispatch_Success_SetsSentAt(t *testing.T) {
    d, repo, _, nid := setupQueuedIOSPush(t, stubAPNS(t, 200, "", nil))
    d.RunOnce(ctx)
    require.Equal(t, "sent", repo.status(t, nid))
    require.NotNil(t, repo.sentAt(t, nid))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `repo.go` (ClaimIOSPushBatch trên schema sẵn của TASK-NOTIF-001, tái dùng MarkSent/MarkFailed/InvalidateToken của TASK-NOTIF-002) -> `jwt.go` (ký JWT provider token ES256 từ .p8, cache <60 phút) -> `payload.go` (dựng aps JSON dưới 4KB) -> `client.go` (Send + classify) -> `pool.go` (nhiều kết nối HTTP/2 song song + round-robin) + `backoff.go` (retry 500/503) -> `dispatcher.go` (vòng RunOnce: claim -> Pick kết nối -> Send song song -> mark) -> tests. Dispatcher chạy thành pool worker, mỗi worker `ClaimIOSPushBatch` rồi gửi qua `Pool.Pick()`; `FOR UPDATE SKIP LOCKED` giữ an toàn khi scale ngang. Fan-out (TASK-NOTIF-003) đẩy dòng push iOS vào `status='queued'` và lo rải đỉnh; dispatcher chỉ nhặt và gửi. Khóa `.p8`, `team_id`, `key_id`, `bundle_id` nạp từ cấu hình/secret store, không hard-code.

---

## §7 - Phụ thuộc

- **TASK-NOTIF-001** - bảng `notification` (đọc `channel='push'`, cập nhật `status`/`sent_at`) và `user_channel_token` (đọc `device_token` `verified=true` của thiết bị iOS, set `verified=false` khi token chết). Dispatcher KHÔNG routing/render - hai việc đó thuộc TASK-NOTIF-001.
- **TASK-NOTIF-003 (upstream)** - fan-out chuyển dòng `notification` push sang `status='queued'`, rải đỉnh (flatten-the-curve), và route dòng push của thiết bị iOS sang dispatcher này; dispatcher là consumer downstream. Dead-letter các dòng `status='failed'` cũng do fan-out lo.
- **TASK-NOTIF-002 (anh em)** - dispatcher FCM cho Android/Web là anh em của task này: cùng vòng đời `notification`, cùng pattern claim/mark, cùng triết lý backoff và đánh dấu token chết (`UNREGISTERED` của FCM tương ứng `Unregistered` 410 của APNs); khác nhà cung cấp và cách auth (OAuth2 service-account vs JWT provider token .p8). `MarkSent`/`MarkFailed`/`InvalidateToken` được TASK-NOTIF-002 thêm vào `repo.go` và task này tái dùng.
- **TASK-INFRA-002** - `app_user` cho FK `user_id` (gián tiếp qua schema TASK-NOTIF-001).
- Lib: `golang.org/x/net/http2` (kết nối HTTP/2 tới APNs), `crypto/ecdsa` + `crypto/x509` (nạp .p8, ký ES256), `encoding/json`, `pgx` (FOR UPDATE SKIP LOCKED).

---

## §8 - Payload ví dụ

### Request gửi APNs HTTP/2 (dispatcher dựng từ dòng notification push iOS)

```
POST /3/device/0a1b2c3d-device-token-cua-iphone HTTP/2
Host: api.push.apple.com
authorization: bearer eyJhbGciOiJFUzI1NiIsImtpZCI6IkFCQzEyM30...
apns-topic: vn.sandeal.app
apns-push-type: alert
apns-priority: 10
apns-id: 9d9b1c4e-0f2a-4d3b-8a7e-1c2d3e4f5a6b
apns-expiration: 0

{
  "aps": {
    "alert": {
      "title": "Giá đã giảm về mức bạn chờ",
      "body": "Sản phẩm bạn theo dõi còn 79.000 VND."
    },
    "sound": "default",
    "badge": 1
  },
  "product_id": "90112",
  "deeplink": "sandeal://product/90112"
}
```

### Dòng notification: queued -> sent sau khi gửi thành công

```sql
-- Trước (fan-out đã đẩy vào hàng đợi, thiết bị iOS):
-- 55021 | push | price_below | queued | sent_at = NULL
UPDATE notification SET status='sent', sent_at=now() WHERE id=55021 AND status='queued';
-- Sau: 55021 | push | price_below | sent | sent_at = 2026-06-28 00:00:03+07
```

### Token chết (410 Unregistered): tắt verified, dòng -> failed

```sql
UPDATE user_channel_token SET verified=false, updated_at=now() WHERE user_id=4022 AND channel='push';
UPDATE notification        SET status='failed'                  WHERE id=55022 AND status='queued';
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Đọc phản hồi feedback hàng loạt / kênh thu hồi token APNs ngoài 410 trên đường gửi - hiện đủ dùng 410 inline; thêm cơ chế quét định kỳ nếu cần dọn token chết chủ động.
- Tinh chỉnh số kết nối pool theo tải thực (autoscale số kết nối HTTP/2 theo độ sâu hàng đợi) - bắt đầu với hằng số cấu hình, tối ưu sau khi có số đo lúc đỉnh.
- Live Activities / push tới Apple Watch / `apns-push-type` khác (`background`, `voip`) - ngoài phạm vi cảnh báo giá, mở khi có tính năng cần.
- Đa app/topic (nhiều bundle id) ký bằng cùng một .p8 - khóa .p8 dùng được cho nhiều topic; mở rộng map `apns-topic` khi có app thứ hai.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Dùng certificate/binary legacy API | test `UsesHTTP2DeviceEndpoint_TokenAuth` | vận hành nặng, lệch chuẩn | Chỉ gọi HTTP/2 `/3/device/{token}` token-based (DEC-NOTIF-50) |
| JWT provider token > 1 giờ | HTTP 403 ExpiredProviderToken | gửi bị từ chối | Cache <60 phút, ký lại trước ~55 phút (§1 #2) |
| 410 Unregistered (app đã gỡ/token xoay) | HTTP 410 + classify | đốt kết nối vào token chết | `InvalidateToken` verified=false + status=failed (DEC-NOTIF-52) |
| 400 BadDeviceToken | HTTP 400 + reason | gửi hỏng lặp lại | Coi như token chết, ngừng gửi (cùng nhánh 410) |
| 500/503 lỗi server APNs lúc đỉnh 00:00 | `apns_send_total{retry}` | nguy cơ mất push nếu drop | Backoff + Retry-After, giữ queued, thử lại (DEC-NOTIF-53) |
| Bỏ qua Retry-After, thử lại quá sớm | test `RespectsRetryAfter` | bị siết tiếp, vòng lặp lỗi | `nextDelay` ưu tiên Retry-After của APNs |
| Một kết nối HTTP/2 nghẽn lúc đỉnh | `apns_connections_active` thấp | throughput iOS nghẽn cổ chai | Pool nhiều kết nối song song + multiplex (DEC-NOTIF-51) |
| Hai worker gửi trùng một dòng | test concurrency | push đúp tới user iOS | FOR UPDATE SKIP LOCKED (§1 #10) |
| Payload vượt 4KB | APNs từ chối PayloadTooLarge | không gửi được | Giữ aps payload gọn dưới 4KB (§1 #6) |

---

## §11 - Ghi chú

- Dispatcher APNs là cái miệng cuối của kênh push cho iOS - anh em của TASK-NOTIF-002 (FCM Android/Web); cùng vòng đời `notification`, cùng pattern claim/mark, khác nhà cung cấp và cách auth.
- Token-based auth (JWT provider token ký bằng .p8) nhẹ vận hành hơn certificate-based: một khóa .p8 cho nhiều topic, ký JWT ngắn hạn theo nhu cầu, không phải canh ngày hết hạn chứng chỉ; phải ký lại trước 60 phút.
- Nhiều kết nối HTTP/2 song song + multiplexing là cách APNs khuyến nghị đạt throughput cao; một kết nối đơn tự thành cổ chai lúc đỉnh 00:00.
- 410 set `verified=false` thay vì xóa hàng: routing TASK-NOTIF-001 tự thôi chọn push, vẫn còn vết chẩn đoán; thiết bị đăng ký lại sẽ ghi đè token và bật lại verified - đối xứng hoàn toàn với cách TASK-NOTIF-002 xử lý `UNREGISTERED` của FCM.
- 500/503 backoff + tôn trọng Retry-After, không drop: lỗi server APNs thường thoáng qua, drop thẳng làm người dùng iOS mất cảnh báo họ chờ cả ngày; jitter +/-50% chống thundering herd sau đợt lỗi hàng loạt.
- Ranh giới cứng với TASK-NOTIF-001: dispatcher không routing, không render lại nội dung; chỉ gửi push iOS cho dòng đã định kênh. FCM (Android/Web), email, SMS có dispatcher riêng.

---

*Hết TASK-NOTIF-005. Status: ready_to_implement (mục tiêu audit 10/10).*
