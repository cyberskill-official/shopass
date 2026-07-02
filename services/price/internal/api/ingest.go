package api

import (
	"encoding/json"
	"net/http"
	"time"

	"shopass/services/price/internal/price"
)

// IngestHandler accepts price snapshots from the scrape service and persists
// them delta-only. price owns price_snapshot (one-table-one-owner, DATA-MODEL),
// so every write goes through this internal endpoint rather than a direct
// cross-service INSERT.
type IngestHandler struct {
	snaps *price.SnapshotRepo
}

func NewIngestHandler(snaps *price.SnapshotRepo) *IngestHandler {
	return &IngestHandler{snaps: snaps}
}

func (h *IngestHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/price/snapshots", h.HandleIngest)
}

type ingestRequest struct {
	ProductID int64      `json:"product_id"`
	TS        *time.Time `json:"ts,omitempty"`
	Price     int64      `json:"price"`
	ListPrice *int64     `json:"list_price,omitempty"`
	Stock     *int32     `json:"stock,omitempty"`
	Sold      *int32     `json:"sold,omitempty"`
	FlashSale bool       `json:"flash_sale,omitempty"`
}

type ingestResponse struct {
	Written bool `json:"written"`
}

// HandleIngest serves POST /v1/price/snapshots.
func (h *IngestHandler) HandleIngest(w http.ResponseWriter, r *http.Request) {
	var req ingestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.ProductID <= 0 {
		writeErr(w, http.StatusBadRequest, "product_id must be > 0")
		return
	}
	// Money is BIGINT VND, strictly positive (DEC-PRICE-05 + CHECK price > 0).
	if req.Price <= 0 {
		writeErr(w, http.StatusBadRequest, "price must be > 0")
		return
	}
	if req.ListPrice != nil && *req.ListPrice < req.Price {
		writeErr(w, http.StatusBadRequest, "list_price must be >= price")
		return
	}

	ts := time.Now().UTC()
	if req.TS != nil {
		ts = req.TS.UTC()
	}
	snap := price.PriceSnapshot{
		ProductID: req.ProductID,
		TS:        ts,
		Price:     req.Price,
		ListPrice: req.ListPrice,
		Stock:     req.Stock,
		Sold:      req.Sold,
		FlashSale: req.FlashSale,
	}
	written, err := h.snaps.InsertSnapshot(r.Context(), snap)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "insert failed")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if written {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK) // delta-only skip, not an error
	}
	_ = json.NewEncoder(w).Encode(ingestResponse{Written: written})
}
