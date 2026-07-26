package trend

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// CellRepo persists and reads anonymized market trend cells.
type CellRepo interface {
	UpsertCells(ctx context.Context, cells []MarketTrendCell) error
	// QueryCells returns published cells only (suppressed=false) by default.
	QueryCells(ctx context.Context, categoryID int64, platformID int16, from, to time.Time) ([]MarketTrendCell, error)
	// QueryAll includes suppressed cells (audit / idempotency checks).
	QueryAll(ctx context.Context, categoryID int64, platformID int16, from, to time.Time) ([]MarketTrendCell, error)
}

type cellKey struct {
	cat int64
	pf  int16
	day string
}

func keyOf(c MarketTrendCell) cellKey {
	return cellKey{cat: c.CategoryID, pf: c.PlatformID, day: c.Day.UTC().Format("2006-01-02")}
}

// MemoryRepo is an in-process CellRepo for unit tests and noop deployments.
type MemoryRepo struct {
	mu    sync.Mutex
	cells map[cellKey]MarketTrendCell
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{cells: make(map[cellKey]MarketTrendCell)}
}

func (r *MemoryRepo) UpsertCells(_ context.Context, cells []MarketTrendCell) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	for _, c := range cells {
		if c.SKUCount < 0 {
			return fmt.Errorf("trend: sku_count must be >= 0")
		}
		if !c.Suppressed {
			if c.P25P == nil || c.MedianP == nil || c.P75P == nil {
				return fmt.Errorf("trend: published cell missing percentiles")
			}
			if *c.P25P > *c.MedianP || *c.MedianP > *c.P75P {
				return fmt.Errorf("trend: percentile order violated")
			}
		}
		cp := c
		if cp.ComputedAt.IsZero() {
			cp.ComputedAt = now
		}
		// Normalize day to UTC midnight for stable keys.
		cp.Day = time.Date(c.Day.Year(), c.Day.Month(), c.Day.Day(), 0, 0, 0, 0, time.UTC)
		r.cells[keyOf(cp)] = cp
	}
	return nil
}

func (r *MemoryRepo) QueryCells(ctx context.Context, categoryID int64, platformID int16, from, to time.Time) ([]MarketTrendCell, error) {
	all, err := r.QueryAll(ctx, categoryID, platformID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]MarketTrendCell, 0, len(all))
	for _, c := range all {
		if !c.Suppressed {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *MemoryRepo) QueryAll(_ context.Context, categoryID int64, platformID int16, from, to time.Time) ([]MarketTrendCell, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	from = truncateDay(from)
	to = truncateDay(to)
	out := make([]MarketTrendCell, 0)
	for _, c := range r.cells {
		if c.CategoryID != categoryID || c.PlatformID != platformID {
			continue
		}
		d := truncateDay(c.Day)
		if (d.Equal(from) || d.After(from)) && d.Before(to) {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Day.Before(out[j].Day) })
	return out, nil
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// UpsertSQL is the idempotent write used by PG-backed repos (DEC-B2B-04).
const UpsertSQL = `
INSERT INTO market_trend_daily (
  category_id, platform_id, day, median_p, p25_p, p75_p,
  avg_discount_pct, sku_count, suppressed, computed_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,now())
ON CONFLICT (category_id, platform_id, day) DO UPDATE SET
  median_p = EXCLUDED.median_p,
  p25_p = EXCLUDED.p25_p,
  p75_p = EXCLUDED.p75_p,
  avg_discount_pct = EXCLUDED.avg_discount_pct,
  sku_count = EXCLUDED.sku_count,
  suppressed = EXCLUDED.suppressed,
  computed_at = now()
`

// QueryPublishedSQL is the default consumer read (filters suppressed).
const QueryPublishedSQL = `
SELECT category_id, platform_id, day, median_p, p25_p, p75_p,
       avg_discount_pct, sku_count, suppressed, computed_at
FROM market_trend_daily
WHERE category_id = $1 AND platform_id = $2
  AND day >= $3 AND day < $4
  AND suppressed = false
ORDER BY day
`
