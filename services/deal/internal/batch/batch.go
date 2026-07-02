package batch

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotifItem struct {
	UserID    int64
	ProductID int64
	Reason    string
	Payload   map[string]any
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
