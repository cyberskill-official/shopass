package pay

import (
	"context"
	"fmt"
)

type vietqr struct{}

func (v *vietqr) Code() string { return "vietqr" }
func (v *vietqr) CreatePayment(ctx context.Context, r PaymentRequest) (PaymentResult, error) {
	return PaymentResult{
		OrderRef:  r.OrderRef,
		Gateway:   v.Code(),
		Amount:    r.Amount,
		QRPayload: fmt.Sprintf("vietqr://970415/113113113?amount=%d&addInfo=%s", r.Amount, r.OrderRef),
	}, nil
}

func NewVietQR() PaymentGateway { return &vietqr{} }
