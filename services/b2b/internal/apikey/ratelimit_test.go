package apikey

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRate_ExceedsLimit_429(t *testing.T) {
	rl := NewRateLimiter()
	fixed := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	rl.now = func() time.Time { return fixed }
	for i := 0; i < 60; i++ {
		ok, _ := rl.Allow(7, 60)
		require.True(t, ok)
	}
	ok, retry := rl.Allow(7, 60)
	require.False(t, ok)
	require.Greater(t, retry, time.Duration(0))
}

func TestAllowedEndpoint_ByTier(t *testing.T) {
	require.True(t, AllowedEndpoint("free", "/public/v1/trends"))
	require.True(t, AllowedEndpoint("pro", "/public/v1/trends"))
	require.True(t, AllowedEndpoint("enterprise", "/public/v1/trends"))
	require.False(t, AllowedEndpoint("free", "/v1/b2b/reports"))
	require.True(t, AllowedEndpoint("enterprise", "/v1/admin"))
}
