package api

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler, postbackHandler *PostbackHandler) {
	mux.HandleFunc("POST /v1/affiliate/link", handler.HandleCreateLink)
	// TASK-AFFIL-003: Postback webhook (no auth middleware, uses signature)
	if postbackHandler != nil {
		mux.HandleFunc("POST /v1/affiliate/postback/{network}", postbackHandler.HandlePostback)
	}
}
