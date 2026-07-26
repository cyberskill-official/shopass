package cashback

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type memCB struct {
	byConv  map[int64]Entry
	payouts []int64
}

func (m *memCB) InsertHeld(ctx context.Context, e Entry) error {
	if m.byConv == nil {
		m.byConv = map[int64]Entry{}
	}
	m.byConv[e.ConversionID] = e
	return nil
}
func (m *memCB) GetByConversion(ctx context.Context, id int64) (Entry, bool, error) {
	e, ok := m.byConv[id]
	return e, ok, nil
}
func (m *memCB) MarkReleased(ctx context.Context, id int64) error {
	e := m.byConv[id]
	e.Status = "released"
	m.byConv[id] = e
	return nil
}
func (m *memCB) MarkClawedBack(ctx context.Context, id int64) error {
	e := m.byConv[id]
	e.Status = "clawed_back"
	m.byConv[id] = e
	return nil
}
func (m *memCB) SumReleasedUnpaid(ctx context.Context, userID int64) (int64, error) {
	var sum int64
	for _, e := range m.byConv {
		if e.UserID == userID && e.Status == "released" {
			sum += e.UserShare
		}
	}
	return sum, nil
}
func (m *memCB) CreatePayoutRequest(ctx context.Context, userID, amount int64) error {
	m.payouts = append(m.payouts, amount)
	return nil
}

type holdStub struct{ blocked bool }

func (h holdStub) Blocked(ctx context.Context, conversionID int64) (bool, error) {
	return h.blocked, nil
}

func TestSplit_FloorShare(t *testing.T) {
	share, kept := Split(10_001, 5000)
	require.Equal(t, int64(5000), share)
	require.Equal(t, int64(5001), kept)
}

func TestLedger_ReleaseBlockedByTrust(t *testing.T) {
	st := &memCB{}
	l := &Ledger{Cfg: DefaultConfig(), Store: st, Hold: holdStub{blocked: true}}
	_, err := l.OnConfirmed(context.Background(), 1, 9, 100_000)
	require.NoError(t, err)
	require.Error(t, l.TryRelease(context.Background(), 1))
}

func TestLedger_PayoutThreshold(t *testing.T) {
	st := &memCB{}
	l := &Ledger{Cfg: DefaultConfig(), Store: st, Hold: holdStub{}}
	_, _ = l.OnConfirmed(context.Background(), 1, 9, 120_000)
	require.NoError(t, l.TryRelease(context.Background(), 1))
	ok, err := l.MaybeRequestPayout(context.Background(), 9)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []int64{60_000}, st.payouts)
}
