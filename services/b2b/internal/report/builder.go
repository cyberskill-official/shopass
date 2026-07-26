package report

import (
	"context"
	"time"

	"shopass/services/b2b/internal/trend"
)

// TrendReader is the TASK-B2B-001 QueryCells surface (published cells only).
type TrendReader interface {
	QueryCells(ctx context.Context, categoryID int64, platformID int16, from, to time.Time) ([]trend.MarketTrendCell, error)
}

type TrendPoint struct {
	CategoryID     int64     `json:"category_id"`
	PlatformID     int16     `json:"platform_id"`
	Day            string    `json:"day"`
	MedianP        int64     `json:"median_p"`
	P25P           int64     `json:"p25_p"`
	P75P           int64     `json:"p75_p"`
	AvgDiscountPct float64   `json:"avg_discount_pct"`
	SKUCount       int32     `json:"sku_count"`
	ComputedAt     time.Time `json:"-"`
}

type Report struct {
	Scope            ReportScope  `json:"scope"`
	Cells            []TrendPoint `json:"cells"`
	Reason           string       `json:"reason,omitempty"`
	CachedAt         time.Time    `json:"cached_at"`
	SourceComputedAt *time.Time   `json:"source_computed_at"`
}

type Builder struct {
	Trend TrendReader
	Now   func() time.Time
}

func (b *Builder) Build(ctx context.Context, s ReportScope) (Report, error) {
	now := time.Now().UTC()
	if b.Now != nil {
		now = b.Now()
	}
	var pts []TrendPoint
	var maxComputed time.Time
	for _, cat := range s.CategoryIDs {
		for _, pf := range s.PlatformIDs {
			cells, err := b.Trend.QueryCells(ctx, cat, pf, s.From, s.To)
			if err != nil {
				return Report{}, err
			}
			for _, c := range cells {
				pts = append(pts, toPoint(c))
				if c.ComputedAt.After(maxComputed) {
					maxComputed = c.ComputedAt
				}
			}
		}
	}
	r := Report{Scope: s, Cells: pts, CachedAt: now}
	if len(pts) == 0 {
		r.Reason = "insufficient_data"
	} else {
		t := maxComputed
		r.SourceComputedAt = &t
	}
	return r, nil
}

func toPoint(c trend.MarketTrendCell) TrendPoint {
	p := TrendPoint{
		CategoryID: c.CategoryID,
		PlatformID: c.PlatformID,
		Day:        c.Day.UTC().Format("2006-01-02"),
		SKUCount:   c.SKUCount,
		ComputedAt: c.ComputedAt,
	}
	if c.MedianP != nil {
		p.MedianP = *c.MedianP
	}
	if c.P25P != nil {
		p.P25P = *c.P25P
	}
	if c.P75P != nil {
		p.P75P = *c.P75P
	}
	if c.AvgDiscountPct != nil {
		p.AvgDiscountPct = *c.AvgDiscountPct
	}
	return p
}
