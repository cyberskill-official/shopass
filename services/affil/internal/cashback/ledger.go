package cashback

import (
	"context"
	"fmt"
)

type Entry struct {
	ConversionID int64
	UserID       int64
	Commission   int64
	UserShare    int64
	KeptMargin   int64
	Status       string
}

type Store interface {
	InsertHeld(ctx context.Context, e Entry) error
	GetByConversion(ctx context.Context, conversionID int64) (Entry, bool, error)
	MarkReleased(ctx context.Context, conversionID int64) error
	MarkClawedBack(ctx context.Context, conversionID int64) error
	SumReleasedUnpaid(ctx context.Context, userID int64) (int64, error)
	CreatePayoutRequest(ctx context.Context, userID, amount int64) error
}

type HoldChecker interface {
	// Blocked returns true when TRUST-005 still holds / investigates the conversion.
	Blocked(ctx context.Context, conversionID int64) (bool, error)
}

type Config struct {
	ShareRateBPS    int64 // basis points, e.g. 5000 = 50%
	PayoutThreshold int64
}

func DefaultConfig() Config {
	return Config{ShareRateBPS: 5000, PayoutThreshold: 50_000}
}

type Ledger struct {
	Cfg   Config
	Store Store
	Hold  HoldChecker
}

func Split(commission, shareRateBPS int64) (userShare, kept int64) {
	userShare = commission * shareRateBPS / 10_000
	kept = commission - userShare
	return userShare, kept
}

// OnConfirmed records a held cashback entry (BIGINT VND). Idempotent per conversion.
func (l *Ledger) OnConfirmed(ctx context.Context, conversionID, userID, commission int64) (Entry, error) {
	if existing, ok, err := l.Store.GetByConversion(ctx, conversionID); err != nil {
		return Entry{}, err
	} else if ok {
		return existing, nil
	}
	share, kept := Split(commission, l.Cfg.ShareRateBPS)
	e := Entry{
		ConversionID: conversionID,
		UserID:       userID,
		Commission:   commission,
		UserShare:    share,
		KeptMargin:   kept,
		Status:       "held",
	}
	if err := l.Store.InsertHeld(ctx, e); err != nil {
		return Entry{}, err
	}
	return e, nil
}

func (l *Ledger) TryRelease(ctx context.Context, conversionID int64) error {
	if l.Hold != nil {
		blocked, err := l.Hold.Blocked(ctx, conversionID)
		if err != nil {
			return err
		}
		if blocked {
			return fmt.Errorf("cashback: still held by trust")
		}
	}
	return l.Store.MarkReleased(ctx, conversionID)
}

func (l *Ledger) Clawback(ctx context.Context, conversionID int64) error {
	return l.Store.MarkClawedBack(ctx, conversionID)
}

func (l *Ledger) MaybeRequestPayout(ctx context.Context, userID int64) (bool, error) {
	sum, err := l.Store.SumReleasedUnpaid(ctx, userID)
	if err != nil || sum < l.Cfg.PayoutThreshold {
		return false, err
	}
	if err := l.Store.CreatePayoutRequest(ctx, userID, sum); err != nil {
		return false, err
	}
	return true, nil
}
