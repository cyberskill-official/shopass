package tiktok

import (
	"context"
	"fmt"
	"shopass/services/scrape/internal/orchestrator"
	"time"
)

// PLATFORM_TIKTOK maps to platform_id = 2
const PLATFORM_TIKTOK int16 = 2

type TikTokAdapter struct {
	// httpClient, farm config etc.
}

func NewTikTokAdapter() *TikTokAdapter {
	return &TikTokAdapter{}
}

func (a *TikTokAdapter) PlatformID() int16 {
	return PLATFORM_TIKTOK
}

func (a *TikTokAdapter) Fetch(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error) {
	// Proxy enterprise được gán ở tầng orchestrator hoặc worker/farm.
	// Adapter Go này sẽ gọi sang farm (Playwright Node.js) qua HTTP/gRPC.
	// Ở bước này ta mock pass qua để orchestrator integration test chạy.
	fmt.Printf("[TikTokAdapter] dispatching job %d to farm (tier %s)...\n", job.ProductID, job.Tier)

	return orchestrator.PriceSnapshot{
		ProductID: job.ProductID,
		TS:        time.Now(),
		Price:     99000,
		FlashSale: true,
	}, nil
}
