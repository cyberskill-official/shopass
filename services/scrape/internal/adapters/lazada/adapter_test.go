package lazada

import (
	"context"
	"errors"
	"testing"

	"shopass/services/scrape/internal/orchestrator"
)

// fakeFarm records the job it was handed and returns a canned result, standing
// in for the real Playwright farm (TASK-SCRAPE-003) in unit tests.
type fakeFarm struct {
	snap orchestrator.PriceSnapshot
	err  error
	got  orchestrator.ScrapeJob
}

func (f *fakeFarm) RenderPrice(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error) {
	f.got = job
	return f.snap, f.err
}

func TestLazadaAdapter_PlatformID(t *testing.T) {
	if a := NewLazadaAdapter(nil); a.PlatformID() != PLATFORM_LAZADA {
		t.Fatalf("platform id = %d, want %d", a.PlatformID(), PLATFORM_LAZADA)
	}
}

func TestLazadaAdapter_DispatchesToFarm(t *testing.T) {
	farm := &fakeFarm{snap: orchestrator.PriceSnapshot{Price: 199000, FlashSale: true}}
	a := NewLazadaAdapter(farm)
	job := orchestrator.ScrapeJob{
		ProductID:      42,
		PlatformID:     PLATFORM_LAZADA,
		PlatformItemID: "laz-1",
		Tier:           orchestrator.TierHot,
	}

	snap, err := a.Fetch(context.Background(), job)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if farm.got.PlatformItemID != "laz-1" {
		t.Errorf("farm was not handed the job: got item %q", farm.got.PlatformItemID)
	}
	if snap.Price != 199000 || !snap.FlashSale {
		t.Errorf("farm snapshot not passed through: %+v", snap)
	}
	if snap.ProductID != 42 {
		t.Errorf("productID = %d, want 42 (stamped from the job)", snap.ProductID)
	}
}

func TestLazadaAdapter_NoFarmIsError(t *testing.T) {
	a := NewLazadaAdapter(nil)
	if _, err := a.Fetch(context.Background(), orchestrator.ScrapeJob{ProductID: 1}); err == nil {
		t.Fatal("expected an error when no farm is configured, got nil (must not fabricate a price)")
	}
}

func TestLazadaAdapter_FarmErrorPropagates(t *testing.T) {
	farm := &fakeFarm{err: errors.New("akamai challenge")}
	a := NewLazadaAdapter(farm)
	if _, err := a.Fetch(context.Background(), orchestrator.ScrapeJob{ProductID: 1}); err == nil {
		t.Fatal("expected the farm error to propagate")
	}
}
