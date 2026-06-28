---
id: FR-NOTIF-004
title: "Midnight-spike scheduler - flatten the curve cho đỉnh 00:00: jitter [-90s,+180s] rải đều sự kiện, gom bucket theo phút, spread overflow sang phút kế, tránh dồn vào mốc tròn :00/:15/:30/:45, không bao giờ vượt giới hạn 600k/phút của FCM để khỏi ăn 429"
module: NOTIF
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-NOTIF-001, FR-NOTIF-002, FR-NOTIF-003, FR-TRACK-004, FR-DEAL-006]
depends_on: [FR-NOTIF-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.5(5) (pseudo-code scheduleMidnightAlerts: jitter + bucket theo phút + spread overflow)"
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.6 (giới hạn FCM 600.000 tin/phút/project; đỉnh 00:00 traffic tăng >2x trong 30s-2 phút đầu mỗi giờ; FCM khuyến nghị flatten the curve, tránh dồn vào mốc tròn :00/:15/:30/:45)"
source_decisions:
  - "DEC-NOTIF-31: scheduler đặt notification.scheduled_at = event_time + jitter, jitter ngẫu nhiên trong [-90s, +180s] (bất đối xứng: 90s trước tới 180s sau thời điểm sự kiện) để rải đều quanh sự kiện"
  - "DEC-NOTIF-32: gom alert vào bucket theo phút bằng floor(dispatch_at / 60s); mỗi bucket là một phút lịch"
  - "DEC-NOTIF-33: FCM_RATE_LIMIT_PER_MIN = 600.000 (§3.6); bucket nào > giới hạn thì spread phần dư sang các phút kế tiếp (spreadAcrossNextMinutes) cho tới khi mọi bucket <= giới hạn"
  - "DEC-NOTIF-34: tránh dồn vào mốc tròn :00/:15/:30/:45 - jitter bất đối xứng + spread phải kéo khối ra khỏi đúng các giây :00 của mốc tròn, không để cụm dồn lại đó"
  - "DEC-NOTIF-35: rng tiêm được (inject) để jitter tất định khi test (seeded); scheduler chỉ đặt scheduled_at rồi giao fan-out (FR-NOTIF-003) phát đúng giờ, KHÔNG tự gọi nhà cung cấp"

language: "Go 1.22 (notif-svc)"
service: shopass/services/notif/
new_files:
  - services/notif/internal/notif/scheduler.go
  - services/notif/internal/notif/scheduler_test.go
modified_files:
  - services/notif/internal/notif/repo.go        # thêm BatchSetScheduledAt(ids, times) ghi scheduled_at hàng loạt
allowed_tools:
  - file_read: services/notif/**
  - file_write: services/notif/**
  - bash: cd services/notif && go test ./...
disallowed_tools:
  - dùng time.Now()/rand toàn cục trực tiếp trong ScheduleAlerts (vi phạm DEC-NOTIF-35, test không tất định)
  - enqueue bucket vượt 600.000/phút mà không spread (vi phạm DEC-NOTIF-33, ăn 429 lúc đỉnh)
  - tự gọi FCM/APNs trong scheduler (vi phạm DEC-NOTIF-35; gửi thuộc fan-out FR-NOTIF-003)

effort_hours: 6
sub_tasks:
  - "0.5h: scheduler.go - hằng FCM_RATE_LIMIT_PER_MIN, JitterMin=-90s, JitterMax=+180s, interface RNG tiêm được"
  - "1.5h: scheduler.go - applyJitter (dispatch_at = event_time + jitter trong [-90s,+180s]) + bucketByMinute (floor(dispatch_at/60s))"
  - "1.5h: scheduler.go - spreadAcrossNextMinutes: bucket > giới hạn thì đẩy phần dư sang phút kế cho tới khi mọi bucket <= giới hạn; né mốc tròn :00/:15/:30/:45"
  - "1.0h: scheduler.go - ScheduleAlerts ghép jitter -> bucket -> spread; trả map id->scheduled_at; gọi repo.BatchSetScheduledAt; KHÔNG gọi nhà cung cấp"
  - "1.5h: scheduler_test.go - 6 test: jitter bounds [-90s,+180s], bucket không vượt giới hạn, overflow spread sang phút kế, né mốc tròn, batch lớn được flatten, seeded rng tất định"

risk_if_skipped: "Đỉnh 00:00 là khoảnh khắc nguy hiểm nhất của module thông báo. §3.6 ghi rõ traffic tăng hơn 2x trong 30 giây tới 2 phút đầu mỗi giờ, và đặc biệt dồn cục lúc nửa đêm khi hàng loạt flash sale Shopee/TikTok/Lazada cùng mở và hàng loạt rule giá cùng kích. Nếu không có scheduler này, mọi alert tới hạn cùng một thời điểm sẽ đập thẳng vào FCM trong cùng một phút; một bucket phút duy nhất dễ dàng vượt trần 600.000 tin/phút/project, FCM trả về 429 Too Many Requests, và toàn bộ phần dư bị từ chối hoặc phải retry dồn vào phút sau - tạo hiệu ứng dây chuyền đẩy phút kế tiếp cũng vỡ trần, cứ thế cuộn lại thành một cơn bão lỗi kéo dài cả chục phút quanh nửa đêm. Hệ quả là đúng lúc deal nóng nhất thì người dùng không nhận được cảnh báo, hoặc nhận trễ tới mức deal đã hết - phá thẳng giá trị cốt lõi của SănDeal là báo kịp lúc. Tệ hơn, gửi dồn cụm vào đúng các giây :00 của mốc tròn (điều FCM khuyến nghị tránh) khiến chính hạ tầng FCM coi traffic của ta là spike bất thường và siết thêm. Không flatten the curve thì cũng đốt vô ích phần quota retry và làm nhiễu metric. Scheduler này là cái van duy nhất giữ cho dòng thông báo nửa đêm chảy đều dưới trần, nên thiếu nó là chấp nhận hỏng đúng vào giờ vàng."
---

## §1 - Mô tả (BCP-14 normative)

Service NOTIF **MUST** cung cấp một midnight-spike scheduler san phẳng đỉnh 00:00 (flatten the curve) trước khi giao cho fan-out. Scheduler nhận một lô alert tới hạn, đặt `notification.scheduled_at` cho từng dòng bằng jitter trong `[-90s, +180s]`, gom theo bucket phút, và spread phần dư khi một bucket vượt trần FCM, sao cho KHÔNG bucket phút nào vượt 600.000 tin/phút/project. Scheduler chỉ đặt thời điểm; việc phát đúng giờ thuộc fan-out (FR-NOTIF-003). Hợp đồng:

1. **MUST** với mỗi alert, tính `dispatch_at = event_time + jitter`, trong đó `jitter` là số ngẫu nhiên trong khoảng `[-90s, +180s]` (bất đối xứng: từ 90 giây trước tới 180 giây sau `event_time`) theo §3.5(5) và DEC-NOTIF-31. `jitter` **MUST** nằm trong cận này, không bao giờ ngoài.
2. **MUST** gom các alert vào bucket theo phút bằng khóa `floor(dispatch_at / 60s)` (DEC-NOTIF-32). Mỗi bucket đại diện đúng một phút lịch.
3. **MUST** giữ hằng `FCM_RATE_LIMIT_PER_MIN = 600_000` (§3.6, DEC-NOTIF-33). Đây là trần cứng cho số tin trong một phút của một project FCM.
4. **MUST** khi `size(bucket) > FCM_RATE_LIMIT_PER_MIN` thì gọi `spreadAcrossNextMinutes(bucket)`: giữ lại tối đa `FCM_RATE_LIMIT_PER_MIN` tin trong phút hiện tại, đẩy phần dư sang phút kế tiếp, lặp cho tới khi mọi bucket `<= FCM_RATE_LIMIT_PER_MIN` (DEC-NOTIF-33).
5. **MUST** đảm bảo bất biến sau khi spread: với mọi bucket phút, `size(bucket) <= FCM_RATE_LIMIT_PER_MIN`. Đây là bất biến cốt lõi - vi phạm nó là ăn 429.
6. **MUST** tránh dồn cụm vào các mốc tròn `:00`, `:15`, `:30`, `:45` (DEC-NOTIF-34): jitter bất đối xứng làm khối alert lệch khỏi đúng giây sự kiện, và bước spread **MUST KHÔNG** đẩy phần dư về đúng giây `:00` của một phút mốc tròn theo cách tạo cụm mới ở đó. Phân bố sau lập lịch không được có đỉnh nhọn tại các mốc tròn.
7. **MUST** cung cấp `ScheduleAlerts(now time.Time, alerts []Alert, rng RNG) map[int64]time.Time`: nhận lô alert, trả ánh xạ `notification_id -> scheduled_at` đã jitter + bucket + spread. Hàm **MUST** thuần (deterministic) khi `rng` cố định.
8. **MUST** nhận `rng` qua tham số (dependency injection) để jitter tất định khi test (DEC-NOTIF-35); code production tiêm một `RNG` thật, test tiêm `RNG` seeded để tái lập đúng phân bố. `ScheduleAlerts` **MUST KHÔNG** dùng `rand` toàn cục hay `time.Now()` ẩn bên trong. Sau khi tính, **MUST** ghi `scheduled_at` cho từng dòng `notification` qua `repo.BatchSetScheduledAt(ctx, map[int64]time.Time)` bằng một câu UPDATE hàng loạt (không N+1).
9. **MUST** giữ ranh giới với tầng gửi (DEC-NOTIF-35): scheduler **MUST KHÔNG** gọi FCM/APNs/SES/SMS. Nó chỉ đặt `scheduled_at`; fan-out (FR-NOTIF-003) đọc các dòng tới hạn theo `scheduled_at` rồi mới phát.
10. **MUST** giữ ổn định thứ tự khi spread: trong một bucket vượt trần, phần được giữ lại và phần bị đẩy sang phút sau **MUST** theo thứ tự tất định (vd theo `notification_id`) để kết quả tái lập được, không phụ thuộc thứ tự map.
11. **SHOULD** ưu tiên giữ lại alert high-value (OTP, deal giá trị cao) trong phút sớm khi phải spread, để cảnh báo quan trọng không bị đẩy lùi; alert thường nhường chỗ trước.
12. **SHOULD** phát OTel: `notif_scheduler_batch_size` (histogram - số alert mỗi lô), `notif_scheduler_buckets_total` (counter), `notif_scheduler_overflow_spread_total` (counter - số tin phải spread vì vượt trần), `notif_scheduler_max_bucket_size` (gauge - bucket lớn nhất sau spread, phải `<= 600_000`).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao phải flatten the curve cho đỉnh 00:00?** §3.6 đo được traffic thông báo tăng hơn 2x trong 30 giây tới 2 phút đầu mỗi giờ, và nửa đêm là đỉnh của đỉnh: flash sale các sàn cùng mở lúc 00:00, hàng loạt rule giá cùng tới ngưỡng, batch dự đoán đáy giá (FR-DEAL-006) cùng bắn. Nếu để tất cả đập vào FCM trong một phút, một bucket phút dễ vượt trần 600.000 và FCM trả 429. Một khi 429 xảy ra, phần dư retry dồn sang phút sau làm phút đó cũng vỡ - hiệu ứng cuộn tuyết. San phẳng đường cong là cách FCM khuyến nghị để chính ta không tự bắn vào chân mình lúc cao điểm.

**Vì sao jitter bất đối xứng `[-90s, +180s]`?** Ý tưởng là rải đều các alert quanh thời điểm sự kiện thay vì để chúng chụm đúng một giây. Bất đối xứng (90 giây trước, 180 giây sau) cố ý nghiêng về phía sau: với phần lớn cảnh báo, gửi sớm tối đa 90 giây vẫn kịp giá trị, còn cho phép trễ tới 180 giây mở rộng cửa sổ để hạ mật độ tin/phút mà không làm người dùng cảm thấy chậm. Cửa sổ rộng hơn về sau cũng là nơi có chỗ trống để hứng phần spread khi phút sự kiện quá tải. Đây là cách rẻ nhất để biến một cột nhọn thành một dải phẳng trước cả khi cần tới bước spread.

**Vì sao gom bucket theo phút (`floor(dispatch_at / 60s)`)?** Trần của FCM tính theo phút (600.000 tin/phút/project), nên đơn vị kiểm soát tự nhiên là phút. Quy mỗi `dispatch_at` về phút chứa nó cho ta đúng đại lượng cần khống chế: số tin trong mỗi phút lịch. Bucket phút biến bài toán "đừng vượt rate limit" thành một phép đếm đơn giản trên từng nhóm, dễ kiểm và dễ chứng minh bất biến.

**Vì sao spread phần dư sang phút kế khi vượt trần?** Jitter một mình không bảo đảm tuyệt đối: nếu lô quá lớn, ngay cả sau khi rải, một bucket vẫn có thể vượt 600.000. spread là chốt chặn cuối: giữ lại đúng trần trong phút hiện tại, đẩy phần thừa sang phút sau, lặp cho tới khi mọi phút đều dưới trần. Nhờ vậy bất biến `size(bucket) <= 600_000` luôn đúng bất kể lô lớn cỡ nào - ta đánh đổi một chút độ trễ phát của phần dư để đổi lấy việc không bao giờ ăn 429.

**Vì sao tránh dồn vào mốc tròn `:00/:15/:30/:45`?** FCM khuyến nghị tránh các mốc tròn vì rất nhiều hệ thống lập lịch "ngây thơ" đều bắn đúng các giây này, làm hạ tầng FCM thấy spike đồng pha và siết lại. Nếu jitter và spread của ta lại vô tình kéo khối tin về đúng giây `:00` của mốc tròn, ta tự đặt mình vào nhóm bị siết. Vì thế jitter cố ý lệch, và spread cố ý không nhắm phần dư về đúng đầu phút mốc tròn - mục tiêu là một phân bố trơn, không có gai tại các mốc đó.

**Vì sao rng tiêm được và scheduler chỉ đặt scheduled_at, không tự gửi?** Lập lịch dựa trên ngẫu nhiên mà không kiểm soát được nguồn ngẫu nhiên thì không test được: ta không thể khẳng định "bucket không vượt trần" nếu mỗi lần chạy ra một phân bố khác. Tiêm `RNG` cho phép test seeded tái lập đúng phân bố và kiểm bất biến tất định. Và vì gửi ở quy mô (rate-limit, dead-letter, retry) là việc của fan-out FR-NOTIF-003, scheduler dừng đúng ở "đặt thời điểm": nó ghi `scheduled_at` rồi trao lại, giữ một ranh giới trách nhiệm sạch và tránh trộn logic lập lịch với gọi nhà cung cấp.

---

## §3 - Hợp đồng API / DDL

Không có DDL mới. Bảng `notification` và cột `scheduled_at` đã do FR-NOTIF-001 định nghĩa; FR này chỉ ghi vào cột đó qua một UPDATE hàng loạt.

### Hằng số và kiểu (Go)

```go
// services/notif/internal/notif/scheduler.go

const (
    // FCM_RATE_LIMIT_PER_MIN: trần cứng 600.000 tin/phút/project (§3.6).
    // Không bucket phút nào được vượt giá trị này, nếu không sẽ ăn 429.
    FCM_RATE_LIMIT_PER_MIN = 600_000

    // Jitter bất đối xứng [-90s, +180s] (§3.5(5), DEC-NOTIF-31).
    JitterMin = -90 * time.Second  // 90 giây TRƯỚC event_time
    JitterMax = 180 * time.Second  // 180 giây SAU event_time
)

// RNG tiêm được để jitter tất định khi test (DEC-NOTIF-35).
// Production tiêm rng thật; test tiêm rng seeded.
type RNG interface {
    // Int63n trả số trong [0, n); chuẩn như math/rand.
    Int63n(n int64) int64
}

type Alert struct {
    NotificationID int64
    EventTime      time.Time
    HighValue      bool // OTP / deal giá trị cao: ưu tiên giữ phút sớm khi spread (§1 #11)
}

// minuteKey quy một thời điểm về phút chứa nó: floor(t / 60s) (DEC-NOTIF-32).
func minuteKey(t time.Time) int64 { return t.Unix() / 60 }

// roundMarkSecond cho biết giây-trong-giờ của t có rơi đúng mốc tròn
// :00/:15/:30/:45 không (DEC-NOTIF-34).
func isRoundMark(t time.Time) bool {
    m := t.Minute()
    return t.Second() == 0 && (m == 0 || m == 15 || m == 30 || m == 45)
}
```

### applyJitter + bucketByMinute (§1 #1, #2)

```go
// applyJitter đặt dispatch_at = event_time + jitter, jitter trong [-90s,+180s].
// Span = JitterMax - JitterMin (= 270s); offset = rng trong [0, span], rồi + JitterMin.
func applyJitter(a Alert, rng RNG) time.Time {
    span := int64((JitterMax - JitterMin) / time.Second) // 270
    off := time.Duration(rng.Int63n(span+1)) * time.Second
    return a.EventTime.Add(JitterMin + off) // [event-90s, event+180s]
}

// bucketByMinute gom alert (đã có dispatch_at) vào bucket theo phút.
func bucketByMinute(items []scheduled) map[int64][]scheduled {
    buckets := make(map[int64][]scheduled)
    for _, it := range items {
        k := minuteKey(it.At)
        buckets[k] = append(buckets[k], it)
    }
    return buckets
}

type scheduled struct {
    ID        int64
    At        time.Time
    HighValue bool
}
```

### spreadAcrossNextMinutes (§1 #4, #5, #10, #11)

```go
// spreadAcrossNextMinutes bảo đảm mọi bucket phút <= FCM_RATE_LIMIT_PER_MIN.
// Bucket vượt trần thì giữ lại đúng trần (ưu tiên high-value + thứ tự id),
// đẩy phần dư sang phút kế; lặp cho tới khi không còn bucket nào vượt.
func spreadAcrossNextMinutes(buckets map[int64][]scheduled) map[int64][]scheduled {
    keys := sortedKeys(buckets) // tăng dần, lặp tất định
    for i := 0; i < len(keys); i++ {
        k := keys[i]
        b := buckets[k]
        if len(b) <= FCM_RATE_LIMIT_PER_MIN {
            continue
        }
        sortStable(b) // high-value trước, rồi theo ID (§1 #10, #11)
        keep, overflow := b[:FCM_RATE_LIMIT_PER_MIN], b[FCM_RATE_LIMIT_PER_MIN:]
        buckets[k] = keep

        next := k + 1
        // Đẩy overflow xuống phút kế; gán At = đầu phút kế + 1s để KHÔNG
        // rơi đúng giây :00 mốc tròn (DEC-NOTIF-34).
        nextStart := time.Unix(next*60, 0).Add(1 * time.Second)
        for j := range overflow {
            overflow[j].At = nextStart
        }
        buckets[next] = append(buckets[next], overflow...)
        if !containsKey(keys, next) {
            keys = append(keys, next) // phút mới sinh ra cũng phải duyệt
        }
    }
    return buckets
}
```

### ScheduleAlerts (§1 #7, #8, #9)

```go
// ScheduleAlerts: jitter -> bucket theo phút -> spread overflow.
// Trả map notification_id -> scheduled_at. Thuần khi rng cố định.
// KHÔNG gọi nhà cung cấp (DEC-NOTIF-35); chỉ đặt thời điểm cho fan-out.
func ScheduleAlerts(now time.Time, alerts []Alert, rng RNG) map[int64]time.Time {
    items := make([]scheduled, 0, len(alerts))
    for _, a := range alerts {
        items = append(items, scheduled{
            ID: a.NotificationID, At: applyJitter(a, rng), HighValue: a.HighValue,
        })
    }
    buckets := spreadAcrossNextMinutes(bucketByMinute(items))

    out := make(map[int64]time.Time, len(alerts))
    for _, b := range buckets {
        for _, it := range b {
            out[it.ID] = it.At
        }
    }
    return out
}

// Caller (orchestrator) ghi xuống DB:
//   sched := ScheduleAlerts(time.Now(), alerts, prodRNG)
//   err := repo.BatchSetScheduledAt(ctx, sched) // 1 UPDATE hàng loạt
// Fan-out (FR-NOTIF-003) sau đó nhặt dòng theo scheduled_at và phát.
```

---

## §4 - Acceptance criteria

1. Với mọi alert, `scheduled_at - event_time` nằm trong `[-90s, +180s]` (không bao giờ ngoài cận).
2. jitter bất đối xứng đúng hướng: cận dưới là `-90s`, cận trên là `+180s` (không phải `+-90s` hay `+-180s`).
3. `minuteKey(dispatch_at)` = `floor(dispatch_at / 60s)`; hai alert cùng phút lịch rơi cùng bucket.
4. Sau `ScheduleAlerts`, KHÔNG bucket phút nào có `size > 600_000` (bất biến §1 #5).
5. Khi một phút có `> 600_000` alert, đúng `600_000` được giữ lại, phần dư xuất hiện ở phút kế (hoặc các phút kế tiếp nếu vẫn tràn).
6. Spread lặp tới hội tụ: lô lớn bằng `N x 600_000 + r` được rải qua đủ số phút để mọi bucket dưới trần.
7. Phân bố sau lập lịch KHÔNG có đỉnh nhọn tại các giây `:00` của mốc tròn `:00/:15/:30/:45`; phần dư bị spread không bị gán đúng giây `:00` mốc tròn.
8. `ScheduleAlerts` tất định: cùng `alerts` + cùng `rng` seeded -> cùng map kết quả qua nhiều lần chạy.
9. Thứ tự keep/overflow tất định (theo high-value rồi ID), không phụ thuộc thứ tự duyệt map.
10. Alert `HighValue=true` được ưu tiên giữ ở phút sớm khi bucket phải spread.
11. `ScheduleAlerts` KHÔNG gọi FCM/APNs/SES/SMS (kiểm qua review + không import client gửi trong scheduler).
12. `repo.BatchSetScheduledAt` (§1 #8) ghi `scheduled_at` cho đúng tập `notification_id` bằng một UPDATE hàng loạt (không N+1); metric `notif_scheduler_max_bucket_size <= 600_000`.

---

## §5 - Kiểm thử (verification)

```go
// services/notif/internal/notif/scheduler_test.go

// seededRNG: RNG tất định cho test (DEC-NOTIF-35).
type seededRNG struct{ r *rand.Rand }
func newSeededRNG(seed int64) RNG { return &seededRNG{rand.New(rand.NewSource(seed))} }
func (s *seededRNG) Int63n(n int64) int64 { return s.r.Int63n(n) }

func TestJitter_WithinAsymmetricBounds(t *testing.T) {
    rng := newSeededRNG(1)
    base := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC) // 00:00
    for i := 0; i < 10_000; i++ {
        at := applyJitter(Alert{NotificationID: int64(i), EventTime: base}, rng)
        d := at.Sub(base)
        require.GreaterOrEqual(t, d, JitterMin) // >= -90s
        require.LessOrEqual(t, d, JitterMax)    // <= +180s
    }
}

func TestBucket_NeverExceedsLimit(t *testing.T) {
    rng := newSeededRNG(7)
    base := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
    alerts := makeAlerts(t, base, FCM_RATE_LIMIT_PER_MIN*3+123) // 3 phút + dư
    sched := ScheduleAlerts(base, alerts, rng)

    counts := map[int64]int{}
    for _, at := range sched {
        counts[minuteKey(at)]++
    }
    for k, c := range counts {
        require.LessOrEqualf(t, c, FCM_RATE_LIMIT_PER_MIN,
            "bucket phút %d có %d tin > trần", k, c) // bất biến §1 #5
    }
}

func TestOverflow_SpreadsToNextMinute(t *testing.T) {
    rng := newSeededRNG(3)
    base := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
    // Ép mọi alert vào cùng một phút bằng event_time trùng + rng cố định,
    // rồi kiểm phần dư chảy sang phút kế.
    alerts := makeAlerts(t, base, FCM_RATE_LIMIT_PER_MIN+5_000)
    sched := ScheduleAlerts(base, alerts, rng)

    counts := map[int64]int{}
    for _, at := range sched {
        counts[minuteKey(at)]++
    }
    require.GreaterOrEqual(t, len(counts), 2)          // tràn sang ít nhất 1 phút nữa
    for _, c := range counts {
        require.LessOrEqual(t, c, FCM_RATE_LIMIT_PER_MIN)
    }
}

func TestSpread_AvoidsRoundMarks(t *testing.T) {
    rng := newSeededRNG(9)
    base := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
    alerts := makeAlerts(t, base, FCM_RATE_LIMIT_PER_MIN*2+10)
    sched := ScheduleAlerts(base, alerts, rng)
    for _, at := range sched {
        // Không tin nào bị spread rơi đúng giây :00 của mốc tròn (DEC-NOTIF-34).
        require.Falsef(t, isRoundMark(at),
            "tin bị gán đúng mốc tròn %s", at.Format("15:04:05"))
    }
}

func TestLargeBatch_Flattened(t *testing.T) {
    rng := newSeededRNG(42)
    base := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
    alerts := makeAlerts(t, base, FCM_RATE_LIMIT_PER_MIN*5) // 5x trần trong 1 sự kiện
    sched := ScheduleAlerts(base, alerts, rng)
    require.Len(t, sched, FCM_RATE_LIMIT_PER_MIN*5) // không mất alert nào
    counts := map[int64]int{}
    for _, at := range sched {
        counts[minuteKey(at)]++
    }
    require.GreaterOrEqual(t, len(counts), 5) // rải qua >= 5 phút
    for _, c := range counts {
        require.LessOrEqual(t, c, FCM_RATE_LIMIT_PER_MIN)
    }
}

func TestScheduleAlerts_Deterministic(t *testing.T) {
    base := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
    alerts := makeAlerts(t, base, 50_000)
    a := ScheduleAlerts(base, alerts, newSeededRNG(100))
    b := ScheduleAlerts(base, alerts, newSeededRNG(100))
    require.Equal(t, a, b) // cùng seed -> cùng kết quả (DEC-NOTIF-35)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `scheduler.go` (hằng `FCM_RATE_LIMIT_PER_MIN`, `JitterMin`, `JitterMax`, interface `RNG`, các helper `minuteKey`/`isRoundMark`) -> `applyJitter` + `bucketByMinute` -> `spreadAcrossNextMinutes` -> `ScheduleAlerts` -> `repo.BatchSetScheduledAt` -> tests. Orchestrator (engine alert FR-TRACK-004 / batch đáy giá FR-DEAL-006) sau khi đã có các dòng `notification` pending sẽ gọi `ScheduleAlerts(time.Now(), alerts, prodRNG)` rồi `BatchSetScheduledAt`; fan-out (FR-NOTIF-003) nhặt theo `scheduled_at`. Không có lời gọi nhà cung cấp nào trong package scheduler. Production tiêm một `RNG` mỏng bọc `math/rand` (hoặc `crypto/rand` rút gọn); test tiêm `seededRNG`.

---

## §7 - Phụ thuộc

- **FR-NOTIF-001 (upstream)** - định nghĩa bảng `notification` và cột `scheduled_at` mà scheduler ghi vào; cũng là nơi routing đã quyết kênh + tạo dòng pending trước khi lập lịch.
- **FR-NOTIF-003 (downstream, depends_on)** - fan-out đọc các dòng `notification` theo `scheduled_at` (qua `idx_notif_dispatch`) và phát đúng giờ; scheduler đặt thời điểm, fan-out thực thi rate-limit khi gửi. Scheduler là van san phẳng đặt TRƯỚC fan-out.
- **FR-TRACK-004 / FR-DEAL-006 (producer)** - sinh lô alert tới hạn (đặc biệt cụm 00:00); gọi `ScheduleAlerts` rồi `BatchSetScheduledAt`.
- Lib: `math/rand` (bọc trong `RNG` tiêm được), `time`, `pgx` (cho `BatchSetScheduledAt`). Không import client FCM/APNs trong package này.

---

## §8 - Payload ví dụ

### Trước lập lịch: 1.500.000 alert cùng tới hạn lúc 00:00:00

```text
event_time = 2026-06-28T00:00:00Z cho cả 1.500.000 alert
(flash sale Shopee + rule giá + batch đáy giá cùng bắn)
Nếu enqueue thô: bucket phút 00:00 = 1.500.000 tin > 600.000 -> 429.
```

### Sau ScheduleAlerts: phân bố dispatch_at theo phút

```text
phút 00:00  -> 600.000 tin  (jitter rải trong [-90s,+180s], giữ tối đa trần)
phút 00:01  -> 600.000 tin  (phần dư spread xuống, gán 00:01:01 né mốc tròn)
phút 00:02  -> 300.000 tin  (phần dư còn lại)
-----------------------------------------------
tổng        = 1.500.000 tin, KHÔNG phút nào > 600.000  (bất biến §1 #5)
không tin nào rơi đúng giây :00 của mốc tròn :00/:15/:30/:45 (DEC-NOTIF-34)
```

### Ghi xuống DB (một UPDATE hàng loạt)

```sql
-- repo.BatchSetScheduledAt dựng UPDATE ... FROM (VALUES ...) theo id
UPDATE notification AS n
SET    scheduled_at = v.sched
FROM   (VALUES (55012, TIMESTAMPTZ '2026-06-28 00:00:37Z'),
               (55013, TIMESTAMPTZ '2026-06-28 00:01:01Z'))
       AS v(id, sched)
WHERE  n.id = v.id;
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Lập lịch theo từng project FCM nếu sau này tách nhiều project (trần 600k là per-project; hiện một project nên một van là đủ).
- Tham số hóa cận jitter theo loại alert (deal cực nóng cho jitter hẹp hơn để gửi sớm) - tinh chỉnh sau khi có dữ liệu thực về đỉnh 00:00.
- Backpressure liên giờ: khi cụm 00:00 spread lấn sang 00:03+, có nên hợp nhất với cụm 01:00 của giờ kế - giai đoạn sau, hiện cửa sổ spread đủ ngắn.
- Quiet-hours phối hợp với preference center (không push ban đêm trừ khi user đồng ý) - gắn với FR-NOTIF-001 mục đã hoãn.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Một bucket phút vượt 600.000 mà không spread | test bất biến + metric max_bucket_size | FCM trả 429, phần dư bị từ chối | spreadAcrossNextMinutes lặp tới khi mọi bucket <= trần (§1 #4,#5) |
| jitter ra ngoài [-90s,+180s] | TestJitter_WithinAsymmetricBounds | rải sai, cụm lệch | applyJitter dùng span=270s + offset trong [0,span], + JitterMin |
| jitter đối xứng nhầm (+-90s) | review + test cận trên | cửa sổ sau hẹp, dễ tràn | Cận đúng: dưới -90s, trên +180s (DEC-NOTIF-31) |
| Phần dư spread bị gán đúng giây :00 mốc tròn | TestSpread_AvoidsRoundMarks | FCM siết vì spike đồng pha | Gán nextStart = đầu phút + 1s, né mốc tròn (DEC-NOTIF-34) |
| Spread không hội tụ (lô khổng lồ) | TestLargeBatch_Flattened | treo/tràn vô hạn | Vòng lặp duyệt cả phút mới sinh ra, mỗi vòng đẩy đúng phần dư |
| Kết quả không tất định | TestScheduleAlerts_Deterministic | test rung, không kiểm được | RNG tiêm được + sortStable theo id (DEC-NOTIF-35, §1 #10) |
| Mất alert khi spread (rơi rớt) | TestLargeBatch_Flattened đếm len | cảnh báo biến mất | keep + overflow hợp lại đủ, không drop |
| Scheduler tự gọi FCM | code review (không import client) | vượt rate-limit, lấn việc fan-out | Chỉ đặt scheduled_at (DEC-NOTIF-35); gửi là FR-NOTIF-003 |
| Ghi scheduled_at N+1 | review SQL + metric | DB nghẽn lúc đỉnh | BatchSetScheduledAt một UPDATE hàng loạt (§1 #8) |
| Alert high-value bị đẩy lùi khi spread | test ưu tiên | OTP/deal nóng tới trễ | sortStable ưu tiên HighValue giữ phút sớm (§1 #11) |

---

## §11 - Ghi chú

- Scheduler là cái van san phẳng đặt TRƯỚC fan-out: nó chỉ đặt `scheduled_at`, còn rate-limit lúc gửi thật là việc của FR-NOTIF-003 - hai lớp phòng thủ nối tiếp, không trùng vai.
- Bất biến sống còn: sau `ScheduleAlerts`, mọi bucket phút `<= 600.000`. Mọi thứ khác (jitter, né mốc tròn, ưu tiên high-value) là để đường cong đẹp; bất biến này là để không ăn 429.
- jitter bất đối xứng `[-90s, +180s]` nghiêng về sau có chủ đích: cửa sổ sau rộng hơn vừa hạ mật độ tin/phút vừa chừa chỗ hứng phần spread, mà vẫn kịp giá trị cảnh báo.
- Né mốc tròn `:00/:15/:30/:45` không phải tiểu tiết: rất nhiều scheduler ngây thơ bắn đúng các giây này nên FCM cảnh giác với chúng; lệch ra là tự tách khỏi nhóm bị siết.
- RNG tiêm được là điều kiện để FR này verify được: không kiểm soát ngẫu nhiên thì không chứng minh được bất biến bucket. Production bọc `math/rand`; test seeded.
- Tiền/giá trong payload vẫn là int64 VND theo DEC-PRICE-05; scheduler không đụng tới nội dung, chỉ đụng thời điểm.

---

*Hết FR-NOTIF-004. Status: ready_to_implement (mục tiêu audit 10/10).*
