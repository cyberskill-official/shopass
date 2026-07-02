package pay

import (
	"context"
	"testing"
)

func TestRegistry_SelectsAdapter(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewMoMo())
	reg.Register(NewZaloPay())
	reg.Register(NewVNPay())
	reg.Register(NewVietQR())

	for _, code := range []string{"momo", "zalopay", "vnpay", "vietqr"} {
		g, ok := reg.Get(code)
		if !ok {
			t.Fatalf("expected to find gateway %s", code)
		}
		if g.Code() != code {
			t.Fatalf("expected code %s, got %s", code, g.Code())
		}
	}
	_, ok := reg.Get("paypal")
	if ok {
		t.Fatalf("expected not to find paypal")
	}
}

func TestVietQR_ReturnsQRPayload(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewVietQR())
	g, _ := reg.Get("vietqr")
	res, err := g.CreatePayment(context.Background(), PaymentRequest{OrderRef: "o1", Amount: 29000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.QRPayload == "" {
		t.Fatalf("expected QR payload")
	}
	if res.Amount != 29000 {
		t.Fatalf("expected amount 29000, got %d", res.Amount)
	}
}
