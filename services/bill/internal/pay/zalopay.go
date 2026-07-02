package pay

import (
	"context"
	"fmt"
)

type zalopay struct{}

func (z *zalopay) Code() string { return "zalopay" }
func (z *zalopay) CreatePayment(ctx context.Context, r PaymentRequest) (PaymentResult, error) {
	return PaymentResult{
		OrderRef: r.OrderRef,
		Gateway:  z.Code(),
		Amount:   r.Amount,
		PayURL:   fmt.Sprintf("https://sandbox.zalopay.vn/pay?amount=%d&orderId=%s", r.Amount, r.OrderRef),
	}, nil
}

func NewZaloPay() PaymentGateway { return &zalopay{} }
