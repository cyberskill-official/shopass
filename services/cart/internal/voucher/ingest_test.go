package voucher

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/cart/internal/cart"
)

func setupIngest(t *testing.T) (*Ingestor, context.Context) {
	pool, ctx := setupVoucherDB(t)
	return NewIngestor(pool), ctx
}

func countVouchers(t *testing.T, i *Ingestor, pid int16, code string) int {
	var c int
	err := i.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM voucher_catalog WHERE platform_id = $1 AND code = $2`, pid, code).Scan(&c)
	require.NoError(t, err)
	return c
}

func TestUpsert_RejectShopVoucherWithoutShopID(t *testing.T) {
	i, ctx := setupIngest(t)
	t0 := time.Now()
	err := i.Upsert(ctx, Voucher{PlatformID: 1, Code: "X", Type: TypeShop, DiscountType: DiscountAmount,
		DiscountValue: 30_000, ValidFrom: t0, ValidTo: t0.AddDate(0, 1, 0)}) // missing ShopID
	require.ErrorIs(t, err, cart.ErrShopVoucherNeedsShopID)
}

func TestUpsert_RejectPercentOver100(t *testing.T) {
	i, ctx := setupIngest(t)
	t0 := time.Now()
	err := i.Upsert(ctx, Voucher{PlatformID: 1, Code: "P", Type: TypePlatform, DiscountType: DiscountPercent,
		DiscountValue: 120, ValidFrom: t0, ValidTo: t0.AddDate(0, 1, 0)})
	require.ErrorIs(t, err, cart.ErrPercentOutOfRange)
}

func TestUpsert_Idempotent(t *testing.T) {
	i, ctx := setupIngest(t)
	t0 := time.Now()
	v := Voucher{PlatformID: 1, Code: "DUP", Type: TypePlatform, DiscountType: DiscountAmount,
		DiscountValue: 50_000, ValidFrom: t0, ValidTo: t0.AddDate(0, 1, 0)}
	require.NoError(t, i.Upsert(ctx, v))
	v.DiscountValue = 60_000
	require.NoError(t, i.Upsert(ctx, v)) // same (platform_id, code) → UPDATE
	require.Equal(t, 1, countVouchers(t, i, 1, "DUP"))

	var dval int64
	err := i.pool.QueryRow(ctx, `SELECT discount_value FROM voucher_catalog WHERE code = 'DUP'`).Scan(&dval)
	require.NoError(t, err)
	require.Equal(t, int64(60_000), dval) // ensure it updated
}
