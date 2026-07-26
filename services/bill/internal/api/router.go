package api

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, checkoutHandler *Handler, ipnHandler *IPNHandler, waitlist *WaitlistHandler, referralH *ReferralHandler) {
	mux.HandleFunc("POST /v1/billing/checkout", checkoutHandler.HandleCheckout)
	mux.HandleFunc("POST /v1/billing/ipn/{gateway}", ipnHandler.HandleIPN)
	if waitlist != nil {
		mux.HandleFunc("POST /v1/leads/waitlist", waitlist.HandleWaitlist)
	}
	if referralH != nil {
		mux.HandleFunc("GET /v1/referral/me", referralH.HandleMe)
		mux.HandleFunc("POST /v1/referral/attribute", referralH.HandleAttribute)
	}
}
