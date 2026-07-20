package api

import (
	"net/http"
)

// RegisterRoutes registers track endpoints.
func RegisterRoutes(mux *http.ServeMux, h *Handler, wh *WishlistHandler, ah *AlertRuleHandler) {
	// The production mux is wrapped by gateway identity middleware before it
	// reaches these handlers.
	mux.HandleFunc("POST /v1/track", h.HandleTrack)
	mux.HandleFunc("GET /v1/tracked-products", h.HandleListTrackedProducts)
	mux.HandleFunc("POST /v1/products/{id}/browser-snapshot", h.HandleBrowserSnapshot)

	if wh != nil {
		wh.RegisterRoutes(mux)
	}
	if ah != nil {
		ah.RegisterRoutes(mux)
	}
}
