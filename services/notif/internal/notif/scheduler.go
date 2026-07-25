package notif

import (
	"sort"
	"time"
)

const (
	// FCM_RATE_LIMIT_PER_MIN: trần cứng 600.000 tin/phút/project (§3.6).
	// Không bucket phút nào được vượt giá trị này, nếu không sẽ ăn 429.
	FCM_RATE_LIMIT_PER_MIN = 600_000

	// Jitter bất đối xứng [-90s, +180s] (§3.5(5), DEC-NOTIF-31).
	JitterMin = -90 * time.Second // 90 giây TRƯỚC event_time
	JitterMax = 180 * time.Second // 180 giây SAU event_time
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

type scheduled struct {
	ID        int64
	At        time.Time
	HighValue bool
}

// minuteKey quy một thời điểm về phút chứa nó: floor(t / 60s) (DEC-NOTIF-32).
func minuteKey(t time.Time) int64 { return t.Unix() / 60 }

// isRoundMark cho biết giây-trong-giờ của t có rơi đúng mốc tròn
// :00/:15/:30/:45 không (DEC-NOTIF-34).
func isRoundMark(t time.Time) bool {
	m := t.Minute()
	return t.Second() == 0 && (m == 0 || m == 15 || m == 30 || m == 45)
}

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

func sortedKeys(m map[int64][]scheduled) []int64 {
	var keys []int64
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortStable(b []scheduled) {
	sort.SliceStable(b, func(i, j int) bool {
		if b[i].HighValue != b[j].HighValue {
			return b[i].HighValue // true comes first
		}
		return b[i].ID < b[j].ID
	})
}

func containsKey(keys []int64, k int64) bool {
	for _, key := range keys {
		if key == k {
			return true
		}
	}
	return false
}

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
