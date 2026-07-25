package tiktok

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"shopass/services/scrape/internal/orchestrator"
)

// PLATFORM_TIKTOK maps to platform_id = 2
const PLATFORM_TIKTOK int16 = 2

// EnvStubPrice enables the demo-only stub price path when set to "1".
// Production MUST leave this unset so Fetch fails closed instead of writing
// fabricated prices into price_snapshot.
const EnvStubPrice = "SCRAPE_TIKTOK_STUB_PRICE"

// ErrDOMPriceNotImplemented is returned when stub prices are disabled (default).
var ErrDOMPriceNotImplemented = fmt.Errorf(
	"tiktok: DOM price extraction is not implemented; set %s=1 only for demos/tests",
	EnvStubPrice,
)

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

// stubPriceEnabled is true only when SCRAPE_TIKTOK_STUB_PRICE=1 (explicit opt-in).
func stubPriceEnabled() bool {
	return strings.TrimSpace(os.Getenv(EnvStubPrice)) == "1"
}

func (a *TikTokAdapter) Fetch(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error) {
	// TikTok Shop is a content-commerce SPA (TASK-SCRAPE-004): render the product
	// page in a browser, then read the price from the DOM. DOM parsing is not
	// implemented in this Go adapter yet.
	targetURL := fmt.Sprintf("https://www.tiktok.com/t/%d", job.ProductID)
	if a.renderer != nil {
		if err := a.renderer.Render(ctx, targetURL); err != nil {
			return orchestrator.PriceSnapshot{}, err
		}
	}

	// Fail closed by default: never write fabricated prices into price_snapshot.
	// Opt in with SCRAPE_TIKTOK_STUB_PRICE=1 for demos/tests only.
	if !stubPriceEnabled() {
		return orchestrator.PriceSnapshot{}, ErrDOMPriceNotImplemented
	}

	return orchestrator.PriceSnapshot{
		ProductID: job.ProductID,
		TS:        time.Now(),
		Price:     99000,
		FlashSale: true,
	}, nil
}
