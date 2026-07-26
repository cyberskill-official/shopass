package notif

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
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
	Attempts    int
	LeaseUntil  *time.Time
	LastError   string
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) InsertNotification(ctx context.Context, n Notification) (int64, error) {
	status := n.Status
	if status == "" {
		status = "pending"
	}
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notification (user_id, channel, template, payload, scheduled_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, n.UserID, n.Channel, n.Template, n.Payload, n.ScheduledAt, status).Scan(&id)
	return id, err
}

// MarkQueued transitions pending → queued so the FCM dispatcher can claim the row.
func (r *Repo) MarkQueued(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notification SET status='queued' WHERE id=$1 AND status='pending'`, id)
	return err
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
		 AND t.platform IN ('android', 'web')
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

// ClaimIOSPushBatch claims queued push rows for APNs (platform=ios).
func (r *Repo) ClaimIOSPushBatch(ctx context.Context, n int) ([]PushJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT n.id, n.user_id, t.address AS token, n.payload
		FROM notification n
		JOIN user_channel_token t
		  ON t.user_id = n.user_id AND t.channel = 'push' AND t.verified = true
		 AND t.platform = 'ios'
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

// ClaimEmailBatch claims queued email notifications with verified email addresses.
func (r *Repo) ClaimEmailBatch(ctx context.Context, n int) ([]PushJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT n.id, n.user_id, t.address AS token, n.payload
		FROM notification n
		JOIN user_channel_token t
		  ON t.user_id = n.user_id AND t.channel = 'email' AND t.verified = true
		WHERE n.channel = 'email' AND n.status = 'queued'
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

// ClaimSMSBatch claims queued SMS notifications with verified phone addresses.
func (r *Repo) ClaimSMSBatch(ctx context.Context, n int) ([]PushJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT n.id, n.user_id, t.address AS token, n.payload
		FROM notification n
		JOIN user_channel_token t
		  ON t.user_id = n.user_id AND t.channel = 'sms' AND t.verified = true
		WHERE n.channel = 'sms' AND n.status = 'queued'
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

// InvalidateEmail marks email channel tokens unverified (hard bounce / complaint).
func (r *Repo) InvalidateEmail(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_channel_token SET verified=false, updated_at=now()
		 WHERE user_id=$1 AND channel='email'`, userID)
	return err
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

// BatchSetScheduledAt writes scheduled_at for multiple notifications in one query
func (r *Repo) BatchSetScheduledAt(ctx context.Context, scheduled map[int64]time.Time) error {
	if len(scheduled) == 0 {
		return nil
	}

	b := &pgx.Batch{}
	for id, t := range scheduled {
		b.Queue("UPDATE notification SET scheduled_at = $1 WHERE id = $2", t, id)
	}

	br := r.pool.SendBatch(ctx, b)
	defer br.Close()

	for range scheduled {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}
