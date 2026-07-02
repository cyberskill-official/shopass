package pay

import (
	"context"
	"fmt"
)

type momo struct{}

func (m *momo) Code() string { return "momo" }
func (m *momo) CreatePayment(ctx context.Context, r PaymentRequest) (PaymentResult, error) {
	return PaymentResult{
		OrderRef: r.OrderRef,
		Gateway:  m.Code(),
		Amount:   r.Amount,
		PayURL:   fmt.Sprintf("https://test-payment.momo.vn/pay?amount=%d&orderId=%s", r.Amount, r.OrderRef),
	}, nil
}

func NewMoMo() PaymentGateway { return &momo{} }
