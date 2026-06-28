---
id: FR-NOTIF-003
title: "Fan-out pipeline - producer -> Kafka/Redis Streams -> fan-out worker -> per-channel dispatcher, at-least-once + idempotent (claim/lease trên notification), backoff + jitter + dead-letter queue cho thông báo lỗi"
module: NOTIF
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-NOTIF-001, FR-NOTIF-002, FR-NOTIF-004, FR-NOTIF-005, FR-NOTIF-006, FR-NOTIF-007]
depends_on: [FR-NOTIF-001]
blocks: [FR-NOTIF-004, FR-NOTIF-005, FR-NOTIF-006, FR-NOTIF-007]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.6 (kiến trúc fan-out: producer -> Kafka/Redis Streams -> fan-out worker -> per-channel dispatcher FCM/APNs/Email/SMS; backoff + jitter + DLQ)"
  - "docs/... §3.6 (at-least-once + idempotent consumer; mô hình chi phí thông báo push > email > sms)"
source_decisions:
  - "DEC-NOTIF-20: pipeline là producer -> hàng đợi bền (Kafka hoặc Redis Streams) -> fan-out worker -> per-channel dispatcher; fan-out CHỈ định tuyến/điều phối, KHÔNG tự gọi nhà cung cấp (gọi thật ở FR-NOTIF-002/005/006/007)"
  - "DEC-NOTIF-21: ngữ nghĩa at-least-once với consumer idempotent - một dòng notification chỉ được gửi đúng một lần, đảm bảo bằng claim/lease (UPDATE pending -> queued WHERE status='pending' RETURNING) trên chính dòng notification"
  - "DEC-NOTIF-22: retry với exponential backoff + jitter; vượt max_attempts -> đẩy sang dead-letter queue (DLQ) để soi, KHÔNG retry vô hạn"
  - "DEC-NOTIF-23: fan-out định tuyến theo notification.channel (đã được FR-NOTIF-001 routing chốt) tới đúng handler kênh; dispatcher trả Permanent vs Transient để pipeline quyết DLQ ngay hay retry"
  - "DEC-NOTIF-24: lease có thời hạn (lease_until); job claim mà worker crash trước khi ack -> hết hạn lease được re-claim, không kẹt vĩnh viễn ở 'queued'"
  - "DEC-NOTIF-25: FR-NOTIF-004 (scheduler chống dồn 00:00) đẩy việc VÀO pipeline này; fan-out là tầng tiêu thụ, scheduler là tầng điều tiết đầu vào"

language: "Go 1.22 (notif-svc); Kafka hoặc Redis Streams"
service: shopass/services/notif/
new_files:
  - services/notif/migrations/0003_notification_lease.sql
  - services/notif/migrations/0004_notification_dlq.sql
  - services/notif/internal/fanout/producer.go
  - services/notif/internal/fanout/worker.go
  - services/notif/internal/fanout/dispatch.go
  - services/notif/internal/fanout/backoff.go
  - services/notif/internal/fanout/dlq.go
  - services/notif/internal/fanout/worker_test.go
  - services/notif/internal/fanout/backoff_test.go
  - services/notif/internal/fanout/dispatch_test.go
modified_files:
  - services/notif/internal/notif/repo.go               # thêm ClaimPending (lease) + MarkSent/MarkFailed
