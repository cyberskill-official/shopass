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
}

func DefaultConfig() Config {
	return Config{Delay: 7 * 24 * time.Hour, InvestigateMin: 70}
}

type Service struct {
	Cfg  Config
	Store Store
	Risk RiskReader
}

// OnConversionConfirmed creates a payout hold — never pays immediately.
func (s *Service) OnConversionConfirmed(ctx context.Context, conversionID, userID, amount int64) (Hold, error) {
	if existing, ok, err := s.Store.GetByConversion(ctx, conversionID); err != nil {
		return Hold{}, err
	} else if ok {
		return existing, nil // idempotent
	}
	now := time.Now().UTC()
	h := Hold{
		ConversionID: conversionID,
		UserID:       userID,
		Amount:       amount,
		Status:       "held",
		EligibleAt:   now.Add(s.Cfg.Delay),
	}
	if s.Risk != nil {
		if score, err := s.Risk.RiskScore(ctx, userID); err == nil && score >= s.Cfg.InvestigateMin {
			h.Status = "under_investigation"
			h.HoldReason = "elevated_risk_score"
		}
	}
	if err := s.Store.InsertHold(ctx, h); err != nil {
		return Hold{}, err
	}
	return h, nil
}

// TryRelease pays once when eligible and not under investigation.
func (s *Service) TryRelease(ctx context.Context, conversionID int64, now time.Time) (bool, error) {
	h, ok, err := s.Store.GetByConversion(ctx, conversionID)
	if err != nil || !ok {
		return false, err
	}
	if h.Status == "released" || h.Status == "denied" || h.Status == "under_investigation" {
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
