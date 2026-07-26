package fraud

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGDB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PGEventCounter struct {
	db PGDB
}

func NewPGEventCounter(db PGDB) *PGEventCounter {
	return &PGEventCounter{db: db}
}

func (c *PGEventCounter) CountRedeems(ctx context.Context, userID int64, windowMinutes int) (int, error) {
	if c == nil || c.db == nil {
		return 0, nil
	}
	if windowMinutes <= 0 {
		windowMinutes = 1
	}
	var count int
	err := c.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM referral_code rc
		JOIN app_user referee ON referee.referral_code_id = rc.id
		WHERE rc.user_id = $1
		  AND referee.created_at >= now() - ($2::int * interval '1 minute')
	`, userID, windowMinutes).Scan(&count)
	return count, err
}

type PGClusterSizer struct {
	db PGDB
}

func NewPGClusterSizer(db PGDB) *PGClusterSizer {
	return &PGClusterSizer{db: db}
}

func (s *PGClusterSizer) ClusterSize(ctx context.Context, userID int64) (int, error) {
	if s == nil || s.db == nil {
		return 1, nil
	}
	var size int
	err := s.db.QueryRow(ctx, `
		WITH RECURSIVE reach(uid) AS (
			SELECT $1::bigint
			UNION
			SELECT CASE WHEN e.a_user = reach.uid THEN e.b_user ELSE e.a_user END
			FROM account_link_edge e
			JOIN reach ON reach.uid IN (e.a_user, e.b_user)
		)
		SELECT COUNT(*) FROM reach
	`, userID).Scan(&size)
	return size, err
}

type PGSignalStore struct {
	db PGDB
}

func NewPGSignalStore(db PGDB) *PGSignalStore {
	return &PGSignalStore{db: db}
}

func (s *PGSignalStore) UpsertOpen(ctx context.Context, userID int64, kind string, score int, reasons []Reason) error {
	if s == nil || s.db == nil {
		return nil
	}
	payload, err := json.Marshal(reasons)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO fraud_signal (subject_user_id, kind, risk_score, reasons, status, updated_at)
		VALUES ($1, $2, $3, $4::jsonb, 'open', now())
		ON CONFLICT (subject_user_id, kind) DO UPDATE
		SET risk_score = EXCLUDED.risk_score,
		    reasons = EXCLUDED.reasons,
		    status = CASE
		      WHEN fraud_signal.status IN ('confirmed_fraud', 'cleared') THEN fraud_signal.status
		      ELSE 'open'
		    END,
		    updated_at = now()
	`, userID, kind, score, string(payload))
	return err
}

type PGAccountLinkStore struct {
	db PGDB
}

func NewPGAccountLinkStore(db PGDB) *PGAccountLinkStore {
	return &PGAccountLinkStore{db: db}
}

func (s *PGAccountLinkStore) UpsertReferralEdge(ctx context.Context, referrerID, refereeID int64) error {
	return s.UpsertAccountLinkEdge(ctx, referrerID, refereeID, "referral", 1.0)
}

func (s *PGAccountLinkStore) UpsertAccountLinkEdge(ctx context.Context, userA, userB int64, linkType string, weight float64) error {
	if s == nil || s.db == nil || userA == userB {
		return nil
	}
	if userA > userB {
		userA, userB = userB, userA
	}
	if weight <= 0 {
		weight = 1.0
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO account_link_edge (a_user, b_user, link_type, weight)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (a_user, b_user, link_type) DO UPDATE
		SET weight = GREATEST(account_link_edge.weight, EXCLUDED.weight)
	`, userA, userB, linkType, weight)
	return err
}

type PGRewardHolder struct {
	db         PGDB
	log        *slog.Logger
	holdReason string
}

func NewPGRewardHolder(db PGDB, log *slog.Logger) *PGRewardHolder {
	if log == nil {
		log = slog.Default()
	}
	return &PGRewardHolder{
		db:         db,
		log:        log,
		holdReason: "fraud_risk_score_threshold",
	}
}

func (h *PGRewardHolder) Hold(ctx context.Context, userID int64) error {
	if h == nil || h.db == nil {
		return nil
	}
	tag, err := h.db.Exec(ctx, `
		UPDATE payout_hold
		SET status = 'under_investigation',
		    hold_reason = COALESCE(NULLIF(hold_reason, ''), $2)
		WHERE user_id = $1
		  AND status IN ('held', 'eligible')
	`, userID, h.holdReason)
	if missingPayoutHoldSchema(err) {
		h.log.Info("fraud.reward_hold_requested", "user_id", userID, "reason", h.holdReason, "mode", "log_only_schema_absent")
		return nil
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		h.log.Info("fraud.reward_hold_requested", "user_id", userID, "reason", h.holdReason, "mode", "log_only_no_open_hold")
	}
	return nil
}

func missingPayoutHoldSchema(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "42P01" || pgErr.Code == "42703"
}

var _ EventCounter = (*PGEventCounter)(nil)
var _ ClusterSizer = (*PGClusterSizer)(nil)
var _ SignalStore = (*PGSignalStore)(nil)
var _ RewardHolder = (*PGRewardHolder)(nil)
