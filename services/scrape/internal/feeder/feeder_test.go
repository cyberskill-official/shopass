package feeder

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestSyncJobs_RegistersTrackedProducts(t *testing.T) {
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()
	_, err = pool.Exec(ctx, `
		DROP TABLE IF EXISTS scrape_job;
		DROP TABLE IF EXISTS tracked_product;
		DROP TABLE IF EXISTS platform;
		DROP TYPE  IF EXISTS scrape_tier;
		CREATE TYPE scrape_tier AS ENUM ('hot','warm','cold');
		CREATE TABLE platform (id SMALLINT PRIMARY KEY);
		CREATE TABLE tracked_product (id BIGINT PRIMARY KEY, platform_id SMALLINT NOT NULL, platform_item_id TEXT NOT NULL);
		CREATE TABLE scrape_job (
			product_id BIGINT PRIMARY KEY REFERENCES tracked_product(id),
			platform_id SMALLINT NOT NULL REFERENCES platform(id),
			tier scrape_tier NOT NULL DEFAULT 'cold',
			next_run_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			attempts INTEGER NOT NULL DEFAULT 0,
			last_status TEXT NOT NULL DEFAULT 'pending',
			locked_until TIMESTAMPTZ);
		INSERT INTO platform (id) VALUES (1);
		INSERT INTO tracked_product (id, platform_id, platform_item_id) VALUES (100,1,'555:777'),(101,1,'888:999');
	`)
	require.NoError(t, err)

	n, err := SyncJobs(ctx, pool)
	require.NoError(t, err)
	require.EqualValues(t, 2, n) // both tracked products registered

	var cnt int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM scrape_job").Scan(&cnt))
	require.Equal(t, 2, cnt)

	n2, err := SyncJobs(ctx, pool)
	require.NoError(t, err)
	require.EqualValues(t, 0, n2) // idempotent
}
