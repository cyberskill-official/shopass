package payout

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type stubCluster map[[2]int64]bool

func (s stubCluster) SameCluster(ctx context.Context, a, b int64) (bool, error) {
	if a > b {
		a, b = b, a
	}
	return s[[2]int64{a, b}], nil
}

func TestGuard_LastClickManipulation(t *testing.T) {
	g := &Guard{Cfg: DefaultGuardConfig()}
	now := time.Now().UTC()
	reason, err := g.Inspect(context.Background(), Conversion{
		BuyerID:       1,
		ReferrerID:    2,
		OrderedAt:     now,
		ClickAt:       now.Add(-1 * time.Minute),
		CartSince:     now.Add(-48 * time.Hour),
		UserInitiated: true,
	})
	require.NoError(t, err)
	require.Equal(t, "last_click_manipulation", reason)
}

func TestGuard_SelfReferral_SameCluster(t *testing.T) {
	g := &Guard{
		Cfg:     DefaultGuardConfig(),
		Cluster: stubCluster{[2]int64{10, 11}: true},
	}
	now := time.Now().UTC()
	reason, err := g.Inspect(context.Background(), Conversion{
		BuyerID:       10,
		ReferrerID:    11,
		OrderedAt:     now,
		ClickAt:       now.Add(-2 * time.Hour),
		CartSince:     now.Add(-30 * time.Minute),
		UserInitiated: true,
	})
	require.NoError(t, err)
	require.Equal(t, "self_referral", reason)
}

func TestGuard_CookieStuffing(t *testing.T) {
	g := &Guard{Cfg: DefaultGuardConfig()}
	reason, err := g.Inspect(context.Background(), Conversion{
		BuyerID:       1,
		ReferrerID:    2,
		UserInitiated: false,
	})
	require.NoError(t, err)
	require.Equal(t, "cookie_stuffing_signal", reason)
}

func TestGuard_CleanConversion_NoHold(t *testing.T) {
	g := &Guard{
		Cfg:     DefaultGuardConfig(),
		Cluster: stubCluster{},
	}
	now := time.Now().UTC()
	reason, err := g.Inspect(context.Background(), Conversion{
		BuyerID:       1,
		ReferrerID:    2,
		OrderedAt:     now,
		ClickAt:       now.Add(-2 * time.Hour),
		CartSince:     now.Add(-30 * time.Minute),
		UserInitiated: true,
	})
	require.NoError(t, err)
	require.Equal(t, "", reason)
}
