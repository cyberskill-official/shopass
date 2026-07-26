// Package payout re-exports TRUST-005 payout delay / hold APIs for sibling services
// (TASK-AFFIL-005 HoldChecker.Blocked). Same facade pattern as services/trust/fraud.
package payout

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	internal "shopass/services/trust/internal/payout"
)

type Hold = internal.Hold
type Config = internal.Config
type Store = internal.Store
type Service = internal.Service
type Guard = internal.Guard
type Conversion = internal.Conversion
type ConfirmInput = internal.ConfirmInput
type RiskReader = internal.RiskReader
type DueHold = internal.DueHold
type DueStore = internal.DueStore
type NetworkConfirmReader = internal.NetworkConfirmReader
type Releaser = internal.Releaser
type Payer = internal.Payer
type PGStore = internal.PGStore
type PGNetworkConfirm = internal.PGNetworkConfirm

type PGDB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func DefaultConfig() Config {
	return internal.DefaultConfig()
}

func NewPGStore(db PGDB) *PGStore {
	return internal.NewPGStore(db)
}

func NewPGNetworkConfirm(db PGDB) *PGNetworkConfirm {
	return internal.NewPGNetworkConfirm(db)
}

// CreateHoldAdapter adapts Service to affil's error-only hold hook.
type CreateHoldAdapter struct {
	Svc *Service
}

func (a CreateHoldAdapter) OnConversionConfirmed(ctx context.Context, conversionID, beneficiaryID, amount int64) error {
	if a.Svc == nil {
		return nil
	}
	_, err := a.Svc.OnConversionConfirmed(ctx, conversionID, beneficiaryID, amount)
	return err
}

// HoldChecker adapts Service.Blocked for cashback.HoldChecker.
type HoldChecker struct {
	Svc *Service
}

func (h HoldChecker) Blocked(ctx context.Context, conversionID int64) (bool, error) {
	if h.Svc == nil {
		return false, nil
	}
	return h.Svc.Blocked(ctx, conversionID)
}

// TryRelease is exported for jobs that release trust-side holds.
func TryRelease(s *Service, ctx context.Context, conversionID int64, now time.Time) (bool, error) {
	if s == nil {
		return false, nil
	}
	return s.TryRelease(ctx, conversionID, now)
}
