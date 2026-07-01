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
