package pay

import (
	"context"
	"fmt"
)

type vnpay struct{}

func (v *vnpay) Code() string { return "vnpay" }
func (v *vnpay) CreatePayment(ctx context.Context, r PaymentRequest) (PaymentResult, error) {
	return PaymentResult{
		OrderRef: r.OrderRef,
		Gateway:  v.Code(),
		Amount:   r.Amount,
		PayURL:   fmt.Sprintf("https://sandbox.vnpayment.vn/paymentv2/vpcpay.html?vnp_Amount=%d&vnp_TxnRef=%s", r.Amount*100, r.OrderRef), // VNPay expects amount * 100
	}, nil
}

func NewVNPay() PaymentGateway { return &vnpay{} }
