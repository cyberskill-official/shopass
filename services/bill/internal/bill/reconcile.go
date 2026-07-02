package bill

import (
	"context"
	"time"
)

type ReconcileJob struct {
	payments      PaymentRepo
	subs          SubscriptionActivator
	gatewayClient gatewayClient // mock
}

type SubscriptionActivator interface {
	ActivateSubscription(ctx context.Context, subID int64, duration time.Duration) error
}

type gatewayClient interface {
	CheckStatus(ctx context.Context, orderRef string) (status string, transactionID string, err error)
}

func NewReconcileJob(payments PaymentRepo, subs SubscriptionActivator, gw gatewayClient) *ReconcileJob {
	return &ReconcileJob{
		payments:      payments,
		subs:          subs,
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
			if p.SubscriptionID != nil && j.subs != nil {
				// Activate Subscription for 30 days
				j.subs.ActivateSubscription(ctx, *p.SubscriptionID, 30*24*time.Hour)
			}
		} else if status == "failed" {
			j.payments.MarkFailed(ctx, p.ID)
		}
	}
}
