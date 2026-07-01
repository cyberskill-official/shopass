package notif

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// ClaimPending claims a single pending notification or reclaims an expired queued lease.
func (r *Repo) ClaimPending(ctx context.Context, id int64, lease time.Duration) (Notification, bool, error) {
	var n Notification
	err := r.pool.QueryRow(ctx, `
		UPDATE notification
		   SET status='queued',
			   attempts=attempts+1,
			   lease_until=now() + $2::interval
		 WHERE id=$1
		   AND (status='pending'
				OR (status='queued' AND lease_until < now()))
		RETURNING id, user_id, channel, template, payload, attempts`,
		id, lease.String()).Scan(&n.ID, &n.UserID, &n.Channel, &n.Template, &n.Payload, &n.Attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return Notification{}, false, nil
	}
	return n, err == nil, err
}

// Requeue puts a notification back into queued state with an extended lease and saves the error.
func (r *Repo) Requeue(ctx context.Context, id int64, backoff time.Duration, lastErr string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notification 
		 SET status='queued',
			 lease_until=now() + $2::interval,
			 last_error=$3
		 WHERE id=$1 AND status='queued'`, id, backoff.String(), lastErr)
	return err
}

// PublishDLQ marks the notification as failed and inserts a record into notification_dlq.
func (r *Repo) PublishDLQ(ctx context.Context, n Notification, reason, msg string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	payloadBytes, _ := json.Marshal(n.Payload)

	_, err = tx.Exec(ctx, `
		INSERT INTO notification_dlq (notification_id, channel, payload, attempts, last_error, reason)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, n.ID, n.Channel, payloadBytes, n.Attempts, msg, reason)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE notification 
		SET status='failed', last_error=$2
		WHERE id=$1 AND status='queued'
	`, n.ID, msg)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
