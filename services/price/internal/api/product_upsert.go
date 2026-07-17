package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"shopass/services/price/internal/price"
)

const (
	productUpsertPath    = "/internal/v1/products/upsert"
	maxProductUpsertBody = 64 << 10
)

// ProductUpserter is the part of the price registry used by the private
// product-upsert endpoint. Keeping this small makes the HTTP boundary easy to
// exercise without a database.
type ProductUpserter interface {
	Upsert(ctx context.Context, p price.TrackedProduct) (price.TrackedProduct, error)
}

// ProductUpsertHandler accepts registry writes from trusted internal services.
// The route is intentionally outside /v1 so the public gateway will not route
// to it. Network isolation is defense in depth; every request must also carry
// the configured service token.
type ProductUpsertHandler struct {
	products          ProductUpserter
	serviceTokenHash  [sha256.Size]byte
	tokenIsConfigured bool
}

// NewProductUpsertHandler builds the private product endpoint. An empty token
// deliberately leaves the endpoint unavailable rather than accepting a blank
// header as credentials.
func NewProductUpsertHandler(products ProductUpserter, serviceToken string) *ProductUpsertHandler {
	return &ProductUpsertHandler{
		products:          products,
		serviceTokenHash:  sha256.Sum256([]byte(serviceToken)),
		tokenIsConfigured: strings.TrimSpace(serviceToken) != "",
	}
}

func (h *ProductUpsertHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST "+productUpsertPath, h.HandleUpsert)
}

type productUpsertRequest struct {
	PlatformID     int16   `json:"platform_id"`
	PlatformItemID string  `json:"platform_item_id"`
	ShopID         *string `json:"shop_id,omitempty"`
	Title          *string `json:"title,omitempty"`
	CategoryID     *int64  `json:"category_id,omitempty"`
}

type productUpsertResponse struct {
	ID int64 `json:"id"`
}

func (h *ProductUpsertHandler) validServiceToken(got string) bool {
	if !h.tokenIsConfigured {
		return false
	}
	// Hash both values first so the comparison always has a fixed-length input.
	// This avoids a direct string comparison of the service credential.
	gotHash := sha256.Sum256([]byte(got))
	return subtle.ConstantTimeCompare(h.serviceTokenHash[:], gotHash[:]) == 1
}

// HandleUpsert serves POST /internal/v1/products/upsert.
func (h *ProductUpsertHandler) HandleUpsert(w http.ResponseWriter, r *http.Request) {
	if !h.tokenIsConfigured {
		writeErr(w, http.StatusServiceUnavailable, "internal endpoint unavailable")
		return
	}
	if !h.validServiceToken(r.Header.Get("X-Service-Token")) {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req productUpsertRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxProductUpsertBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}

	req.PlatformItemID = strings.TrimSpace(req.PlatformItemID)
	if req.PlatformID <= 0 || req.PlatformItemID == "" {
		writeErr(w, http.StatusBadRequest, "invalid product")
		return
	}

	product, err := h.products.Upsert(r.Context(), price.TrackedProduct{
		PlatformID:     req.PlatformID,
		PlatformItemID: req.PlatformItemID,
		ShopID:         req.ShopID,
		Title:          req.Title,
		CategoryID:     req.CategoryID,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(productUpsertResponse{ID: product.ID})
}
