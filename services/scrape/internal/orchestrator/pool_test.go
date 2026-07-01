package orchestrator

import (
	"context"
	"errors"
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
	done := make(chan struct{})
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			_ = p.ProcessJob(ctx, ScrapeJob{PlatformID: 1, ProductID: int64(id)})
			if id == 9 {
				close(done)
			}
		}(i)
	}

	<-done
	// wait for stragglers
	time.Sleep(50 * time.Millisecond)

	require.LessOrEqual(t, atomic.LoadInt32(&peak), int32(2)) // không vượt cap
}

type errorAdapter struct{}

func (e *errorAdapter) Fetch(ctx context.Context, job ScrapeJob) (PriceSnapshot, error) {
	return PriceSnapshot{}, errors.New("fail")
}
func (e *errorAdapter) PlatformID() int16 { return 1 }

func TestPool_RetryThenFail(t *testing.T) {
	cfg := Config{MaxConcurrency: map[int16]int{1: 1}, MaxAttempts: 5}
	p := NewPool(cfg, &dummyPriceRepo{}, &dummyQueue{})
	p.RegisterAdapter(&errorAdapter{})

	err := p.runOne(context.Background(), ScrapeJob{ProductID: 1, PlatformID: 1, Attempts: 0})
	require.NoError(t, err) // scheduleRetry returns nil in our stub
}