allowed_tools:
  - file_read: services/notif/**
  - file_write: services/notif/**
  - bash: cd services/notif && go test ./...
disallowed_tools:
  - tự gọi FCM/APNs/SES/SendGrid/SMS trong package fanout (vi phạm DEC-NOTIF-20; gọi thật là việc của FR-NOTIF-002/005/006/007)
  - gửi một dòng notification 2 lần (vi phạm DEC-NOTIF-21; phải claim/lease idempotent)
  - retry vô hạn một job hỏng vĩnh viễn (vi phạm DEC-NOTIF-22; vượt max_attempts -> DLQ)
  - backoff cố định không jitter (vi phạm DEC-NOTIF-22; gây đồng pha retry, dồn tải nhà cung cấp lúc đỉnh 00:00)

effort_hours: 8
sub_tasks:
  - "0.5h: 0003_notification_lease.sql - thêm cột attempts, lease_until, last_error vào notification + index claim"
  - "0.5h: 0004_notification_dlq.sql - bảng notification_dlq (notification_id, channel, payload, attempts, last_error, dead_at)"
  - "1.0h: producer.go - Enqueue(notificationID) đẩy message vào Kafka/Redis Streams (key=user_id để giữ thứ tự per-user)"
  - "1.5h: worker.go - vòng lặp claim (UPDATE pending->queued RETURNING) -> dispatch -> ack/retry/DLQ; re-claim lease hết hạn"
  - "1.0h: dispatch.go - interface ChannelDispatcher + bảng định tuyến theo channel; phân loại Permanent vs Transient error"
  - "1.0h: backoff.go - exponential backoff + full jitter, trần (cap) thời gian chờ"
  - "0.5h: dlq.go - PublishDLQ(notificationID, reason) ghi notification_dlq + status='failed'"
  - "1.5h: worker_test.go - at-least-once giao hàng, idempotent không gửi đôi, re-claim lease hết hạn, định tuyến đúng kênh"
  - "0.5h: backoff_test.go + dispatch_test.go - backoff tăng + jitter trong biên; DLQ sau max_attempts; Permanent -> DLQ ngay"

risk_if_skipped: "Fan-out pipeline là tầng vận chuyển nối engine alert với người dùng cuối; thiếu nó thì FR-NOTIF-001 ghi được dòng notification pending nhưng không có gì nhặt ra và đẩy tới dispatcher, mọi cảnh báo giá nằm chết trong bảng và sản phẩm mất đi tính năng lõi của nó là báo đúng lúc. Nếu pipeline không idempotent (không claim/lease) thì một dòng notification có thể bị hai worker cùng nhặt và gửi đôi: người dùng nhận hai push giống hệt cho một lần giảm giá, niềm tin rơi và app bị tắt thông báo - mà tắt thông báo là mất luôn affiliate, nguồn sống của mô hình free-tier. Nếu không có backoff + jitter thì khi FCM/SES chập chờn lúc đỉnh 00:00 (giờ flash sale các sàn), hàng nghìn worker retry đồng pha sẽ tự dồn tải lên nhà cung cấp, ăn 429 dây chuyền, và cơn bão retry kéo dài cả pipeline. Nếu không có DLQ thì một dòng hỏng vĩnh viễn (token rác, payload sai) sẽ retry vô hạn, chiếm slot worker, đẩy lùi mọi thông báo hợp lệ phía sau và che mất tín hiệu lỗi cần con người soi. Đây là khúc xương sống vận chuyển của module thông báo; vỡ ở đây là vỡ ở khâu cuối cùng người dùng thực sự thấy."
---

## §1 - Mô tả (BCP-14 normative)

Service NOTIF **MUST** cung cấp một fan-out pipeline tiêu thụ các dòng `notification` ở trạng thái `pending`, định tuyến mỗi dòng tới đúng per-channel dispatcher, với ngữ nghĩa at-least-once + consumer idempotent, retry backoff có jitter, và dead-letter queue cho job hỏng. Pipeline CHỈ định tuyến và điều phối; nó KHÔNG tự gọi nhà cung cấp. Hợp đồng:

1. **MUST** dựng đường ống `producer -> hàng đợi bền (Kafka hoặc Redis Streams) -> fan-out worker -> per-channel dispatcher` (DEC-NOTIF-20). Producer chỉ đẩy `notification_id` (con trỏ tới dòng đã ghi bởi FR-NOTIF-001), KHÔNG nhồi toàn bộ nội dung vào message - nguồn sự thật là bảng `notification`.
2. **MUST** giữ ngữ nghĩa at-least-once: một message có thể được giao lại (worker crash trước khi ack, re-balance partition). Pipeline **MUST** chịu được giao lại mà KHÔNG gửi đôi nội dung tới người dùng (DEC-NOTIF-21).
3. **MUST** đảm bảo idempotent bằng claim/lease trên chính dòng `notification`: worker **MUST** giành quyền xử lý qua `UPDATE notification SET status='queued', lease_until=now()+lease, attempts=attempts+1 WHERE id=$1 AND status='pending' RETURNING ...`. Chỉ worker nhận được dòng RETURNING (CAS thắng) mới được gửi; worker thua coi như đã có người khác xử lý và bỏ qua. Một dòng `notification` được gửi đúng một lần.
4. **MUST** định tuyến theo `notification.channel` (đã được FR-NOTIF-001 routing chốt thành đúng một kênh trong enum `{push, email, sms}`) tới đúng `ChannelDispatcher` đã đăng ký (DEC-NOTIF-23): `push` tới dispatcher push (FR-NOTIF-002 cho Android/Web, FR-NOTIF-005 cho iOS - phân theo `user_channel_token.platform`, KHÔNG phải theo một giá trị channel riêng); `email -> FR-NOTIF-006`; `sms -> FR-NOTIF-007`. `channel` chỉ nhận `{push, email, sms}` theo CHECK của FR-NOTIF-001 - KHÔNG có giá trị `apns` riêng. Fan-out KHÔNG chứa mã gọi FCM/APNs/SES/SendGrid/nhà-cung-cấp-SMS.
5. **MUST** dùng exponential backoff + jitter cho retry (DEC-NOTIF-22): lần thử thứ `n` chờ `base * 2^(n-1)` rồi cộng jitter ngẫu nhiên, có trần `cap`. Backoff **MUST** có thành phần jitter để chống đồng pha retry giữa nhiều worker (tránh dồn tải nhà cung cấp lúc đỉnh).
6. **MUST** có dead-letter queue: khi `attempts` vượt `max_attempts` (mặc định 5), pipeline **MUST** ghi dòng vào bảng `notification_dlq`, đặt `notification.status='failed'`, và NGỪNG retry. DLQ giữ `notification_id`, `channel`, `payload` snapshot, `attempts`, `last_error`, `dead_at` để con người soi.
7. **MUST** phân biệt lỗi Permanent với Transient từ dispatcher (DEC-NOTIF-23): dispatcher trả `Permanent` (vd token không hợp lệ, payload sai, 4xx không-retry) -> đẩy DLQ NGAY không retry; trả `Transient` (vd 429/5xx/timeout mạng) -> retry theo backoff cho tới `max_attempts`.
8. **MUST** cập nhật trạng thái cuối: dispatch thành công -> `UPDATE notification SET status='sent', sent_at=now()`; thất bại tạm -> giữ `queued` cho lần retry (đặt lại `lease_until`); thất bại vĩnh viễn hoặc cạn retry -> `status='failed'` + DLQ. Vòng đời trạng thái khớp FR-NOTIF-001: `pending -> queued -> sent | failed`.
9. **MUST** re-claim job có lease hết hạn (DEC-NOTIF-24): dòng `status='queued'` mà `lease_until < now()` (worker giữ nó đã crash) **MUST** được một worker khác claim lại qua `UPDATE ... WHERE id=$1 AND status='queued' AND lease_until < now() RETURNING`. Không dòng nào kẹt vĩnh viễn ở `queued`.
10. **MUST** dùng `idx_notif_dispatch` (đã tạo ở FR-NOTIF-001, partial trên `status IN ('pending','queued')`) làm nguồn lấy việc cho producer/scheduler; thêm cột `attempts`, `lease_until`, `last_error` vào `notification` qua migration của FR này.
11. **MUST** giữ thứ tự per-user khi dùng hàng đợi phân vùng: message **MUST** được phân vùng theo khóa `user_id` (Kafka partition key / Redis Streams consumer group theo user-hash) để các thông báo của cùng một người không bị đảo thứ tự bất thường giữa các partition.
12. **SHOULD** phát OTel: `notif_fanout_dispatched_total{channel}` (counter), `notif_fanout_retry_total{channel, reason}` (counter), `notif_fanout_dlq_total{channel, reason}` (counter), `notif_fanout_dispatch_duration_ms{channel}` (histogram), `notif_fanout_inflight{channel}` (gauge), `notif_fanout_double_claim_total` (counter - đếm lần CAS thua, kỳ vọng 0 trong điều kiện thường).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao hàng đợi bền Kafka/Redis Streams (DEC-NOTIF-20)?** Đỉnh thông báo của SănDeal dồn vào 00:00 khi các sàn mở flash sale double-date (1.1, 2.2 ... 12.12): hàng chục nghìn alert giá bung trong vài phút. Gọi dispatcher trực tiếp từ engine alert (đồng bộ) sẽ vỡ ngay vì rate-limit FCM/SES và độ trễ mạng. Một hàng đợi bền hấp thụ cơn sóng, tách "lúc sinh ra alert" khỏi "lúc gửi đi", cho phép worker tiêu thụ theo nhịp nhà cung cấp chịu được. Bền (không in-memory) vì worker sẽ crash; mất message là mất thông báo người dùng đang chờ.

**Vì sao at-least-once + idempotent thay vì exactly-once (DEC-NOTIF-21)?** Exactly-once xuyên qua ranh giới mạng (queue + DB + nhà cung cấp ngoài) là ảo tưởng đắt đỏ. Cách vững là chấp nhận giao lại (at-least-once, rẻ và đáng tin) rồi làm consumer idempotent để giao lại không gây hại. Claim/lease bằng `UPDATE ... WHERE status='pending' RETURNING` là một compare-and-swap nguyên tử trong Postgres: chỉ một worker thắng, mọi giao lại sau đó thấy `status<>'pending'` và lặng lẽ bỏ qua. Người dùng nhận đúng một push cho một lần giảm giá - gửi đôi là cách nhanh nhất bị tắt thông báo.

**Vì sao DLQ thay vì retry mãi (DEC-NOTIF-22)?** Một số lỗi không bao giờ tự khỏi: token push đã gỡ app, payload dị dạng, số điện thoại sai định dạng. Retry vô hạn các job này đốt slot worker, đẩy lùi thông báo hợp lệ phía sau, và biến lỗi "cần người sửa" thành nhiễu nền vô hình. DLQ tách dòng chết ra một chỗ riêng để soi, giải phóng pipeline, và biến mỗi mục DLQ thành một tín hiệu cụ thể (token rác hàng loạt? adapter kênh hỏng?) thay vì một vòng lặp câm.

**Vì sao backoff cần jitter (DEC-NOTIF-22, §1 #5)?** Nếu FCM trả 429 lúc 00:00, hàng nghìn worker cùng đánh dấu "thử lại sau 2 giây" rồi cùng bắn lại ở giây thứ 2 - một đợt sóng đồng pha tự gây ra 429 tiếp, lặp mãi (thundering herd). Cộng jitter ngẫu nhiên vào mỗi khoảng chờ làm các lần retry rải đều theo thời gian, để nhà cung cấp kịp hồi và pipeline thoát khỏi cộng hưởng. Đây là cùng một bài học pacing/jitter của scraping farm (FR-SCRAPE-001/005), áp vào phía gửi thông báo.

**Vì sao tách per-channel dispatcher khỏi fan-out (DEC-NOTIF-20, DEC-NOTIF-23)?** Mỗi kênh có giao thức, SDK, lược đồ lỗi, và rate-limit riêng: FCM/APNs là push protocol với token; SES/SendGrid là email với bounce; SMS là gateway với mã lỗi nhà mạng. Nhồi hết vào fan-out tạo một khối khổng lồ khó test và khó thay nhà cung cấp. Fan-out chỉ giữ phần chung (claim, retry, backoff, DLQ, định tuyến); mỗi dispatcher (FR-NOTIF-002/005/006/007) là một implementation `ChannelDispatcher` cắm vào theo `channel`. Thêm kênh mới không đụng vào xương sống pipeline.

**Vì sao phân vùng theo user_id (§1 #11)?** Một người có thể nhận chuỗi thông báo liên quan trong khoảng ngắn (giá chạm ngưỡng, rồi giảm sâu thêm). Nếu các message của cùng người rải ngẫu nhiên qua nhiều partition và worker, thứ tự gửi có thể đảo, gây trải nghiệm khó hiểu. Khóa phân vùng `user_id` giữ thông báo của một người đi qua cùng một partition theo thứ tự, mà vẫn cho phép phân tải ngang giữa nhiều người khác nhau.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/notif/migrations/0003_notification_lease.sql
-- Bổ sung cột phục vụ claim/lease + retry cho fan-out (FR-NOTIF-003).
ALTER TABLE notification
  ADD COLUMN attempts    INTEGER     NOT NULL DEFAULT 0,
  ADD COLUMN lease_until TIMESTAMPTZ,
  ADD COLUMN last_error  TEXT;

-- idx_notif_dispatch (status, scheduled_at) WHERE status IN ('pending','queued')
-- đã tạo ở FR-NOTIF-001; phục vụ producer/scheduler lấy việc và re-claim lease hết hạn.

-- services/notif/migrations/0004_notification_dlq.sql
-- Dead-letter: dòng hỏng vĩnh viễn hoặc cạn retry, tách ra để con người soi.
CREATE TABLE notification_dlq (
  id              BIGSERIAL   PRIMARY KEY,
  notification_id BIGINT      NOT NULL REFERENCES notification(id),
  channel         TEXT        NOT NULL,
  payload         JSONB,                       -- snapshot lúc chết
  attempts        INTEGER     NOT NULL,
  last_error      TEXT        NOT NULL,
  reason          TEXT        NOT NULL,         -- 'permanent' | 'max_attempts'
  dead_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_dlq_channel ON notification_dlq (channel, dead_at DESC);
```

### Dispatcher interface + định tuyến (Go)

```go
// services/notif/internal/fanout/dispatch.go

// errClass phân loại lỗi để pipeline quyết DLQ-ngay vs retry.
type errClass int

const (
    ClassOK        errClass = iota // gửi thành công
    ClassTransient                 // 429/5xx/timeout -> retry theo backoff
    ClassPermanent                 // token rác/payload sai/4xx -> DLQ ngay
)

// ChannelDispatcher là per-channel gửi thật, hiện thực ở FR-NOTIF-002/005/006/007.
// Fan-out KHÔNG hiện thực interface này; nó chỉ định tuyến tới đúng impl theo channel.
type ChannelDispatcher interface {
    // Dispatch gửi một notification đã render. Trả errClass để pipeline xử lý.
    Dispatch(ctx context.Context, n notif.Notification) (errClass, error)
    Channel() string // "push" | "email" | "sms" (enum FR-NOTIF-001; iOS push phân theo platform, không phải channel riêng)
}

// Router map channel -> dispatcher. Đăng ký lúc khởi động service.
type Router struct{ byChannel map[string]ChannelDispatcher }

func (r *Router) Route(channel string) (ChannelDispatcher, bool) {
    d, ok := r.byChannel[channel]
    return d, ok // không có handler cho channel -> Permanent (cấu hình sai)
}
```

### Fan-out worker: claim/lease + dispatch + retry/DLQ (Go)

```go
// services/notif/internal/fanout/worker.go

// handle xử lý đúng một notification_id: claim (CAS) -> dispatch -> ack/retry/DLQ.
func (w *Worker) handle(ctx context.Context, id int64) error {
    // 1) Claim/lease: chỉ worker thắng CAS (RETURNING có dòng) mới được gửi (§1 #3).
    n, claimed, err := w.repo.ClaimPending(ctx, id, w.lease)
    if err != nil {
        return err
    }
    if !claimed {
        metrics.DoubleClaim() // có worker khác đã giành -> bỏ qua, idempotent (§1 #2)
        return nil
    }

    // 2) Định tuyến theo channel đã chốt bởi routing FR-NOTIF-001.
    d, ok := w.router.Route(n.Channel)
    if !ok {
        return w.dead(ctx, n, "permanent", "không có dispatcher cho kênh "+n.Channel)
    }

    // 3) Gửi (gọi thật nằm trong dispatcher FR-NOTIF-002/005/006/007).
    class, derr := d.Dispatch(ctx, n)
    switch class {
    case ClassOK:
        return w.repo.MarkSent(ctx, n.ID) // status='sent', sent_at=now()
    case ClassPermanent:
        return w.dead(ctx, n, "permanent", errStr(derr)) // DLQ ngay (§1 #7)
    default: // ClassTransient
        if n.Attempts >= w.maxAttempts {
            return w.dead(ctx, n, "max_attempts", errStr(derr)) // cạn retry -> DLQ (§1 #6)
        }
        // Giữ 'queued', đặt lại lease theo backoff để re-claim sau (§1 #5, #9).
        return w.repo.Requeue(ctx, n.ID, backoff(n.Attempts, w.base, w.cap), errStr(derr))
    }
}

func (w *Worker) dead(ctx context.Context, n notif.Notification, reason, msg string) error {
    metrics.DLQ(n.Channel, reason)
    return w.dlq.Publish(ctx, n, reason, msg) // ghi notification_dlq + status='failed'
}
```

### Claim/lease trong repo (Go)

```go
// services/notif/internal/notif/repo.go (bổ sung)

// ClaimPending giành quyền xử lý một dòng pending HOẶC re-claim một dòng queued
// có lease đã hết hạn. CAS nguyên tử: chỉ một worker nhận được dòng RETURNING.
func (r *Repo) ClaimPending(ctx context.Context, id int64, lease time.Duration) (Notification, bool, error) {
    var n Notification
    err := r.pool.QueryRow(ctx, `
        UPDATE notification
           SET status='queued',
               attempts=attempts+1,
               lease_until=now() + $2::interval
         WHERE id=$1
           AND (status='pending'
                OR (status='queued' AND lease_until < now()))   -- re-claim lease hết hạn (§1 #9)
        RETURNING id, user_id, channel, template, payload, attempts`,
        id, lease.String()).Scan(&n.ID, &n.UserID, &n.Channel, &n.Template, &n.Payload, &n.Attempts)
    if errors.Is(err, pgx.ErrNoRows) {
        return Notification{}, false, nil // worker khác đã giành -> không claim được
    }
    return n, err == nil, err
}
```

### Exponential backoff + full jitter (Go)

```go
// services/notif/internal/fanout/backoff.go

// backoff tính khoảng chờ cho lần thử thứ n: base*2^(n-1), trần cap, cộng full jitter.
// Full jitter (AWS): chờ = random[0, min(cap, base*2^(n-1))] -> rải đều, chống đồng pha (§1 #5).
func backoff(attempt int, base, cap time.Duration) time.Duration {
    exp := base << uint(max(0, attempt-1)) // base * 2^(attempt-1)
    if exp > cap || exp <= 0 {
        exp = cap
    }
    return time.Duration(rand.Int63n(int64(exp) + 1)) // full jitter trong [0, exp]
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `notification` có thêm `attempts`, `lease_until`, `last_error`; bảng `notification_dlq` tồn tại với FK tới `notification(id)`.
2. Producer `Enqueue(id)` đẩy message chứa `notification_id` vào Kafka/Redis Streams; message KHÔNG nhồi toàn bộ nội dung (chỉ con trỏ).
3. Một dòng `notification` pending được fan-out nhặt, dispatch thành công -> `status='sent'`, `sent_at` set, đúng dispatcher của `channel` được gọi.
4. At-least-once: giao lại cùng `notification_id` 2 lần -> dispatcher chỉ được gọi gửi đúng 1 lần (lần thứ 2 thua CAS, bỏ qua); `notif_fanout_double_claim_total` tăng.
5. Idempotent dưới cạnh tranh: hai worker chạy `handle` đồng thời trên cùng `id` -> đúng một worker `MarkSent`, một worker bỏ qua; bảng chỉ có một lần chuyển sang `sent`.
6. Dispatcher trả `ClassTransient` -> job giữ `queued`, `attempts` tăng, `lease_until` lùi theo backoff; KHÔNG vào DLQ khi chưa cạn retry.
7. Dispatcher trả `ClassTransient` lặp tới khi `attempts > max_attempts` -> ghi `notification_dlq` với `reason='max_attempts'`, `notification.status='failed'`, ngừng retry.
8. Dispatcher trả `ClassPermanent` -> vào `notification_dlq` NGAY với `reason='permanent'`, KHÔNG retry dù `attempts` còn thấp.
9. `channel` không có dispatcher đăng ký -> coi là Permanent -> DLQ với lý do "không có dispatcher".
10. Re-claim: dòng `queued` có `lease_until < now()` (worker giả lập crash) được worker khác `ClaimPending` lại; không kẹt vĩnh viễn ở `queued`.
11. Định tuyến đúng kênh: `notification.channel='push'` -> dispatcher push (FR-NOTIF-002 cho Android/Web, FR-NOTIF-005 cho iOS - phân theo `user_channel_token.platform`); `email` -> email (FR-NOTIF-006); `sms` -> sms (FR-NOTIF-007). `channel` không nhận giá trị `apns` (CHECK FR-NOTIF-001). Không có lời gọi nhà cung cấp nào trong package `fanout`.
12. `backoff(n)` tăng theo `n` (kỳ vọng trung bình tăng), luôn nằm trong `[0, cap]`, và hai lần gọi liên tiếp cho giá trị khác nhau (có jitter), không cố định.

---

## §5 - Kiểm thử (verification)

```go
// services/notif/internal/fanout/worker_test.go

// At-least-once: dispatch đúng 1 lần dù message giao lại 2 lần (CAS bảo vệ).
func TestFanout_AtLeastOnce_NoDoubleSend(t *testing.T) {
    w, repo := newTestWorker(t)
    id := seedPending(t, repo, "push")
    var sent int32
    w.router = routeAll(spyDispatcher(func() (errClass, error) {
        atomic.AddInt32(&sent, 1)
        return ClassOK, nil
    }))

    require.NoError(t, w.handle(ctx, id)) // lần giao 1
    require.NoError(t, w.handle(ctx, id)) // lần giao 2 (giao lại) -> thua CAS, bỏ qua

    require.Equal(t, int32(1), atomic.LoadInt32(&sent)) // gửi đúng 1 lần
    require.Equal(t, "sent", statusOf(t, repo, id))
}

// Idempotent dưới cạnh tranh: 2 worker đồng thời -> đúng 1 lần sent.
func TestFanout_ConcurrentClaim_SingleSend(t *testing.T) {
    w, repo := newTestWorker(t)
    id := seedPending(t, repo, "email")
    var sent int32
    w.router = routeAll(spyDispatcher(func() (errClass, error) {
        atomic.AddInt32(&sent, 1)
        return ClassOK, nil
    }))
    var wg sync.WaitGroup
    for i := 0; i < 8; i++ {
        wg.Add(1)
        go func() { defer wg.Done(); _ = w.handle(ctx, id) }()
    }
    wg.Wait()
    require.Equal(t, int32(1), atomic.LoadInt32(&sent)) // chỉ một worker thắng CAS
}

// Transient -> retry (giữ queued, attempts tăng), chưa vào DLQ.
func TestFanout_TransientRetries(t *testing.T) {
    w, repo := newTestWorker(t)
    id := seedPending(t, repo, "push")
    w.router = routeAll(constDispatcher(ClassTransient, errors.New("429")))
    require.NoError(t, w.handle(ctx, id))
    require.Equal(t, "queued", statusOf(t, repo, id))
    require.GreaterOrEqual(t, attemptsOf(t, repo, id), 1)
    require.Equal(t, 0, dlqCount(t, repo, id)) // chưa cạn retry
}

// Cạn retry -> DLQ với reason=max_attempts, status=failed.
func TestFanout_MaxAttempts_ToDLQ(t *testing.T) {
    w, repo := newTestWorker(t)
    w.maxAttempts = 3
    id := seedPending(t, repo, "sms")
    w.router = routeAll(constDispatcher(ClassTransient, errors.New("timeout")))
    for i := 0; i < 5; i++ {
        expireLease(t, repo, id) // giả lập tới hạn re-claim giữa các lần
        _ = w.handle(ctx, id)
    }
    require.Equal(t, "failed", statusOf(t, repo, id))
    require.Equal(t, "max_attempts", dlqReason(t, repo, id))
}

// Permanent -> DLQ ngay, không retry.
func TestFanout_PermanentStraightToDLQ(t *testing.T) {
    w, repo := newTestWorker(t)
    id := seedPending(t, repo, "push")
    w.router = routeAll(constDispatcher(ClassPermanent, errors.New("token gỡ app")))
    require.NoError(t, w.handle(ctx, id))
    require.Equal(t, "failed", statusOf(t, repo, id))
    require.Equal(t, "permanent", dlqReason(t, repo, id))
    require.Equal(t, 1, attemptsOf(t, repo, id)) // không retry thêm
}

// Định tuyến đúng kênh tới đúng dispatcher (không gọi nhà cung cấp thật).
func TestFanout_RoutesByChannel(t *testing.T) {
    w, repo := newTestWorker(t)
    hit := map[string]bool{}
    w.router = &Router{byChannel: map[string]ChannelDispatcher{
        "push":  recordDispatcher("push", hit),
        "email": recordDispatcher("email", hit),
        "sms":   recordDispatcher("sms", hit),
    }}
    for _, ch := range []string{"push", "email", "sms"} {
        _ = w.handle(ctx, seedPending(t, repo, ch))
    }
    require.True(t, hit["push"] && hit["email"] && hit["sms"])
}

// Re-claim: dòng queued lease hết hạn được claim lại, không kẹt vĩnh viễn.
func TestFanout_ReclaimExpiredLease(t *testing.T) {
    _, repo := newTestWorker(t)
    id := seedPending(t, repo, "push")
    _, ok1, _ := repo.ClaimPending(ctx, id, time.Minute)
    require.True(t, ok1)                 // worker A giành
    expireLease(t, repo, id)             // A "crash", lease quá hạn
    _, ok2, _ := repo.ClaimPending(ctx, id, time.Minute)
    require.True(t, ok2)                 // worker B re-claim
}

// services/notif/internal/fanout/backoff_test.go

func TestBackoff_GrowsAndCaps(t *testing.T) {
    base, cap := 200*time.Millisecond, 30*time.Second
    for n := 1; n <= 10; n++ {
        d := backoff(n, base, cap)
        require.GreaterOrEqual(t, d, time.Duration(0))
        require.LessOrEqual(t, d, cap) // luôn trong [0, cap]
    }
}

func TestBackoff_HasJitter(t *testing.T) {
    base, cap := 1*time.Second, 60*time.Second
    a := backoff(5, base, cap)
    b := backoff(5, base, cap)
    require.NotEqual(t, a, b) // jitter -> hai lần khác nhau, không đồng pha (§1 #5)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0003_notification_lease.sql` (cột claim/lease) + `0004_notification_dlq.sql` (bảng DLQ) -> `repo.go` bổ sung `ClaimPending`/`MarkSent`/`Requeue` -> `backoff.go` (thuần, test trước) -> `dispatch.go` (interface + Router) -> `dlq.go` (Publish) -> `worker.go` (vòng claim -> dispatch -> ack/retry/DLQ) -> `producer.go` (Enqueue) -> tests. Dispatcher thật (push/APNs/email/sms) đến ở FR-NOTIF-002/005/006/007; ở FR này dùng dispatcher giả (stub) để kiểm thử pipeline độc lập. Chọn Kafka khi triển khai đa-node có sẵn Kafka; Redis Streams cho khởi đầu nhẹ - cả hai cùng cung cấp claim/ack + consumer group, interface producer ẩn lựa chọn này.

---

## §7 - Phụ thuộc

- **FR-NOTIF-001 (upstream)** - định nghĩa bảng `notification`, routing chốt `channel`, và `idx_notif_dispatch`; fan-out tiêu thụ các dòng `pending` này.
- **FR-NOTIF-002 (downstream)** - dispatcher push (FCM) hiện thực `ChannelDispatcher` cho `channel='push'`.
- **FR-NOTIF-005 (downstream)** - dispatcher APNs cho push iOS: nhặt dòng `channel='push'` lọc thêm `user_channel_token.platform='ios'` (KHÔNG phải một giá trị `channel='apns'` riêng).
- **FR-NOTIF-006 (downstream)** - dispatcher email (SES/SendGrid) cho `channel='email'`.
- **FR-NOTIF-007 (downstream)** - dispatcher SMS cho `channel='sms'`.
- **FR-NOTIF-004 (upstream điều tiết)** - scheduler chống dồn 00:00 đẩy việc vào pipeline này (rải `scheduled_at`); fan-out là tầng tiêu thụ.
- Hạ tầng: Kafka hoặc Redis Streams (hàng đợi bền, consumer group, claim/ack); Postgres (`notification`, `notification_dlq`); OTel.

---

## §8 - Payload ví dụ

### Message trên hàng đợi (con trỏ, không nhồi nội dung)

```json
{ "notification_id": 55012, "user_id": 4021, "channel": "push",
  "enqueued_at": "2026-06-28T17:00:03Z", "trace_id": "0af7651916cd43dd8448eb211c80319c" }
```

### Một mục dead-letter sau khi cạn retry

```sql
SELECT notification_id, channel, attempts, reason, last_error
FROM notification_dlq WHERE notification_id = 55012;
-- 55012 | push | 6 | max_attempts | FCM 503 Service Unavailable (sau 5 lần thử + backoff)
```

```json
{ "id": 88001, "notification_id": 55012, "channel": "push",
  "payload": { "template": "price_below", "price": 79000 },
  "attempts": 6, "reason": "max_attempts",
  "last_error": "FCM 503 Service Unavailable", "dead_at": "2026-06-28T17:04:11Z" }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Cơ chế replay tự động từ DLQ (sau khi sự cố nhà cung cấp hồi) - hiện soi thủ công; thêm worker replay có nút bấm ở giai đoạn vận hành.
- Circuit breaker per-channel (khi một nhà cung cấp sập diện rộng, tạm ngắt để khỏi đốt retry) - phối hợp với dispatcher FR-NOTIF-002/006 sau.
- Ưu tiên hàng đợi (priority lane) cho OTP/high-value tách khỏi thông báo giá thường - gắn khi BILL/Premium (FR-BILL-001) lên.
- Exactly-once xuyên nhà cung cấp qua idempotency-key của FCM/SES - tăng cường idempotent phía ngoài ở giai đoạn sau; hiện claim/lease phía DB là đủ.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Worker crash sau claim, trước ack | `lease_until` quá hạn | dòng kẹt `queued` | Re-claim sau lease hết hạn (§1 #9) |
| Message giao lại (re-balance/at-least-once) | CAS thua RETURNING | nguy cơ gửi đôi | Claim/lease idempotent, bỏ qua lần thua (§1 #3) |
| Hai worker cùng nhặt một id | `double_claim_total` tăng | gửi đôi | Đúng một thắng CAS; còn lại no-op |
| Nhà cung cấp 429/5xx lúc đỉnh 00:00 | dispatcher trả Transient | trễ gửi | Backoff + jitter, retry tới max (§1 #5) |
| Retry đồng pha (thundering herd) | spike retry metric | tự dồn tải | Full jitter rải đều khoảng chờ (§1 #5) |
| Token push gỡ app / payload sai | dispatcher trả Permanent | gửi hỏng | DLQ ngay, không retry (§1 #7) |
| Job hỏng retry vô hạn | `attempts > max` | chiếm slot worker | Cạn retry -> DLQ, ngừng (§1 #6) |
| `channel` không có dispatcher | Route trả false | không gửi được | Coi Permanent -> DLQ (cấu hình sai) (§1 #9) |
| Fan-out tự gọi nhà cung cấp | code review | vượt rate-limit, lệch ranh giới | Chỉ định tuyến; gọi thật ở FR-NOTIF-002/005/006/007 (§1 #4) |
| Thông báo cùng user đảo thứ tự | quan sát trải nghiệm | khó hiểu | Phân vùng theo `user_id` (§1 #11) |
| DLQ phình (một nhà cung cấp sập diện rộng) | kích thước `notification_dlq` | nhiều thông báo trượt | Alert vận hành, replay sau khi hồi (§9) |

---

## §11 - Ghi chú

- Fan-out là xương sống vận chuyển của module thông báo: nó nối engine alert (qua dòng `notification` pending của FR-NOTIF-001) tới người dùng cuối, nhưng KHÔNG tự gọi nhà cung cấp.
- Claim/lease bằng `UPDATE ... WHERE status='pending' RETURNING` là một CAS nguyên tử trong Postgres - đó là cốt lõi biến at-least-once thành "gửi đúng một lần" mà không cần exactly-once đắt đỏ.
- Backoff full jitter là tuyến phòng thủ chống thundering herd lúc đỉnh 00:00, song song với pacing/jitter của scraping farm (FR-SCRAPE-001/005) ở phía thu thập.
- DLQ tách dòng chết khỏi vòng retry, vừa giải phóng slot worker vừa biến lỗi thành tín hiệu cụ thể để con người soi - không có vòng lặp câm.
- Ranh giới cứng: fan-out giữ phần chung (claim, retry, backoff, DLQ, định tuyến); mỗi `ChannelDispatcher` (FR-NOTIF-002/005/006/007) là một implementation cắm vào theo `channel`, thêm kênh không đụng xương sống.
- FR-NOTIF-004 (scheduler chống dồn 00:00) là tầng điều tiết đầu vào; fan-out là tầng tiêu thụ - hai mảnh khớp nhau ở hàng đợi.

---

*Hết FR-NOTIF-003. Status: ready_to_implement (mục tiêu audit 10/10).*
