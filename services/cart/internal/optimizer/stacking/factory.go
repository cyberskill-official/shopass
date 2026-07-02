package stacking

import (
	"shopass/region"
	"shopass/services/cart/internal/optimizer"
)

// RulesForCountry chọn StackRules theo policy của nước; mặc định no-stack (DEC-CART-23).
func RulesForCountry(country string, policy region.CountryPolicy) optimizer.StackRules {
	switch country {
	case "VN":
		return newVNRules()
	case "MY", "PH":
		return newMYPHRules()
	default:
		// nước khác: đọc CountryPolicy; chưa cấu hình -> hạn chế nhất (no-stack)
		return PolicyStackRules{
			StackingAllowed:       policy.VoucherStackingAllowed, // mặc định false nếu chưa set
			FreeshipGroupedWithPF: policy.FreeshipGroupedWithPlatform || !policy.VoucherStackingAllowed,
		}
	}
}
