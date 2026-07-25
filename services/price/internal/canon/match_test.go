package canon

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	dsn := "postgres://postgres:postgres@localhost:5432/shopass_test?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skip("Postgres not available, skipping canon match tests")
	}

	if err := pool.Ping(ctx); err != nil {
		t.Skip("Postgres ping failed, skipping canon match tests")
	}

	// Create extension and tables for test
	_, err = pool.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pg_trgm;
		CREATE TABLE IF NOT EXISTS tracked_product (
			id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			canonical_key TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_tp_title_trgm_test ON tracked_product USING gin (title gin_trgm_ops);
		
		CREATE TABLE IF NOT EXISTS canonical_review_queue (
			id BIGSERIAL PRIMARY KEY,
			product_id BIGINT NOT NULL,
			candidate_key TEXT NOT NULL,
			confidence REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (product_id, candidate_key)
		);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Exec(ctx, `DROP TABLE IF EXISTS canonical_review_queue CASCADE; DROP TABLE IF EXISTS tracked_product CASCADE;`)
		pool.Close()
	})

	return pool
}

func setupWithProduct(t *testing.T, platform, title string) (*Matcher, string) {
	pool := setupTestDB(t)
	queue := NewReviewQueue(pool)
	matcher := NewMatcher(pool, queue)

	// Add initial product
	norm := Normalize(title)
	attrs := Extract(norm)
	key := CanonicalKey(attrs.Brand, attrs.Model, attrs.Salient)

	_, err := pool.Exec(context.Background(), `INSERT INTO tracked_product (title, canonical_key) VALUES ($1, $2)`, norm, key)
	require.NoError(t, err)

	return matcher, key
}

func countPending(t *testing.T, m *Matcher) int {
	var c int
	err := m.pool.QueryRow(context.Background(), `SELECT count(*) FROM canonical_review_queue WHERE status = 'pending'`).Scan(&c)
	require.NoError(t, err)
	return c
}

func TestMatch_SameProductDifferentPlatform_Merges(t *testing.T) {
	m, key := setupWithProduct(t, "shopee", "Tai nghe Sony WH-1000XM5 Chính Hãng")
	ctx := context.Background()

	// Trigram sim between "tai nghe sony wh 1000xm5" and "sony wh 1000xm5 headphone" is high
	res, err := m.Match(ctx, Candidate{
		ProductID: 1,
		Title:     "[LAZMALL] Sony WH 1000XM5 Freeship - Headphone",
		Attrs:     Extract(Normalize("[LAZMALL] Sony WH 1000XM5 Freeship - Headphone")),
	})

	require.NoError(t, err)
	require.Equal(t, "merge", res.Action)
	require.Equal(t, key, res.CanonicalKey)
}

func TestMatch_DifferentProduct_DoesNotMerge(t *testing.T) {
	m, _ := setupWithProduct(t, "shopee", "iPhone 15 128GB")
	ctx := context.Background()

	c := Candidate{
		ProductID: 2,
		Title:     "Samsung Galaxy S24 Ultra 256GB",
	}
	c.Attrs = Extract(Normalize(c.Title))

	res, err := m.Match(ctx, c)
	require.NoError(t, err)
	require.Equal(t, "skip", res.Action)
}

func TestMatch_LowConfidence_GoesToReviewQueue(t *testing.T) {
	m, _ := setupWithProduct(t, "shopee", "iPhone 15 128GB")
	ctx := context.Background()

	// "iphone 15 pro 128gb" vs "iphone 15 128gb" has high trigram similarity but slightly different
	// Let's force similarity to be between 0.60 and 0.82 in the test DB by manually inserting it
	// Actually, pg_trgm similarity("iphone 15 128gb", "iphone 15 pro 128gb") is approx 0.65-0.75.
	// So it naturally falls in the review queue.
	c := Candidate{
		ProductID: 2,
		Title:     "iPhone 15 Pro 128GB",
	}
	c.Attrs = Extract(Normalize(c.Title))

	res, err := m.Match(ctx, c)

	require.NoError(t, err)
	require.Equal(t, "review", res.Action)
	require.Equal(t, 1, countPending(t, m)) // 1 dòng review, KHÔNG auto-merge
}
