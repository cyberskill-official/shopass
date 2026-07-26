package trend

import "time"

// Cell is an anonymized market aggregate — never includes product/user/shop ids.
type Cell struct {
	CategoryID int
	PlatformID int16
	Day        time.Time
	SKUCount   int
	P25        *int64
	Median     *int64
	P75        *int64
	Suppressed bool
}

// BuildCell applies k-anonymity suppression.
func BuildCell(categoryID int, platformID int16, day time.Time, skuCount int, p25, median, p75 int64) Cell {
	c := Cell{CategoryID: categoryID, PlatformID: platformID, Day: day, SKUCount: skuCount}
	if !Publishable(skuCount) {
		c.Suppressed = true
		return c
	}
	c.P25, c.Median, c.P75 = &p25, &median, &p75
	return c
}
