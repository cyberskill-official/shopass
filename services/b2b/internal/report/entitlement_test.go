package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var dNow = time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)

func ids(n int) []int64 {
	out := make([]int64, n)
	for i := range out {
		out[i] = int64(i + 1)
	}
	return out
}

func TestScope_WithinTier_OK(t *testing.T) {
	e := Entitlement{Tier: "pro", MaxCategories: 10, HistoryDays: 180, CanExport: true}
	s := ReportScope{CategoryIDs: ids(5), From: dNow.AddDate(0, 0, -90), To: dNow}
	require.NoError(t, CheckScope(e, s))
}

func TestScope_TooManyCategories_403(t *testing.T) {
	e := Entitlement{Tier: "basic", MaxCategories: 3, HistoryDays: 30}
	s := ReportScope{CategoryIDs: ids(4), From: dNow.AddDate(0, 0, -7), To: dNow}
	err := CheckScope(e, s)
	var se ErrScopeExceeded
	require.ErrorAs(t, err, &se)
	require.Equal(t, "categories", se.Field)
}

func TestScope_HistoryTooDeep_403(t *testing.T) {
	e := Entitlement{Tier: "basic", MaxCategories: 3, HistoryDays: 30}
	s := ReportScope{CategoryIDs: ids(1), From: dNow.AddDate(0, 0, -90), To: dNow}
	err := CheckScope(e, s)
	var se ErrScopeExceeded
	require.ErrorAs(t, err, &se)
	require.Equal(t, "history_days", se.Field)
}

func TestEntitlement_PaymentRequired(t *testing.T) {
	err := AssertActive(Subscription{Status: "past_due", ExpiresAt: time.Now().Add(time.Hour)}, time.Now())
	require.ErrorIs(t, err, ErrPaymentRequired)
	err = AssertActive(Subscription{Status: "active", ExpiresAt: time.Now().Add(-time.Hour)}, time.Now())
	require.ErrorIs(t, err, ErrPaymentRequired)
}

func TestEntitlement_Export(t *testing.T) {
	require.ErrorIs(t, AssertExport(Entitlement{CanExport: false}), ErrNoExport)
	require.NoError(t, AssertExport(Entitlement{CanExport: true}))
}

func TestAssertScope_Legacy(t *testing.T) {
	sub := Subscription{Status: "active", MaxCategories: 3, HistoryDays: 30, ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, AssertActive(sub, time.Now()))
	require.Error(t, AssertScope(sub, 4, 7))
}
