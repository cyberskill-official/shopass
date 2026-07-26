package payout

import (
	"context"
	"time"
)

// Conversion is the attribution snapshot inspected for gaming signals.
type Conversion struct {
	ConversionID  int64
	BuyerID       int64
	ReferrerID    int64
	OrderedAt     time.Time
	ClickAt       time.Time
	CartSince     time.Time
	UserInitiated bool
}

type ClusterChecker interface {
	SameCluster(ctx context.Context, a, b int64) (bool, error)
}

type GuardConfig struct {
	SuspiciousClickGap time.Duration // click too close to order
	OldCartWindow      time.Duration // item already in cart this long before click
}

func DefaultGuardConfig() GuardConfig {
	return GuardConfig{
		SuspiciousClickGap: 5 * time.Minute,
		OldCartWindow:      24 * time.Hour,
	}
}

// Guard detects attribution gaming. It only returns hold_reason strings — never denies payout.
type Guard struct {
	Cfg     GuardConfig
	Cluster ClusterChecker
}

// Inspect returns a hold_reason when a gaming signal is present, or "" when clean.
func (g *Guard) Inspect(ctx context.Context, conv Conversion) (string, error) {
	if g == nil {
		return "", nil
	}
	cfg := g.Cfg
	if cfg.SuspiciousClickGap <= 0 {
		cfg.SuspiciousClickGap = DefaultGuardConfig().SuspiciousClickGap
	}
	if cfg.OldCartWindow <= 0 {
		cfg.OldCartWindow = DefaultGuardConfig().OldCartWindow
	}

	// (c) cookie-stuffing: click not user-initiated
	if !conv.UserInitiated {
		return "cookie_stuffing_signal", nil
	}

	// (a) last-click manipulation: click immediately before order on a long-lived cart item
	if !conv.ClickAt.IsZero() && !conv.OrderedAt.IsZero() && !conv.CartSince.IsZero() {
		gap := conv.OrderedAt.Sub(conv.ClickAt)
		if gap >= 0 && gap < cfg.SuspiciousClickGap {
			if conv.CartSince.Before(conv.ClickAt.Add(-cfg.OldCartWindow)) {
				return "last_click_manipulation", nil
			}
		}
	}

	// (b) self-referral: buyer and referrer in the same relationship cluster
	if g.Cluster != nil && conv.BuyerID > 0 && conv.ReferrerID > 0 && conv.BuyerID != conv.ReferrerID {
		same, err := g.Cluster.SameCluster(ctx, conv.BuyerID, conv.ReferrerID)
		if err != nil {
			return "", err
		}
		if same {
			return "self_referral", nil
		}
	}
	if conv.BuyerID > 0 && conv.ReferrerID > 0 && conv.BuyerID == conv.ReferrerID {
		return "self_referral", nil
	}

	return "", nil
}
