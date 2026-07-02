package gating

import (
	"context"
	"testing"
)

func TestTriggers_WishlistFull_ShowsCTA(t *testing.T) {
	g, repo, _ := setupGate()
	repo.limits["free:wishlist_items"] = 20
	repo.usage[1] = 20

	sig, err := g.EvaluateTriggers(context.Background(), 1, UsageEvent{FeatureKey: "wishlist_items"})
	if err != nil {
		t.Fatalf("unexpected err")
	}
	if sig == nil || sig.Trigger != "wishlist_limit_reached" {
		t.Fatalf("expected wishlist_limit_reached trigger, got %v", sig)
	}
}

func TestTriggers_NoSignalWhenWithinLimit(t *testing.T) {
	g, repo, _ := setupGate()
	repo.limits["free:wishlist_items"] = 20
	repo.usage[1] = 5

	sig, _ := g.EvaluateTriggers(context.Background(), 1, UsageEvent{FeatureKey: "wishlist_items"})
	if sig != nil {
		t.Fatalf("expected nil trigger, got %v", sig)
	}
}

func TestTriggers_PremiumTouch(t *testing.T) {
	g, repo, _ := setupGate()
	repo.limits["free:bottom_predict"] = 0

	sig, _ := g.EvaluateTriggers(context.Background(), 1, UsageEvent{FeatureKey: "bottom_predict"})
	if sig == nil || sig.Trigger != "premium_feature_touch" {
		t.Fatalf("expected premium touch trigger")
	}
}
