package ecom

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupSeeded() *Service {
	repo := NewRepo()
	return NewService(repo)
}

func setTxCount(s *Service, year int, count int64) {
	s.repo.txCounts[year] = count
}

func seedThresholdVersion(s *Service, key string, version int, count int64) {
	s.repo.thresholds[key] = count
}

func TestThreshold_BelowDoesNotFlag(t *testing.T) {
	s := setupSeeded()
	ctx := context.Background()
	setTxCount(s, 2026, 50_000)
	st, _ := s.Threshold(ctx, 2026)
	require.False(t, st.MustRegister)
}

func TestThreshold_AboveFlagsRegister(t *testing.T) {
	s := setupSeeded()
	ctx := context.Background()
	setTxCount(s, 2026, 120_000)
	st, _ := s.Threshold(ctx, 2026)
	require.True(t, st.MustRegister) // vuot 100.000
}

func TestThreshold_FromVersionedConfig(t *testing.T) {
	s := setupSeeded()
	ctx := context.Background()
	seedThresholdVersion(s, "foreign_platform_yearly_tx", 2, 80_000)
	s.Reload(ctx)
	setTxCount(s, 2026, 90_000)
	st, _ := s.Threshold(ctx, 2026)
	require.True(t, st.MustRegister) // nguong moi 80.000
}
