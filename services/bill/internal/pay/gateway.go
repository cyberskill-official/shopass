package pay

import (
	"context"
	"fmt"
)

type PaymentRequest struct {
	OrderRef string
	Amount   int64 // VND, BIGINT
	UserID   int64
	PlanTier string
}

type PaymentResult struct {
	OrderRef  string `json:"order_ref"`
	Gateway   string `json:"gateway"`
	Amount    int64  `json:"amount"`               // VND
	PayURL    string `json:"pay_url,omitempty"`    // MoMo/ZaloPay/VNPay
	QRPayload string `json:"qr_payload,omitempty"` // VietQR
}

type PaymentGateway interface {
	Code() string
	CreatePayment(ctx context.Context, r PaymentRequest) (PaymentResult, error)
}

type Registry struct {
	byCode map[string]PaymentGateway
}

func NewRegistry() *Registry {
	return &Registry{
		byCode: make(map[string]PaymentGateway),
	}
}

func (reg *Registry) Register(g PaymentGateway) {
	reg.byCode[g.Code()] = g
}

func (reg *Registry) Get(code string) (PaymentGateway, bool) {
	g, ok := reg.byCode[code]
	return g, ok
}

// OrderRef generator logic placeholder, but the task says:
// orderRef := pay.NewOrderRef(userID, plan.Tier)
func NewOrderRef(userID int64, planTier string) string {
	return fmt.Sprintf("order_%d_%s", userID, planTier) // simpler deterministic id for demo
}
