package bill

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool    *pgxpool.Pool
	metrics MetricsClient
}

func NewRepo(pool *pgxpool.Pool, metrics MetricsClient) *Repo {
	return &Repo{
		pool:    pool,
		metrics: metrics,
	}
}

// CreateSubscription tạo subscription active; lỗi nếu user đã có active
func (r *Repo) CreateSubscription(ctx context.Context, userID int64, planID int16, renewsAt time.Time) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO subscription (user_id, plan_id, renews_at)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		userID, planID, renewsAt).Scan(&id)
	return id, err
}

func (r *Repo) GetActive(ctx context.Context, userID int64) (Subscription, bool, error) {
	var sub Subscription
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, plan_id, started_at, renews_at, status
		 FROM subscription
		 WHERE user_id = $1 AND status = 'active'`, userID).Scan(
		&sub.ID, &sub.UserID, &sub.PlanID, &sub.StartedAt, &sub.RenewsAt, &sub.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return sub, false, nil
	}
	return sub, err == nil, err
}

func (r *Repo) SetRenewsAt(ctx context.Context, subID int64, t time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE subscription SET renews_at=$1 WHERE id=$2`, t, subID)
	return err
}

func (r *Repo) statusOf(ctx context.Context, subID int64) (string, error) {
	var status string
	err := r.pool.QueryRow(ctx, `SELECT status FROM subscription WHERE id=$1`, subID).Scan(&status)
	return status, err
}
