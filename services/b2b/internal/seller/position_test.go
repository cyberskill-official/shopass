package seller

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPosition_RequiresVerified(t *testing.T) {
	_, err := ComputePosition(100, MarketBand{1, 2, 3}, false, false)
	require.ErrorIs(t, err, ErrNotOwner)
}

func TestPosition_SuppressesThinMarket(t *testing.T) {
	_, err := ComputePosition(100, MarketBand{}, true, true)
	require.ErrorIs(t, err, ErrInsufficientMarket)
}

func TestPosition_RankMonotonic(t *testing.T) {
	band := MarketBand{P25: 100, Median: 200, P75: 300}
	low, err := ComputePosition(100, band, false, true)
	require.NoError(t, err)
	mid, err := ComputePosition(200, band, false, true)
	require.NoError(t, err)
	high, err := ComputePosition(300, band, false, true)
	require.NoError(t, err)
	require.Less(t, low.PercentileRank, mid.PercentileRank)
	require.Less(t, mid.PercentileRank, high.PercentileRank)
	require.GreaterOrEqual(t, low.PercentileRank, 0.0)
	require.LessOrEqual(t, high.PercentileRank, 100.0)
}

func TestRank_Boundaries(t *testing.T) {
	require.InDelta(t, 25.0, Rank(100, 100, 200, 300), 0.01)
	require.InDelta(t, 50.0, Rank(200, 100, 200, 300), 0.01)
	require.InDelta(t, 75.0, Rank(300, 100, 200, 300), 0.01)
}
