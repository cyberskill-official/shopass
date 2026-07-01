package api

import (
	"net/http"
)

// RegisterRoutes registers track endpoints.
func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	// Giả sử mux đã được wrap bằng middleware JWT của gateway
	mux.HandleFunc("POST /v1/track", h.HandleTrack)
}
