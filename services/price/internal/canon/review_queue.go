package canon

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReviewQueue queue items that need manual review for merging
type ReviewQueue struct {
	pool *pgxpool.Pool
}

// NewReviewQueue creates a new review queue
func NewReviewQueue(pool *pgxpool.Pool) *ReviewQueue {
	return &ReviewQueue{pool: pool}
}

// Enqueue adds an item to the manual review queue
func (q *ReviewQueue) Enqueue(ctx context.Context, productID int64, candidateKey string, confidence float64) error {
	_, err := q.pool.Exec(ctx, `
		INSERT INTO canonical_review_queue (product_id, candidate_key, confidence, status)
		VALUES ($1, $2, $3, 'pending')
		ON CONFLICT (product_id, candidate_key) DO UPDATE
		SET confidence = EXCLUDED.confidence, status = 'pending'
	`, productID, candidateKey, confidence)
	return err
}
