package orchestrator

import (
	"context"
)

type JobRepo interface {
	GetDueJobs(ctx context.Context, limit int) ([]ScrapeJob, error)
	MarkFailed(ctx context.Context, productID int64) error
}

type Scheduler struct {
	repo  JobRepo
	queue Queue
}

func NewScheduler(repo JobRepo, queue Queue) *Scheduler {
	return &Scheduler{
		repo:  repo,
		queue: queue,
	}
}

// Tick finds due jobs and enqueues them. In reality this would run on a timer.
func (s *Scheduler) Tick(ctx context.Context) error {
	jobs, err := s.repo.GetDueJobs(ctx, 100)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if err := s.queue.Enqueue(ctx, job); err != nil {
			// Log error, continue with next
			continue
		}
	}
	return nil
}
