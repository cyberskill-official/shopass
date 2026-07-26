package api

import (
	"errors"
	"net/http"

	"shopass/services/comply/internal/consent"
)

type ConsentHandler struct {
	service ConsentService
}

func NewConsentHandler(service ConsentService) *ConsentHandler {
	return &ConsentHandler{service: service}
}

func (h *ConsentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/consent/grant", h.HandleGrant)
	mux.HandleFunc("POST /v1/consent/withdraw", h.HandleWithdraw)
	mux.HandleFunc("GET /v1/consent/history", h.HandleHistory)
}

type consentRequest struct {
	Purpose string `json:"purpose"`
	Source  string `json:"source"`
}

func (h *ConsentHandler) HandleGrant(w http.ResponseWriter, r *http.Request) {
	h.handleMutation(w, r, true)
}

func (h *ConsentHandler) HandleWithdraw(w http.ResponseWriter, r *http.Request) {
	h.handleMutation(w, r, false)
}

func (h *ConsentHandler) handleMutation(w http.ResponseWriter, r *http.Request, granted bool) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	var in consentRequest
	if !decodeJSON(r, &in) {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if in.Source == "" {
		in.Source = "web"
	}

	purpose := consent.Purpose(in.Purpose)
	meta := requestMeta(r)
	if granted {
		err = h.service.Grant(r.Context(), userID, purpose, in.Source, meta)
	} else {
		err = h.service.Withdraw(r.Context(), userID, purpose, in.Source, meta)
	}
	if errors.Is(err, consent.ErrUnknownPurpose) {
		writeError(w, http.StatusBadRequest, "unknown purpose")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "consent write failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id": userID,
		"purpose": in.Purpose,
		"granted": granted,
	})
}

func (h *ConsentHandler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing user identity")
		return
	}
	history, err := h.service.HistoryAll(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "consent history failed")
		return
	}
	writeJSON(w, http.StatusOK, history)
}
