package trend

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	d0 = time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	d1 = time.Date(2026, 6, 21, 0, 0, 0, 0, time.UTC)
)

func ptr[T any](v T) *T { return &v }

func TestKAnon_BelowThreshold_Suppressed(t *testing.T) {
	raw := aggRow{CategoryID: 7, PlatformID: 1, Day: d0, SKUCount: 49,
		MedianP: 120_000, P25P: 100_000, P75P: 140_000, AvgDiscountPct: 12.5}
	c := applyKAnon(raw)
	require.True(t, c.Suppressed)
	require.Nil(t, c.MedianP)
	require.Nil(t, c.P25P)
	require.Nil(t, c.P75P)
	require.Nil(t, c.AvgDiscountPct)
	require.EqualValues(t, 49, c.SKUCount)
}

func TestKAnon_AtThreshold_Published(t *testing.T) {
	raw := aggRow{CategoryID: 7, PlatformID: 1, Day: d0, SKUCount: 50,
		MedianP: 120_000, P25P: 100_000, P75P: 140_000, AvgDiscountPct: 12.5}
	c := applyKAnon(raw)
	require.False(t, c.Suppressed)
	require.NotNil(t, c.MedianP)
	require.EqualValues(t, 120_000, *c.MedianP)
	require.EqualValues(t, 12.5, *c.AvgDiscountPct)
}

func TestQueryCells_SkipsSuppressed(t *testing.T) {
	r := NewMemoryRepo()
	ctx := context.Background()
	require.NoError(t, r.UpsertCells(ctx, []MarketTrendCell{
		{CategoryID: 7, PlatformID: 1, Day: d0, SKUCount: 60,
			MedianP: ptr(int64(120_000)), P25P: ptr(int64(100_000)), P75P: ptr(int64(140_000)),
			AvgDiscountPct: ptr(12.5)},
		{CategoryID: 7, PlatformID: 1, Day: d1, SKUCount: 10, Suppressed: true},
	}))
	cells, err := r.QueryCells(ctx, 7, 1, d0, d1.AddDate(0, 0, 1))
	require.NoError(t, err)
	require.Len(t, cells, 1)
	require.Equal(t, d0.Unix(), cells[0].Day.Unix())
}

func TestJob_Idempotent(t *testing.T) {
	r := NewMemoryRepo()
	ctx := context.Background()
	src := PointSource{Points: map[string]map[cellKey][]PricePoint{
		"2026-06-20": {
			{cat: 7, pf: 1}: makePoints(60, 100_000, 200_000),
		},
	}}
	job := &Job{Source: src, Repo: r, Metrics: NewMetrics(), LagDays: 1}
	from, to := d0, d0.AddDate(0, 0, 1) // settled past window (before today)
	require.NoError(t, job.Run(ctx, from, to))
	snap1, err := r.QueryAll(ctx, 7, 1, from, to)
	require.NoError(t, err)
	require.NoError(t, job.Run(ctx, from, to))
	snap2, err := r.QueryAll(ctx, 7, 1, from, to)
	require.NoError(t, err)
	require.Equal(t, len(snap1), len(snap2))
	require.Equal(t, snap1[0].SKUCount, snap2[0].SKUCount)
	require.Equal(t, snap1[0].Suppressed, snap2[0].Suppressed)
	require.Equal(t, *snap1[0].MedianP, *snap2[0].MedianP)
	pub, sup, runs, _ := job.Metrics.Snapshot()
	require.EqualValues(t, 2, runs)
	require.EqualValues(t, 2, pub)
	require.EqualValues(t, 0, sup)
}

func makePoints(n int, minP, maxP int64) []PricePoint {
	out := make([]PricePoint, n)
	span := maxP - minP
	for i := 0; i < n; i++ {
		closeP := minP + span*int64(i)/int64(n-1)
		out[i] = PricePoint{ProductID: int64(i + 1), CloseP: closeP, MaxP: maxP}
	}
	return out
}
