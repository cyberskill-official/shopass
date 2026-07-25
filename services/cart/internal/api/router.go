package api

import (
	"net/http"
	"shopass/services/cart/internal/cart"
)

func NewRouter(repo *cart.SnapshotRepo) *http.ServeMux {
	mux := http.NewServeMux()

	snapHandler := NewSnapshotHandler(repo)
	optHandler := NewOptimizeHandler()

	mux.HandleFunc("POST /v1/cart/snapshot", snapHandler.CreateSnapshot)
	mux.HandleFunc("POST /v1/cart/optimize", optHandler.OptimizeCart)

	return mux
}
