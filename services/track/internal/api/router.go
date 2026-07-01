package api

import (
	"net/http"
)

// RegisterRoutes registers track endpoints.
func RegisterRoutes(mux *http.ServeMux, h *Handler, wh *WishlistHandler, ah *AlertRuleHandler) {
	// Giả sử mux đã được wrap bằng middleware JWT của gateway
	mux.HandleFunc("POST /v1/track", h.HandleTrack)
	
	if wh != nil {
		wh.RegisterRoutes(mux)
	}
	if ah != nil {
		ah.RegisterRoutes(mux)
	}
}
