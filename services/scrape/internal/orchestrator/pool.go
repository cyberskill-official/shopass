package orchestrator

import (
	"context"
	"fmt"
	"time"
)

const (
	defaultMaxAttempts = 5
	defaultBackoffBase = 500 * time.Millisecond
	maxRetryBackoff    = 24 * time.Hour
)

type Config struct {
	MaxConcurrency map[int16]int
	MaxAttempts    int
	BackoffBaseMs  int
}

type Pool struct {
	cfg      Config
	adapters map[int16]PlatformAdapter
	price    PriceRepo
	queue    Queue
	resched  Rescheduler
	inflight map[int16]chan struct{}
	now      func() time.Time
}

// SetRescheduler enables persisted tier-based rescheduling on successful scrapes.
func (p *Pool) SetRescheduler(r Rescheduler) { p.resched = r }

func NewPool(cfg Config, price PriceRepo, q Queue) *Pool {
	p := &Pool{
		cfg:      cfg,
		adapters: make(map[int16]PlatformAdapter),
		price:    price,
		queue:    q,
		inflight: make(map[int16]chan struct{}),
		now:      time.Now,
	}
	for pid, cap := range cfg.MaxConcurrency {
		p.inflight[pid] = make(chan struct{}, cap)
	}
	return p
}

func (p *Pool) RegisterAdapter(a PlatformAdapter) {
	p.adapters[a.PlatformID()] = a
}

// ProcessJob persists one of three outcomes: success (and Ack), a deferred
// retry, or terminal failure. A non-nil error means that outcome itself could
// not be persisted; callers must not log it as a successfully handled job.
func (p *Pool) ProcessJob(ctx context.Context, job ScrapeJob) (ProcessResult, error) {
	sem := p.inflight[job.PlatformID]
	if sem != nil {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		case <-ctx.Done():
			return ProcessResult{}, ctx.Err()
		}
	}

	result, err := p.runOne(ctx, job)
	if err != nil {
		return result, err
	}
	if result.Outcome != JobSucceeded {
		return result, nil
	}
	if err := p.queue.Ack(ctx, job.ProductID); err != nil {
		return result, fmt.Errorf("ack scrape job %d: %w", job.ProductID, err)
	}
	return result, nil
}

func (p *Pool) runOne(ctx context.Context, job ScrapeJob) (ProcessResult, error) {
	a := p.adapters[job.PlatformID]
	if a == nil {
		return p.scheduleRetry(ctx, job, fmt.Errorf("no adapter registered for platform %d", job.PlatformID))
	}
	snap, err := a.Fetch(ctx, job)
	if err != nil {
		return p.scheduleRetry(ctx, job, err) // tăng attempts + backoff
	}
	written, err := p.price.InsertSnapshot(ctx, snap) // delta-only (TASK-PRICE-002)
	if err != nil {
		return p.scheduleRetry(ctx, job, err)
	}
	nextTier := ReTier(job.Tier, written, snap.FlashSale)
	if err := p.commit(ctx, job.ProductID, nextTier, NextRunAt(nextTier, p.now())); err != nil {
		return p.scheduleRetry(ctx, job, fmt.Errorf("reschedule scrape job: %w", err))
	}
	return ProcessResult{Outcome: JobSucceeded, Attempts: job.Attempts}, nil
}

func (p *Pool) scheduleRetry(ctx context.Context, job ScrapeJob, cause error) (ProcessResult, error) {
	attempts := job.Attempts
	if attempts < 1 {
		attempts = 1
	}
	job.Attempts = attempts
	result := ProcessResult{Attempts: attempts, Cause: cause}

	if attempts >= p.maxAttempts() {
		if err := p.queue.Fail(ctx, job); err != nil {
			return result, fmt.Errorf("mark scrape job %d failed: %w", job.ProductID, err)
		}
		result.Outcome = JobFailed
		return result, nil
	}

	nextRunAt := p.now().Add(p.retryBackoff(attempts))
	if err := p.queue.Retry(ctx, job, nextRunAt); err != nil {
		return result, fmt.Errorf("defer scrape job %d: %w", job.ProductID, err)
	}
	result.Outcome = JobDeferred
	result.RetryAt = nextRunAt
	return result, nil
}

func (p *Pool) maxAttempts() int {
	if p.cfg.MaxAttempts > 0 {
		return p.cfg.MaxAttempts
	}
	return defaultMaxAttempts
}

// retryBackoff is exponential and capped both by MaxAttempts and a practical
// upper bound, so malformed configuration cannot produce an overflowing delay.
func (p *Pool) retryBackoff(attempts int) time.Duration {
	base := defaultBackoffBase
	if p.cfg.BackoffBaseMs > 0 {
		if p.cfg.BackoffBaseMs > int(maxRetryBackoff/time.Millisecond) {
			return maxRetryBackoff
		}
		base = time.Duration(p.cfg.BackoffBaseMs) * time.Millisecond
	}
	if base > maxRetryBackoff {
		return maxRetryBackoff
	}
	if attempts < 1 {
		attempts = 1
	}
	if maxAttempts := p.maxAttempts(); attempts > maxAttempts {
		attempts = maxAttempts
	}

	backoff := base
	for i := 1; i < attempts; i++ {
		if backoff >= maxRetryBackoff/2 {
			return maxRetryBackoff
		}
		backoff *= 2
	}
	return backoff
}

func (p *Pool) commit(ctx context.Context, productID int64, tier Tier, nextRun time.Time) error {
	if p.resched != nil {
		return p.resched.Reschedule(ctx, productID, tier, nextRun)
	}
	return nil
}
