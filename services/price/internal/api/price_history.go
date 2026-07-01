package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"shopass/services/price/internal/price"
)

type Handler struct {
	repo *price.Repo
}

func NewHandler(repo *price.Repo) *Handler {
	return &Handler{repo: repo}
}

type HistoryResponse struct {
	ProductID int64              `json:"product_id"`
	Range     string             `json:"range"`
	Daily     []price.DailyPoint `json:"daily"`
	Tail      []price.TailPoint  `json:"tail"`
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// HandlePriceHistory phục vụ GET /v1/products/{id}/price-history?range=90d.
func (h *Handler) HandlePriceHistory(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid product id")
		return
	}

	rangeRaw := req.URL.Query().Get("range")
	window, ok := price.ParseRange(rangeRaw)
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid range")
		return
	}
	if rangeRaw == "" {
		rangeRaw = "90d"
	}

	exists, err := h.repo.ProductExists(req.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !exists {
		writeErr(w, http.StatusNotFound, "product not found") // DEC-PRICE-34
		return
	}

	from := time.Now().Add(-window)
	daily, err := h.repo.QueryDailyBody(req.Context(), id, from)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Handle nil slice safely by coercing to empty slice for JSON output
	if daily == nil {
		daily = []price.DailyPoint{}
	}

	tail, err := h.repo.QueryRawTail(req.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	if tail == nil {
		tail = []price.TailPoint{}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(HistoryResponse{
		ProductID: id,
		Range:     rangeRaw,
		Daily:     daily,
		Tail:      tail,
	})
}
