package tiktok

import (
	"context"
	"errors"
	"shopass/services/scrape/internal/orchestrator"
	"testing"
)

func TestTikTokAdapter_PlatformID(t *testing.T) {
	adapter := NewTikTokAdapter()
	if adapter.PlatformID() != PLATFORM_TIKTOK {
		t.Errorf("Expected platform ID %d, got %d", PLATFORM_TIKTOK, adapter.PlatformID())
	}
}

type fakeRenderer struct{}

func (fakeRenderer) Render(ctx context.Context, url string) error { return nil }

func TestTikTokAdapter_Fetch_StubDisabledByDefault(t *testing.T) {
	t.Setenv(EnvStubPrice, "")
	adapter := NewTikTokAdapterWithRenderer(fakeRenderer{})
	ctx := context.Background()
	job := orchestrator.ScrapeJob{
		ProductID:  123,
		PlatformID: PLATFORM_TIKTOK,
		Tier:       orchestrator.TierHot,
	}

	snap, err := adapter.Fetch(ctx, job)
	if !errors.Is(err, ErrDOMPriceNotImplemented) {
		t.Fatalf("Expected ErrDOMPriceNotImplemented, got err=%v snap=%+v", err, snap)
	}
	if snap.Price != 0 {
		t.Errorf("Expected zero price on error, got %d", snap.Price)
	}
}

func TestTikTokAdapter_Fetch_StubRequiresExplicitOne(t *testing.T) {
	// Any value other than "1" must remain fail-closed.
	for _, v := range []string{"", "0", "true", "yes", "TRUE"} {
		t.Run("env="+v, func(t *testing.T) {
			t.Setenv(EnvStubPrice, v)
			adapter := NewTikTokAdapterWithRenderer(fakeRenderer{})
			_, err := adapter.Fetch(context.Background(), orchestrator.ScrapeJob{
				ProductID:  1,
				PlatformID: PLATFORM_TIKTOK,
			})
			if !errors.Is(err, ErrDOMPriceNotImplemented) {
				t.Fatalf("Expected ErrDOMPriceNotImplemented for %s=%q, got %v", EnvStubPrice, v, err)
			}
		})
	}
}

func TestTikTokAdapter_Fetch_StubEnabled(t *testing.T) {
	t.Setenv(EnvStubPrice, "1")
	adapter := NewTikTokAdapterWithRenderer(fakeRenderer{})
	ctx := context.Background()
	job := orchestrator.ScrapeJob{
		ProductID:  123,
		PlatformID: PLATFORM_TIKTOK,
		Tier:       orchestrator.TierHot,
	}

	snap, err := adapter.Fetch(ctx, job)
	if err != nil {
		t.Fatalf("Expected no error with stub enabled, got %v", err)
	}
	if snap.ProductID != 123 {
		t.Errorf("Expected product ID 123, got %d", snap.ProductID)
	}
	if snap.Price != 99000 {
		t.Errorf("Expected price 99000, got %d", snap.Price)
	}
	if !snap.FlashSale {
		t.Error("Expected FlashSale true from stub")
	}
}
