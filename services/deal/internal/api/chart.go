package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"shopass/services/deal/internal/chart"
)

type ChartResponse struct {
	ProductID   int64              `json:"product_id"`
	Range       string             `json:"range"`
	Maturity    string             `json:"maturity"` // MATURE | WARMING | NEW (FR-DEAL-002)
	Daily       []chart.DailyPoint `json:"daily"`
	Annotations chart.Annotations  `json:"annotations"`
}

var rangeWindows = map[string]time.Duration{
	"7d": 7 * 24 * time.Hour, "30d": 30 * 24 * time.Hour,
	"90d": 90 * 24 * time.Hour, "180d": 180 * 24 * time.Hour,
	"1y": 365 * 24 * time.Hour,
}

type Repo interface {
	// UserCanViewProduct is an ownership check, not a product-existence check.
	// Returning false must not reveal whether the product exists to the caller.
	UserCanViewProduct(ctx context.Context, userID, productID int64) (bool, error)
	QueryDaily(ctx context.Context, productID int64, from time.Time) ([]chart.DailyPoint, error)
	QueryRawTail(ctx context.Context, productID int64) ([]SnapshotPoint, error)
}

// SnapshotPoint is a recent raw price point that has not necessarily reached
// the price_daily continuous aggregate yet.
type SnapshotPoint struct {
	TS    time.Time
	Price int64
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func gatewayUserID(req *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(req.Header.Get("X-User-Id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func dayBucket(ts time.Time) time.Time {
	return ts.UTC().Truncate(24 * time.Hour)
}

// mergeRawTail makes the chart immediately useful after the first scrape. The
// continuous aggregate intentionally excludes its most recent window, so the
// fresh raw snapshots are folded into the daily series until the cagg catches
// up.
func mergeRawTail(daily []chart.DailyPoint, tail []SnapshotPoint) []chart.DailyPoint {
	for _, snapshot := range tail {
		bucket := dayBucket(snapshot.TS)
		matched := false
		for i := range daily {
			if !dayBucket(daily[i].Day).Equal(bucket) {
				continue
			}
			if snapshot.Price < daily[i].MinP {
				daily[i].MinP = snapshot.Price
			}
			if snapshot.Price > daily[i].MaxP {
				daily[i].MaxP = snapshot.Price
			}
			// The tail query is timestamp ordered, so the last value assigned is
			// the current close for this day.
			daily[i].CloseP = snapshot.Price
			matched = true
			break
		}
		if !matched {
			daily = append(daily, chart.DailyPoint{
				Day: bucket, MinP: snapshot.Price, MaxP: snapshot.Price, CloseP: snapshot.Price,
			})
		}
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day.Before(daily[j].Day) })
	return daily
}

// HandleChart phục vụ GET /v1/products/{id}/chart?range=90d.
func (h *Handler) HandleChart(w http.ResponseWriter, req *http.Request) {
	userID, ok := gatewayUserID(req)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid product id")
		return
	}

	// Check the user-product link before any chart work. A false result maps to
	// the same 404 as an unknown product, preventing product-ID enumeration.
	allowed, err := h.repo.UserCanViewProduct(req.Context(), userID, id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !allowed {
		writeErr(w, http.StatusNotFound, "product not found")
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

	from, to := time.Now().Add(-window), time.Now()
	daily, err := h.repo.QueryDaily(req.Context(), id, from) // đọc price_daily
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if daily == nil {
		daily = []chart.DailyPoint{}
	}
	tail, err := h.repo.QueryRawTail(req.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	daily = mergeRawTail(daily, tail)
	ann := chart.Build(daily, from, to)
	mat := h.deal.Maturity(req.Context(), id) // FR-DEAL-002
	ann.Accumulating = mat == "WARMING"
	ann.Verdict = h.deal.Verdict(req.Context(), id) // FR-DEAL-001
	if mat == "NEW" {
		ann.Verdict = "UNKNOWN" // <14 ngày: không kết luận (DEC-DEAL-23)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(ChartResponse{
		ProductID: id, Range: rng, Maturity: mat, Daily: daily, Annotations: ann,
	})
}
