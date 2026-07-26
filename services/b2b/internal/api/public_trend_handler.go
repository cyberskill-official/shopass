package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"shopass/services/b2b/internal/apikey"
	"shopass/services/b2b/internal/trend"
)

type PublicTrendHandler struct {
	Auth    *apikey.Auth
	Limit   *apikey.RateLimiter
	Usage   apikey.UsageStore
	Trend   reportTrend
	Now     func() time.Time
}

func (h *PublicTrendHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /public/v1/trends", h.handleTrends)
}

func (h *PublicTrendHandler) handleTrends(w http.ResponseWriter, r *http.Request) {
	const endpoint = "/public/v1/trends"
	now := time.Now().UTC()
	if h.Now != nil {
		now = h.Now()
	}
	raw := r.Header.Get("X-API-Key")
	key, err := h.Auth.Authenticate(r.Context(), raw)
	if err != nil {
		h.record(0, endpoint, http.StatusUnauthorized, now)
		writeErr(w, http.StatusUnauthorized, "unauthorized", "invalid or revoked API key", nil)
		return
	}
	if !apikey.AllowedEndpoint(key.Tier, endpoint) {
		h.record(key.ID, endpoint, http.StatusForbidden, now)
		writeErr(w, http.StatusForbidden, "forbidden", "tier cannot access endpoint", nil)
		return
	}
	if ok, retry := h.Limit.Allow(key.ID, key.RatePerMin); !ok {
		h.record(key.ID, endpoint, http.StatusTooManyRequests, now)
		w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())+1))
		writeErr(w, http.StatusTooManyRequests, "rate_limited", "per-minute rate exceeded", nil)
		return
	}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	used, err := h.Usage.CountMonth(r.Context(), key.ID, monthStart)
	if err == nil && used >= key.MonthlyQuota {
		h.record(key.ID, endpoint, http.StatusTooManyRequests, now)
		writeErr(w, http.StatusTooManyRequests, "quota_exceeded", "monthly quota exceeded", nil)
		return
	}

	catID, err := strconv.ParseInt(r.URL.Query().Get("category_id"), 10, 64)
	if err != nil {
		h.record(key.ID, endpoint, http.StatusBadRequest, now)
		writeErr(w, http.StatusBadRequest, "bad_params", "category_id required", nil)
		return
	}
	pf, err := strconv.ParseInt(r.URL.Query().Get("platform_id"), 10, 16)
	if err != nil {
		h.record(key.ID, endpoint, http.StatusBadRequest, now)
		writeErr(w, http.StatusBadRequest, "bad_params", "platform_id required", nil)
		return
	}
	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		h.record(key.ID, endpoint, http.StatusBadRequest, now)
		writeErr(w, http.StatusBadRequest, "bad_params", "from must be YYYY-MM-DD", nil)
		return
	}
	to, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil || !to.After(from) {
		h.record(key.ID, endpoint, http.StatusBadRequest, now)
		writeErr(w, http.StatusBadRequest, "bad_params", "to must be after from", nil)
		return
	}
	// free tier: max 7-day window
	if key.Tier == "free" && to.Sub(from) > 7*24*time.Hour {
		h.record(key.ID, endpoint, http.StatusForbidden, now)
		writeErr(w, http.StatusForbidden, "forbidden", "free tier limited to 7-day windows", nil)
		return
	}

	cells, err := h.Trend.QueryCells(r.Context(), catID, int16(pf), from.UTC(), to.UTC())
	if err != nil {
		h.record(key.ID, endpoint, http.StatusInternalServerError, now)
		writeErr(w, http.StatusInternalServerError, "trend_error", err.Error(), nil)
		return
	}
	out := make([]map[string]any, 0, len(cells))
	for _, c := range cells {
		out = append(out, publicCell(c))
	}
	h.record(key.ID, endpoint, http.StatusOK, now)
	writeJSON(w, http.StatusOK, map[string]any{"cells": out})
}

func publicCell(c trend.MarketTrendCell) map[string]any {
	m := map[string]any{
		"category_id": c.CategoryID,
		"platform_id": c.PlatformID,
		"day":         c.Day.UTC().Format("2006-01-02"),
		"sku_count":   c.SKUCount,
	}
	if c.MedianP != nil {
		m["median_p"] = *c.MedianP
	}
	if c.P25P != nil {
		m["p25_p"] = *c.P25P
	}
	if c.P75P != nil {
		m["p75_p"] = *c.P75P
	}
	if c.AvgDiscountPct != nil {
		m["avg_discount_pct"] = *c.AvgDiscountPct
	}
	return m
}

func (h *PublicTrendHandler) record(keyID int64, endpoint string, status int, ts time.Time) {
	if h.Usage == nil {
		return
	}
	_ = h.Usage.Record(context.Background(), apikey.UsageEvent{
		APIKeyID: keyID, Endpoint: endpoint, StatusCode: status, TS: ts,
	})
}

// Ensure public router never registers raw/user-level paths (test helper).
func PublicRoutePrefixes() []string {
	return []string{"/public/v1/trends"}
}

func HasForbiddenPublicRoute(paths []string) bool {
	forbidden := []string{"price_snapshot", "wishlist", "alert", "cart", "/v1/auth"}
	for _, p := range paths {
		low := strings.ToLower(p)
		for _, f := range forbidden {
			if strings.Contains(low, f) {
				return true
			}
		}
	}
	return false
}
