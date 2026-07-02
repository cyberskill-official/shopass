package pgqueue

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"shopass/services/scrape/internal/orchestrator"
)

func setup(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set; skipping pgqueue integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(), `
		DROP TABLE IF EXISTS scrape_job;
		DROP TABLE IF EXISTS tracked_product;
		DROP TABLE IF EXISTS platform;
		DROP TYPE  IF EXISTS scrape_tier;
		CREATE TYPE scrape_tier AS ENUM ('hot','warm','cold');
		CREATE TABLE platform (id SMALLINT PRIMARY KEY);
		CREATE TABLE tracked_product (id BIGINT PRIMARY KEY, platform_item_id TEXT NOT NULL);
		CREATE TABLE scrape_job (
			product_id BIGINT PRIMARY KEY REFERENCES tracked_product(id),
			platform_id SMALLINT NOT NULL REFERENCES platform(id),
			tier scrape_tier NOT NULL DEFAULT 'cold',
			next_run_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			attempts INTEGER NOT NULL DEFAULT 0,
			last_status TEXT NOT NULL DEFAULT 'pending',
			locked_until TIMESTAMPTZ);
		INSERT INTO platform (id) VALUES (1);
		INSERT INTO tracked_product (id, platform_item_id) VALUES (100,'555:777'),(101,'888:999');
	`)
	require.NoError(t, err)
	return pool
}

func job(pid int64) orchestrator.ScrapeJob {
	return orchestrator.ScrapeJob{ProductID: pid, PlatformID: 1, Tier: orchestrator.TierHot}
}

func TestEnqueueClaimAck(t *testing.T) {
	pool := setup(t)
	defer pool.Close()
	q := New(pool, time.Minute)
	ctx := context.Background()

	require.NoError(t, q.Enqueue(ctx, job(100)))

	got, ok, err := q.Claim(ctx, 1)
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 100, got.ProductID)
	require.Equal(t, "555:777", got.PlatformItemID) // joined from tracked_product
	require.EqualValues(t, 1, got.Attempts)

	// leased -> second claim finds nothing
	_, ok2, err := q.Claim(ctx, 1)
	require.NoError(t, err)
	require.False(t, ok2)

	require.NoError(t, q.Ack(ctx, 100))
}

func TestClaimReturnsDistinctThenEmpty(t *testing.T) {
	pool := setup(t)
	defer pool.Close()
	q := New(pool, time.Minute)
	ctx := context.Background()
	require.NoError(t, q.Enqueue(ctx, job(100)))
	require.NoError(t, q.Enqueue(ctx, job(101)))

	seen := map[int64]bool{}
	for i := 0; i < 2; i++ {
		g, ok, err := q.Claim(ctx, 1)
		require.NoError(t, err)
		require.True(t, ok)
		seen[g.ProductID] = true
	}
	require.Len(t, seen, 2) // SKIP LOCKED -> two distinct jobs

	_, ok, err := q.Claim(ctx, 1)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestReclaimExpiredLease(t *testing.T) {
	pool := setup(t)
	defer pool.Close()
	q := New(pool, time.Minute)
	ctx := context.Background()
	require.NoError(t, q.Enqueue(ctx, job(100)))
	_, ok, _ := q.Claim(ctx, 1)
	require.True(t, ok)

	// simulate a crashed worker: lease already expired
	_, err := pool.Exec(ctx, `UPDATE scrape_job SET locked_until = now() - INTERVAL '5 minutes' WHERE product_id=100`)
	require.NoError(t, err)

	got, ok, err := q.Reclaim(ctx, 1, time.Minute)
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 100, got.ProductID)
}
