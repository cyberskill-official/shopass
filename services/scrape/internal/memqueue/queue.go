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
	mu   sync.Mutex
	jobs map[int16][]orchestrator.ScrapeJob
}

func New() *Queue { return &Queue{jobs: make(map[int16][]orchestrator.ScrapeJob)} }

func (q *Queue) Enqueue(_ context.Context, job orchestrator.ScrapeJob) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs[job.PlatformID] = append(q.jobs[job.PlatformID], job)
	return nil
}

func (q *Queue) Claim(_ context.Context, platformID int16) (orchestrator.ScrapeJob, bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	l := q.jobs[platformID]
	if len(l) == 0 {
		return orchestrator.ScrapeJob{}, false, nil
	}
	job := l[0]
	q.jobs[platformID] = l[1:]
	return job, true, nil
}

func (q *Queue) Ack(_ context.Context, _ int64) error { return nil }

func (q *Queue) Reclaim(_ context.Context, _ int16, _ time.Duration) (orchestrator.ScrapeJob, bool, error) {
	return orchestrator.ScrapeJob{}, false, nil
}
