package payout

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type memDue struct {
	holds map[int64]Hold
	pays  int
}

func (m *memDue) ListDue(ctx context.Context, now time.Time) ([]DueHold, error) {
	var out []DueHold
	for _, h := range m.holds {
		if h.Status != "held" {
			continue
		}
		if h.HoldReason != "" {
			continue
		}
		if now.Before(h.EligibleAt) {
			continue
		}
		out = append(out, DueHold{ConversionID: h.ConversionID, UserID: h.UserID, Amount: h.Amount})
	}
	return out, nil
}

func (m *memDue) MarkReleased(ctx context.Context, id int64) error {
	h := m.holds[id]
	h.Status = "released"
	m.holds[id] = h
	return nil
}

func (m *memDue) MarkUnderInvestigation(ctx context.Context, id int64, reason string) error {
	h := m.holds[id]
	h.Status = "under_investigation"
	h.HoldReason = reason
	m.holds[id] = h
	return nil
}

func (m *memDue) Pay(ctx context.Context, h DueHold) error {
	m.pays++
	return nil
}

type stubNet map[int64]bool

func (s stubNet) NetworkConfirmed(ctx context.Context, id int64) (bool, error) {
	return s[id], nil
}

func TestRelease_HighRisk_HoldsNotDenies(t *testing.T) {
	now := time.Now().UTC()
	st := &memDue{holds: map[int64]Hold{
		1: {ConversionID: 1, UserID: 9, Amount: 100, Status: "held", EligibleAt: now.Add(-time.Hour)},
	}}
	r := &Releaser{
		Store:   st,
		Network: stubNet{1: true},
		Risk:    stubRisk{score: 90},
		Payer:   st,
		Cfg:     DefaultConfig(),
	}
	n, err := r.ReleaseDue(context.Background(), now)
	require.NoError(t, err)
	require.Equal(t, 0, n)
	require.Equal(t, "under_investigation", st.holds[1].Status)
	require.NotEqual(t, "denied", st.holds[1].Status)
}

func TestRelease_Idempotent_NoDoublePay(t *testing.T) {
	now := time.Now().UTC()
	st := &memDue{holds: map[int64]Hold{
		2: {ConversionID: 2, UserID: 1, Amount: 50, Status: "held", EligibleAt: now.Add(-time.Hour)},
	}}
	r := &Releaser{
		Store:   st,
		Network: stubNet{2: true},
		Payer:   st,
		Cfg:     DefaultConfig(),
	}
	n, err := r.ReleaseDue(context.Background(), now)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	n, err = r.ReleaseDue(context.Background(), now)
	require.NoError(t, err)
	require.Equal(t, 0, n)
	require.Equal(t, 1, st.pays)
}

func TestRelease_RequiresNetworkConfirm(t *testing.T) {
	now := time.Now().UTC()
	st := &memDue{holds: map[int64]Hold{
		3: {ConversionID: 3, UserID: 1, Amount: 50, Status: "held", EligibleAt: now.Add(-time.Hour)},
	}}
	r := &Releaser{
		Store:   st,
		Network: stubNet{3: false},
		Payer:   st,
		Cfg:     DefaultConfig(),
	}
	n, err := r.ReleaseDue(context.Background(), now)
	require.NoError(t, err)
	require.Equal(t, 0, n)
	require.Equal(t, "held", st.holds[3].Status)
}

func TestConfirm_GamingHoldReason(t *testing.T) {
	st := &memStore{}
	svc := &Service{
		Cfg:   DefaultConfig(),
		Store: st,
		Guard: &Guard{Cfg: DefaultGuardConfig()},
	}
	now := time.Now().UTC()
	h, err := svc.Confirm(context.Background(), ConfirmInput{
		ConversionID: 9,
		Beneficiary:  1,
		Amount:       1000,
		ConfirmedAt:  now,
		Inspect: Conversion{
			BuyerID:       1,
			ReferrerID:    2,
			OrderedAt:     now,
			ClickAt:       now.Add(-30 * time.Second),
			CartSince:     now.Add(-72 * time.Hour),
			UserInitiated: true,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "last_click_manipulation", h.HoldReason)
	ok, err := svc.TryRelease(context.Background(), 9, now.Add(30*24*time.Hour))
	require.NoError(t, err)
	require.False(t, ok)
}
