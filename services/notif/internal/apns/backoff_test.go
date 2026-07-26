package apns

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBackoffDelay_ExponentialBounded(t *testing.T) {
	d := BackoffDelay(0, 0)
	require.True(t, d >= 500*time.Millisecond)
	require.True(t, d <= 2*time.Second)

	d = BackoffDelay(20, 0)
	require.True(t, d <= maxBackoff)
}
