package trend

import (
	"context"
	"time"
)

// DaySource loads raw price points for a day window from price_daily + tracked_product.
// Implementations MUST NOT read user-level tables (DEC-B2B-01).
type DaySource interface {
	// LoadDays returns raw cells (pre k-anon) keyed by category/platform/day.
	LoadDays(ctx context.Context, from, to time.Time) ([]aggRow, error)
}

// PointSource is a test-friendly source that aggregates in-process.
type PointSource struct {
	// PointsByDay maps YYYY-MM-DD -> (category, platform) -> points
	Points map[string]map[cellKey][]PricePoint
}

func (s PointSource) LoadDays(_ context.Context, from, to time.Time) ([]aggRow, error) {
	from = truncateDay(from)
	to = truncateDay(to)
	var out []aggRow
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		dayKey := d.Format("2006-01-02")
		byCell, ok := s.Points[dayKey]
		if !ok {
			continue
		}
		for k, pts := range byCell {
			day, _ := time.ParseInLocation("2006-01-02", dayKey, time.UTC)
			raw := aggregatePoints(k.cat, k.pf, day, pts)
			out = append(out, raw)
		}
	}
	return out, nil
}

// Job is the idempotent nightly batch that writes market_trend_daily.
type Job struct {
	Source  DaySource
	Repo    CellRepo
	Metrics *Metrics
	// LagDays: only compute days at least this many days before "now" (default 1).
	LagDays int
}

// Run aggregates [from, to) and UPSERTs cells. Safe to re-run (DEC-B2B-04).
// With LagDays=1 (default), the exclusive upper bound is clamped to today UTC
// so only fully settled (yesterday-and-earlier) days are computed (§1 #11).
func (j *Job) Run(ctx context.Context, from, to time.Time) error {
	start := time.Now()
	lag := j.LagDays
	if lag <= 0 {
		lag = 1
	}
	// maxTo exclusive: today when lag=1 → window ends at start of today.
	maxTo := truncateDay(time.Now().UTC()).AddDate(0, 0, 1-lag)
	from = truncateDay(from)
	to = truncateDay(to)
	if to.After(maxTo) {
		to = maxTo
	}
	if !from.Before(to) {
		return nil
	}

	rows, err := j.Source.LoadDays(ctx, from, to)
	if err != nil {
		return err
	}
	cells := make([]MarketTrendCell, 0, len(rows))
	var published, suppressed int64
	for _, raw := range rows {
		c := applyKAnon(raw)
		if c.Suppressed {
			suppressed++
		} else {
			published++
		}
		cells = append(cells, c)
	}
	if err := j.Repo.UpsertCells(ctx, cells); err != nil {
		return err
	}
	if j.Metrics != nil {
		j.Metrics.RecordJob(ctx, published, suppressed, time.Since(start))
	}
	return nil
}
