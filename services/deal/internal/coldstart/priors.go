package coldstart

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type CategoryPrior struct {
    CategoryID    int64   `db:"category_id"`
    MedianPrice   int64   `db:"median_price"`
    DiscountDepth float64 `db:"typical_discount_depth"`
    SampleCount   int     `db:"sample_count"`
}

type Queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repo struct {
	pool Queryer
}

func NewRepo(pool Queryer) *Repo {
	return &Repo{pool: pool}
}

// PriorFor đọc prior của 1 category; ok=false khi không có category,
// hoặc sample_count dưới sàn min_sample_count (prior quá mỏng, §1 #12).
func (r *Repo) PriorFor(ctx context.Context, categoryID int64) (CategoryPrior, bool, error) {
    var cp CategoryPrior
    err := r.pool.QueryRow(ctx,
        `SELECT category_id, median_price, typical_discount_depth, sample_count
         FROM category_prior WHERE category_id = $1`, categoryID).
        Scan(&cp.CategoryID, &cp.MedianPrice, &cp.DiscountDepth, &cp.SampleCount)
    
    if errors.Is(err, pgx.ErrNoRows) {
        return CategoryPrior{}, false, nil
    }
    if err != nil {
        return CategoryPrior{}, false, err
    }
    if cp.SampleCount < minSamples {
        return cp, false, nil // prior mỏng -> coi như không đủ tin
    }
    return cp, true, nil
}
