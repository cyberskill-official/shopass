package gating

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQLRepo reads plan_feature limits from Postgres.
type SQLRepo struct {
	pool *pgxpool.Pool
}

func NewSQLRepo(pool *pgxpool.Pool) *SQLRepo {
	return &SQLRepo{pool: pool}
}

func (r *SQLRepo) LimitFor(ctx context.Context, tier string, featureKey string) (int64, error) {
	var limit int64
	err := r.pool.QueryRow(ctx, `
		SELECT limit_value FROM plan_feature
		WHERE tier = $1 AND feature_key = $2`, tier, featureKey).Scan(&limit)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errors.New("feature not configured")
	}
	return limit, err
}

// CountUsage is unused when callers supply usage via AllowWithUsage; kept for
// Gate.Allow compatibility (returns 0).
func (r *SQLRepo) CountUsage(ctx context.Context, userID int64, featureKey string) (int64, error) {
	return 0, nil
}
