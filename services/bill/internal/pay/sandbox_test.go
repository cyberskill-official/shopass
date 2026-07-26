package pay

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMoMoSandbox_StubWithoutCreds(t *testing.T) {
	t.Setenv("MOMO_SECRET_KEY", "")
	t.Setenv("MOMO_PARTNER_CODE", "")
	gw := NewMoMoSandbox(NewEnvSecrets())
	res, err := gw.CreatePayment(context.Background(), PaymentRequest{
		OrderRef: "order_1_premium_basic", Amount: 29000, UserID: 1, PlanTier: "premium_basic",
	})
	require.NoError(t, err)
	require.Contains(t, res.PayURL, "test-payment.momo.vn")
	require.NotContains(t, res.PayURL, "signature=")
}

func TestMoMoSandbox_SignedWhenCredsPresent(t *testing.T) {
	t.Setenv("MOMO_SECRET_KEY", "test-secret")
	t.Setenv("MOMO_PARTNER_CODE", "MOMOXX")
	gw := NewMoMoSandbox(NewEnvSecrets())
	res, err := gw.CreatePayment(context.Background(), PaymentRequest{
		OrderRef: "order_9_premium_plus", Amount: 49000, UserID: 9, PlanTier: "premium_plus",
	})
	require.NoError(t, err)
	require.Contains(t, res.PayURL, "signature=")
	require.True(t, strings.Contains(res.PayURL, "orderId=order_9_premium_plus") || strings.Contains(res.PayURL, "orderId%3Dorder_9_premium_plus"))
}
