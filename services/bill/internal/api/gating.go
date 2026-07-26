package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"shopass/services/bill/internal/gating"
)

// GateChecker is the narrow surface the private check endpoint needs.
type GateChecker interface {
	AllowWithUsage(ctx context.Context, userID int64, featureKey string, usage *int64) (bool, error)
}

type GatingHandler struct {
	gate              GateChecker
	serviceTokenHash  [sha256.Size]byte
	tokenIsConfigured bool
}

func NewGatingHandler(gate GateChecker, serviceToken string) *GatingHandler {
	return &GatingHandler{
		gate:              gate,
		serviceTokenHash:  sha256.Sum256([]byte(serviceToken)),
		tokenIsConfigured: strings.TrimSpace(serviceToken) != "",
	}
}

func (h *GatingHandler) validServiceToken(got string) bool {
	if !h.tokenIsConfigured || strings.TrimSpace(got) == "" {
		return false
	}
	sum := sha256.Sum256([]byte(got))
	return subtle.ConstantTimeCompare(sum[:], h.serviceTokenHash[:]) == 1
}

func (h *GatingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /internal/v1/gating/check", h.HandleCheck)
}

type checkRequest struct {
	UserID     int64  `json:"user_id"`
	FeatureKey string `json:"feature_key"`
	Usage      *int64 `json:"usage"`
}

func (h *GatingHandler) HandleCheck(w http.ResponseWriter, req *http.Request) {
	if !h.validServiceToken(req.Header.Get("X-Service-Token")) {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body checkRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.UserID <= 0 || body.FeatureKey == "" {
		writeErr(w, http.StatusBadRequest, "user_id and feature_key required")
		return
	}
	ok, err := h.gate.AllowWithUsage(req.Context(), body.UserID, body.FeatureKey, body.Usage)
	allowed := ok && err == nil
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"allowed":       allowed,
		"limit_reached": errors.Is(err, gating.ErrLimitReached),
		"feature_key":   body.FeatureKey,
	})
}
