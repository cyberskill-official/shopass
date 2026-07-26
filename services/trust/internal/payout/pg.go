package payout

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGDB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// PGStore persists payout_hold rows.
type PGStore struct {
	db PGDB
}

func NewPGStore(db PGDB) *PGStore {
	return &PGStore{db: db}
}

func (s *PGStore) InsertHold(ctx context.Context, h Hold) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO payout_hold (conversion_id, user_id, amount, status, hold_reason, confirmed_at, eligible_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7)
		ON CONFLICT (conversion_id) DO NOTHING
	`, h.ConversionID, h.UserID, h.Amount, h.Status, h.HoldReason, h.ConfirmedAt, h.EligibleAt)
	return err
}

func (s *PGStore) GetByConversion(ctx context.Context, conversionID int64) (Hold, bool, error) {
	var h Hold
	var reason *string
	err := s.db.QueryRow(ctx, `
		SELECT conversion_id, user_id, amount, status, hold_reason, eligible_at, confirmed_at
		FROM payout_hold WHERE conversion_id = $1
	`, conversionID).Scan(&h.ConversionID, &h.UserID, &h.Amount, &h.Status, &reason, &h.EligibleAt, &h.ConfirmedAt)
	if err == pgx.ErrNoRows {
		return Hold{}, false, nil
	}
	if err != nil {
		return Hold{}, false, err
	}
	if reason != nil {
		h.HoldReason = *reason
	}
	return h, true, nil
}

func (s *PGStore) MarkReleased(ctx context.Context, conversionID int64) error {
	_, err := s.db.Exec(ctx, `
		UPDATE payout_hold
		SET status = 'released', released_at = now()
		WHERE conversion_id = $1 AND status IN ('held', 'eligible')
	`, conversionID)
	return err
}

func (s *PGStore) ExtendInvestigation(ctx context.Context, conversionID int64, reason string) error {
	return s.MarkUnderInvestigation(ctx, conversionID, reason)
}

func (s *PGStore) MarkUnderInvestigation(ctx context.Context, conversionID int64, reason string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE payout_hold
		SET status = 'under_investigation',
		    hold_reason = COALESCE(NULLIF($2, ''), hold_reason)
		WHERE conversion_id = $1 AND status IN ('held', 'eligible')
	`, conversionID, reason)
	return err
}

func (s *PGStore) ListDue(ctx context.Context, now time.Time) ([]DueHold, error) {
	rows, err := s.db.Query(ctx, `
		SELECT conversion_id, user_id, amount
		FROM payout_hold
		WHERE status = 'held'
		  AND hold_reason IS NULL
		  AND eligible_at <= $1
		ORDER BY eligible_at
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DueHold
	for rows.Next() {
		var h DueHold
		if err := rows.Scan(&h.ConversionID, &h.UserID, &h.Amount); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// PGNetworkConfirm reads affiliate_conversion.status.
type PGNetworkConfirm struct {
	db PGDB
}

func NewPGNetworkConfirm(db PGDB) *PGNetworkConfirm {
	return &PGNetworkConfirm{db: db}
}

func (n *PGNetworkConfirm) NetworkConfirmed(ctx context.Context, conversionID int64) (bool, error) {
	var status string
	err := n.db.QueryRow(ctx, `
		SELECT status FROM affiliate_conversion WHERE id = $1
	`, conversionID).Scan(&status)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == "confirmed", nil
}

var _ Store = (*PGStore)(nil)
var _ DueStore = (*PGStore)(nil)
var _ NetworkConfirmReader = (*PGNetworkConfirm)(nil)
