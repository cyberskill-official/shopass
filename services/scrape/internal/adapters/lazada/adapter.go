package lazada

import (
	"context"
	"fmt"
	"shopass/services/scrape/internal/orchestrator"
	"time"
)

// PLATFORM_LAZADA maps to platform_id = 3
const PLATFORM_LAZADA int16 = 3

type LazadaAdapter struct {
}

func NewLazadaAdapter() *LazadaAdapter {
	return &LazadaAdapter{}
}

func (a *LazadaAdapter) PlatformID() int16 {
	return PLATFORM_LAZADA
}

func (a *LazadaAdapter) Fetch(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error) {
	fmt.Printf("[LazadaAdapter] dispatching job %d to farm (tier %s)...\n", job.ProductID, job.Tier)

	return orchestrator.PriceSnapshot{
		ProductID: job.ProductID,
		TS:        time.Now(),
		Price:     120000,
		FlashSale: true,
	}, nil
}
