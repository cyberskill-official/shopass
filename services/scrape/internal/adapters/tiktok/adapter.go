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

// Renderer renders a TikTok product page. Production uses Playwright; unit tests
// inject a fake so the adapter runs without launching a real browser.
type Renderer interface {
	Render(ctx context.Context, url string) error
}

type playwrightRenderer struct{}

func (playwrightRenderer) Render(ctx context.Context, url string) error {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %w", err)
	}
	defer pw.Stop()
	browser, err := pw.Chromium.Launch()
	if err != nil {
		return fmt.Errorf("could not launch browser: %w", err)
	}
	defer browser.Close()
	page, err := browser.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}
	if _, err := page.Goto(url); err != nil {
		return fmt.Errorf("could not goto: %w", err)
	}
	return nil
}

type TikTokAdapter struct {
	renderer Renderer
}

// NewTikTokAdapter builds a production adapter backed by Playwright.
func NewTikTokAdapter() *TikTokAdapter {
	return &TikTokAdapter{renderer: playwrightRenderer{}}
}

// NewTikTokAdapterWithRenderer injects a custom renderer (used by unit tests).
func NewTikTokAdapterWithRenderer(r Renderer) *TikTokAdapter {
	return &TikTokAdapter{renderer: r}
}

func (a *TikTokAdapter) PlatformID() int16 {
	return PLATFORM_TIKTOK
}

func (a *TikTokAdapter) Fetch(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error) {
	// TikTok Shop is a content-commerce SPA (TASK-SCRAPE-004): render the product
	// page in a browser, then read the price from the DOM. DOM parsing is a
	// deferred integration step; the rendered price is stubbed for now.
	targetURL := fmt.Sprintf("https://www.tiktok.com/t/%d", job.ProductID)
	if a.renderer != nil {
		if err := a.renderer.Render(ctx, targetURL); err != nil {
			return orchestrator.PriceSnapshot{}, err
		}
	}
	return orchestrator.PriceSnapshot{
		ProductID: job.ProductID,
		TS:        time.Now(),
		Price:     99000,
		FlashSale: true,
	}, nil
}
