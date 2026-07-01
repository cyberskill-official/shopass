package proxy

import (
	"testing"
)

func TestSelectTier_ByDifficulty(t *testing.T) {
	if got := SelectTier(DiffAkamai); got != TierEnterprise {
		t.Errorf("SelectTier(DiffAkamai) = %v, want %v", got, TierEnterprise)
	}
	if got := SelectTier(DiffByteDance); got != TierEnterprise {
		t.Errorf("SelectTier(DiffByteDance) = %v, want %v", got, TierEnterprise)
	}
	if got := SelectTier(DiffShopeeJSON); got != TierBudget {
		t.Errorf("SelectTier(DiffShopeeJSON) = %v, want %v", got, TierBudget)
	}
	if got := SelectTier(DiffUnknown); got != TierMid {
		t.Errorf("SelectTier(DiffUnknown) = %v, want %v", got, TierMid)
	}
}
