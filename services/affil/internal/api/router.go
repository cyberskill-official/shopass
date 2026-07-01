package api

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("POST /v1/affiliate/link", handler.HandleCreateLink)
}
