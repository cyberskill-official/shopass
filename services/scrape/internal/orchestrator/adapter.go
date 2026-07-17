package orchestrator

import (
	"context"
	"time"
)

type PriceSnapshot struct {
	ProductID int64
	TS        time.Time
	Price     int64
	ListPrice *int64
	Stock     *int32
	Sold      *int32
	FlashSale bool
}

type ScrapeJob struct {
	ProductID      int64
	PlatformID     int16
	PlatformItemID string // platform-specific ref, e.g. "itemID:shopID" for Shopee
	Tier           Tier
	Attempts       int
	// NextRunAt is populated for deferred in-memory jobs. The durable queue keeps
	// the source of truth in scrape_job.next_run_at.
	NextRunAt time.Time
}

// JobOutcome describes a job whose resulting queue state was persisted. A
// processing error is represented by Deferred or Failed rather than by the
// error return from Pool.ProcessJob.
type JobOutcome string

const (
	JobSucceeded JobOutcome = "succeeded"
	JobDeferred  JobOutcome = "deferred"
	JobFailed    JobOutcome = "failed"
)

// ProcessResult separates a persisted scrape outcome from an infrastructure
// error. When ProcessJob returns a non-nil error, callers must assume the
// queue state could not be persisted and should not report the job as handled.
type ProcessResult struct {
	Outcome  JobOutcome
	Attempts int
	RetryAt  time.Time
	Cause    error
}

type PlatformAdapter interface {
	// Fetch lấy snapshot giá cho một job. Lỗi mạng/parse trả error;
	// orchestrator quyết định retry/backoff dựa trên error.
	Fetch(ctx context.Context, job ScrapeJob) (PriceSnapshot, error)
	PlatformID() int16
}

// PriceRepo is a minimal interface to the Price service to avoid circular dependencies
// In real app, this might be a gRPC client or a direct DB repo if in a monolith
type PriceRepo interface {
	InsertSnapshot(ctx context.Context, snap PriceSnapshot) (bool, error)
}

// Rescheduler persists the next scrape time and tier after a scrape. It is
// optional: a Pool without one treats commit as a no-op (in-memory mode).
type Rescheduler interface {
	Reschedule(ctx context.Context, productID int64, tier Tier, nextRunAt time.Time) error
}
