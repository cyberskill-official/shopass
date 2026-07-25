package gating

import (
	"context"
	"testing"

	"shopass/services/bill/internal/bill"
)

type mockRepo struct {
	limits map[string]int64
	usage  map[int64]int64
}

func (m *mockRepo) LimitFor(ctx context.Context, tier string, featureKey string) (int64, error) {
	if l, ok := m.limits[tier+":"+featureKey]; ok {
		return l, nil
	}
	return 0, nil
}
func (m *mockRepo) CountUsage(ctx context.Context, userID int64, featureKey string) (int64, error) {
	return m.usage[userID], nil
}

type mockSubs struct {
	subs map[int64]bill.Subscription
	err  error
}

func (m *mockSubs) GetActive(ctx context.Context, userID int64) (bill.Subscription, bool, error) {
	if m.err != nil {
		return bill.Subscription{}, false, m.err
	}
	s, ok := m.subs[userID]
	return s, ok, nil
}

type mockPlans struct{}

func (m mockPlans) TierOf(planID int16) string {
	if planID == 1 {
		return "premium_basic"
	}
	return "free"
}

func setupGate() (*Gate, *mockRepo, *mockSubs) {
	repo := &mockRepo{
		limits: make(map[string]int64),
		usage:  make(map[int64]int64),
	}
	subs := &mockSubs{subs: make(map[int64]bill.Subscription)}
	plans := mockPlans{}
	return NewGate(repo, subs, plans), repo, subs
}

func TestAllow_CoreFeatureFreeUser(t *testing.T) {
	g, repo, _ := setupGate()
	for _, f := range []string{"price_tracking", "fake_sale_detect", "price_chart"} {
		repo.limits["free:"+f] = -1
		ok, err := g.Allow(context.Background(), 1, f)
		if err != nil {
			t.Fatalf("unexpected err")
		}
		if !ok {
			t.Fatalf("core feature %s should be allowed for free user", f)
		}
	}
}

func TestAllow_WishlistLimit_Free(t *testing.T) {
	g, repo, _ := setupGate()
	repo.limits["free:wishlist_items"] = 20
	repo.usage[1] = 20 // reached limit
	ok, err := g.Allow(context.Background(), 1, "wishlist_items")
	if err != ErrLimitReached {
		t.Fatalf("expected ErrLimitReached, got %v", err)
	}
	if ok {
		t.Fatalf("should not be allowed")
	}
}

func TestAllow_WishlistLimit_Premium(t *testing.T) {
	g, repo, subs := setupGate()
	subs.subs[2] = bill.Subscription{PlanID: 1} // premium_basic
	repo.limits["premium_basic:wishlist_items"] = 100
	repo.usage[2] = 50

	ok, _ := g.Allow(context.Background(), 2, "wishlist_items")
	if !ok {
		t.Fatalf("should be allowed for premium user under limit")
	}
}

func TestAllow_PremiumOnlyFeature(t *testing.T) {
	g, repo, subs := setupGate()
	repo.limits["free:bottom_predict"] = 0
	repo.limits["premium_basic:bottom_predict"] = -1
	subs.subs[2] = bill.Subscription{PlanID: 1} // premium

	okFree, _ := g.Allow(context.Background(), 1, "bottom_predict")
	if okFree {
		t.Fatalf("free should not access bottom_predict")
	}

	okPrem, _ := g.Allow(context.Background(), 2, "bottom_predict")
	if !okPrem {
		t.Fatalf("premium should access bottom_predict")
	}
}
