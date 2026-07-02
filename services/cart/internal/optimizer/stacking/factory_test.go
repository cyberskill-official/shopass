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

func TestFactory_SelectsByCountry(t *testing.T) {
	require.IsType(t, newVNRules(), RulesForCountry("VN", region.CountryPolicy{}))
	require.IsType(t, newMYPHRules(), RulesForCountry("MY", region.CountryPolicy{}))
}
