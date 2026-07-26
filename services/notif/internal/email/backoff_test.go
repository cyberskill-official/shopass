package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBackoffDelay_ExponentialWithJitter(t *testing.T) {
	for attempt := 0; attempt < 5; attempt++ {
		d := BackoffDelay(attempt, 0)
		base := time.Duration(1<<uint(attempt)) * time.Second
		if base > maxBackoff {
			base = maxBackoff
		}
		require.GreaterOrEqual(t, d, base/2)
		require.Less(t, d, base)
	}
}

func TestBackoffDelay_RespectsRetryAfter(t *testing.T) {
	require.Equal(t, 42*time.Second, BackoffDelay(3, 42*time.Second))
}

func TestBackoffDelay_CappedAtMax(t *testing.T) {
	require.LessOrEqual(t, BackoffDelay(20, 0), maxBackoff)
}
