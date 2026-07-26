package api

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler, postbackHandler *PostbackHandler) {
	RegisterRoutesWithCashback(mux, handler, postbackHandler, nil)
}

func RegisterRoutesWithCashback(mux *http.ServeMux, handler *Handler, postbackHandler *PostbackHandler, cashbackHandler *CashbackHandler) {
	mux.HandleFunc("POST /v1/affiliate/link", handler.HandleCreateLink)
	// TASK-AFFIL-003: Postback webhook (no auth middleware, uses signature)
	if postbackHandler != nil {
		mux.HandleFunc("POST /v1/affiliate/postback/{network}", postbackHandler.HandlePostback)
	}
	// TASK-AFFIL-005: cashback summary + disclosure
	if cashbackHandler != nil {
		mux.HandleFunc("GET /v1/cashback/summary", cashbackHandler.HandleSummary)
	}
}
