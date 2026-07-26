package batch

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotifItem struct {
	UserID    int64          `json:"user_id"`
	ProductID int64          `json:"product_id"`
	Reason    string         `json:"reason"`
	Payload   map[string]any `json:"payload"`
}

type NotifEnqueuer interface {
	Enqueue(ctx context.Context, item NotifItem) error
}

type Batch struct {
	pool  *pgxpool.Pool
	log   *slog.Logger
	notif NotifEnqueuer
}

func New(pool *pgxpool.Pool, log *slog.Logger, notif NotifEnqueuer) *Batch {
	return &Batch{
		pool:  pool,
		log:   log,
		notif: notif,
	}
}
