package memqueue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/scrape/internal/orchestrator"
)

func TestRetryKeepsJobDeferredAndCarriesAttempts(t *testing.T) {
	q := New()
	ctx := context.Background()
	job := orchestrator.ScrapeJob{ProductID: 10, PlatformID: 1, Tier: orchestrator.TierHot}
	require.NoError(t, q.Enqueue(ctx, job))

	claimed, ok, err := q.Claim(ctx, 1)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 1, claimed.Attempts)

	retryAt := time.Now().Add(time.Hour)
	require.NoError(t, q.Retry(ctx, claimed, retryAt))
	_, ok, err = q.Claim(ctx, 1)
	require.NoError(t, err)
	require.False(t, ok, "a deferred retry must not be claimed immediately")

	require.Len(t, q.jobs[1], 1)
	require.Equal(t, 1, q.jobs[1][0].Attempts)
	require.Equal(t, retryAt, q.jobs[1][0].NextRunAt)
}

func TestFailRemovesJobFromFutureClaims(t *testing.T) {
	q := New()
	ctx := context.Background()
	job := orchestrator.ScrapeJob{ProductID: 10, PlatformID: 1, Tier: orchestrator.TierHot}
	require.NoError(t, q.Enqueue(ctx, job))

	claimed, ok, err := q.Claim(ctx, 1)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, q.Fail(ctx, claimed))

	_, ok, err = q.Claim(ctx, 1)
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, claimed, q.failed[job.ProductID])
}
