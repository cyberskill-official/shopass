package bill

import (
	"context"
	"time"
)

type PaymentRecord struct {
	ID             int64
	OrderRef       string
	SubscriptionID *int64
	Gateway        string
	Amount         int64
	Fee            int64
	Status         string
	TransactionID  *string
	PaidAt         *time.Time
	CreatedAt      time.Time
}

type PaymentRepo interface {
	ByOrderRef(ctx context.Context, orderRef string) (PaymentRecord, bool)
	InsertPending(ctx context.Context, orderRef string, userID int64, amount int64, gateway string)
	MarkPaid(ctx context.Context, id int64, transactionID string)
	MarkFailed(ctx context.Context, id int64)
	MarkMismatch(ctx context.Context, id int64, gatewayAmount int64)
	GetPendingOlderThan(ctx context.Context, d time.Duration) []PaymentRecord
}
