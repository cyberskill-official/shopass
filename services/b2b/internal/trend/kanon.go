package trend

// applyKAnon returns a publishable cell, or a suppressed cell with metrics NULLed
// when sku_count < KMin (DEC-B2B-02, DEC-B2B-05).
func applyKAnon(raw aggRow) MarketTrendCell {
	c := MarketTrendCell{
		CategoryID: raw.CategoryID,
		PlatformID: raw.PlatformID,
		Day:        raw.Day,
		SKUCount:   raw.SKUCount,
	}
	if raw.SKUCount < KMin {
		c.Suppressed = true
		return c
	}
	median := raw.MedianP
	p25 := raw.P25P
	p75 := raw.P75P
	disc := raw.AvgDiscountPct
	c.MedianP = &median
	c.P25P = &p25
	c.P75P = &p75
	c.AvgDiscountPct = &disc
	return c
}

// Publishable reports whether skuCount clears the k-anonymity gate.
func Publishable(skuCount int) bool {
	return skuCount >= KMin
}
