package fcm

import (
	"math/rand"
	"time"
)

const maxBackoff = 5 * time.Minute

// nextDelay computes backoff with jitter; respects FCM Retry-After when present (§1 #5).
func nextDelay(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter // respect FCM's directive (DEC-NOTIF-12)
	}
	base := time.Duration(1<<uint(attempt)) * time.Second // 1s,2s,4s,8s...
	if base > maxBackoff {
		base = maxBackoff
	}
	// ±50% jitter to avoid thundering herd at 00:00 flash sale peak
	jitter := time.Duration(rand.Int63n(int64(base / 2)))
	return base/2 + jitter
}
