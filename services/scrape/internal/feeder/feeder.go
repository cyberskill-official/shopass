// Package feeder keeps the durable scrape queue fed: it registers a scrape_job
// for every tracked_product that has none, so newly tracked products get picked
// up. Re-tiering of existing jobs happens in the orchestrator after each scrape
// (Pool.commit -> Rescheduler); this only handles registration.
package feeder

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncJobs inserts a due 'cold' scrape_job for each tracked_product missing one.
// Returns the number of newly registered products. Idempotent.
func SyncJobs(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	tag, err := pool.Exec(ctx, `
		INSERT INTO scrape_job (product_id, platform_id, tier, next_run_at, last_status)
		SELECT tp.id, tp.platform_id, 'cold'::scrape_tier, now(), 'pending'
		FROM tracked_product tp
		LEFT JOIN scrape_job sj ON sj.product_id = tp.id
		WHERE sj.product_id IS NULL
		ON CONFLICT (product_id) DO NOTHING`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
