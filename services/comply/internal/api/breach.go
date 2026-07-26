package api

import (
	"errors"
	"net/http"
	"strconv"

	"shopass/services/comply/internal/breach"
)

type BreachHandler struct {
	service BreachService
}

func NewBreachHandler(service BreachService) *BreachHandler {
	return &BreachHandler{service: service}
}

func (h *BreachHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/comply/breach/open", h.HandleOpen)
	mux.HandleFunc("POST /v1/comply/breach/{id}/advance", h.HandleAdvance)
	mux.HandleFunc("POST /v1/comply/breach/{id}/close", h.HandleClose)
	mux.HandleFunc("GET /v1/comply/breach/overdue", h.HandleOverdue)
}

type breachAdvanceRequest struct {
	Status string `json:"status"`
}

func (h *BreachHandler) HandleOpen(w http.ResponseWriter, r *http.Request) {
	var in breach.BreachInput
	if !decodeJSON(r, &in) {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	id, err := h.service.Open(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "breach open failed")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *BreachHandler) HandleAdvance(w http.ResponseWriter, r *http.Request) {
	id, ok := incidentID(w, r)
	if !ok {
		return
	}
	var in breachAdvanceRequest
	if !decodeJSON(r, &in) {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	err := h.service.Advance(r.Context(), id, breach.Status(in.Status))
	if errors.Is(err, breach.ErrInvalidTransition) {
		writeError(w, http.StatusBadRequest, "invalid transition")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "breach advance failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": in.Status})
}

func (h *BreachHandler) HandleClose(w http.ResponseWriter, r *http.Request) {
	id, ok := incidentID(w, r)
	if !ok {
		return
	}
	err := h.service.Close(r.Context(), id)
	if errors.Is(err, breach.ErrInvalidTransition) {
		writeError(w, http.StatusBadRequest, "invalid transition")
		return
	}
	if errors.Is(err, breach.ErrSubjectsNotNotified) {
		writeError(w, http.StatusConflict, "subjects not notified")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "breach close failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": "closed"})
}

func (h *BreachHandler) HandleOverdue(w http.ResponseWriter, r *http.Request) {
	incidents, err := h.service.Overdue(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "breach overdue failed")
		return
	}
	writeJSON(w, http.StatusOK, incidents)
}

func incidentID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid breach id")
		return 0, false
	}
	return id, true
}
