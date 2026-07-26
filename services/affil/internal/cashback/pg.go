package cashback

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGDB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// PGStore persists cashback_entry + payout_request.
type PGStore struct {
	db PGDB
}

func NewPGStore(db PGDB) *PGStore {
	return &PGStore{db: db}
}

func (s *PGStore) InsertPending(ctx context.Context, e Entry) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO cashback_entry
		  (user_id, conversion_id, commission, user_share, kept_margin, status, available_at, created_at)
		VALUES ($1,$2,$3,$4,$5,'pending',$6,$7)
		ON CONFLICT (conversion_id) DO NOTHING
	`, e.UserID, e.ConversionID, e.Commission, e.UserShare, e.KeptMargin, e.AvailableAt, e.CreatedAt)
	return err
}

func (s *PGStore) GetByConversion(ctx context.Context, conversionID int64) (Entry, bool, error) {
	var e Entry
	var paidAt *time.Time
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, conversion_id, commission, user_share, kept_margin, status,
		       available_at, paid_at, created_at
		FROM cashback_entry WHERE conversion_id = $1
	`, conversionID).Scan(
		&e.ID, &e.UserID, &e.ConversionID, &e.Commission, &e.UserShare, &e.KeptMargin, &e.Status,
		&e.AvailableAt, &paidAt, &e.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return Entry{}, false, nil
	}
	if err != nil {
		return Entry{}, false, err
	}
	e.PaidAt = paidAt
	return e, true, nil
}

func (s *PGStore) ListDuePending(ctx context.Context, now time.Time) ([]Entry, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, conversion_id, commission, user_share, kept_margin, status,
		       available_at, paid_at, created_at
		FROM cashback_entry
		WHERE status = 'pending' AND available_at <= $1
		ORDER BY available_at ASC
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		var paidAt *time.Time
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.ConversionID, &e.Commission, &e.UserShare, &e.KeptMargin, &e.Status,
			&e.AvailableAt, &paidAt, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		e.PaidAt = paidAt
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *PGStore) MarkAvailable(ctx context.Context, conversionID int64) error {
	_, err := s.db.Exec(ctx, `
		UPDATE cashback_entry SET status = 'available'
		WHERE conversion_id = $1 AND status = 'pending'
	`, conversionID)
	return err
}

func (s *PGStore) MarkClawedBack(ctx context.Context, conversionID int64) error {
	_, err := s.db.Exec(ctx, `
		UPDATE cashback_entry SET status = 'clawed_back'
		WHERE conversion_id = $1 AND status IN ('pending','available')
	`, conversionID)
	return err
}

func (s *PGStore) SumAvailable(ctx context.Context, userID int64) (int64, error) {
	var sum int64
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(user_share), 0) FROM cashback_entry
		WHERE user_id = $1 AND status = 'available'
	`, userID).Scan(&sum)
	return sum, err
}

func (s *PGStore) ListAvailable(ctx context.Context, userID int64) ([]Entry, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, conversion_id, commission, user_share, kept_margin, status,
		       available_at, paid_at, created_at
		FROM cashback_entry
		WHERE user_id = $1 AND status = 'available'
		ORDER BY available_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		var paidAt *time.Time
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.ConversionID, &e.Commission, &e.UserShare, &e.KeptMargin, &e.Status,
			&e.AvailableAt, &paidAt, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		e.PaidAt = paidAt
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *PGStore) MarkPaid(ctx context.Context, conversionIDs []int64, paidAt time.Time) error {
	for _, id := range conversionIDs {
		if _, err := s.db.Exec(ctx, `
			UPDATE cashback_entry SET status = 'paid', paid_at = $2
			WHERE conversion_id = $1 AND status = 'available'
		`, id, paidAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *PGStore) CreatePayoutRequest(ctx context.Context, userID, amount int64, gatewayRef string) (int64, error) {
	status := "queued"
	switch {
	case strings.HasPrefix(gatewayRef, "failed:"):
		status = "failed"
	case gatewayRef != "":
		status = "sent"
	}
	var id int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO payout_request (user_id, amount, status, gateway_ref)
		VALUES ($1,$2,$3,$4) RETURNING id
	`, userID, amount, status, gatewayRef).Scan(&id)
	return id, err
}

func (s *PGStore) Summary(ctx context.Context, userID int64) (UserSummary, error) {
	var out UserSummary
	err := s.db.QueryRow(ctx, `
		SELECT
		  COUNT(*) FILTER (WHERE status = 'pending'),
		  COALESCE(SUM(user_share) FILTER (WHERE status = 'pending'), 0),
		  MIN(available_at) FILTER (WHERE status = 'pending'),
		  COUNT(*) FILTER (WHERE status = 'available'),
		  COALESCE(SUM(user_share) FILTER (WHERE status = 'available'), 0),
		  COALESCE(SUM(user_share) FILTER (WHERE status = 'paid'), 0)
		FROM cashback_entry WHERE user_id = $1
	`, userID).Scan(
		&out.PendingCount, &out.PendingAmount, &out.NextAvailableAt,
		&out.AvailableCount, &out.AvailableAmount, &out.PaidTotal,
	)
	if err != nil {
		return UserSummary{}, err
	}
	out.Note = DisclosureNote
	return out, nil
}

// EnsurePGStore implements Store.
var _ Store = (*PGStore)(nil)

// ApplyDDL applies the cashback tables (for integration tests).
func ApplyDDL(ctx context.Context, db PGDB) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS cashback_entry (
		  id             BIGSERIAL PRIMARY KEY,
		  user_id        BIGINT NOT NULL,
		  conversion_id  BIGINT NOT NULL UNIQUE,
		  commission     BIGINT NOT NULL CHECK (commission >= 0),
		  user_share     BIGINT NOT NULL CHECK (user_share >= 0),
		  kept_margin    BIGINT NOT NULL CHECK (kept_margin >= 0),
		  status         TEXT NOT NULL DEFAULT 'pending'
		    CHECK (status IN ('pending','available','paid','clawed_back')),
		  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
		  available_at   TIMESTAMPTZ NOT NULL,
		  paid_at        TIMESTAMPTZ
		);
		CREATE TABLE IF NOT EXISTS payout_request (
		  id          BIGSERIAL PRIMARY KEY,
		  user_id     BIGINT NOT NULL,
		  amount      BIGINT NOT NULL CHECK (amount > 0),
		  status      TEXT NOT NULL DEFAULT 'queued'
		    CHECK (status IN ('queued','sent','failed')),
		  gateway_ref TEXT,
		  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		return fmt.Errorf("cashback ddl: %w", err)
	}
	return nil
}
