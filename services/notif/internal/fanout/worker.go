package fanout

import (
	"context"
	"fmt"
	"time"

	"shopass/services/notif/internal/notif"
)

// NotifRepo defines the database methods needed by the worker.
type NotifRepo interface {
	ClaimPending(ctx context.Context, id int64, lease time.Duration) (notif.Notification, bool, error)
	MarkSent(ctx context.Context, id int64) error
	Requeue(ctx context.Context, id int64, backoff time.Duration, lastErr string) error
	PublishDLQ(ctx context.Context, n notif.Notification, reason, msg string) error
}

// Worker consumes notification IDs, claims them idempotently, and dispatches them.
type Worker struct {
	repo        NotifRepo
	router      *Router
	lease       time.Duration
	base        time.Duration
	cap         time.Duration
	maxAttempts int
}

// NewWorker creates a new fanout worker.
func NewWorker(repo NotifRepo, router *Router, lease, base, cap time.Duration, maxAttempts int) *Worker {
	return &Worker{
		repo:        repo,
		router:      router,
		lease:       lease,
		base:        base,
		cap:         cap,
		maxAttempts: maxAttempts,
	}
}

// Handle processes a single notification ID. It is safe for concurrent use.
func (w *Worker) Handle(ctx context.Context, id int64) error {
	// 1. Claim/lease (CAS) to ensure at-least-once with idempotency
	n, claimed, err := w.repo.ClaimPending(ctx, id, w.lease)
	if err != nil {
		return fmt.Errorf("claim failed: %w", err)
	}
	if !claimed {
		// Another worker claimed it, or it was already processed
		// metrics.DoubleClaim() could be recorded here
		return nil
	}

	// 2. Route based on the channel
	d, ok := w.router.Route(n.Channel)
	if !ok {
		return w.dead(ctx, n, "permanent", "no dispatcher for channel: "+n.Channel)
	}

	// 3. Dispatch
	class, derr := d.Dispatch(ctx, n)
	switch class {
	case ClassOK:
		return w.repo.MarkSent(ctx, n.ID)
	case ClassPermanent:
		return w.dead(ctx, n, "permanent", errStr(derr))
	default: // ClassTransient
		if n.Attempts >= w.maxAttempts {
			return w.dead(ctx, n, "max_attempts", errStr(derr))
		}
		delay := NextDelay(n.Attempts, w.base, w.cap)
		return w.repo.Requeue(ctx, n.ID, delay, errStr(derr))
	}
}

func (w *Worker) dead(ctx context.Context, n notif.Notification, reason, msg string) error {
	// metrics.DLQ(n.Channel, reason)
	return w.repo.PublishDLQ(ctx, n, reason, msg)
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
