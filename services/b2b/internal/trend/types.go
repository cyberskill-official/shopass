package trend

import "time"

// KMin is the k-anonymity threshold: cells with fewer distinct SKUs are suppressed.
const KMin = 50

// MarketTrendCell is one anonymized (category × platform × day) aggregate.
// It MUST NOT carry product_id, shop_id, or user_id (DEC-B2B-06).
type MarketTrendCell struct {
	CategoryID       int64     `db:"category_id" json:"category_id"`
	PlatformID       int16     `db:"platform_id" json:"platform_id"`
	Day              time.Time `db:"day" json:"day"`
	MedianP          *int64    `db:"median_p" json:"median_p"`
	P25P             *int64    `db:"p25_p" json:"p25_p"`
	P75P             *int64    `db:"p75_p" json:"p75_p"`
	AvgDiscountPct   *float64  `db:"avg_discount_pct" json:"avg_discount_pct"`
	SKUCount         int32     `db:"sku_count" json:"sku_count"`
	Suppressed       bool      `db:"suppressed" json:"suppressed"`
	ComputedAt       time.Time `db:"computed_at" json:"computed_at,omitempty"`
}

// aggRow is the raw grouped result before the k-anonymity gate.
type aggRow struct {
	CategoryID     int64
	PlatformID     int16
	Day            time.Time
	MedianP        int64
	P25P           int64
	P75P           int64
	AvgDiscountPct float64
	SKUCount       int32
}
