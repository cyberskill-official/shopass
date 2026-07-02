package tiktok

import (
	"context"
	"shopass/services/scrape/internal/orchestrator"
	"testing"
)

func TestTikTokAdapter_PlatformID(t *testing.T) {
	adapter := NewTikTokAdapter()
	if adapter.PlatformID() != PLATFORM_TIKTOK {
		t.Errorf("Expected platform ID %d, got %d", PLATFORM_TIKTOK, adapter.PlatformID())
	}
}

func TestTikTokAdapter_Fetch(t *testing.T) {
	adapter := NewTikTokAdapter()
	ctx := context.Background()
	job := orchestrator.ScrapeJob{
		ProductID:  123,
		PlatformID: PLATFORM_TIKTOK,
		Tier:       orchestrator.TierHot,
	}

	snap, err := adapter.Fetch(ctx, job)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if snap.ProductID != 123 {
		t.Errorf("Expected product ID 123, got %d", snap.ProductID)
	}
	if snap.Price != 99000 {
		t.Errorf("Expected price 99000, got %d", snap.Price)
	}
}
