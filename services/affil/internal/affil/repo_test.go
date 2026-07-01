package affil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// setupWithUserProduct creates a test database connection, sets up tables,
// and inserts dummy rows for foreign keys.
func setupWithUserProduct(t *testing.T) (*Repo, int64, int64) {
	// In a real environment, this connects to a test DB using testcontainers.
	// For this test, we assume a local DB or we can mock.
	// Assuming an existing test DB is available at POSTGRES_URL or we skip if not set.
	// To comply with the test schema, we will try connecting to localhost postgres.
	
	ctx := context.Background()
	// Use a fallback URL if needed for local testing
	dsn := "postgres://postgres:postgres@localhost:5432/shopass_test?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skip("Postgres not available, skipping repo tests")
	}
	
	// Try to ping
	if err := pool.Ping(ctx); err != nil {
		t.Skip("Postgres ping failed, skipping repo tests")
	}

	// Setup schemas
	_, err = pool.Exec(ctx, `
		DROP TABLE IF EXISTS affiliate_conversion CASCADE;
		DROP TABLE IF EXISTS affiliate_click CASCADE;
		DROP TABLE IF EXISTS tracked_product CASCADE;
		DROP TABLE IF EXISTS platform CASCADE;
		DROP TABLE IF EXISTS app_user CASCADE;

		CREATE TABLE app_user (id BIGSERIAL PRIMARY KEY);
		CREATE TABLE platform (id SMALLSERIAL PRIMARY KEY);
		CREATE TABLE tracked_product (id BIGSERIAL PRIMARY KEY);
		
		CREATE TABLE affiliate_click (
			id          BIGSERIAL   PRIMARY KEY,
			user_id     BIGINT      NOT NULL REFERENCES app_user(id),
			platform_id SMALLINT    NOT NULL REFERENCES platform(id),
			product_id  BIGINT      REFERENCES tracked_product(id),
			sub_id      TEXT        NOT NULL UNIQUE,
			network     TEXT        NOT NULL,
			clicked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE affiliate_conversion (
			id           BIGSERIAL   PRIMARY KEY,
			click_id     BIGINT      NOT NULL UNIQUE REFERENCES affiliate_click(id),
			order_value  BIGINT      NOT NULL CHECK (order_value >= 0),
			commission   BIGINT      NOT NULL CHECK (commission  >= 0),
			status       TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','confirmed','rejected')),
			confirmed_at TIMESTAMPTZ
		);
	`)
	require.NoError(t, err)

	var uid, pid int64
	var platid int16
	err = pool.QueryRow(ctx, `INSERT INTO app_user DEFAULT VALUES RETURNING id`).Scan(&uid)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `INSERT INTO platform DEFAULT VALUES RETURNING id`).Scan(&platid)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `INSERT INTO tracked_product DEFAULT VALUES RETURNING id`).Scan(&pid)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	return NewRepo(pool), uid, pid
}

func statusOf(t *testing.T, r *Repo, cid int64) string {
	var status string
	err := r.pool.QueryRow(context.Background(), `SELECT status FROM affiliate_conversion WHERE id = $1`, cid).Scan(&status)
	require.NoError(t, err)
	return status
}

func countConversions(t *testing.T, r *Repo) int {
	var c int
	err := r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM affiliate_conversion`).Scan(&c)
	require.NoError(t, err)
	return c
}

func TestRecordClick_Insert(t *testing.T) {
	r, uid, pid := setupWithUserProduct(t)
	ctx := context.Background()
	id, err := r.RecordClick(ctx, AffiliateClick{
		UserID: uid, PlatformID: 1, ProductID: &pid,
		SubID: "sd_ab12cd34", Network: "involve_asia"})
	require.NoError(t, err)
	require.Greater(t, id, int64(0))
}

func TestRecordClick_SubIDUnique(t *testing.T) {
	r, uid, _ := setupWithUserProduct(t)
	ctx := context.Background()
	c := AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_dup", Network: "accesstrade"}
	_, err1 := r.RecordClick(ctx, c)
	require.NoError(t, err1)
	_, err2 := r.RecordClick(ctx, c)
	require.Error(t, err2) // UNIQUE(sub_id)
}

func TestConversion_LastClickBySubID(t *testing.T) {
	r, uid, _ := setupWithUserProduct(t)
	ctx := context.Background()
	r.RecordClick(ctx, AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_x", Network: "involve_asia"})
	cid, err := r.RecordConversion(ctx, "sd_x", 250_000, 12_000, "involve_asia")
	require.NoError(t, err)
	require.Greater(t, cid, int64(0))
	require.Equal(t, "pending", statusOf(t, r, cid))
}

func TestConversion_UnknownSubID_NoOrphan(t *testing.T) {
	r, _, _ := setupWithUserProduct(t)
	ctx := context.Background()
	_, err := r.RecordConversion(ctx, "sd_unknown", 100_000, 5_000, "accesstrade")
	require.ErrorIs(t, err, ErrUnknownSubID)
	require.Equal(t, 0, countConversions(t, r))
}

func TestConversion_PostbackIdempotent(t *testing.T) {
	r, uid, _ := setupWithUserProduct(t)
	ctx := context.Background()
	r.RecordClick(ctx, AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_y", Network: "involve_asia"})
	_, err1 := r.RecordConversion(ctx, "sd_y", 250_000, 12_000, "involve_asia")
	require.NoError(t, err1)
	_, err2 := r.RecordConversion(ctx, "sd_y", 250_000, 12_000, "involve_asia")
	require.ErrorIs(t, err2, ErrConversionExists) // một click một conversion
	require.Equal(t, 1, countConversions(t, r))
}

func TestConversion_NegativeMoney_Rejected(t *testing.T) {
	r, uid, _ := setupWithUserProduct(t)
	ctx := context.Background()
	r.RecordClick(ctx, AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_z", Network: "involve_asia"})
	_, err := r.RecordConversion(ctx, "sd_z", -1, 0, "involve_asia")
	require.Error(t, err) // CHECK order_value >= 0
}

func TestConfirm_PendingToConfirmed(t *testing.T) {
	r, uid, _ := setupWithUserProduct(t)
	ctx := context.Background()
	r.RecordClick(ctx, AffiliateClick{UserID: uid, PlatformID: 1, SubID: "sd_c", Network: "involve_asia"})
	cid, _ := r.RecordConversion(ctx, "sd_c", 250_000, 12_000, "involve_asia")
	require.NoError(t, r.ConfirmConversion(ctx, cid))
	require.Equal(t, "confirmed", statusOf(t, r, cid))
}
