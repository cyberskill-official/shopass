package stacking

import (
	"testing"

	"github.com/stretchr/testify/require"
	"shopass/services/cart/internal/optimizer"
)

func exampleVouchersMYPH() optimizer.Vouchers {
	return optimizer.Vouchers{
		Shop:     []optimizer.Voucher{shopV("A", optimizer.DiscountAmount, 30_000, ptr(250_000), "shop-A-grp")},
		Platform: []optimizer.Voucher{platV(optimizer.DiscountAmount, 50_000, ptr(500_000), "platform-grp")},
		Freeship: []optimizer.Voucher{freeV(30_000, nil, "platform-grp")}, // freeship gán cùng stack_group platform
	}
}

func TestMYPH_RejectsPlatformPlusFreeship(t *testing.T) {
	r := newMYPHRules()
	pv := platV(optimizer.DiscountAmount, 50_000, ptr(500_000), "platform-grp")
	fs := freeV(30_000, nil, "platform-grp") // MY/PH: freeship cùng nhóm platform
	shop := []optimizer.Voucher{shopV("A", optimizer.DiscountAmount, 30_000, nil, "shop-A-grp")}
	require.False(t, r.ValidStack(&pv, &fs, shop)) // không stack platform + freeship (DEC-CART-21)
}

func TestMYPH_AllowsShopPlusOne(t *testing.T) {
	r := newMYPHRules()
	pv := platV(optimizer.DiscountAmount, 50_000, ptr(500_000), "platform-grp")
	shop := []optimizer.Voucher{shopV("A", optimizer.DiscountAmount, 30_000, nil, "shop-A-grp")}
	require.True(t, r.ValidStack(&pv, nil, shop)) // shop + platform OK
}

// Ví dụ §3.5(3) MY/PH -> 80k (platform+freeship bị loại, chọn max(50k,30k)+30k shop).
func TestMYPH_OptimizeExample_80k(t *testing.T) {
	items := exampleCart()
	v := exampleVouchersMYPH() // freeship gán cùng stack_group platform
	res := optimizer.OptimizeCart(items, v, newMYPHRules())
	require.Equal(t, int64(80_000), res.Discount) // max(50k, 30k) + 30k shop = 80k
}
