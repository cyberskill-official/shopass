package breach

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRepo struct {
	db *pgxpool.Pool
}

func NewPgRepo(db *pgxpool.Pool) repo {
	return &pgRepo{db: db}
}

func (r *pgRepo) create(ctx context.Context, in BreachInput, ack time.Time) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO breach_incident (summary, severity, occurred_at, acknowledged_at, source_ref)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, in.Summary, in.Severity, in.OccurredAt, ack, in.SourceRef).Scan(&id)
	return id, err
}

func (r *pgRepo) get(ctx context.Context, id int64) (BreachIncident, error) {
	var b BreachIncident
	err := r.db.QueryRow(ctx, `
		SELECT id, summary, severity, status, occurred_at, acknowledged_at, triaged_at, notified_authority_at, notified_subjects_at, closed_at, source_ref, created_at
		FROM breach_incident
		WHERE id = $1
	`, id).Scan(&b.ID, &b.Summary, &b.Severity, &b.Status, &b.OccurredAt, &b.AcknowledgedAt, &b.TriagedAt, &b.NotifiedAuthorityAt, &b.NotifiedSubjectsAt, &b.ClosedAt, &b.SourceRef, &b.CreatedAt)
	return b, err
}

func (r *pgRepo) transition(ctx context.Context, id int64, to Status, t time.Time) error {
	col := ""
	switch to {
	case "triaged":
		col = "triaged_at"
	case "notified_authority":
		col = "notified_authority_at"
	case "notified_subjects":
		col = "notified_subjects_at"
	case "closed":
		col = "closed_at"
	default:
		return ErrInvalidTransition
	}

	q := fmt.Sprintf(`UPDATE breach_incident SET status = $1, %s = $2 WHERE id = $3`, col)
	_, err := r.db.Exec(ctx, q, to, t, id)
	return err
}

func (r *pgRepo) overdue(ctx context.Context) ([]BreachIncident, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, summary, severity, status, occurred_at, acknowledged_at, triaged_at, notified_authority_at, notified_subjects_at, closed_at, source_ref, created_at
		FROM breach_incident
		WHERE notified_authority_at IS NULL
		  AND now() > acknowledged_at + INTERVAL '72 hours'
		ORDER BY acknowledged_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []BreachIncident
	for rows.Next() {
		var b BreachIncident
		if err := rows.Scan(&b.ID, &b.Summary, &b.Severity, &b.Status, &b.OccurredAt, &b.AcknowledgedAt, &b.TriagedAt, &b.NotifiedAuthorityAt, &b.NotifiedSubjectsAt, &b.ClosedAt, &b.SourceRef, &b.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, rows.Err()
}
