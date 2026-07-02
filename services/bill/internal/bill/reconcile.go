package bill

import (
	"context"
	"time"
)

type ReconcileJob struct {
	payments PaymentRepo
	gatewayClient gatewayClient // mock
}

type gatewayClient interface {
	CheckStatus(ctx context.Context, orderRef string) (status string, transactionID string, err error)
}

func NewReconcileJob(payments PaymentRepo, gw gatewayClient) *ReconcileJob {
	return &ReconcileJob{
		payments: payments,
		gatewayClient: gw,
	}
}

func (j *ReconcileJob) Run(ctx context.Context) {
	stale := j.payments.GetPendingOlderThan(ctx, 15 * time.Minute)
	for _, p := range stale {
		status, txID, err := j.gatewayClient.CheckStatus(ctx, p.OrderRef)
		if err != nil {
			continue // skip on network error
		}
		if status == "paid" {
			j.payments.MarkPaid(ctx, p.ID, txID)
			// TODO: Activate Subscription
		} else if status == "failed" {
			j.payments.MarkFailed(ctx, p.ID)
		}
	}
}
