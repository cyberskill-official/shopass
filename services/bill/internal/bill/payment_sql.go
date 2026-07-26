package bill

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Repo) ByOrderRef(ctx context.Context, orderRef string) (PaymentRecord, bool) {
	var p PaymentRecord
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_ref, subscription_id, gateway, amount, fee, status, transaction_id, paid_at, created_at
		FROM payment WHERE order_ref = $1`, orderRef).Scan(
		&p.ID, &p.OrderRef, &p.SubscriptionID, &p.Gateway, &p.Amount, &p.Fee,
		&p.Status, &p.TransactionID, &p.PaidAt, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return PaymentRecord{}, false
	}
	if err != nil {
		return PaymentRecord{}, false
	}
	return p, true
}

func (r *Repo) InsertPending(ctx context.Context, orderRef string, userID int64, amount int64, gateway string) {
	// userID is encoded in order_ref; payment schema has no user_id column.
	_ = userID
	_, _ = r.pool.Exec(ctx, `
		INSERT INTO payment (order_ref, gateway, amount, status)
		VALUES ($1, $2, $3, 'pending')
		ON CONFLICT (order_ref) DO NOTHING`,
		orderRef, gateway, amount)
}

func (r *Repo) MarkPaid(ctx context.Context, id int64, transactionID string) {
	_, _ = r.pool.Exec(ctx, `
		UPDATE payment
		SET status='paid', transaction_id=$2, paid_at=now()
		WHERE id=$1 AND status='pending'`, id, transactionID)
}

func (r *Repo) MarkFailed(ctx context.Context, id int64) {
	_, _ = r.pool.Exec(ctx, `
		UPDATE payment SET status='failed' WHERE id=$1 AND status='pending'`, id)
}

func (r *Repo) MarkMismatch(ctx context.Context, id int64, gatewayAmount int64) {
	_ = gatewayAmount
	_, _ = r.pool.Exec(ctx, `
		UPDATE payment SET status='mismatch' WHERE id=$1 AND status='pending'`, id)
}

func (r *Repo) GetPendingOlderThan(ctx context.Context, d time.Duration) []PaymentRecord {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_ref, subscription_id, gateway, amount, fee, status, transaction_id, paid_at, created_at
		FROM payment
		WHERE status='pending' AND created_at < now() - $1::interval
		ORDER BY created_at`, d.String())
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []PaymentRecord
	for rows.Next() {
		var p PaymentRecord
		if err := rows.Scan(
			&p.ID, &p.OrderRef, &p.SubscriptionID, &p.Gateway, &p.Amount, &p.Fee,
			&p.Status, &p.TransactionID, &p.PaidAt, &p.CreatedAt,
		); err != nil {
			return out
		}
		out = append(out, p)
	}
	return out
}
