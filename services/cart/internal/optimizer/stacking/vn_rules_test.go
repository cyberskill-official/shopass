package stacking

import (
	"testing"

	"github.com/stretchr/testify/require"
	"shopass/services/cart/internal/optimizer"
)

func ptr(v int64) *int64 { return &v }
func sptr(s string) *string { return &s }

func shopV(shopID string, dt optimizer.DiscountType, dv int64, minSpend *int64, stackGrp string) optimizer.Voucher {
	sid := shopID
	return optimizer.Voucher{
		Type:          optimizer.TypeShop,
		ShopID:        &sid,
		DiscountType:  dt,
		DiscountValue: dv,
		MinSpend:      minSpend,
		StackGroup:    &stackGrp,
	}
}

func platV(dt optimizer.DiscountType, dv int64, minSpend *int64, stackGrp string) optimizer.Voucher {
	return optimizer.Voucher{
		Type:          optimizer.TypePlatform,
		DiscountType:  dt,
		DiscountValue: dv,
		MinSpend:      minSpend,
		StackGroup:    &stackGrp,
	}
}

func freeV(dv int64, minSpend *int64, stackGrp string) optimizer.Voucher {
	return optimizer.Voucher{
		Type:          optimizer.TypeFreeship,
		DiscountType:  optimizer.DiscountAmount,
		DiscountValue: dv,
		MinSpend:      minSpend,
		StackGroup:    &stackGrp,
	}
}

func exampleCart() []optimizer.CartItem {
	return []optimizer.CartItem{
		{ShopID: sptr("A"), Qty: 1, UnitPrice: 300_000},
		{ShopID: sptr("B"), Qty: 1, UnitPrice: 250_000},
	}
}

func exampleVouchers() optimizer.Vouchers {
	return optimizer.Vouchers{
		Shop:     []optimizer.Voucher{shopV("A", optimizer.DiscountAmount, 30_000, ptr(250_000), "shop-A-grp")},
		Platform: []optimizer.Voucher{platV(optimizer.DiscountAmount, 50_000, ptr(500_000), "platform-grp")},
		Freeship: []optimizer.Voucher{freeV(30_000, nil, "freeship-grp")},
	}
}

func TestVN_AllowsStack_ShopPlatformFreeship(t *testing.T) {
	r := newVNRules()
	pv := platV(optimizer.DiscountAmount, 50_000, ptr(500_000), "platform-grp")
	fs := freeV(30_000, nil, "freeship-grp") // khác stack_group platform
	shop := []optimizer.Voucher{shopV("A", optimizer.DiscountAmount, 30_000, nil, "shop-A-grp")}
	require.True(t, r.ValidStack(&pv, &fs, shop)) // VN cho stack cả ba (DEC-CART-20)
}

// Ví dụ §3.5(3) VN -> 110k (qua optimizer FR-CART-003).
func TestVN_OptimizeExample_110k(t *testing.T) {
	items := exampleCart() // shop A 300k, shop B 250k
	v := exampleVouchers() // shop A -30k, platform -50k>=500k, freeship 30k
	res := optimizer.OptimizeCart(items, v, newVNRules())
	require.Equal(t, int64(110_000), res.Discount)
}
