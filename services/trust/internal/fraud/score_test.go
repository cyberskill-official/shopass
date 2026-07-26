package fraud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubCounter struct{ n int }

func (s stubCounter) CountRedeems(ctx context.Context, userID int64, windowMinutes int) (int, error) {
	return s.n, nil
}

type stubCluster struct{ size int }

func (s stubCluster) ClusterSize(ctx context.Context, userID int64) (int, error) {
	return s.size, nil
}

type memStore struct {
	last Assessment
}

func (m *memStore) UpsertOpen(ctx context.Context, userID int64, kind string, score int, reasons []Reason) error {
	m.last = Assessment{UserID: userID, RiskScore: score, Reasons: reasons}
	return nil
}

type memHolder struct{ held []int64 }

func (m *memHolder) Hold(ctx context.Context, userID int64) error {
	m.held = append(m.held, userID)
	return nil
}

func TestAssess_NormalUser_LowScore(t *testing.T) {
	cfg := DefaultConfig()
	e := NewEngine(cfg, stubCounter{n: 1}, stubCluster{size: 1}, &memStore{}, &memHolder{})
	a, err := e.Assess(context.Background(), 42, nil)
	require.NoError(t, err)
	require.Equal(t, 0, a.RiskScore)
	require.False(t, a.HoldReward)
	require.Empty(t, a.Reasons)
}

func TestAssess_BurstAndCluster_Holds(t *testing.T) {
	cfg := DefaultConfig()
	holder := &memHolder{}
	store := &memStore{}
	e := NewEngine(cfg, stubCounter{n: 50}, stubCluster{size: 12}, store, holder)
	a, err := e.Assess(context.Background(), 7, map[string]any{"self_referral": true})
	require.NoError(t, err)
	require.GreaterOrEqual(t, a.RiskScore, cfg.HoldThreshold)
	require.True(t, a.HoldReward)
	require.Contains(t, holder.held, int64(7))
	require.NotEmpty(t, a.Reasons)
	require.Equal(t, a.RiskScore, store.last.RiskScore)
}

func TestAssess_NeverExceeds100(t *testing.T) {
	cfg := DefaultConfig()
	cfg.VelocityWeight = 80
	cfg.GraphWeight = 80
	cfg.RuleWeight = 80
	e := NewEngine(cfg, stubCounter{n: 100}, stubCluster{size: 100}, &memStore{}, nil)
	a, err := e.Assess(context.Background(), 1, map[string]any{"self_referral": true})
	require.NoError(t, err)
	require.Equal(t, 100, a.RiskScore)
}
