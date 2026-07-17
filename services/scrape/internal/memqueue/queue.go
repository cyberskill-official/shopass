package memqueue

import (
	"context"
	"sync"
	"time"

	"shopass/services/scrape/internal/orchestrator"
)

// Queue is a minimal in-memory orchestrator.Queue for dev and one-shot runs.
// Production uses a durable queue (Redis Streams / DB) per TASK-SCRAPE-001.
type Queue struct {
	mu     sync.Mutex
	jobs   map[int16][]orchestrator.ScrapeJob
	failed map[int64]orchestrator.ScrapeJob
}

func New() *Queue {
	return &Queue{
		jobs:   make(map[int16][]orchestrator.ScrapeJob),
		failed: make(map[int64]orchestrator.ScrapeJob),
	}
}

func (q *Queue) Enqueue(_ context.Context, job orchestrator.ScrapeJob) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if job.NextRunAt.IsZero() {
		job.NextRunAt = time.Now()
	}
	// Explicitly enqueueing a job starts a fresh processing cycle.
	job.Attempts = 0
	q.removeQueuedLocked(job.ProductID)
	delete(q.failed, job.ProductID)
	q.jobs[job.PlatformID] = append(q.jobs[job.PlatformID], job)
	return nil
}

func (q *Queue) Claim(_ context.Context, platformID int16) (orchestrator.ScrapeJob, bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	l := q.jobs[platformID]
	now := time.Now()
	for i, job := range l {
		if job.NextRunAt.After(now) {
			continue
		}
		q.jobs[platformID] = append(l[:i:i], l[i+1:]...)
		job.Attempts++
		return job, true, nil
	}
	return orchestrator.ScrapeJob{}, false, nil
}

func (q *Queue) Ack(_ context.Context, _ int64) error { return nil }

func (q *Queue) Retry(_ context.Context, job orchestrator.ScrapeJob, nextRunAt time.Time) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	job.NextRunAt = nextRunAt
	q.removeQueuedLocked(job.ProductID)
	delete(q.failed, job.ProductID)
	q.jobs[job.PlatformID] = append(q.jobs[job.PlatformID], job)
	return nil
}

func (q *Queue) Fail(_ context.Context, job orchestrator.ScrapeJob) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.removeQueuedLocked(job.ProductID)
	q.failed[job.ProductID] = job
	return nil
}

func (q *Queue) Reclaim(_ context.Context, _ int16, _ time.Duration) (orchestrator.ScrapeJob, bool, error) {
	return orchestrator.ScrapeJob{}, false, nil
}

func (q *Queue) removeQueuedLocked(productID int64) {
	for platformID, jobs := range q.jobs {
		kept := jobs[:0]
		for _, queued := range jobs {
			if queued.ProductID != productID {
				kept = append(kept, queued)
			}
		}
		q.jobs[platformID] = kept
	}
}
