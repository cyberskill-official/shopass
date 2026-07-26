package stacking

import (
	"shopass/region"
	"shopass/services/cart/internal/optimizer"
)

// RulesForCountry chọn StackRules theo policy của nước; mặc định no-stack (DEC-CART-23).
func RulesForCountry(_ string, policy region.CountryPolicy) optimizer.StackRules {
	return PolicyStackRules{
		StackingAllowed:       policy.VoucherStackingAllowed, // mặc định false nếu chưa set
		FreeshipGroupedWithPF: policy.FreeshipGroupedWithPlatform || !policy.VoucherStackingAllowed,
	}
}
