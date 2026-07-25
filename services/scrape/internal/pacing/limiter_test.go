package pacing

import (
	"context"
	"testing"
	"time"
)

func TestLimiter_RespectsMinDelay(t *testing.T) {
	minD := map[int16]time.Duration{1: 50 * time.Millisecond}
	maxD := map[int16]time.Duration{1: 120 * time.Millisecond}
	l := NewLimiter(minD, maxD)
	ctx := context.Background()

	t0 := time.Now()
	l.Wait(ctx, 1)
	l.Wait(ctx, 1)
	if time.Since(t0) < 50*time.Millisecond {
		t.Errorf("Limiter was too fast, expected at least 50ms")
	}
}

func TestLimiter_HasJitter(t *testing.T) {
	minD := map[int16]time.Duration{1: 20 * time.Millisecond}
	maxD := map[int16]time.Duration{1: 80 * time.Millisecond}
	l := NewLimiter(minD, maxD)
	ctx := context.Background()

	gaps := make([]time.Duration, 8)
	l.Wait(ctx, 1)
	for i := 0; i < 8; i++ {
		t0 := time.Now()
		l.Wait(ctx, 1)
		gaps[i] = time.Since(t0)
	}

	distinct := 0
	for i := 1; i < 8; i++ {
		if gaps[i] != gaps[0] {
			distinct++
		}
	}

	if distinct == 0 {
		t.Errorf("Expected distinct gaps due to jitter, but all were roughly equal")
	}
}
