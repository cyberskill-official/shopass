package gw

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testUpstream(t *testing.T, name string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Upstream", name)
		w.Header().Set("X-User-Id-Echo", r.Header.Get("X-User-Id"))
		w.Header().Set("X-Request-Id-Echo", r.Header.Get("X-Request-Id"))
		w.WriteHeader(http.StatusOK)
	})
}

func testDeps(t *testing.T) Deps {
	t.Helper()
	auth := testUpstream(t, "auth")
	track := testUpstream(t, "track")
	price := testUpstream(t, "price")
	deal := testUpstream(t, "deal")
	notif := testUpstream(t, "notif")
	bff := testUpstream(t, "bff")
	return Deps{
		JWKS: &mockJWKS{},
		Upstreams: Upstreams{
			AuthHandler:  auth,
			TrackHandler: track,
			PriceHandler: price,
			DealHandler:  deal,
			NotifHandler: notif,
			BFFHandler:   bff,
		},
	}
}

func TestWAF_PathTraversal_400(t *testing.T) {
	h := NewHandler(testDeps(t))

	req := httptest.NewRequest("GET", "/v1/../etc/passwd", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, 400, rr.Code)
}

func TestWAF_BodySize_413(t *testing.T) {
	deps := testDeps(t)
	deps.WAFConfig = WAFConfig{MaxBodySize: 10}
	h := NewHandler(deps)

	req := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader("this is a very long body that exceeds 10 bytes"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, 413, rr.Code)
}

func TestRequestID_Generated(t *testing.T) {
	h := NewHandler(testDeps(t))

	req := httptest.NewRequest("POST", "/v1/auth/login", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, 200, rr.Code)
	require.NotEmpty(t, rr.Header().Get("X-Request-Id-Echo"))
	require.NotEmpty(t, rr.Header().Get("X-Request-Id"))
}

func TestRouter_RoutesToAllowlistedUpstream(t *testing.T) {
	h := NewHandler(testDeps(t))

	tests := []struct {
		method string
		path   string
		token  bool
		want   string
	}{
		{http.MethodPost, "/v1/auth/login", false, "auth"},
		{http.MethodGet, "/v1/track", true, "track"},
		{http.MethodGet, "/v1/tracked-products", true, "track"},
		{http.MethodPost, "/v1/products/1/browser-snapshot", true, "track"},
		{http.MethodGet, "/v1/alerts", true, "track"},
		{http.MethodPost, "/v1/alerts", true, "track"},
		{http.MethodPatch, "/v1/alerts/3", true, "track"},
		{http.MethodPost, "/v1/devices", true, "notif"},
		{http.MethodGet, "/v1/products/1/chart", true, "deal"},
		{http.MethodPost, "/graphql", true, "bff"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.token {
				req.Header.Set("Authorization", "Bearer valid-1")
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			require.Equal(t, tc.want, rr.Header().Get("X-Upstream"))
		})
	}
}

func TestGatewayStripsForgedIdentityBeforeSettingVerifiedIdentity(t *testing.T) {
	h := NewHandler(testDeps(t))
	req := httptest.NewRequest(http.MethodGet, "/v1/tracked-products", nil)
	req.Header.Set("Authorization", "Bearer valid-90112")
	req.Header.Set("X-User-Id", "999999")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "90112", rr.Header().Get("X-User-Id-Echo"))
}

func TestGatewayRejectsInternalPriceIngestRoute(t *testing.T) {
	h := NewHandler(testDeps(t))
	req := httptest.NewRequest(http.MethodPost, "/v1/price/snapshots", nil)
	req.Header.Set("Authorization", "Bearer valid-1")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGatewayRejectsUnscopedPriceReadRoutes(t *testing.T) {
	h := NewHandler(testDeps(t))
	for _, path := range []string{"/v1/products/1/price-history", "/v1/compare"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer valid-1")
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			require.Equal(t, http.StatusNotFound, rr.Code)
		})
	}
}

func TestGatewayRejectsUnavailableBetaRoutes(t *testing.T) {
	h := NewHandler(testDeps(t))
	for _, path := range []string{"/v1/wishlists", "/v1/compare"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer valid-1")
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			require.Equal(t, http.StatusNotFound, rr.Code)
		})
	}
}

func TestGatewayAlertsRejectSpoofedUserHeader(t *testing.T) {
	h := NewHandler(testDeps(t))
	req := httptest.NewRequest(http.MethodGet, "/v1/alerts", nil)
	req.Header.Set("Authorization", "Bearer valid-90112")
	req.Header.Set("X-User-Id", "999999")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "track", rr.Header().Get("X-Upstream"))
	// jwtVerify overwrites client-supplied identity with the token subject.
	require.Equal(t, "90112", rr.Header().Get("X-User-Id-Echo"))
}
