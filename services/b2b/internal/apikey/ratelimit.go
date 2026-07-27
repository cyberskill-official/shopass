package apikey

import (
	"sync"
	"time"
)

// RateLimiter is a per-key token bucket intended for gateway enforcement (DEC-B2B-32).
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[int64]*bucket
	now     func() time.Time
}

type bucket struct {
	tokens     float64
	capacity   float64
	refillPerS float64
	last       time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[int64]*bucket),
		now:     time.Now,
	}
}

func (rl *RateLimiter) Allow(keyID int64, ratePerMin int) (ok bool, retryAfter time.Duration) {
	if ratePerMin <= 0 {
		return false, time.Second
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := rl.now()
	b, okb := rl.buckets[keyID]
	if !okb || int(b.capacity) != ratePerMin {
		b = &bucket{
			tokens:     float64(ratePerMin),
			capacity:   float64(ratePerMin),
			refillPerS: float64(ratePerMin) / 60.0,
			last:       now,
		}
		rl.buckets[keyID] = b
	}
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.refillPerS
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.last = now
	}
	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	need := 1 - b.tokens
	sec := need / b.refillPerS
	if sec < 0.001 {
		sec = 0.001
	}
	return false, time.Duration(sec * float64(time.Second))
}

// AllowedEndpoint gates public routes by tier (DEC-B2B-31).
func AllowedEndpoint(tier, endpoint string) bool {
	switch tier {
	case "enterprise":
		return true
	case "pro", "free":
		return endpoint == "/public/v1/trends"
	default:
		return false
	}
}
