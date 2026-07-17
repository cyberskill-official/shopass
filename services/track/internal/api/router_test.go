package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The production wiring registers all three endpoint groups through this
// helper. Keep this regression test so adding /v1/track never drops the
// pre-existing wishlist or alert surface.
func TestRegisterRoutesKeepsTrackWishlistAndAlertEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	trackHandler, _ := setupHandler(t)
	RegisterRoutes(mux, trackHandler, &WishlistHandler{}, &AlertRuleHandler{})

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/v1/track"},
		{method: http.MethodGet, path: "/v1/tracked-products"},
		{method: http.MethodPost, path: "/v1/wishlists"},
		{method: http.MethodGet, path: "/v1/wishlists"},
		{method: http.MethodPost, path: "/v1/alerts"},
		{method: http.MethodGet, path: "/v1/alerts"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		_, pattern := mux.Handler(req)
		if pattern == "" {
			t.Fatalf("%s %s was not registered", tc.method, tc.path)
		}
	}
}
