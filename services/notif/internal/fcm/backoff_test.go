package fcm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNextDelay_ExponentialWithJitter(t *testing.T) {
	for attempt := 0; attempt < 5; attempt++ {
		d := nextDelay(attempt, 0)
		base := time.Duration(1<<uint(attempt)) * time.Second
		if base > maxBackoff {
			base = maxBackoff
		}
		// Delay should be in [base/2, base)
		require.GreaterOrEqual(t, d, base/2)
		require.Less(t, d, base)
	}
}

func TestNextDelay_RespectsRetryAfter(t *testing.T) {
	d := nextDelay(3, 30*time.Second)
	require.Equal(t, 30*time.Second, d) // always uses Retry-After when present
}

func TestNextDelay_CappedAt5Min(t *testing.T) {
	d := nextDelay(20, 0) // 2^20 seconds would be huge without cap
	require.LessOrEqual(t, d, maxBackoff)
}

func TestQuota_BlocksOverLimitThenRefills(t *testing.T) {
	b := NewBucketWithCap(10) // small cap for fast test
	for i := 0; i < 10; i++ {
		require.True(t, b.Allow(), "token %d should be allowed", i)
	}
	require.False(t, b.Allow(), "should block after exhausting quota")

	// Simulate minute passing
	b.mu.Lock()
	b.last = time.Now().Add(-2 * time.Minute)
	b.mu.Unlock()

	require.True(t, b.Allow(), "should refill after minute elapses")
}

func TestQuota_600k_Capacity(t *testing.T) {
	b := NewBucket()
	require.Equal(t, 600_000, b.capacity)
}
