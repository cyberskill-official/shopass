package payout

import (
	"context"
	"time"
)

// NetworkConfirmReader reports whether the affiliate network confirmed a conversion.
type NetworkConfirmReader interface {
	NetworkConfirmed(ctx context.Context, conversionID int64) (bool, error)
}

// DueHold is a hold ready for payout consideration.
type DueHold struct {
	ConversionID int64
	UserID       int64
	Amount       int64
}

// DueStore selects and releases due holds (idempotent via status flip).
type DueStore interface {
	ListDue(ctx context.Context, now time.Time) ([]DueHold, error)
	MarkReleased(ctx context.Context, conversionID int64) error
	MarkUnderInvestigation(ctx context.Context, conversionID int64, reason string) error
}

// Payer performs the actual payout once (bill/cashback). Tests may use a counter.
type Payer interface {
	Pay(ctx context.Context, h DueHold) error
}

// Releaser pays holds that meet all three release conditions.
type Releaser struct {
	Store   DueStore
	Network NetworkConfirmReader
	Risk    RiskReader
	Payer   Payer
	Cfg     Config
}

// ReleaseDue pays each hold that is past eligible_at, has no hold_reason,
// is network-confirmed, and is not under investigation. Never auto-denies.
func (r *Releaser) ReleaseDue(ctx context.Context, now time.Time) (int, error) {
	if r == nil || r.Store == nil {
		return 0, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	due, err := r.Store.ListDue(ctx, now)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, h := range due {
		if r.Risk != nil {
			score, err := r.Risk.RiskScore(ctx, h.UserID)
			if err != nil {
				return released, err
			}
			if score >= r.Cfg.InvestigateMin {
				if err := r.Store.MarkUnderInvestigation(ctx, h.ConversionID, "elevated_risk_score"); err != nil {
					return released, err
				}
				continue // hold, never deny
			}
		}
		if r.Network != nil {
			ok, err := r.Network.NetworkConfirmed(ctx, h.ConversionID)
			if err != nil {
				return released, err
			}
			if !ok {
				continue
			}
		}
		if r.Payer != nil {
			if err := r.Payer.Pay(ctx, h); err != nil {
				return released, err
			}
		}
		if err := r.Store.MarkReleased(ctx, h.ConversionID); err != nil {
			return released, err
		}
		released++
	}
	return released, nil
}

// Blocked reports whether cashback/TRUST still holds a conversion (for AFFIL-005).
func (s *Service) Blocked(ctx context.Context, conversionID int64) (bool, error) {
	h, ok, err := s.Store.GetByConversion(ctx, conversionID)
	if err != nil || !ok {
		return false, err
	}
	switch h.Status {
	case "released", "denied":
		return false, nil
	case "under_investigation":
		return true, nil
	default:
		if h.HoldReason != "" {
			return true, nil
		}
		return time.Now().UTC().Before(h.EligibleAt), nil
	}
}
