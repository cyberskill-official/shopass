package api

import (
	"encoding/json"
	"net/http"

	"shopass/services/affil/internal/auth"
	"shopass/services/affil/internal/cashback"
)

// CashbackHandler serves GET /v1/cashback/summary (TASK-AFFIL-005).
type CashbackHandler struct {
	ledger *cashback.Ledger
}

func NewCashbackHandler(ledger *cashback.Ledger) *CashbackHandler {
	return &CashbackHandler{ledger: ledger}
}

func (h *CashbackHandler) HandleSummary(w http.ResponseWriter, req *http.Request) {
	if h == nil || h.ledger == nil || h.ledger.Store == nil {
		writeErr(w, 503, "cashback unavailable")
		return
	}
	userID := auth.UserID(req.Context())
	if userID == 0 {
		writeErr(w, 401, "unauthorized")
		return
	}
	sum, err := h.ledger.Store.Summary(req.Context(), userID)
	if err != nil {
		writeErr(w, 500, "internal error")
		return
	}
	resp := sum.ToResponse(h.ledger.Cfg.PayoutThreshold)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}
