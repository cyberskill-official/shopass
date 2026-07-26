package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"shopass/services/deal/internal/chart"
	"shopass/services/deal/internal/itemurl"
)

// ProductLookup resolves a marketplace SKU to an internal product id.
type ProductLookup interface {
	FindProductID(ctx context.Context, platformCode, platformItemID string) (productID int64, found bool, err error)
}

type FakeSaleCheckHandler struct {
	lookup ProductLookup
	repo   Repo
	deal   DealService
}

func NewFakeSaleCheckHandler(lookup ProductLookup, repo Repo, deal DealService) *FakeSaleCheckHandler {
	return &FakeSaleCheckHandler{lookup: lookup, repo: repo, deal: deal}
}

type fakeSaleCheckBody struct {
	ItemURL string `json:"item_url"`
}

type FakeSaleCheckResponse struct {
	Tracked      bool               `json:"tracked"`
	Platform     string             `json:"platform,omitempty"`
	ProductID    int64              `json:"product_id,omitempty"`
	Maturity     string             `json:"maturity,omitempty"`
	Verdict      string             `json:"verdict,omitempty"`
	CurrentPrice int64              `json:"current_price,omitempty"`
	Median90     int64              `json:"median90,omitempty"`
	TrailingMin  int64              `json:"trailing_min,omitempty"`
	Daily        []chart.DailyPoint `json:"daily,omitempty"`
	Message      string             `json:"message,omitempty"`
}

func (h *FakeSaleCheckHandler) HandleFakeSaleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body fakeSaleCheckBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	parsed, ok := itemurl.Parse(body.ItemURL)
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid item_url")
		return
	}

	productID, found, err := h.lookup.FindProductID(r.Context(), parsed.Platform, parsed.PlatformItemID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !found {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(FakeSaleCheckResponse{
			Tracked:  false,
			Platform: parsed.Platform,
			Message:  "not_tracked",
		})
		return
	}

	from, to := time.Now().Add(-90*24*time.Hour), time.Now()
	daily, err := h.repo.QueryDaily(r.Context(), productID, from)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if daily == nil {
		daily = []chart.DailyPoint{}
	}
	tail, err := h.repo.QueryRawTail(r.Context(), productID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	daily = mergeRawTail(daily, tail)
	ann := chart.Build(daily, from, to)
	mat := h.deal.Maturity(r.Context(), productID)
	ann.Accumulating = mat == "WARMING"
	ann.Verdict = h.deal.Verdict(r.Context(), productID)
	if mat == "NEW" {
		ann.Verdict = "UNKNOWN"
	}
	var current int64
	if len(daily) > 0 {
		current = daily[len(daily)-1].CloseP
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(FakeSaleCheckResponse{
		Tracked:      true,
		Platform:     parsed.Platform,
		ProductID:    productID,
		Maturity:     mat,
		Verdict:      ann.Verdict,
		CurrentPrice: current,
		Median90:     ann.Median90,
		TrailingMin:  ann.TrailingMin,
		Daily:        daily,
	})
}
