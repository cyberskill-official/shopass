package api

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, checkoutHandler *Handler, ipnHandler *IPNHandler, waitlist *WaitlistHandler) {
	mux.HandleFunc("POST /v1/billing/checkout", checkoutHandler.HandleCheckout)
	mux.HandleFunc("POST /v1/billing/ipn/{gateway}", ipnHandler.HandleIPN)
	if waitlist != nil {
		mux.HandleFunc("POST /v1/leads/waitlist", waitlist.HandleWaitlist)
	}
}
