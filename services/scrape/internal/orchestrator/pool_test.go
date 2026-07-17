package orchestrator

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type dummyPriceRepo struct {
	written bool
}

func (d *dummyPriceRepo) InsertSnapshot(ctx context.Context, snap PriceSnapshot) (bool, error) {
	return d.written, nil
}

type dummyQueue struct{}

func (q *dummyQueue) Enqueue(ctx context.Context, job ScrapeJob) error { return nil }
func (q *dummyQueue) Claim(ctx context.Context, platformID int16) (ScrapeJob, bool, error) {
	return ScrapeJob{}, false, nil
}
func (q *dummyQueue) Ack(ctx context.Context, productID int64) error { return nil }
func (q *dummyQueue) Retry(ctx context.Context, job ScrapeJob, nextRunAt time.Time) error {
	return nil
}
func (q *dummyQueue) Fail(ctx context.Context, job ScrapeJob) error { return nil }
func (q *dummyQueue) Reclaim(ctx context.Context, platformID int16, timeout time.Duration) (ScrapeJob, bool, error) {
	return ScrapeJob{}, false, nil
}

type blockingAdapter struct {
	fn func()
}

func (b *blockingAdapter) Fetch(ctx context.Context, job ScrapeJob) (PriceSnapshot, error) {
	b.fn()
	return PriceSnapshot{}, nil
}
func (b *blockingAdapter) PlatformID() int16 { return 1 }

func TestPool_ConcurrencyCapPerPlatform(t *testing.T) {
	cfg := Config{MaxConcurrency: map[int16]int{1: 2}}
	p := NewPool(cfg, &dummyPriceRepo{}, &dummyQueue{})

	var peak int32
	var current int32

	p.RegisterAdapter(&blockingAdapter{
		fn: func() {
			n := atomic.AddInt32(&current, 1)
			if n > atomic.LoadInt32(&peak) {
				atomic.StoreInt32(&peak, n)
			}
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt32(&current, -1)
		},
	})

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = p.ProcessJob(ctx, ScrapeJob{PlatformID: 1, ProductID: int64(id)})
		}(i)
	}

	wg.Wait()

	require.LessOrEqual(t, atomic.LoadInt32(&peak), int32(2)) // không vượt cap
}

type errorAdapter struct{}

func (e *errorAdapter) Fetch(ctx context.Context, job ScrapeJob) (PriceSnapshot, error) {
	return PriceSnapshot{}, errors.New("fail")
}
func (e *errorAdapter) PlatformID() int16 { return 1 }

type retryCall struct {
	job       ScrapeJob
	nextRunAt time.Time
}

type recordingQueue struct {
	acked    []int64
	retried  []retryCall
	failed   []ScrapeJob
	retryErr error
	failErr  error
}

func (q *recordingQueue) Enqueue(context.Context, ScrapeJob) error { return nil }
func (q *recordingQueue) Claim(context.Context, int16) (ScrapeJob, bool, error) {
	return ScrapeJob{}, false, nil
}
func (q *recordingQueue) Ack(_ context.Context, productID int64) error {
	q.acked = append(q.acked, productID)
	return nil
}
func (q *recordingQueue) Retry(_ context.Context, job ScrapeJob, nextRunAt time.Time) error {
	if q.retryErr != nil {
		return q.retryErr
	}
	q.retried = append(q.retried, retryCall{job: job, nextRunAt: nextRunAt})
	return nil
}
func (q *recordingQueue) Fail(_ context.Context, job ScrapeJob) error {
	if q.failErr != nil {
		return q.failErr
	}
	q.failed = append(q.failed, job)
	return nil
}
func (q *recordingQueue) Reclaim(context.Context, int16, time.Duration) (ScrapeJob, bool, error) {
	return ScrapeJob{}, false, nil
}

func TestPool_RetryDefersWithoutAck(t *testing.T) {
	fixedNow := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	q := &recordingQueue{}
	p := NewPool(Config{MaxConcurrency: map[int16]int{1: 1}, MaxAttempts: 3, BackoffBaseMs: 200}, &dummyPriceRepo{}, q)
	p.now = func() time.Time { return fixedNow }
	p.RegisterAdapter(&errorAdapter{})

	result, err := p.ProcessJob(context.Background(), ScrapeJob{ProductID: 1, PlatformID: 1, Attempts: 1})
	require.NoError(t, err)
	require.Equal(t, JobDeferred, result.Outcome)
	require.Equal(t, 1, result.Attempts)
	require.EqualError(t, result.Cause, "fail")
	require.Equal(t, fixedNow.Add(200*time.Millisecond), result.RetryAt)
	require.Len(t, q.retried, 1)
	require.Equal(t, 1, q.retried[0].job.Attempts)
	require.Equal(t, result.RetryAt, q.retried[0].nextRunAt)
	require.Empty(t, q.acked)
	require.Empty(t, q.failed)
}

func TestPool_RetryBecomesTerminalFailureWithoutAck(t *testing.T) {
	q := &recordingQueue{}
	p := NewPool(Config{MaxConcurrency: map[int16]int{1: 1}, MaxAttempts: 3, BackoffBaseMs: 200}, &dummyPriceRepo{}, q)
	p.RegisterAdapter(&errorAdapter{})

	result, err := p.ProcessJob(context.Background(), ScrapeJob{ProductID: 1, PlatformID: 1, Attempts: 3})
	require.NoError(t, err)
	require.Equal(t, JobFailed, result.Outcome)
	require.Equal(t, 3, result.Attempts)
	require.EqualError(t, result.Cause, "fail")
	require.Len(t, q.failed, 1)
	require.Equal(t, 3, q.failed[0].Attempts)
	require.Empty(t, q.acked)
	require.Empty(t, q.retried)
}

func TestPool_RetryBackoffIsExponentialAndBounded(t *testing.T) {
	p := NewPool(Config{MaxAttempts: 3, BackoffBaseMs: 200}, &dummyPriceRepo{}, &dummyQueue{})
	require.Equal(t, 200*time.Millisecond, p.retryBackoff(1))
	require.Equal(t, 400*time.Millisecond, p.retryBackoff(2))
	require.Equal(t, 800*time.Millisecond, p.retryBackoff(3))
	// scheduleRetry never retries this far, but clamp defensively protects a
	// malformed job record from turning into an unbounded duration.
	require.Equal(t, 800*time.Millisecond, p.retryBackoff(100))
}

func TestPool_RetryPersistenceErrorIsReturned(t *testing.T) {
	q := &recordingQueue{retryErr: errors.New("database unavailable")}
	p := NewPool(Config{MaxAttempts: 3, BackoffBaseMs: 200}, &dummyPriceRepo{}, q)
	p.RegisterAdapter(&errorAdapter{})

	result, err := p.ProcessJob(context.Background(), ScrapeJob{ProductID: 1, PlatformID: 1, Attempts: 1})
	require.Error(t, err)
	require.Empty(t, result.Outcome)
	require.Empty(t, q.acked)
}
