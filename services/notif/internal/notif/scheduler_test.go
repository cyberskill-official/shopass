package notif

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// seededRNG: RNG tất định cho test (DEC-NOTIF-35).
type seededRNG struct{ r *rand.Rand }

func newSeededRNG(seed int64) RNG         { return &seededRNG{rand.New(rand.NewSource(seed))} }
func (s *seededRNG) Int63n(n int64) int64 { return s.r.Int63n(n) }

func makeAlerts(t *testing.T, base time.Time, n int) []Alert {
	var alerts []Alert
	for i := 0; i < n; i++ {
		alerts = append(alerts, Alert{
			NotificationID: int64(i),
			EventTime:      base,
		})
	}
	return alerts
}

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
	require.GreaterOrEqual(t, len(counts), 2) // tràn sang ít nhất 1 phút nữa
	for _, c := range counts {
		require.LessOrEqual(t, c, FCM_RATE_LIMIT_PER_MIN)
	}
}

func TestSpread_AvoidsRoundMarks(t *testing.T) {
	rng := newSeededRNG(9)
	base := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
	alerts := makeAlerts(t, base, FCM_RATE_LIMIT_PER_MIN*5)
	sched := ScheduleAlerts(base, alerts, rng)

	overflows := 0
	for _, at := range sched {
		if at.Second() == 1 { // spread items have Second == 1
			overflows++
			// Không tin nào bị spread rơi đúng giây :00 của mốc tròn (DEC-NOTIF-34).
			require.Falsef(t, isRoundMark(at),
				"tin bị gán đúng mốc tròn %s", at.Format("15:04:05"))
		}
	}
	require.Greater(t, overflows, 0)
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
