package stacking

import "shopass/services/cart/internal/optimizer"

// VN: cho stack 1 shop + 1 platform + freeship (DEC-CART-20).
func newVNRules() optimizer.StackRules {
	return PolicyStackRules{StackingAllowed: true, FreeshipGroupedWithPF: false}
}
