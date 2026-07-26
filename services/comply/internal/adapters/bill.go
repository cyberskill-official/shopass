package adapters

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Bill struct {
	pool *pgxpool.Pool
}

func NewBill(pool *pgxpool.Pool) *Bill {
	return &Bill{pool: pool}
}

func (b *Bill) AnonymizeByUser(ctx context.Context, userID int64) (int, error) {
	orderPrefix := fmt.Sprintf("order_%d_%%", userID)
	tag, err := b.pool.Exec(ctx, `
		UPDATE payment AS p
		SET order_ref = 'erased_' || p.id::text,
		    transaction_id = NULL
		WHERE p.order_ref LIKE $1
		   OR EXISTS (
		      SELECT 1
		      FROM subscription s
		      WHERE s.id = p.subscription_id AND s.user_id = $2
		   )
	`, orderPrefix, userID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}
