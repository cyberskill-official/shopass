package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"shopass/services/b2b/internal/seller"
	"shopass/services/b2b/internal/trend"
)

type SellerPriceSource interface {
	// AvgOwnPrice returns the seller's own verified SKU close price for the day.
	AvgOwnPrice(ctx context.Context, sellerOrgID int64, shopID string, day time.Time) (int64, error)
}

type SellerHandler struct {
	Ownership *seller.Ownership
	Trend     reportTrend
	Prices    SellerPriceSource
	Metrics   *sellerMetrics
}

// reportTrend is the published-cell reader from TASK-B2B-001.
type reportTrend interface {
	QueryCells(ctx context.Context, categoryID int64, platformID int16, from, to time.Time) ([]trend.MarketTrendCell, error)
}

type sellerMetrics struct {
	Served int64
	Denied map[string]int64
}

func newSellerMetrics() *sellerMetrics {
	return &sellerMetrics{Denied: make(map[string]int64)}
}

func (h *SellerHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/b2b/seller/position", h.handlePosition)
}

func (h *SellerHandler) handlePosition(w http.ResponseWriter, r *http.Request) {
	orgID, err := strconv.ParseInt(r.Header.Get("X-B2B-Org-Id"), 10, 64)
	if err != nil || orgID <= 0 {
		writeErr(w, http.StatusBadRequest, "bad_org", "missing X-B2B-Org-Id", nil)
		return
	}
	q := r.URL.Query()
	shopID := q.Get("shop_id")
	if shopID == "" {
		writeErr(w, http.StatusBadRequest, "bad_params", "shop_id required", nil)
		return
	}
	catID, err := strconv.ParseInt(q.Get("category_id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_params", "category_id required", nil)
		return
	}
	pf, err := strconv.ParseInt(q.Get("platform_id"), 10, 16)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_params", "platform_id required", nil)
		return
	}
	day, err := time.Parse("2006-01-02", q.Get("day"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_params", "day must be YYYY-MM-DD", nil)
		return
	}
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)

	if err := h.Ownership.AssertOwned(r.Context(), orgID, shopID); err != nil {
		var nv seller.ErrNotVerifiedOwner
		if errors.As(err, &nv) {
			h.deny("not_owner")
			writeErr(w, http.StatusForbidden, "not_verified_owner", "shop ownership not verified", nil)
			return
		}
		writeErr(w, http.StatusInternalServerError, "ownership_error", err.Error(), nil)
		return
	}

	cells, err := h.Trend.QueryCells(r.Context(), catID, int16(pf), day, day.AddDate(0, 0, 1))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "trend_error", err.Error(), nil)
		return
	}
	if len(cells) == 0 {
		h.deny("insufficient_market_data")
		writeErr(w, http.StatusUnprocessableEntity, "insufficient_market_data", "not enough anonymized market data", nil)
		return
	}
	c := cells[0]
	if c.Suppressed || c.MedianP == nil || c.P25P == nil || c.P75P == nil {
		h.deny("insufficient_market_data")
		writeErr(w, http.StatusUnprocessableEntity, "insufficient_market_data", "not enough anonymized market data", nil)
		return
	}

	ownPrice, err := h.Prices.AvgOwnPrice(r.Context(), orgID, shopID, day)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "price_error", err.Error(), nil)
		return
	}
	pos, err := seller.ComputePosition(ownPrice, seller.MarketBand{
		P25: *c.P25P, Median: *c.MedianP, P75: *c.P75P,
	}, false, true)
	if err != nil {
		if errors.Is(err, seller.ErrInsufficientMarket) {
			writeErr(w, http.StatusUnprocessableEntity, "insufficient_market_data", err.Error(), nil)
			return
		}
		writeErr(w, http.StatusForbidden, "forbidden", err.Error(), nil)
		return
	}
	pos.CategoryID = catID
	pos.PlatformID = int16(pf)
	pos.Day = day.Format("2006-01-02")
	if h.Metrics != nil {
		h.Metrics.Served++
	}
	writeJSON(w, http.StatusOK, pos)
}

func (h *SellerHandler) deny(reason string) {
	if h.Metrics == nil {
		h.Metrics = newSellerMetrics()
	}
	h.Metrics.Denied[reason]++
}
