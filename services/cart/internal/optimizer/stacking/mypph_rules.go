package stacking

import "shopass/services/cart/internal/optimizer"

// MY/PH 2025: bỏ stack đa nhóm; freeship gộp nhóm platform (DEC-CART-21).
func newMYPHRules() optimizer.StackRules {
	return PolicyStackRules{StackingAllowed: true, FreeshipGroupedWithPF: true}
	// StackingAllowed=true cho phép shop + một-trong-{platform,freeship};
	// FreeshipGroupedWithPF=true loại trừ platform & freeship dùng cùng lúc.
}
