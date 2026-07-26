package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/b2b/internal/apikey"
	"shopass/services/b2b/internal/trend"
)

func setupPublic(t *testing.T, cells []trend.MarketTrendCell) (*http.ServeMux, string, *apikey.MemoryUsage) {
	t.Helper()
	store := apikey.NewMemoryKeyStore()
	prefix, secret, hash, err := apikey.NewKey()
	require.NoError(t, err)
	store.Put(&apikey.APIKey{
		ID: 1, Prefix: prefix, SecretHash: hash, OrgName: "acme",
		Tier: "pro", RatePerMin: 60, MonthlyQuota: 1000,
	})
	usage := apikey.NewMemoryUsage()
	h := &PublicTrendHandler{
		Auth:  apikey.NewAuth(store),
		Limit: apikey.NewRateLimiter(),
		Usage: usage,
		Trend: memTrend{cells: cells},
		Now:   func() time.Time { return time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC) },
	}
	mux := http.NewServeMux()
	RegisterRoutes(mux, nil, nil, h)
	return mux, apikey.Format(prefix, secret), usage
}

func TestPublicTrends_Unauthorized(t *testing.T) {
	mux, _, _ := setupPublic(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/public/v1/trends?category_id=7&platform_id=1&from=2026-06-14&to=2026-06-21", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, 401, rec.Code)
}

func TestPublicTrends_OK(t *testing.T) {
	day := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	med := int64(320000)
	mux, raw, usage := setupPublic(t, []trend.MarketTrendCell{{
		CategoryID: 7, PlatformID: 1, Day: day, SKUCount: 80,
		MedianP: &med, P25P: &med, P75P: &med,
	}})
	req := httptest.NewRequest(http.MethodGet, "/public/v1/trends?category_id=7&platform_id=1&from=2026-06-14&to=2026-06-21", nil)
	req.Header.Set("X-API-Key", raw)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)
	require.NotContains(t, rec.Body.String(), "product_id")
	require.NotContains(t, rec.Body.String(), "user_id")
	require.Len(t, usage.Events(), 1)
	require.Equal(t, 200, usage.Events()[0].StatusCode)
}

func TestPublicRoutes_NoRawSurfaces(t *testing.T) {
	require.False(t, HasForbiddenPublicRoute(PublicRoutePrefixes()))
}
