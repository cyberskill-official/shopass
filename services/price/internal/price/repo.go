package price

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SnapshotRepo handles price_snapshot persistence with delta-only writes.
type SnapshotRepo struct {
	pool *pgxpool.Pool
}

// NewSnapshotRepo creates a new repo backed by the given connection pool.
func NewSnapshotRepo(pool *pgxpool.Pool) *SnapshotRepo {
	return &SnapshotRepo{pool: pool}
}

// latest returns the most recent snapshot for product_id, or pgx.ErrNoRows.
func (r *SnapshotRepo) latest(ctx context.Context, productID int64) (PriceSnapshot, error) {
	var s PriceSnapshot
	err := r.pool.QueryRow(ctx, `
		SELECT product_id, ts, price, list_price, stock, sold, flash_sale
		FROM price_snapshot
		WHERE product_id = $1
		ORDER BY ts DESC
		LIMIT 1
	`, productID).Scan(&s.ProductID, &s.TS, &s.Price, &s.ListPrice, &s.Stock, &s.Sold, &s.FlashSale)
	return s, err
}

// InsertSnapshot applies delta-only logic (DEC-PRICE-04):
// only writes when at least one of (price, list_price, stock, flash_sale) changed.
// Returns written=true if a row was actually inserted.
func (r *SnapshotRepo) InsertSnapshot(ctx context.Context, s PriceSnapshot) (bool, error) {
	last, err := r.latest(ctx, s.ProductID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}
	// If we have a previous snapshot and nothing changed, skip.
	if err == nil && !changed(last, s) {
		return false, nil // delta_skipped
	}

	// §1 #11: ON CONFLICT DO NOTHING for idempotent retry
	_, err = r.pool.Exec(ctx, `
		INSERT INTO price_snapshot (product_id, ts, price, list_price, stock, sold, flash_sale)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (product_id, ts) DO NOTHING
	`, s.ProductID, s.TS, s.Price, s.ListPrice, s.Stock, s.Sold, s.FlashSale)
	if err != nil {
		return false, err
	}
	return true, nil
}

// QueryRange returns raw snapshots within [from, to] for a product, ordered by ts.
func (r *SnapshotRepo) QueryRange(ctx context.Context, productID int64, from, to time.Time) ([]PriceSnapshot, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT product_id, ts, price, list_price, stock, sold, flash_sale
		FROM price_snapshot
		WHERE product_id = $1 AND ts >= $2 AND ts <= $3
		ORDER BY ts
	`, productID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PriceSnapshot
	for rows.Next() {
		var s PriceSnapshot
		if err := rows.Scan(&s.ProductID, &s.TS, &s.Price, &s.ListPrice, &s.Stock, &s.Sold, &s.FlashSale); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// QueryDaily returns aggregated daily buckets from the price_daily continuous aggregate.
func (r *SnapshotRepo) QueryDaily(ctx context.Context, productID int64, from, to time.Time) ([]DailyBucket, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT product_id, day, min_p, max_p, close_p
		FROM price_daily
		WHERE product_id = $1 AND day >= $2 AND day <= $3
		ORDER BY day
	`, productID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DailyBucket
	for rows.Next() {
		var b DailyBucket
		if err := rows.Scan(&b.ProductID, &b.Day, &b.MinP, &b.MaxP, &b.CloseP); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}
