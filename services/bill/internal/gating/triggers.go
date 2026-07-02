package gating

import "context"

type UsageEvent struct {
	FeatureKey string
}

type UpgradeSignal struct {
	Trigger       string `json:"trigger"`
	SuggestedTier string `json:"suggested_tier"`
}

func (g *Gate) EvaluateTriggers(ctx context.Context, userID int64, ev UsageEvent) (*UpgradeSignal, error) {
	allowed, _ := g.Allow(ctx, userID, ev.FeatureKey)
	if !allowed && ev.FeatureKey == "wishlist_items" {
		return &UpgradeSignal{Trigger: "wishlist_limit_reached", SuggestedTier: "premium_basic"}, nil
	}
	if !allowed && ev.FeatureKey == "bottom_predict" {
		return &UpgradeSignal{Trigger: "premium_feature_touch", SuggestedTier: "premium_basic"}, nil
	}
	return nil, nil
}
