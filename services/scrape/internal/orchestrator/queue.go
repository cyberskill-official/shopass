package orchestrator

import (
	"context"
	"time"
)

type Queue interface {
	Enqueue(ctx context.Context, job ScrapeJob) error
	Claim(ctx context.Context, platformID int16) (ScrapeJob, bool, error)
	Ack(ctx context.Context, productID int64) error
	Reclaim(ctx context.Context, platformID int16, timeout time.Duration) (ScrapeJob, bool, error)
}
