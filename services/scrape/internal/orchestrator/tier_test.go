package orchestrator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNextRunAt_Tiers(t *testing.T) {
	now := time.Now()

	hot := NextRunAt(TierHot, now).Sub(now)
	require.GreaterOrEqual(t, hot, 3*time.Minute)
	require.LessOrEqual(t, hot, 5*time.Minute)

	warm := NextRunAt(TierWarm, now).Sub(now)
	require.GreaterOrEqual(t, warm, 1*time.Hour)
	require.LessOrEqual(t, warm, 6*time.Hour)

	cold := NextRunAt(TierCold, now).Sub(now)
	require.GreaterOrEqual(t, cold, 23*time.Hour)
	require.LessOrEqual(t, cold, 25*time.Hour)
}

func TestReTier_PromoteOnChange(t *testing.T) {
	require.Equal(t, TierHot, ReTier(TierCold, true, false)) // giá đổi -> nóng
	require.Equal(t, TierHot, ReTier(TierWarm, false, true)) // flash -> nóng
}

func TestReTier_DemoteWhenQuiet(t *testing.T) {
	require.Equal(t, TierWarm, ReTier(TierHot, false, false))
	require.Equal(t, TierCold, ReTier(TierWarm, false, false))
	require.Equal(t, TierCold, ReTier(TierCold, false, false))
}
