package apns

import (
	"math/rand"
	"time"
)

const maxBackoff = 5 * time.Minute

// BackoffDelay computes exponential backoff with jitter and honors Retry-After.
func BackoffDelay(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}
	if attempt < 0 {
		attempt = 0
	}
	base := time.Duration(1<<uint(attempt)) * time.Second
	if base > maxBackoff {
		base = maxBackoff
	}
	jitter := time.Duration(rand.Int63n(int64(base/2) + 1))
	return base/2 + jitter
}
