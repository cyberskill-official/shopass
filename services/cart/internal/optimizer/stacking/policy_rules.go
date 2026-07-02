package stacking

import "shopass/services/cart/internal/optimizer"

// PolicyStackRules đọc CountryPolicy thay vì hardcode luật (DEC-CART-19).
type PolicyStackRules struct {
	StackingAllowed       bool // CountryPolicy.voucher_stacking_allowed
	FreeshipGroupedWithPF bool // freeship gộp nhóm platform (MY/PH 2025)
}

func (r PolicyStackRules) ValidStack(pv *optimizer.Voucher, fs *optimizer.Voucher, shopVouchers []optimizer.Voucher) bool {
	// một shop tối đa một voucher shop (đã đảm bảo bởi chooseBestShopVoucherPerShop)
	if !r.StackingAllowed {
		// no-stack: tối đa một voucher ngoài shop (platform XOR freeship)
		if pv != nil && fs != nil {
			return false
		}
	}
	if r.FreeshipGroupedWithPF && pv != nil && fs != nil {
		return false // MY/PH: freeship gộp nhóm platform -> loại trừ (DEC-CART-21)
	}
	return !hasSameStackGroupConflict(pv, fs, shopVouchers) // stack_group loại trừ (DEC-CART-22)
}

func hasSameStackGroupConflict(pv *optimizer.Voucher, fs *optimizer.Voucher, shopVouchers []optimizer.Voucher) bool {
	groups := make(map[string]bool)
	if pv != nil && pv.StackGroup != nil {
		groups[*pv.StackGroup] = true
	}
	if fs != nil && fs.StackGroup != nil {
		if groups[*fs.StackGroup] {
			return true
		}
		groups[*fs.StackGroup] = true
	}
	for _, sv := range shopVouchers {
		if sv.StackGroup != nil {
			if groups[*sv.StackGroup] {
				return true
			}
			groups[*sv.StackGroup] = true
		}
	}
	return false
}
