package optimizer

import (
	"sort"
)

// optimizeCart bám đúng cấu trúc pseudo-code §3.5(3) (DEC-CART-13):
// duyệt (platformVoucher x freeship x shopCombo), lọc validStack + meetsMinSpend, applyCaps, giữ best.
func OptimizeCart(items []CartItem, vouchers Vouchers, rules StackRules) OptimizeResult {
	cartTotal := totalOf(items)
	applicableShop := filterByMinSpend(vouchers.Shop, items) // voucher shop đạt min_spend
	best := OptimizeResult{Discount: 0, Combo: nil, Breakdown: Breakdown{}}

	platformOptions := append([]*Voucher{nil}, toPtrSlice(vouchers.Platform)...)
	freeshipOptions := append([]*Voucher{nil}, toPtrSlice(vouchers.Freeship)...)

	for _, pv := range platformOptions { // platformVoucher + [none]
		for _, fs := range freeshipOptions { // freeship + [none]
			shopCombo := chooseBestShopVoucherPerShop(applicableShop, items) // tốt nhất per shop
			if !rules.ValidStack(pv, fs, shopCombo) {                        // luật per-country (TASK-CART-004)
				continue
			}
			if !meetsMinSpend(shopCombo, pv, fs, items) { // ngưỡng đơn
				continue
			}
			
			shopDiscount := sumShopDiscounts(shopCombo, items)
			platDiscount := discountOf(pv, cartTotal)
			freeDiscount := freeshipValue(fs, cartTotal)
			
			total := shopDiscount + platDiscount + freeDiscount
			if total > best.Discount || (total == best.Discount && len(best.Combo) == 0 && total > 0) {
				best = OptimizeResult{
					Discount:  total,
					Combo:     combine(shopCombo, pv, fs),
					Breakdown: Breakdown{
						Shop:     shopDiscount,
						Platform: platDiscount,
						Freeship: freeDiscount,
					},
				}
			}
		}
	}
	return best // giỏ rỗng / không voucher -> Discount 0, Combo nil (§1 #9)
}

// chooseBestShopVoucherPerShop chọn voucher shop tốt nhất CHO TỪNG shop (DEC-CART-16).
func chooseBestShopVoucherPerShop(shopVouchers []Voucher, items []CartItem) []Voucher {
	bestPerShop := map[string]Voucher{}
	for _, v := range shopVouchers {
		if v.ShopID == nil {
			continue
		}
		base := shopSubtotal(items, *v.ShopID)
		d := discountAmount(v, base)
		cur, ok := bestPerShop[*v.ShopID]
		if !ok || d > discountAmount(cur, shopSubtotal(items, *cur.ShopID)) {
			bestPerShop[*v.ShopID] = v
		}
	}
	return sortedValues(bestPerShop) // tất định (§1 #10)
}

func filterByMinSpend(vouchers []Voucher, items []CartItem) []Voucher {
	var valid []Voucher
	for _, v := range vouchers {
		if v.ShopID == nil {
			continue
		}
		base := shopSubtotal(items, *v.ShopID)
		if v.MinSpend == nil || base >= *v.MinSpend {
			valid = append(valid, v)
		}
	}
	return valid
}

func meetsMinSpend(shopCombo []Voucher, pv *Voucher, fs *Voucher, items []CartItem) bool {
	// Các shop voucher đã được filter qua filterByMinSpend rồi, nhưng ta cần kiểm tra lại
	// platform & freeship.
	cartTotal := totalOf(items)
	if pv != nil && pv.MinSpend != nil && cartTotal < *pv.MinSpend {
		return false
	}
	if fs != nil && fs.MinSpend != nil && cartTotal < *fs.MinSpend {
		return false
	}
	return true
}

func totalOf(items []CartItem) int64 {
	var total int64
	for _, it := range items {
		total += int64(it.Qty) * it.UnitPrice
	}
	return total
}

func shopSubtotal(items []CartItem, shopID string) int64 {
	var total int64
	for _, it := range items {
		if it.ShopID != nil && *it.ShopID == shopID {
			total += int64(it.Qty) * it.UnitPrice
		}
	}
	return total
}

func sumShopDiscounts(shopCombo []Voucher, items []CartItem) int64 {
	var total int64
	for _, v := range shopCombo {
		if v.ShopID != nil {
			base := shopSubtotal(items, *v.ShopID)
			total += discountAmount(v, base)
		}
	}
	return total
}

func discountOf(v *Voucher, base int64) int64 {
	if v == nil {
		return 0
	}
	return discountAmount(*v, base)
}

func combine(shopCombo []Voucher, pv *Voucher, fs *Voucher) []Voucher {
	var combo []Voucher
	if pv != nil {
		combo = append(combo, *pv)
	}
	if fs != nil {
		combo = append(combo, *fs)
	}
	combo = append(combo, shopCombo...)
	return combo
}

func toPtrSlice(vouchers []Voucher) []*Voucher {
	var ptrs []*Voucher
	for i := range vouchers {
		ptrs = append(ptrs, &vouchers[i])
	}
	return ptrs
}

func sortedValues(m map[string]Voucher) []Voucher {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	var vals []Voucher
	for _, k := range keys {
		vals = append(vals, m[k])
	}
	return vals
}
