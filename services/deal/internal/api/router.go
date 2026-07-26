package api

import (
	"net/http"
)

// RegisterRoutes đăng ký các route cho Deal service
func RegisterRoutes(mux *http.ServeMux, h *Handler, check *FakeSaleCheckHandler) {
	// Giả định JWT middleware đã được áp dụng ở API Gateway hoặc bọc ngoài mux
	mux.HandleFunc("GET /v1/products/{id}/chart", h.HandleChart)
	if check != nil {
		mux.HandleFunc("POST /v1/tools/fake-sale-check", check.HandleFakeSaleCheck)
	}
}
