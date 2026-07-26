package seller

import (
	"errors"
)

var (
	ErrNotOwner           = errors.New("seller: not verified owner")
	ErrInsufficientMarket = errors.New("seller: insufficient_market_data")
)

// Position is the seller's own price vs anonymized market band — never competitor SKUs.
type Position struct {
	CategoryID     int64   `json:"category_id"`
	PlatformID     int16   `json:"platform_id"`
	Day            string  `json:"day"`
	SellerPrice    int64   `json:"seller_price"`
	MarketMedian   int64   `json:"market_median_p"`
	MarketP25      int64   `json:"market_p25_p"`
	MarketP75      int64   `json:"market_p75_p"`
	PercentileRank float64 `json:"percentile_rank"`
}

type MarketBand struct {
	P25    int64
	Median int64
	P75    int64
}

// ComputePosition returns own price vs anonymized band.
func ComputePosition(ownPrice int64, band MarketBand, suppressed bool, verifiedOwner bool) (Position, error) {
	if !verifiedOwner {
		return Position{}, ErrNotOwner
	}
	if suppressed {
		return Position{}, ErrInsufficientMarket
	}
	return Position{
		SellerPrice:    ownPrice,
		MarketMedian:   band.Median,
		MarketP25:      band.P25,
		MarketP75:      band.P75,
		PercentileRank: Rank(ownPrice, band.P25, band.Median, band.P75),
	}, nil
}

// Rank approximates percentile of seller price within p25/median/p75 (0..100).
func Rank(sellerPrice, p25, median, p75 int64) float64 {
	switch {
	case sellerPrice <= p25:
		return 25 * float64(sellerPrice) / float64(maxI(p25, 1))
	case sellerPrice <= median:
		return 25 + 25*float64(sellerPrice-p25)/float64(maxI(median-p25, 1))
	case sellerPrice <= p75:
		return 50 + 25*float64(sellerPrice-median)/float64(maxI(p75-median, 1))
	default:
		extra := 25 * float64(sellerPrice-p75) / float64(maxI(p75, 1))
		if extra > 25 {
			extra = 25
		}
		return 75 + extra
	}
}

func maxI(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
