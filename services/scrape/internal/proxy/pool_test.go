package proxy

import (
	"context"
	"testing"
)

type mockVault struct{}

func (v *mockVault) ProxyCreds(tier Tier) ProxyCreds {
	return ProxyCreds{User: "user", Pass: "pass"}
}

type mockProfile struct {
	country string
}

func (m mockProfile) GetCountry() string {
	return m.country
}

func TestAcquire_GeoMatchesProfile(t *testing.T) {
	guard := NewCostGuard(&mockRepo{}, 10000000)
	p := NewPool(guard, &mockVault{})
	ctx := context.Background()

	sess, err := p.Acquire(ctx, TierBudget, "VN")
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	_, err = BindProfile(sess, mockProfile{country: "VN"})
	if err != nil {
		t.Errorf("BindProfile failed: %v", err)
	}

	_, err = BindProfile(sess, mockProfile{country: "US"})
	if err == nil {
		t.Errorf("Expected error for mismatched country")
	}
}

func TestAcquire_BannedIPCooldown(t *testing.T) {
	guard := NewCostGuard(&mockRepo{}, 10000000)
	p := NewPool(guard, &mockVault{})
	ctx := context.Background()

	s, _ := p.Acquire(ctx, TierBudget, "VN")
	p.MarkBanned(s.IP)
	for i := 0; i < 20; i++ {
		s2, _ := p.Acquire(ctx, TierBudget, "VN")
		if s.IP == s2.IP {
			t.Errorf("IP %s was reacquired despite being banned", s.IP)
		}
	}
}
