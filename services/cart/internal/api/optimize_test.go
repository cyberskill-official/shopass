package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"shopass/services/cart/internal/optimizer"
)

type mapGater map[string]bool

func (g mapGater) Allow(country, gate string) bool {
	return g[country+"|"+gate]
}

func int64ptr(v int64) *int64 {
	return &v
}

func stringptr(v string) *string {
	return &v
}

func optimizeItems() []optimizer.CartItem {
	return []optimizer.CartItem{
		{ShopID: stringptr("A"), Qty: 1, UnitPrice: 300_000},
		{ShopID: stringptr("B"), Qty: 1, UnitPrice: 250_000},
	}
}

func optimizeVouchers(freeshipGroup string) optimizer.Vouchers {
	return optimizer.Vouchers{
		Shop: []optimizer.Voucher{{
			Type:          optimizer.TypeShop,
			ShopID:        stringptr("A"),
			DiscountType:  optimizer.DiscountAmount,
			DiscountValue: 30_000,
			MinSpend:      int64ptr(250_000),
			StackGroup:    stringptr("shop-A-grp"),
		}},
		Platform: []optimizer.Voucher{{
			Type:          optimizer.TypePlatform,
			DiscountType:  optimizer.DiscountAmount,
			DiscountValue: 50_000,
			MinSpend:      int64ptr(500_000),
			StackGroup:    stringptr("platform-grp"),
		}},
		Freeship: []optimizer.Voucher{{
			Type:          optimizer.TypeFreeship,
			DiscountType:  optimizer.DiscountAmount,
			DiscountValue: 30_000,
			StackGroup:    stringptr(freeshipGroup),
		}},
	}
}

func runOptimize(t *testing.T, handler *OptimizeHandler, country string) optimizer.OptimizeResult {
	t.Helper()

	body, err := json.Marshal(OptimizeRequest{
		PlatformID: 1,
		Country:    country,
		Items:      optimizeItems(),
		Vouchers:   optimizeVouchers("freeship-grp"),
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/cart/optimize", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handler.OptimizeCart(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var result optimizer.OptimizeResult
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
	return result
}

func TestOptimizeCart_UsesGatingAllowForVNStacking(t *testing.T) {
	handler := NewOptimizeHandler(mapGater{"VN|" + gateVoucherStacking: true})

	result := runOptimize(t, handler, "VN")

	require.Equal(t, int64(110_000), result.Discount)
	require.Equal(t, int64(30_000), result.Breakdown.Shop)
	require.Equal(t, int64(50_000), result.Breakdown.Platform)
	require.Equal(t, int64(30_000), result.Breakdown.Freeship)
}

func TestOptimizeCart_UsesGatingDenyForMYAndPH(t *testing.T) {
	handler := NewOptimizeHandler(mapGater{
		"MY|" + gateVoucherStacking: false,
		"PH|" + gateVoucherStacking: false,
	})

	require.Equal(t, int64(80_000), runOptimize(t, handler, "MY").Discount)
	require.Equal(t, int64(80_000), runOptimize(t, handler, "PH").Discount)
}

func TestOptimizeCart_DenyByDefaultForUnknownCountry(t *testing.T) {
	result := runOptimize(t, NewOptimizeHandler(mapGater{}), "XX")

	require.Equal(t, int64(80_000), result.Discount)
	require.Equal(t, int64(0), result.Breakdown.Freeship)
}
