package stacking

import (
	"testing"

	"github.com/stretchr/testify/require"
	"shopass/region"
	"shopass/services/cart/internal/optimizer"
)

func TestFactory_DefaultsToNoStack(t *testing.T) {
	r := RulesForCountry("XX", region.CountryPolicy{}) // nước chưa cấu hình
	pv := platV(optimizer.DiscountAmount, 50_000, nil, "p")
	fs := freeV(30_000, nil, "f")
	require.False(t, r.ValidStack(&pv, &fs, nil)) // mặc định no-stack (DEC-CART-23)
}

func TestFactory_UsesPolicyStackingAllowed(t *testing.T) {
	r := RulesForCountry("VN", region.CountryPolicy{VoucherStackingAllowed: true})
	pv := platV(optimizer.DiscountAmount, 50_000, nil, "p")
	fs := freeV(30_000, nil, "f")
	require.True(t, r.ValidStack(&pv, &fs, nil))
}

func TestFactory_UsesPolicyFreeshipGrouping(t *testing.T) {
	r := RulesForCountry("MY", region.CountryPolicy{
		VoucherStackingAllowed:      true,
		FreeshipGroupedWithPlatform: true,
	})
	pv := platV(optimizer.DiscountAmount, 50_000, nil, "p")
	fs := freeV(30_000, nil, "f")
	require.False(t, r.ValidStack(&pv, &fs, nil))
}
