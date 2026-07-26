package api

import (
	"net/http"
)

type DSARHandler struct {
	service DSARService
}

func NewDSARHandler(service DSARService) *DSARHandler {
	return &DSARHandler{service: service}
}

func (h *DSARHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/dsar", h.HandleCreate)
}

type dsarRequest struct {
	Kind string `json:"kind"`
}

func (h *DSARHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	var in dsarRequest
	if !decodeJSON(r, &in) {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	switch in.Kind {
	case "access", "portability":
		requestID, err := h.service.CreateRequest(r.Context(), userID, in.Kind)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "dsar create failed")
			return
		}
		bundle, err := h.service.Export(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "dsar export failed")
			return
		}
		if err := h.service.MarkCompleted(r.Context(), requestID); err != nil {
			writeError(w, http.StatusInternalServerError, "dsar complete failed")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"request_id": requestID,
			"status":     "completed",
			"export":     bundle,
		})
	case "erase":
		result, err := h.service.Erase(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "dsar erase failed")
			return
		}
		writeJSON(w, http.StatusOK, result)
	case "rectify":
		requestID, err := h.service.CreateRequest(r.Context(), userID, in.Kind)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "dsar create failed")
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{
			"request_id": requestID,
			"status":     "open",
		})
	default:
		writeError(w, http.StatusBadRequest, "invalid dsar kind")
	}
}
