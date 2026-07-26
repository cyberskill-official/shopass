package trend

import (
	"math"
	"sort"
	"time"
)

// AggregateSourceSQL is the only allowed source query for B2B trend cells.
// It touches price_daily + tracked_product only (DEC-B2B-01). Kept as a constant
// so tests can grep-guard against user-level tables.
const AggregateSourceSQL = `
SELECT tp.category_id,
       tp.platform_id,
       pd.day,
       percentile_cont(0.5)  WITHIN GROUP (ORDER BY pd.close_p) AS median_p,
       percentile_cont(0.25) WITHIN GROUP (ORDER BY pd.close_p) AS p25_p,
       percentile_cont(0.75) WITHIN GROUP (ORDER BY pd.close_p) AS p75_p,
       avg(CASE WHEN pd.max_p > 0
                THEN (pd.max_p - pd.close_p)::numeric / pd.max_p * 100 END) AS avg_discount_pct,
       count(DISTINCT pd.product_id) AS sku_count
FROM price_daily pd
JOIN tracked_product tp ON tp.id = pd.product_id
WHERE pd.day >= $1 AND pd.day < $2
GROUP BY tp.category_id, tp.platform_id, pd.day
`

// ForbiddenSourceTokens must never appear in AggregateSourceSQL (DEC-B2B-01).
var ForbiddenSourceTokens = []string{
	"price_snapshot",
	"wishlist",
	"alert",
	"cart_snapshot",
}

// PricePoint is one SKU's daily OHLC contribution used by the pure aggregator.
type PricePoint struct {
	ProductID int64
	CloseP    int64
	MaxP      int64
}

// AggregateCell computes median/p25/p75/avg_discount/sku_count then applies k-anon.
func AggregateCell(categoryID int64, platformID int16, day time.Time, points []PricePoint) MarketTrendCell {
	raw := aggregatePoints(categoryID, platformID, day, points)
	return applyKAnon(raw)
}

func aggregatePoints(categoryID int64, platformID int16, day time.Time, points []PricePoint) aggRow {
	raw := aggRow{
		CategoryID: categoryID,
		PlatformID: platformID,
		Day:        day,
	}
	if len(points) == 0 {
		return raw
	}
	seen := make(map[int64]struct{}, len(points))
	closes := make([]int64, 0, len(points))
	var discSum float64
	var discN int
	for _, p := range points {
		seen[p.ProductID] = struct{}{}
		closes = append(closes, p.CloseP)
		if p.MaxP > 0 {
			d := float64(p.MaxP-p.CloseP) / float64(p.MaxP) * 100
			if d < 0 {
				d = 0
			}
			if d > 100 {
				d = 100
			}
			discSum += d
			discN++
		}
	}
	raw.SKUCount = int32(len(seen))
	sort.Slice(closes, func(i, j int) bool { return closes[i] < closes[j] })
	raw.P25P = percentileCont(closes, 0.25)
	raw.MedianP = percentileCont(closes, 0.5)
	raw.P75P = percentileCont(closes, 0.75)
	if discN > 0 {
		raw.AvgDiscountPct = math.Round(discSum/float64(discN)*100) / 100
	}
	return raw
}

// percentileCont mirrors PostgreSQL percentile_cont (linear interpolation).
func percentileCont(sorted []int64, p float64) int64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[n-1]
	}
	pos := p * float64(n-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return sorted[lo]
	}
	frac := pos - float64(lo)
	return int64(math.Round(float64(sorted[lo])*(1-frac) + float64(sorted[hi])*frac))
}

// BuildCell is a convenience wrapper used by callers that already have percentiles.
func BuildCell(categoryID int64, platformID int16, day time.Time, skuCount int, p25, median, p75 int64, avgDiscount float64) MarketTrendCell {
	return applyKAnon(aggRow{
		CategoryID:     categoryID,
		PlatformID:     platformID,
		Day:            day,
		SKUCount:       int32(skuCount),
		P25P:           p25,
		MedianP:        median,
		P75P:           p75,
		AvgDiscountPct: avgDiscount,
	})
}
