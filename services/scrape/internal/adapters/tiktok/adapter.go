package tiktok

import (
	"context"
	"fmt"
	"time"

	"shopass/services/scrape/internal/orchestrator"
	"github.com/playwright-community/playwright-go"
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
	fmt.Printf("[TikTokAdapter] dispatching job %d using local Playwright (tier %s)...\n", job.ProductID, job.Tier)

	pw, err := playwright.Run()
	if err != nil {
		return orchestrator.PriceSnapshot{}, fmt.Errorf("could not start playwright: %w", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch()
	if err != nil {
		return orchestrator.PriceSnapshot{}, fmt.Errorf("could not launch browser: %w", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return orchestrator.PriceSnapshot{}, fmt.Errorf("could not create page: %w", err)
	}

	// This is just a conceptual script logic for tiktok products.
	// Normally we would navigate to tiktok.com/product/job.ProductID
	targetURL := fmt.Sprintf("https://www.tiktok.com/t/%d", job.ProductID)
	if _, err := page.Goto(targetURL); err != nil {
		return orchestrator.PriceSnapshot{}, fmt.Errorf("could not goto: %w", err)
	}

	return orchestrator.PriceSnapshot{
		ProductID: job.ProductID,
		TS:        time.Now(),
		Price:     99000, // Conceptually we would parse the DOM: page.Locator(".price").TextContent()
		FlashSale: true,
	}, nil
}
