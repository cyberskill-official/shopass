package pay

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOrderRef(t *testing.T) {
	uid, tier, err := ParseOrderRef("order_42_premium_basic")
	require.NoError(t, err)
	require.Equal(t, int64(42), uid)
	require.Equal(t, "premium_basic", tier)
}

func TestNewOrderRefRoundTrip(t *testing.T) {
	ref := NewOrderRef(7, "premium_pro")
	uid, tier, err := ParseOrderRef(ref)
	require.NoError(t, err)
	require.Equal(t, int64(7), uid)
	require.Equal(t, "premium_pro", tier)
}
