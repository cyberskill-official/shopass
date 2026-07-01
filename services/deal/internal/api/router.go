package api

import (
	"net/http"
)

// RegisterRoutes đăng ký các route cho Deal service
func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	// Giả định JWT middleware đã được áp dụng ở API Gateway hoặc bọc ngoài mux
	mux.HandleFunc("GET /v1/products/{id}/chart", h.HandleChart)
}
