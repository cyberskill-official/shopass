package dsar

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRepo struct {
	db *pgxpool.Pool
}

func NewPgRepo(db *pgxpool.Pool) repo {
	return &pgRepo{db: db}
}

func (r *pgRepo) create(ctx context.Context, userID int64, kind string, slaDue time.Time) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO dsar_request (user_id, kind, sla_due_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, kind, slaDue).Scan(&id)
	return id, err
}

func (r *pgRepo) markCompleted(ctx context.Context, dsarID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE dsar_request
		SET status = 'completed', completed_at = now()
		WHERE id = $1
	`, dsarID)
	return err
}

func (r *pgRepo) overdue(ctx context.Context) ([]DSARRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, kind, status, requested_at, sla_due_at, note
		FROM dsar_request
		WHERE status <> 'completed' AND now() > sla_due_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []DSARRequest
	for rows.Next() {
		var d DSARRequest
		if err := rows.Scan(&d.ID, &d.UserID, &d.Kind, &d.Status, &d.RequestedAt, &d.SLADueAt, &d.Note); err != nil {
			return nil, err
		}
		res = append(res, d)
	}
	return res, rows.Err()
}
