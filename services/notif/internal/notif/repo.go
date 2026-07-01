package notif

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID          int64
	UserID      int64
	Channel     string
	Template    string
	Payload     map[string]any
	ScheduledAt *time.Time
	SentAt      *time.Time
	Status      string
	CreatedAt   time.Time
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) InsertNotification(ctx context.Context, n Notification) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notification (user_id, channel, template, payload, scheduled_at, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id
	`, n.UserID, n.Channel, n.Template, n.Payload, n.ScheduledAt).Scan(&id)
	return id, err
}

func (r *Repo) GetUserChannels(ctx context.Context, userID int64) (UserChannels, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT channel
		FROM user_channel_token
		WHERE user_id = $1 AND verified = true
	`, userID)
	if err != nil {
		return UserChannels{}, err
	}
	defer rows.Close()

	var caps UserChannels
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return UserChannels{}, err
		}
		switch channel {
		case "push":
			caps.Push = true
		case "email":
			caps.Email = true
		case "sms":
			caps.SMS = true
		}
	}
	return caps, rows.Err()
}

// PushJob represents a claimed notification row ready for FCM dispatch.
type PushJob struct {
	NotifID int64
	UserID  int64
	Token   string
	Payload []byte
}

// ClaimPushBatch claims up to n push notifications using FOR UPDATE SKIP LOCKED
// to prevent duplicate sends across concurrent workers (§1 #8).
func (r *Repo) ClaimPushBatch(ctx context.Context, n int) ([]PushJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT n.id, n.user_id, t.address AS token, n.payload
		FROM notification n
		JOIN user_channel_token t
		  ON t.user_id = n.user_id AND t.channel = 'push' AND t.verified = true
		WHERE n.channel = 'push' AND n.status = 'queued'
		ORDER BY n.scheduled_at NULLS FIRST, n.id
		FOR UPDATE OF n SKIP LOCKED
		LIMIT $1`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []PushJob
	for rows.Next() {
		var j PushJob
		if err := rows.Scan(&j.NotifID, &j.UserID, &j.Token, &j.Payload); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// MarkSent transitions a notification from queued→sent (idempotent: only updates if still queued).
func (r *Repo) MarkSent(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notification SET status='sent', sent_at=now()
		 WHERE id=$1 AND status='queued'`, id)
	return err
}

// MarkFailed transitions a notification from queued→failed.
func (r *Repo) MarkFailed(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notification SET status='failed'
		 WHERE id=$1 AND status='queued'`, id)
	return err
}

// InvalidateToken marks a user's push token as unverified (UNREGISTERED/dead token).
func (r *Repo) InvalidateToken(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_channel_token SET verified=false, updated_at=now()
		 WHERE user_id=$1 AND channel='push'`, userID)
	return err
}
