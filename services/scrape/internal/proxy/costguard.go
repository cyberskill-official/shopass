package proxy

import (
	"context"
	"time"
)

type Decision int

const (
	Proceed       Decision = iota // dưới ngân sách
	DowngradeTier                 // hạ tier cho request không tới hạn
	BlockCold                     // dừng tier cold, giữ hot
)

type UsageRepo interface {
	SpentMicroUSD(ctx context.Context, day time.Time) (int64, error)
}

type CostGuard struct {
	repo             UsageRepo
	dailyBudgetMicro int64
}

func NewCostGuard(repo UsageRepo, dailyBudgetMicro int64) *CostGuard {
	return &CostGuard{
		repo:             repo,
		dailyBudgetMicro: dailyBudgetMicro,
	}
}

// Evaluate quyết định dựa trên chi phí ngày so ngân sách (micro-USD số nguyên).
func (g *CostGuard) Evaluate(ctx context.Context, day time.Time) (Decision, error) {
	spent, err := g.repo.SpentMicroUSD(ctx, day)
	if err != nil {
		return Proceed, err
	}
	switch {
	case spent >= g.dailyBudgetMicro:
		return BlockCold, nil // chạm trần -> chỉ giữ hot
	case spent >= g.dailyBudgetMicro*8/10:
		return DowngradeTier, nil // 80% -> hạ tier
	default:
		return Proceed, nil
	}
}

func CostMicroUSD(usdPerGBMicro int64, bytes int64) int64 {
	// (bytes / 1GB) * usdPerGBMicro
	// 1GB = 1 << 30 bytes
	// => (bytes * usdPerGBMicro) / (1 << 30)
	
	// Avoid overflow by dividing first if bytes is large, 
	// but bytes is usually small enough for int64.
	// Actually floating point logic on ints: bytes * usdPerGBMicro / (1<<30)
	return (bytes * usdPerGBMicro) / (1 << 30)
}
