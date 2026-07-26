package cashback

import (
	"context"
	"fmt"
	"time"
)

const (
	StatusPending    = "pending"
	StatusAvailable  = "available"
	StatusPaid       = "paid"
	StatusClawedBack = "clawed_back"

	TierFree    = "free"
	TierPremium = "premium"

	DisclosureNote = "Cashback pending co the bi huy neu don bi huy hoac hoan."
)

// Conversion is the confirmed affiliate conversion payload for cashback.
type Conversion struct {
	ID          int64
	UserID      int64
	Commission  int64
	UserTier    string
	ConfirmedAt time.Time
}

type Entry struct {
	ID           int64
	ConversionID int64
	UserID       int64
	Commission   int64
	UserShare    int64
	KeptMargin   int64
	Status       string
	AvailableAt  time.Time
	PaidAt       *time.Time
	CreatedAt    time.Time
}

type Store interface {
	InsertPending(ctx context.Context, e Entry) error
	GetByConversion(ctx context.Context, conversionID int64) (Entry, bool, error)
	ListDuePending(ctx context.Context, now time.Time) ([]Entry, error)
	MarkAvailable(ctx context.Context, conversionID int64) error
	MarkClawedBack(ctx context.Context, conversionID int64) error
	SumAvailable(ctx context.Context, userID int64) (int64, error)
	ListAvailable(ctx context.Context, userID int64) ([]Entry, error)
	MarkPaid(ctx context.Context, conversionIDs []int64, paidAt time.Time) error
	CreatePayoutRequest(ctx context.Context, userID, amount int64, gatewayRef string) (int64, error)
	Summary(ctx context.Context, userID int64) (UserSummary, error)
}

// HoldChecker is TRUST-005: true while payout/cashback must stay held.
type HoldChecker interface {
	Blocked(ctx context.Context, conversionID int64) (bool, error)
}

// Metrics records cashback counters (VND). Optional; may be nil.
type Metrics interface {
	Pending(vnd int64)
	Released(vnd int64)
	Clawback(vnd int64)
	Paid(vnd int64)
}

type Config struct {
	ShareRateFreeBPS    int64 // basis points, e.g. 3000 = 30%
	ShareRatePremiumBPS int64 // e.g. 5000 = 50%
	PayoutThreshold     int64
	InvestigationWindow time.Duration
}

func DefaultConfig() Config {
	return Config{
		ShareRateFreeBPS:    3000,
		ShareRatePremiumBPS: 5000,
		PayoutThreshold:     50_000,
		InvestigationWindow: 7 * 24 * time.Hour,
	}
}

type Ledger struct {
	Cfg     Config
	Store   Store
	Hold    HoldChecker
	Metrics Metrics
	Payer   Payer
}

func Split(commission, shareRateBPS int64) (userShare, kept int64) {
	userShare = commission * shareRateBPS / 10_000
	kept = commission - userShare
	return userShare, kept
}

func (c Config) ShareRateBPS(tier string) int64 {
	switch tier {
	case TierPremium:
		if c.ShareRatePremiumBPS > 0 {
			return c.ShareRatePremiumBPS
		}
	default:
		if c.ShareRateFreeBPS > 0 {
			return c.ShareRateFreeBPS
		}
	}
	return DefaultConfig().ShareRateFreeBPS
}

// OnConfirmed records a pending cashback entry (BIGINT VND). Idempotent per conversion.
func (l *Ledger) OnConfirmed(ctx context.Context, c Conversion) (Entry, error) {
	if c.Commission < 0 {
		return Entry{}, fmt.Errorf("cashback: negative commission")
	}
	if existing, ok, err := l.Store.GetByConversion(ctx, c.ID); err != nil {
		return Entry{}, err
	} else if ok {
		return existing, nil
	}
	confirmedAt := c.ConfirmedAt.UTC()
	if confirmedAt.IsZero() {
		confirmedAt = time.Now().UTC()
	}
	window := l.Cfg.InvestigationWindow
	if window <= 0 {
		window = DefaultConfig().InvestigationWindow
	}
	share, kept := Split(c.Commission, l.Cfg.ShareRateBPS(c.UserTier))
	e := Entry{
		ConversionID: c.ID,
		UserID:       c.UserID,
		Commission:   c.Commission,
		UserShare:    share,
		KeptMargin:   kept,
		Status:       StatusPending,
		AvailableAt:  confirmedAt.Add(window),
		CreatedAt:    confirmedAt,
	}
	if err := l.Store.InsertPending(ctx, e); err != nil {
		return Entry{}, err
	}
	if l.Metrics != nil {
		l.Metrics.Pending(share)
	}
	return e, nil
}

// Clawback marks a pending/available entry clawed_back (network reject / fraud).
func (l *Ledger) Clawback(ctx context.Context, conversionID int64) error {
	e, ok, err := l.Store.GetByConversion(ctx, conversionID)
	if err != nil || !ok {
		return err
	}
	if e.Status == StatusPaid || e.Status == StatusClawedBack {
		return nil
	}
	if err := l.Store.MarkClawedBack(ctx, conversionID); err != nil {
		return err
	}
	if l.Metrics != nil {
		l.Metrics.Clawback(e.UserShare)
	}
	return nil
}
