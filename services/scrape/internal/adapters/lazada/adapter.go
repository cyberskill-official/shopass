package lazada

import (
	"context"
	"fmt"

	"shopass/services/scrape/internal/orchestrator"
)

// PLATFORM_LAZADA maps to platform.id = 3.
const PLATFORM_LAZADA int16 = 3

// Farm is the Playwright farm interface (TASK-SCRAPE-003). Lazada sits behind
// Akamai, which fingerprints at the TLS/HTTP2 layer before any JS runs, so a
// raw HTTP client is rejected at the handshake (TASK-SCRAPE-008 §1 #2). Every
// Lazada fetch MUST render on the farm and read the embedded JSON / DOM there;
// the real extraction logic lives in services/scrape/farm/src/adapters/lazada.
type Farm interface {
	RenderPrice(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error)
}

// LazadaAdapter implements orchestrator.PlatformAdapter by dispatching every job
// to the Playwright farm (TASK-SCRAPE-008 §1 #1). It holds no HTTP client of its
// own: bypassing the farm's TLS match would be blocked by Akamai.
type LazadaAdapter struct {
	farm Farm
}

// NewLazadaAdapter builds an adapter that renders each job on the given farm.
func NewLazadaAdapter(farm Farm) *LazadaAdapter {
	return &LazadaAdapter{farm: farm}
}

// PlatformID returns the Lazada platform identifier.
func (a *LazadaAdapter) PlatformID() int16 { return PLATFORM_LAZADA }

// Fetch renders the Lazada PDP on the farm and returns the extracted snapshot.
// It never fabricates a price: with no farm configured it returns an error so
// the orchestrator retries or backs off instead of writing a fake value
// (TASK-SCRAPE-008 §1 #7 - challenges surface as errors, not empty snapshots).
func (a *LazadaAdapter) Fetch(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error) {
	if a.farm == nil {
		return orchestrator.PriceSnapshot{}, fmt.Errorf("lazada: no farm configured; Akamai requires farm render (TASK-SCRAPE-008)")
	}
	snap, err := a.farm.RenderPrice(ctx, job)
	if err != nil {
		return orchestrator.PriceSnapshot{}, fmt.Errorf("lazada: farm render for product %d: %w", job.ProductID, err)
	}
	// The farm keys work by URL; stamp the tracked product id the orchestrator asked for.
	snap.ProductID = job.ProductID
	return snap, nil
}
