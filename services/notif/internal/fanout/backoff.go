package fanout

import (
	"math/rand"
	"time"
)

// NextDelay calculates the backoff for the n-th attempt using full jitter.
// Wait time = random[0, min(cap, base * 2^(n-1))].
func NextDelay(attempt int, base, cap time.Duration) time.Duration {
	if attempt <= 0 {
		return 0
	}
	shift := attempt - 1
	if shift < 0 {
		shift = 0
	}
	
	exp := base << uint(shift)
	if exp > cap || exp <= 0 {
		exp = cap
	}
	
	return time.Duration(rand.Int63n(int64(exp) + 1))
}
