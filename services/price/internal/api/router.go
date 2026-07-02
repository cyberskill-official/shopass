package api

import (
	"net/http"
)

// RegisterRoutes configures HTTP routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/products/{id}/price-history", h.HandlePriceHistory)
	mux.HandleFunc("GET /v1/compare", h.HandleCompare) // FR-PRICE-004 cross-platform compare
}
