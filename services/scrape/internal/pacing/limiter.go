package pacing

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type Limiter struct {
	minDelay map[int16]time.Duration // per platform_id
	maxDelay map[int16]time.Duration
	last     map[int16]time.Time
	mu       sync.Mutex
}

func NewLimiter(minDelay, maxDelay map[int16]time.Duration) *Limiter {
	return &Limiter{
		minDelay: minDelay,
		maxDelay: maxDelay,
		last:     make(map[int16]time.Time),
	}
}

// Wait chặn tới khi đủ pacing cho platform: delay ngẫu nhiên [min,max] + jitter.
func (l *Limiter) Wait(ctx context.Context, platformID int16) error {
	l.mu.Lock()
	minD := l.minDelay[platformID]
	maxD := l.maxDelay[platformID]
	if minD == 0 {
		minD = 50 * time.Millisecond // default safety
	}
	if maxD < minD {
		maxD = minD
	}

	target := minD
	if maxD > minD {
		target = minD + time.Duration(rand.Int63n(int64(maxD-minD)))
	}

	elapsed := time.Since(l.last[platformID])
	
	var wait time.Duration
	if elapsed < target {
		wait = target - elapsed
	}
	
	l.last[platformID] = time.Now().Add(wait)
	l.mu.Unlock()

	if wait > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return nil
}
