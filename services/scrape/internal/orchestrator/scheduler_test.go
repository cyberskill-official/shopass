package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockJobRepo struct {
	jobs []ScrapeJob
}

func (m *mockJobRepo) GetDueJobs(ctx context.Context, limit int) ([]ScrapeJob, error) {
	return m.jobs, nil
}
func (m *mockJobRepo) MarkFailed(ctx context.Context, productID int64) error {
	return nil
}

type mockQueue struct {
	enqueued []ScrapeJob
}

func (m *mockQueue) Enqueue(ctx context.Context, job ScrapeJob) error {
	m.enqueued = append(m.enqueued, job)
	return nil
}
func (m *mockQueue) Claim(ctx context.Context, platformID int16) (ScrapeJob, bool, error) {
	return ScrapeJob{}, false, nil
}
func (m *mockQueue) Ack(ctx context.Context, productID int64) error { return nil }
func (m *mockQueue) Retry(ctx context.Context, job ScrapeJob, nextRunAt time.Time) error {
	return nil
}
func (m *mockQueue) Fail(ctx context.Context, job ScrapeJob) error { return nil }
func (m *mockQueue) Reclaim(ctx context.Context, platformID int16, timeout time.Duration) (ScrapeJob, bool, error) {
	return ScrapeJob{}, false, nil
}

func TestScheduler_TickEnqueuesDueJobs(t *testing.T) {
	repo := &mockJobRepo{
		jobs: []ScrapeJob{{ProductID: 1}, {ProductID: 2}},
	}
	q := &mockQueue{}
	s := NewScheduler(repo, q)

	err := s.Tick(context.Background())
	require.NoError(t, err)
	require.Len(t, q.enqueued, 2)
}
