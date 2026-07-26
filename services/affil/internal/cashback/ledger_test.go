package cashback

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type holdStub struct{ blocked bool }

func (h holdStub) Blocked(ctx context.Context, conversionID int64) (bool, error) {
	return h.blocked, nil
}

func TestSplit_FloorShare(t *testing.T) {
	share, kept := Split(10_001, 5000)
	require.Equal(t, int64(5000), share)
	require.Equal(t, int64(5001), kept)
}

func TestCashback_ConfirmedCreatesPending(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st}
	t0 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	e, err := l.OnConfirmed(context.Background(), Conversion{
		ID: 1, UserID: 9, Commission: 100_000, UserTier: TierPremium, ConfirmedAt: t0,
	})
	require.NoError(t, err)
	require.Equal(t, StatusPending, e.Status)
	require.Equal(t, int64(50_000), e.UserShare)
	require.Equal(t, int64(50_000), e.KeptMargin)
	require.Equal(t, e.Commission, e.UserShare+e.KeptMargin)
	require.Equal(t, t0.Add(7*24*time.Hour), e.AvailableAt)
}

func TestCashback_FreeTierShare(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st}
	e, err := l.OnConfirmed(context.Background(), Conversion{
		ID: 2, UserID: 9, Commission: 100_000, UserTier: TierFree, ConfirmedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
	require.Equal(t, int64(30_000), e.UserShare)
	require.Equal(t, int64(70_000), e.KeptMargin)
}

func TestCashback_Idempotent(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st}
	c := Conversion{ID: 3, UserID: 9, Commission: 100_000, UserTier: TierFree, ConfirmedAt: time.Now().UTC()}
	_, err := l.OnConfirmed(context.Background(), c)
	require.NoError(t, err)
	_, err = l.OnConfirmed(context.Background(), c)
	require.NoError(t, err)
	require.Equal(t, 1, len(st.byConv))
}

func TestRelease_HoldsBeforeWindow(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st, Hold: holdStub{}}
	t0 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	_, err := l.OnConfirmed(context.Background(), Conversion{
		ID: 4, UserID: 9, Commission: 100_000, UserTier: TierFree, ConfirmedAt: t0,
	})
	require.NoError(t, err)
	n, err := NewReleaser(l).ReleaseDue(context.Background(), t0.Add(24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, 0, n)
	e, ok, _ := st.GetByConversion(context.Background(), 4)
	require.True(t, ok)
	require.Equal(t, StatusPending, e.Status)
}

func TestRelease_AfterWindowAvailable(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st, Hold: holdStub{}}
	t0 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	_, err := l.OnConfirmed(context.Background(), Conversion{
		ID: 5, UserID: 9, Commission: 100_000, UserTier: TierFree, ConfirmedAt: t0,
	})
	require.NoError(t, err)
	n, err := NewReleaser(l).ReleaseDue(context.Background(), t0.Add(30*24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, 1, n)
	e, _, _ := st.GetByConversion(context.Background(), 5)
	require.Equal(t, StatusAvailable, e.Status)
}

func TestRelease_FlaggedUserNotReleased(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st, Hold: holdStub{blocked: true}}
	t0 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	_, err := l.OnConfirmed(context.Background(), Conversion{
		ID: 6, UserID: 9, Commission: 100_000, UserTier: TierFree, ConfirmedAt: t0,
	})
	require.NoError(t, err)
	n, err := NewReleaser(l).ReleaseDue(context.Background(), t0.Add(30*24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, 0, n)
	e, _, _ := st.GetByConversion(context.Background(), 6)
	require.Equal(t, StatusPending, e.Status)
}

func TestClawback_OnReject(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st}
	_, err := l.OnConfirmed(context.Background(), Conversion{
		ID: 7, UserID: 9, Commission: 100_000, UserTier: TierPremium, ConfirmedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
	require.NoError(t, l.Clawback(context.Background(), 7))
	e, _, _ := st.GetByConversion(context.Background(), 7)
	require.Equal(t, StatusClawedBack, e.Status)
}

func TestPayout_BelowThreshold(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st, Hold: holdStub{}, Payer: NewVietQRStub(nil)}
	t0 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	// free 30% of 100_000 = 30_000 < 50_000
	_, err := l.OnConfirmed(context.Background(), Conversion{
		ID: 8, UserID: 9, Commission: 100_000, UserTier: TierFree, ConfirmedAt: t0,
	})
	require.NoError(t, err)
	_, err = NewReleaser(l).ReleaseDue(context.Background(), t0.Add(30*24*time.Hour))
	require.NoError(t, err)
	ok, err := l.MaybeRequestPayout(context.Background(), 9)
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, st.Payouts())
}

func TestPayout_AtThreshold(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st, Hold: holdStub{}, Payer: NewVietQRStub(nil)}
	t0 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	// premium 50% of 120_000 = 60_000 >= 50_000
	_, err := l.OnConfirmed(context.Background(), Conversion{
		ID: 9, UserID: 9, Commission: 120_000, UserTier: TierPremium, ConfirmedAt: t0,
	})
	require.NoError(t, err)
	_, err = NewReleaser(l).ReleaseDue(context.Background(), t0.Add(30*24*time.Hour))
	require.NoError(t, err)
	ok, err := l.MaybeRequestPayout(context.Background(), 9)
	require.NoError(t, err)
	require.True(t, ok)
	require.Len(t, st.Payouts(), 1)
	require.Equal(t, int64(60_000), st.Payouts()[0].Amount)
	e, _, _ := st.GetByConversion(context.Background(), 9)
	require.Equal(t, StatusPaid, e.Status)
}

func TestSummary_Disclosure(t *testing.T) {
	st := NewMemStore()
	l := &Ledger{Cfg: DefaultConfig(), Store: st}
	t0 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	_, err := l.OnConfirmed(context.Background(), Conversion{
		ID: 10, UserID: 9, Commission: 100_000, UserTier: TierPremium, ConfirmedAt: t0,
	})
	require.NoError(t, err)
	sum, err := st.Summary(context.Background(), 9)
	require.NoError(t, err)
	resp := sum.ToResponse(l.Cfg.PayoutThreshold)
	require.Equal(t, 1, resp.Pending.Count)
	require.Equal(t, int64(50_000), resp.Pending.AmountVND)
	require.NotNil(t, resp.Pending.NextAvailableAt)
	require.Equal(t, "2026-07-08", *resp.Pending.NextAvailableAt)
	require.Equal(t, DisclosureNote, resp.Note)
	require.Equal(t, int64(50_000), resp.PayoutThresholdVND)
}
