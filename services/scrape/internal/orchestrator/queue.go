package orchestrator

import (
	"context"
	"time"
)

type Queue interface {
	Enqueue(ctx context.Context, job ScrapeJob) error
	Claim(ctx context.Context, platformID int16) (ScrapeJob, bool, error)
	Ack(ctx context.Context, productID int64) error
	// Retry records a failed attempt and releases the job for a later claim.
	Retry(ctx context.Context, job ScrapeJob, nextRunAt time.Time) error
	// Fail records the terminal state of a job that exhausted its attempts.
	Fail(ctx context.Context, job ScrapeJob) error
	Reclaim(ctx context.Context, platformID int16, timeout time.Duration) (ScrapeJob, bool, error)
}
