package api

import (
	"context"

	"shopass/services/bill/internal/bill"
	"shopass/services/bill/internal/pay"
)

// checkoutPayments adapts bill.PaymentRepo to the thinner checkout PaymentRepo.
type checkoutPayments struct {
	repo bill.PaymentRepo
}

func NewCheckoutPayments(repo bill.PaymentRepo) PaymentRepo {
	return checkoutPayments{repo: repo}
}

func (c checkoutPayments) ByOrderRef(ctx context.Context, orderRef string) (pay.PaymentResult, bool) {
	rec, ok := c.repo.ByOrderRef(ctx, orderRef)
	if !ok {
		return pay.PaymentResult{}, false
	}
	return pay.PaymentResult{
		OrderRef: rec.OrderRef,
		Gateway:  rec.Gateway,
		Amount:   rec.Amount,
	}, true
}

func (c checkoutPayments) InsertPending(ctx context.Context, orderRef string, userID int64, amount int64, gateway string) {
	c.repo.InsertPending(ctx, orderRef, userID, amount, gateway)
}
