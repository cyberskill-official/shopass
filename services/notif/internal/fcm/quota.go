package fcm

import (
	"sync"
	"time"
)

// Bucket implements a per-minute token bucket for FCM quota (DEC-NOTIF-11).
// Default capacity: 600,000 messages/minute/project.
type Bucket struct {
	mu       sync.Mutex
	tokens   int
	capacity int
	last     time.Time
}

// NewBucket creates a bucket with 600k/minute capacity.
func NewBucket() *Bucket {
	return &Bucket{
		tokens:   600_000,
		capacity: 600_000,
		last:     time.Now(),
	}
}

// NewBucketWithCap creates a bucket with custom capacity (for testing).
func NewBucketWithCap(cap int) *Bucket {
	return &Bucket{
		tokens:   cap,
		capacity: cap,
		last:     time.Now(),
	}
}

// Allow returns false when the per-minute quota is exhausted.
// The bucket refills every minute.
func (b *Bucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if time.Since(b.last) >= time.Minute {
		b.tokens = b.capacity
		b.last = time.Now()
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}
