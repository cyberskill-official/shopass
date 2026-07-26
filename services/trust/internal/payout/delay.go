package payout

import (
	"context"
	"time"
)

type Hold struct {
	ConversionID int64
	UserID       int64
	Amount       int64
	Status       string
	HoldReason   string
	EligibleAt   time.Time
	ConfirmedAt  time.Time
}

type Store interface {
	InsertHold(ctx context.Context, h Hold) error
	GetByConversion(ctx context.Context, conversionID int64) (Hold, bool, error)
	MarkReleased(ctx context.Context, conversionID int64) error
	ExtendInvestigation(ctx context.Context, conversionID int64, reason string) error
}

type RiskReader interface {
	RiskScore(ctx context.Context, userID int64) (int, error)
}

type Config struct {
	Delay          time.Duration
	InvestigateMin int // risk_score >= this → under_investigation (never auto-denied)
	RiskExtraDelay time.Duration
}

func DefaultConfig() Config {
	return Config{
		Delay:          7 * 24 * time.Hour,
		InvestigateMin: 70,
		RiskExtraDelay: 7 * 24 * time.Hour,
	}
}

type Service struct {
	Cfg   Config
	Store Store
	Risk  RiskReader
	Guard *Guard
}

// ConfirmInput is the conversion payload used when creating a payout hold.
type ConfirmInput struct {
	ConversionID int64
	Beneficiary  int64
	Amount       int64
	ConfirmedAt  time.Time
	Inspect      Conversion // optional; used by Guard when set
}

// OnConversionConfirmed creates a payout hold — never pays immediately.
func (s *Service) OnConversionConfirmed(ctx context.Context, conversionID, userID, amount int64) (Hold, error) {
	return s.Confirm(ctx, ConfirmInput{
		ConversionID: conversionID,
		Beneficiary:  userID,
		Amount:       amount,
		ConfirmedAt:  time.Now().UTC(),
	})
}

// CreateHold is the affil postback hook shape (error-only).
func (s *Service) CreateHold(ctx context.Context, conversionID, userID, amount int64) error {
	_, err := s.OnConversionConfirmed(ctx, conversionID, userID, amount)
	return err
}

// Confirm creates a payout_hold with delay + optional gaming hold_reason / risk stretch.
func (s *Service) Confirm(ctx context.Context, in ConfirmInput) (Hold, error) {
	if existing, ok, err := s.Store.GetByConversion(ctx, in.ConversionID); err != nil {
		return Hold{}, err
	} else if ok {
		return existing, nil // idempotent
	}
	now := in.ConfirmedAt.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	delay := s.Cfg.Delay
	if delay <= 0 {
		delay = DefaultConfig().Delay
	}
	h := Hold{
		ConversionID: in.ConversionID,
		UserID:       in.Beneficiary,
		Amount:       in.Amount,
		Status:       "held",
		EligibleAt:   now.Add(delay),
		ConfirmedAt:  now,
	}

	if s.Guard != nil {
		inspect := in.Inspect
		if inspect.ConversionID == 0 {
			inspect.ConversionID = in.ConversionID
		}
		if inspect.BuyerID == 0 {
			inspect.BuyerID = in.Beneficiary
		}
		reason, err := s.Guard.Inspect(ctx, inspect)
		if err != nil {
			return Hold{}, err
		}
		if reason != "" {
			h.HoldReason = reason
		}
	}

	if s.Risk != nil {
		if score, err := s.Risk.RiskScore(ctx, in.Beneficiary); err == nil && score >= s.Cfg.InvestigateMin {
			h.Status = "under_investigation"
			if h.HoldReason == "" {
				h.HoldReason = "elevated_risk_score"
			}
			extra := s.Cfg.RiskExtraDelay
			if extra <= 0 {
				extra = DefaultConfig().RiskExtraDelay
			}
			h.EligibleAt = now.Add(delay + extra)
		}
	}

	if err := s.Store.InsertHold(ctx, h); err != nil {
		return Hold{}, err
	}
	return h, nil
}

// TryRelease pays once when eligible, not under investigation, and no hold_reason.
func (s *Service) TryRelease(ctx context.Context, conversionID int64, now time.Time) (bool, error) {
	h, ok, err := s.Store.GetByConversion(ctx, conversionID)
	if err != nil || !ok {
		return false, err
	}
	if h.Status == "released" || h.Status == "denied" || h.Status == "under_investigation" {
		return false, nil
	}
	if h.HoldReason != "" {
		return false, nil
	}
	if now.Before(h.EligibleAt) {
		return false, nil
	}
	if err := s.Store.MarkReleased(ctx, conversionID); err != nil {
		return false, err
	}
	return true, nil
}
