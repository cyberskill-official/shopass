package voucher

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func setupVoucherDB(t *testing.T) (*pgxpool.Pool, context.Context) {
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
		DROP TABLE IF EXISTS voucher_catalog CASCADE;
		DROP TABLE IF EXISTS platform CASCADE;

		CREATE TABLE platform (id SMALLSERIAL PRIMARY KEY);
		CREATE TABLE voucher_catalog (
			id             BIGSERIAL   PRIMARY KEY,
			platform_id    SMALLINT    NOT NULL REFERENCES platform(id),
			code           TEXT        NOT NULL,
			type           TEXT        NOT NULL CHECK (type IN ('shop','platform','freeship')),
			discount_type  TEXT        NOT NULL CHECK (discount_type IN ('amount','percent')),
			discount_value BIGINT      NOT NULL CHECK (discount_value > 0 AND (discount_type <> 'percent' OR discount_value <= 100)),
			min_spend      BIGINT      CHECK (min_spend IS NULL OR min_spend >= 0),
			cap            BIGINT      CHECK (cap IS NULL OR cap > 0),
			shop_id        TEXT,
			valid_from     TIMESTAMPTZ NOT NULL,
			valid_to       TIMESTAMPTZ NOT NULL CHECK (valid_to >= valid_from),
			stack_group    TEXT,
			CONSTRAINT shop_id_by_type CHECK ((type = 'shop' AND shop_id IS NOT NULL) OR (type <> 'shop' AND shop_id IS NULL)),
			UNIQUE (platform_id, code)
		);
	`)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `INSERT INTO platform (id) VALUES (1) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})
	return pool, ctx
}

func insert(t *testing.T, r *Repo, v Voucher) {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO voucher_catalog (platform_id, code, type, discount_type, discount_value, min_spend, cap, shop_id, valid_from, valid_to, stack_group)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		v.PlatformID, v.Code, v.Type, v.DiscountType, v.DiscountValue, v.MinSpend, v.Cap, v.ShopID, v.ValidFrom, v.ValidTo, v.StackGroup)
	require.NoError(t, err)
}

func ptr(v int64) *int64    { return &v }
func ptrStr(v string) *string { return &v }

func shopVoucher(pid int16, code string, shop string, now time.Time) Voucher {
	return Voucher{
		PlatformID: pid, Code: code, Type: TypeShop, DiscountType: DiscountAmount, DiscountValue: 10000,
		ShopID: &shop, ValidFrom: now.Add(-time.Hour), ValidTo: now.Add(time.Hour),
	}
}

func codesOf(vs []Voucher) []string {
	var c []string
	for _, v := range vs {
		c = append(c, v.Code)
	}
	return c
}

func TestListActive_FiltersByWindow(t *testing.T) {
	pool, ctx := setupVoucherDB(t)
	r := NewRepo(pool)
	now := time.Now()
	insert(t, r, Voucher{PlatformID: 1, Code: "EXPIRED", Type: TypePlatform, DiscountType: DiscountAmount,
		DiscountValue: 50_000, ValidFrom: now.AddDate(0, 0, -10), ValidTo: now.AddDate(0, 0, -1)})
	insert(t, r, Voucher{PlatformID: 1, Code: "ACTIVE", Type: TypePlatform, DiscountType: DiscountAmount,
		DiscountValue: 50_000, ValidFrom: now.AddDate(0, 0, -1), ValidTo: now.AddDate(0, 0, 5)})
	
	out, err := r.ListActive(ctx, 1, nil, now)
	require.NoError(t, err)
	codes := codesOf(out)
	require.Contains(t, codes, "ACTIVE")
	require.NotContains(t, codes, "EXPIRED")
}

func TestListActive_ShopVoucherScopedToCartShops(t *testing.T) {
	pool, ctx := setupVoucherDB(t)
	r := NewRepo(pool)
	now := time.Now()
	insert(t, r, shopVoucher(1, "SHOPA", "shopA", now))
	insert(t, r, shopVoucher(1, "SHOPB", "shopB", now))
	
	out, err := r.ListActive(ctx, 1, []string{"shopA"}, now)
	require.NoError(t, err)
	codes := codesOf(out)
	require.Contains(t, codes, "SHOPA")
	require.NotContains(t, codes, "SHOPB")
}

func TestBigintRoundTrip(t *testing.T) {
	pool, ctx := setupVoucherDB(t)
	r := NewRepo(pool)
	t0 := time.Now()
	insert(t, r, Voucher{PlatformID: 1, Code: "CAP", Type: TypePlatform, DiscountType: DiscountAmount,
		DiscountValue: 70_000, Cap: ptr(int64(70_000)), MinSpend: ptr(int64(500_000)),
		ValidFrom: t0, ValidTo: t0.AddDate(0, 1, 0)})
	
	out, err := r.ListActive(ctx, 1, nil, t0.AddDate(0, 0, 1))
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(70_000), *out[0].Cap)
	require.Equal(t, int64(500_000), *out[0].MinSpend)
}
