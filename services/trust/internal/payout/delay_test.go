package payout

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type memStore struct {
	byConv map[int64]Hold
}

func (m *memStore) InsertHold(ctx context.Context, h Hold) error {
	if m.byConv == nil {
		m.byConv = map[int64]Hold{}
	}
	m.byConv[h.ConversionID] = h
	return nil
}
func (m *memStore) GetByConversion(ctx context.Context, id int64) (Hold, bool, error) {
	h, ok := m.byConv[id]
	return h, ok, nil
}
func (m *memStore) MarkReleased(ctx context.Context, id int64) error {
	h := m.byConv[id]
	h.Status = "released"
	m.byConv[id] = h
	return nil
}
func (m *memStore) ExtendInvestigation(ctx context.Context, id int64, reason string) error {
	h := m.byConv[id]
	h.Status = "under_investigation"
	h.HoldReason = reason
	m.byConv[id] = h
	return nil
}

type stubRisk struct{ score int }

func (s stubRisk) RiskScore(ctx context.Context, userID int64) (int, error) { return s.score, nil }

func TestOnConfirm_CreatesHoldNotPay(t *testing.T) {
	st := &memStore{}
	svc := &Service{Cfg: DefaultConfig(), Store: st, Risk: stubRisk{score: 10}}
	h, err := svc.OnConversionConfirmed(context.Background(), 1, 9, 5000)
	require.NoError(t, err)
	require.Equal(t, "held", h.Status)
	require.True(t, h.EligibleAt.After(time.Now()))
}

func TestOnConfirm_HighRiskInvestigates_NotDenied(t *testing.T) {
	st := &memStore{}
	svc := &Service{Cfg: DefaultConfig(), Store: st, Risk: stubRisk{score: 90}}
	h, err := svc.OnConversionConfirmed(context.Background(), 2, 9, 5000)
	require.NoError(t, err)
	require.Equal(t, "under_investigation", h.Status)
	require.NotEqual(t, "denied", h.Status)
}

func TestTryRelease_IdempotentWindow(t *testing.T) {
	st := &memStore{}
	svc := &Service{Cfg: DefaultConfig(), Store: st}
	h, _ := svc.OnConversionConfirmed(context.Background(), 3, 1, 100)
	ok, err := svc.TryRelease(context.Background(), 3, h.EligibleAt.Add(-time.Hour))
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = svc.TryRelease(context.Background(), 3, h.EligibleAt.Add(time.Hour))
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = svc.TryRelease(context.Background(), 3, h.EligibleAt.Add(2*time.Hour))
	require.NoError(t, err)
	require.False(t, ok)
}
