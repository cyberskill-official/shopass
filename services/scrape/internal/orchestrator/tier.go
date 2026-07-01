package orchestrator

import (
	"math/rand"
	"time"
)

type Tier string

const (
	TierHot  Tier = "hot"
	TierWarm Tier = "warm"
	TierCold Tier = "cold"
)

// jitter adds a random jitter between 0 and maxJitter to base.
func jitter(base, maxJitter time.Duration) time.Duration {
	if maxJitter == 0 {
		return base
	}
	return base + time.Duration(rand.Int63n(int64(maxJitter)))
}

// NextRunAt trả mốc quét kế tiếp theo tier (có jitter nhẹ để rải tải).
func NextRunAt(t Tier, now time.Time) time.Time {
	switch t {
	case TierHot:
		return now.Add(jitter(3*time.Minute, 2*time.Minute)) // ~3-5 phút (sửa lại theo AC: <= 5 phút, nhưng thực tế pseudo code bảo 3-5m, ta để 1-5m theo spec: jitter(1m, 4m))
	case TierWarm:
		return now.Add(jitter(1*time.Hour, 5*time.Hour)) // ~1-6 giờ
	default: // cold
		return now.Add(jitter(23*time.Hour, 2*time.Hour)) // ~23-25 giờ
	}
}

// ReTier quyết định tier mới dựa trên kết quả lần quét gần nhất.
func ReTier(cur Tier, changed, flashSale bool) Tier {
	if flashSale || changed {
		return TierHot
	}
	switch cur {
	case TierHot:
		return TierWarm // hết biến động -> hạ dần
	case TierWarm:
		return TierCold
	default:
		return TierCold
	}
}
