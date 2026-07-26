package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEntitlement_PaymentRequired(t *testing.T) {
	err := AssertActive(Subscription{Status: "expired", ExpiresAt: time.Now().Add(time.Hour)}, time.Now())
	require.ErrorIs(t, err, ErrPaymentRequired)
}

func TestEntitlement_Scope(t *testing.T) {
	sub := Subscription{Status: "active", MaxCategories: 3, HistoryDays: 30, ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, AssertActive(sub, time.Now()))
	require.ErrorIs(t, AssertScope(sub, 4, 7), ErrForbiddenScope)
	require.ErrorIs(t, AssertExport(sub), ErrNoExport)
}
