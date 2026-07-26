package seller

import "errors"

var (
	ErrNotOwner             = errors.New("seller: not verified owner")
	ErrInsufficientMarket   = errors.New("seller: insufficient_market_data")
)

type MarketBand struct {
	P25    int64
	Median int64
	P75    int64
}

type Position struct {
	OwnPrice       int64
	Band           MarketBand
	PercentileRank float64
}

// ComputePosition returns own price vs anonymized band — never competitor SKUs.
func ComputePosition(ownPrice int64, band MarketBand, suppressed bool, verifiedOwner bool) (Position, error) {
	if !verifiedOwner {
		return Position{}, ErrNotOwner
	}
	if suppressed {
		return Position{}, ErrInsufficientMarket
	}
	rank := 0.5
	if ownPrice <= band.P25 {
		rank = 0.25
	} else if ownPrice >= band.P75 {
		rank = 0.75
	}
	return Position{OwnPrice: ownPrice, Band: band, PercentileRank: rank}, nil
}
