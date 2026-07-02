package optimizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type VNStackRules struct{}

func (VNStackRules) ValidStack(pv *Voucher, fs *Voucher, shopVouchers []Voucher) bool {
	return true
}

func shopV(shopID string, dt DiscountType, dv int64, minSpend *int64) Voucher {
	sid := shopID
	return Voucher{
		Type:          TypeShop,
		ShopID:        &sid,
		DiscountType:  dt,
		DiscountValue: dv,
		MinSpend:      minSpend,
	}
}

func platV(dt DiscountType, dv int64, minSpend *int64) Voucher {
	return Voucher{
		Type:          TypePlatform,
		DiscountType:  dt,
		DiscountValue: dv,
		MinSpend:      minSpend,
	}
}

func freeV(dv int64, minSpend *int64) Voucher {
	return Voucher{
		Type:          TypeFreeship,
		DiscountType:  DiscountAmount,
		DiscountValue: dv,
		MinSpend:      minSpend,
	}
}

func sptr(s string) *string { return &s }

// Ví dụ minh họa §3.5(3): VN-stack -> 110k.
func TestOptimize_VNStackExample_110k(t *testing.T) {
	items := []CartItem{
		{ShopID: sptr("A"), Qty: 1, UnitPrice: 300_000},
		{ShopID: sptr("B"), Qty: 1, UnitPrice: 250_000},
	}
	vouchers := Vouchers{
		Shop:     []Voucher{shopV("A", DiscountAmount, 30_000, ptr(250_000))}, // -30k đơn>=250k
		Platform: []Voucher{platV(DiscountAmount, 50_000, ptr(500_000))},      // -50k đơn>=500k
		Freeship: []Voucher{freeV(30_000, nil)},                               // freeship <=30k
	}
	res := OptimizeCart(items, vouchers, VNStackRules{}) // VN cho stack
	require.Equal(t, int64(110_000), res.Discount)       // 30k + 50k + 30k
}

func TestOptimize_ApplyCap_KepsDiscount(t *testing.T) {
	items := []CartItem{{ShopID: sptr("A"), Qty: 1, UnitPrice: 400_000}}
	v1 := shopV("A", DiscountPercent, 20, ptr(50_000))
	v1.Cap = ptr(50_000)
	vouchers := Vouchers{
		Shop: []Voucher{v1}, // 20% tối đa 50k
	}
	res := OptimizeCart(items, vouchers, VNStackRules{})
	require.Equal(t, int64(50_000), res.Discount) // 20% của 400k = 80k, kẹp về cap 50k
}

func TestOptimize_MinSpendGate_ExcludesVoucher(t *testing.T) {
	items := []CartItem{{ShopID: sptr("A"), Qty: 1, UnitPrice: 400_000}} // <500k
	vouchers := Vouchers{
		Platform: []Voucher{platV(DiscountAmount, 50_000, ptr(500_000))}, // cần >=500k
	}
	res := OptimizeCart(items, vouchers, VNStackRules{})
	require.Equal(t, int64(0), res.Discount) // không đạt min_spend -> không áp
}

func TestOptimize_BestShopVoucherPerShop(t *testing.T) {
	items := []CartItem{{ShopID: sptr("A"), Qty: 1, UnitPrice: 300_000}}
	vouchers := Vouchers{Shop: []Voucher{
		shopV("A", DiscountAmount, 20_000, nil),
		shopV("A", DiscountAmount, 35_000, nil), // tốt hơn cho shop A
	}}
	combo := chooseBestShopVoucherPerShop(vouchers.Shop, items)
	require.Len(t, combo, 1)
	require.Equal(t, int64(35_000), combo[0].DiscountValue) // chọn voucher tốt hơn
}

func TestOptimize_EmptyCart_ZeroDiscount(t *testing.T) {
	res := OptimizeCart(nil, Vouchers{}, VNStackRules{})
	require.Equal(t, int64(0), res.Discount)
	require.Empty(t, res.Combo)
}

func TestOptimize_Deterministic(t *testing.T) {
	items := []CartItem{{ShopID: sptr("A"), Qty: 1, UnitPrice: 300_000}}
	v := Vouchers{Shop: []Voucher{shopV("A", DiscountAmount, 30_000, nil)}}
	a := OptimizeCart(items, v, VNStackRules{})
	b := OptimizeCart(items, v, VNStackRules{})
	require.Equal(t, a.Combo, b.Combo) // cùng input -> cùng combo
}
