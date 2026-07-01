package proxy

import (
	"context"
	"testing"
	"time"
)

type mockRepo struct {
	spent int64
}

func (m *mockRepo) SpentMicroUSD(ctx context.Context, day time.Time) (int64, error) {
	return m.spent, nil
}

func TestCostGuard_DowngradeThenBlock(t *testing.T) {
	repo := &mockRepo{}
	g := NewCostGuard(repo, 10_000_000) // 10 USD/ngày
	ctx := context.Background()
	today := time.Now()

	repo.spent = 5_000_000
	d, _ := g.Evaluate(ctx, today)
	if d != Proceed {
		t.Errorf("Expected Proceed, got %v", d)
	}

	repo.spent = 8_500_000 // 85%
	d, _ = g.Evaluate(ctx, today)
	if d != DowngradeTier {
		t.Errorf("Expected DowngradeTier, got %v", d)
	}

	repo.spent = 10_000_000 // chạm trần
	d, _ = g.Evaluate(ctx, today)
	if d != BlockCold {
		t.Errorf("Expected BlockCold, got %v", d)
	}
}

func TestCost_IntegerMicroUSD(t *testing.T) {
	// 1,75 USD/GB x 2GB = 3,50 USD = 3_500_000 micro-USD, chính xác
	got := CostMicroUSD(/*usdPerGBMicro=*/ 1_750_000, /*bytes=*/ 2<<30)
	if got != 3_500_000 {
		t.Errorf("Expected 3_500_000, got %d", got)
	}
}
