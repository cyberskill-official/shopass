package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"shopass/services/track/internal/engine"
)

// ProductEvaluator is the alert engine surface used after a written ingest.
type ProductEvaluator interface {
	EvaluateForProduct(ctx context.Context, snap engine.Snapshot) error
}

// PriceChangedHandler is the private pricesvc → tracksvc hop. It is not
// allowlisted on the public gateway.
type PriceChangedHandler struct {
	eval              ProductEvaluator
	serviceTokenHash  [sha256.Size]byte
	tokenIsConfigured bool
}

func NewPriceChangedHandler(eval ProductEvaluator, serviceToken string) *PriceChangedHandler {
	return &PriceChangedHandler{
		eval:              eval,
		serviceTokenHash:  sha256.Sum256([]byte(serviceToken)),
		tokenIsConfigured: strings.TrimSpace(serviceToken) != "",
	}
}

func (h *PriceChangedHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /internal/v1/price-changed", h.HandlePriceChanged)
}

type priceChangedRequest struct {
	ProductID int64  `json:"product_id"`
	Price     int64  `json:"price"`
	ListPrice *int64 `json:"list_price,omitempty"`
}

func (h *PriceChangedHandler) validServiceToken(got string) bool {
	if !h.tokenIsConfigured {
		return false
	}
	gotHash := sha256.Sum256([]byte(got))
	return subtle.ConstantTimeCompare(h.serviceTokenHash[:], gotHash[:]) == 1
}

func (h *PriceChangedHandler) HandlePriceChanged(w http.ResponseWriter, r *http.Request) {
	if h.eval == nil || !h.tokenIsConfigured {
		writeErr(w, http.StatusServiceUnavailable, "internal endpoint unavailable")
		return
	}
	if !h.validServiceToken(r.Header.Get("X-Service-Token")) {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req priceChangedRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.ProductID <= 0 || req.Price <= 0 {
		writeErr(w, http.StatusBadRequest, "product_id and price must be > 0")
		return
	}

	err := h.eval.EvaluateForProduct(r.Context(), engine.Snapshot{
		ProductID: req.ProductID,
		Price:     req.Price,
		ListPrice: req.ListPrice,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "evaluate failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
