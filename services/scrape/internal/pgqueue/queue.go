// Package pgqueue is a durable Postgres-backed work queue (TASK-SCRAPE-001)
// implementing orchestrator.Queue via SELECT ... FOR UPDATE SKIP LOCKED and a
// lease (scrape_job.locked_until). It replaces the in-memory queue in production.
package pgqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/scrape/internal/orchestrator"
)

type Queue struct {
	pool  *pgxpool.Pool
	lease time.Duration
}

func New(pool *pgxpool.Pool, lease time.Duration) *Queue {
	if lease <= 0 {
		lease = 2 * time.Minute
	}
	return &Queue{pool: pool, lease: lease}
}

// Enqueue upserts a due job. platform_item_id lives on tracked_product, so it is
// not stored here; Claim joins it back.
func (q *Queue) Enqueue(ctx context.Context, job orchestrator.ScrapeJob) error {
	nextRunAt := job.NextRunAt
	if nextRunAt.IsZero() {
		nextRunAt = time.Now()
	}
	_, err := q.pool.Exec(ctx, `
		INSERT INTO scrape_job (product_id, platform_id, tier, next_run_at, last_status)
		VALUES ($1, $2, $3::scrape_tier, $4, 'pending')
		ON CONFLICT (product_id) DO UPDATE SET
			platform_id = EXCLUDED.platform_id,
			tier        = EXCLUDED.tier,
			next_run_at = EXCLUDED.next_run_at,
			attempts    = 0,
			last_status = 'pending',
			locked_until = NULL
	`, job.ProductID, job.PlatformID, string(job.Tier), nextRunAt)
	return err
}

const claimSQL = `
UPDATE scrape_job j
SET locked_until = now() + $2::interval, attempts = attempts + 1
WHERE j.product_id = (
	SELECT product_id FROM scrape_job
	WHERE platform_id = $1
	  AND last_status <> 'failed'
	  AND next_run_at <= now()
	  AND (locked_until IS NULL OR locked_until <= now())
	ORDER BY next_run_at
	FOR UPDATE SKIP LOCKED
	LIMIT 1
)
RETURNING j.product_id, j.platform_id, j.tier::text, j.attempts, j.next_run_at,
	COALESCE((SELECT platform_item_id FROM tracked_product WHERE id = j.product_id), '')`

// Claim leases the next due job for a platform. ok=false when none is available.
func (q *Queue) Claim(ctx context.Context, platformID int16) (orchestrator.ScrapeJob, bool, error) {
	return q.scanJob(ctx, claimSQL, platformID, q.lease.String())
}

// Ack marks a job done and pushes next_run_at out so a drain terminates. Finer
// tier-based rescheduling is the orchestrator's commit step.
func (q *Queue) Ack(ctx context.Context, productID int64) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE scrape_job
		SET attempts = 0, last_status = 'ok', locked_until = NULL,
		    next_run_at = GREATEST(next_run_at, now() + INTERVAL '1 minute')
		WHERE product_id = $1`, productID)
	return err
}

// Retry persists a failed attempt and makes the job eligible only at
// nextRunAt. Attempts are carried by the claimed job and are not reset until
// a later successful Ack.
func (q *Queue) Retry(ctx context.Context, job orchestrator.ScrapeJob, nextRunAt time.Time) error {
	tag, err := q.pool.Exec(ctx, `
		UPDATE scrape_job
		SET attempts = GREATEST(attempts, $2),
		    last_status = 'retry',
		    next_run_at = $3,
		    locked_until = NULL
		WHERE product_id = $1`, job.ProductID, job.Attempts, nextRunAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("scrape job %d not found while deferring", job.ProductID)
	}
	return nil
}

// Fail records the terminal queue state after the retry budget is exhausted.
// Failed jobs are excluded by Claim and by the due-job index.
func (q *Queue) Fail(ctx context.Context, job orchestrator.ScrapeJob) error {
	tag, err := q.pool.Exec(ctx, `
		UPDATE scrape_job
		SET attempts = GREATEST(attempts, $2),
		    last_status = 'failed',
		    locked_until = NULL
		WHERE product_id = $1`, job.ProductID, job.Attempts)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("scrape job %d not found while failing", job.ProductID)
	}
	return nil
}

// Reschedule persists the next scrape time + tier (orchestrator.Rescheduler).
func (q *Queue) Reschedule(ctx context.Context, productID int64, tier orchestrator.Tier, nextRunAt time.Time) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE scrape_job
		SET tier = $2::scrape_tier, next_run_at = $3, attempts = 0,
		    last_status = 'ok', locked_until = NULL
		WHERE product_id = $1`, productID, string(tier), nextRunAt)
	return err
}

const reclaimSQL = `
UPDATE scrape_job j
SET locked_until = now() + $2::interval, attempts = attempts + 1
WHERE j.product_id = (
	SELECT product_id FROM scrape_job
	WHERE platform_id = $1
	  AND locked_until IS NOT NULL AND locked_until <= now()
	  AND last_status NOT IN ('failed','ok')
	ORDER BY locked_until
	FOR UPDATE SKIP LOCKED
	LIMIT 1
)
RETURNING j.product_id, j.platform_id, j.tier::text, j.attempts, j.next_run_at,
	COALESCE((SELECT platform_item_id FROM tracked_product WHERE id = j.product_id), '')`

// Reclaim re-leases a job whose lease expired (crashed worker). timeout is the
// new lease duration.
func (q *Queue) Reclaim(ctx context.Context, platformID int16, timeout time.Duration) (orchestrator.ScrapeJob, bool, error) {
	return q.scanJob(ctx, reclaimSQL, platformID, timeout.String())
}

func (q *Queue) scanJob(ctx context.Context, sql string, platformID int16, lease string) (orchestrator.ScrapeJob, bool, error) {
	var job orchestrator.ScrapeJob
	var tier string
	err := q.pool.QueryRow(ctx, sql, platformID, lease).
		Scan(&job.ProductID, &job.PlatformID, &tier, &job.Attempts, &job.NextRunAt, &job.PlatformItemID)
	if err == pgx.ErrNoRows {
		return orchestrator.ScrapeJob{}, false, nil
	}
	if err != nil {
		return orchestrator.ScrapeJob{}, false, err
	}
	job.Tier = orchestrator.Tier(tier)
	return job, true, nil
}
