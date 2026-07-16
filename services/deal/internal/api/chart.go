package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"shopass/services/deal/internal/chart"
)

type ChartResponse struct {
    ProductID   int64              `json:"product_id"`
    Range       string             `json:"range"`
    Maturity    string             `json:"maturity"` // MATURE | WARMING | NEW (TASK-DEAL-002)
    Daily       []chart.DailyPoint `json:"daily"`
    Annotations chart.Annotations  `json:"annotations"`
}

var rangeWindows = map[string]time.Duration{
    "7d": 7 * 24 * time.Hour, "30d": 30 * 24 * time.Hour,
    "90d": 90 * 24 * time.Hour, "180d": 180 * 24 * time.Hour,
    "1y": 365 * 24 * time.Hour,
}

type Repo interface {
	ProductExists(ctx context.Context, productID int64) (bool, error)
	QueryDaily(ctx context.Context, productID int64, from time.Time) ([]chart.DailyPoint, error)
}

type DealService interface {
	Maturity(ctx context.Context, productID int64) string
	Verdict(ctx context.Context, productID int64) string
}

type Handler struct {
	repo Repo
	deal DealService
}

func NewHandler(repo Repo, deal DealService) *Handler {
	return &Handler{
		repo: repo,
		deal: deal,
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// HandleChart phục vụ GET /v1/products/{id}/chart?range=90d.
func (h *Handler) HandleChart(w http.ResponseWriter, req *http.Request) {
    id, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
    if err != nil {
        writeErr(w, http.StatusBadRequest, "invalid product id")
        return
    }
    rng := req.URL.Query().Get("range")
    if rng == "" {
        rng = "90d"
    }
    window, ok := rangeWindows[rng]
    if !ok {
        writeErr(w, http.StatusBadRequest, "invalid range")
        return
    }
    exists, err := h.repo.ProductExists(req.Context(), id)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    if !exists {
        writeErr(w, http.StatusNotFound, "product not found")
        return
    }
    from, to := time.Now().Add(-window), time.Now()
    daily, err := h.repo.QueryDaily(req.Context(), id, from) // đọc price_daily
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    if daily == nil {
        daily = []chart.DailyPoint{}
    }
    ann := chart.Build(daily, from, to)
    mat := h.deal.Maturity(req.Context(), id)        // TASK-DEAL-002
    ann.Accumulating = mat == "WARMING"
    ann.Verdict = h.deal.Verdict(req.Context(), id)  // TASK-DEAL-001
    if mat == "NEW" {
        ann.Verdict = "UNKNOWN" // <14 ngày: không kết luận (DEC-DEAL-23)
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    _ = json.NewEncoder(w).Encode(ChartResponse{
        ProductID: id, Range: rng, Maturity: mat, Daily: daily, Annotations: ann,
    })
}
