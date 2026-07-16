package price

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

func countProducts(t *testing.T, r *Repo) int {
	var c int
	err := r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM tracked_product`).Scan(&c)
	require.NoError(t, err)
	return c
}

func setupRepoWithPlatform(t *testing.T) (*Repo, int16) {
	ctx := context.Background()
	dsn := "postgres://postgres:postgres@localhost:5432/shopass_test?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skip("Postgres not available, skipping repo tests")
	}

	if err := pool.Ping(ctx); err != nil {
		t.Skip("Postgres ping failed, skipping repo tests")
	}

	_, err = pool.Exec(ctx, `
		DROP TABLE IF EXISTS tracked_product CASCADE;
		DROP TABLE IF EXISTS platform CASCADE;

		CREATE TABLE platform (id SMALLSERIAL PRIMARY KEY);
		CREATE TABLE tracked_product (
			id               BIGSERIAL   PRIMARY KEY,
			platform_id      SMALLINT    NOT NULL REFERENCES platform(id),
			platform_item_id TEXT        NOT NULL,
			shop_id          TEXT,
			title            TEXT,
			category_id      BIGINT,
			canonical_key    TEXT,
			first_seen       TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (platform_id, platform_item_id)
		);
		CREATE INDEX idx_tp_canonical ON tracked_product (canonical_key);
	`)
	require.NoError(t, err)

	var platid int16
	err = pool.QueryRow(ctx, `INSERT INTO platform DEFAULT VALUES RETURNING id`).Scan(&platid)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	return NewRepo(pool), platid
}

func TestUpsert_New(t *testing.T) {
	r, plat := setupRepoWithPlatform(t)
	ctx := context.Background()
	p := TrackedProduct{PlatformID: plat, PlatformItemID: "i-123", Title: ptr("Tai nghe X")}
	out, err := r.Upsert(ctx, p)
	require.NoError(t, err)
	require.Greater(t, out.ID, int64(0))
	require.False(t, out.FirstSeen.IsZero())
	require.Equal(t, 1, countProducts(t, r))
}

func TestUpsert_Conflict_Idempotent(t *testing.T) {
	r, plat := setupRepoWithPlatform(t)
	ctx := context.Background()
	a, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-123", Title: ptr("Tên cũ")})
	b, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-123", Title: ptr("Tên mới")})
	require.Equal(t, a.ID, b.ID)              // cùng một dòng
	require.Equal(t, "Tên mới", *b.Title)     // metadata đã cập nhật
	require.Equal(t, a.FirstSeen.Unix(), b.FirstSeen.Unix()) // first_seen bất biến
	require.Equal(t, 1, countProducts(t, r))   // không nhân bản
}

func TestUnique_PlatformItem(t *testing.T) {
	r, plat := setupRepoWithPlatform(t)
	ctx := context.Background()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tracked_product (platform_id, platform_item_id) VALUES ($1,$2),($1,$2)`,
		plat, "dup")
	require.Error(t, err) // vi phạm UNIQUE(platform_id, platform_item_id)
}

func TestGetByCanonicalKey(t *testing.T) {
	r, plat := setupRepoWithPlatform(t)
	ctx := context.Background()
	a, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-1"})
	c, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-2"})
	// mô phỏng TASK-PRICE-005 gán cùng canonical_key
	r.pool.Exec(ctx, `UPDATE tracked_product SET canonical_key='k-xyz' WHERE id = ANY($1)`,
		[]int64{a.ID, c.ID})
	rows, err := r.GetByCanonicalKey(ctx, "k-xyz")
	require.NoError(t, err)
	require.Len(t, rows, 2)

	_, err = r.GetByCanonicalKey(ctx, "")
	require.Error(t, err) // chuỗi rỗng -> lỗi tham số
}

func TestCanonicalKey_NullOnInsert(t *testing.T) {
	r, plat := setupRepoWithPlatform(t)
	ctx := context.Background()
	out, _ := r.Upsert(ctx, TrackedProduct{PlatformID: plat, PlatformItemID: "i-9"})
	require.Nil(t, out.CanonicalKey) // TASK này KHÔNG điền canonical_key
}

func TestGetByIDAndFindByPlatformItem(t *testing.T) {
	r, plat := setupRepoWithPlatform(t)
	ctx := context.Background()
	p := TrackedProduct{PlatformID: plat, PlatformItemID: "i-get", Title: ptr("test")}
	out, err := r.Upsert(ctx, p)
	require.NoError(t, err)

	outID, err := r.GetByID(ctx, out.ID)
	require.NoError(t, err)
	require.Equal(t, out.ID, outID.ID)

	outPI, err := r.FindByPlatformItem(ctx, plat, "i-get")
	require.NoError(t, err)
	require.Equal(t, out.ID, outPI.ID)
}
