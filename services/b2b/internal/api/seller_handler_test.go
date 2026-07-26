package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/b2b/internal/seller"
	"shopass/services/b2b/internal/trend"
)

type fixedPrice int64

func (p fixedPrice) AvgOwnPrice(_ context.Context, _ int64, _ string, _ time.Time) (int64, error) {
	return int64(p), nil
}

func setupSeller(t *testing.T, owned []seller.OwnedSKU, cells []trend.MarketTrendCell) *http.ServeMux {
	t.Helper()
	h := &SellerHandler{
		Ownership: &seller.Ownership{Store: seller.NewMemoryOwnership(owned...)},
		Trend:     memTrend{cells: cells},
		Prices:    fixedPrice(180_000),
		Metrics:   newSellerMetrics(),
	}
	mux := http.NewServeMux()
	RegisterRoutes(mux, nil, h)
	return mux
}

func TestSellerPosition_NotOwner_403(t *testing.T) {
	mux := setupSeller(t, []seller.OwnedSKU{{SellerOrgID: 1, ShopID: "shopA", ProductID: 100, Verified: false}}, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/b2b/seller/position?shop_id=shopA&category_id=7&platform_id=1&day=2026-06-20", nil)
	req.Header.Set("X-B2B-Org-Id", "1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, 403, rec.Code)
}

func TestSellerPosition_Suppressed_422(t *testing.T) {
	mux := setupSeller(t,
		[]seller.OwnedSKU{{SellerOrgID: 1, ShopID: "shopA", ProductID: 100, Verified: true}},
		nil, // QueryCells empty → insufficient
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/b2b/seller/position?shop_id=shopA&category_id=991&platform_id=1&day=2026-06-20", nil)
	req.Header.Set("X-B2B-Org-Id", "1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, 422, rec.Code)
}

func TestSellerPosition_OK(t *testing.T) {
	day := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	med := int64(200_000)
	mux := setupSeller(t,
		[]seller.OwnedSKU{{SellerOrgID: 1, ShopID: "shopA", ProductID: 100, Verified: true}},
		[]trend.MarketTrendCell{{
			CategoryID: 7, PlatformID: 1, Day: day, SKUCount: 80,
			MedianP: &med, P25P: ptrI(100_000), P75P: ptrI(300_000),
		}},
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/b2b/seller/position?shop_id=shopA&category_id=7&platform_id=1&day=2026-06-20", nil)
	req.Header.Set("X-B2B-Org-Id", "1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)
	require.NotContains(t, rec.Body.String(), "competitor")
	require.Contains(t, rec.Body.String(), "seller_price")
	require.Contains(t, rec.Body.String(), "percentile_rank")
}

func ptrI(v int64) *int64 { return &v }
