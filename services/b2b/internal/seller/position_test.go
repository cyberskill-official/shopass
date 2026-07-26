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
