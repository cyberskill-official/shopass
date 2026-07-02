package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"shopass/services/price/internal/price"
)

// CompareItem is one platform's current price in a cross-platform comparison.
type CompareItem struct {
	PlatformCode string `json:"platform_code"`
	PlatformName string `json:"platform_name"`
	ProductID    int64  `json:"product_id"`
	Price        int64  `json:"price"` // VND
	Currency     string `json:"currency"`
	TS           string `json:"ts"` // RFC3339, freshness of this platform's price
	ItemURL      string `json:"item_url"`
	IsCheapest   bool   `json:"is_cheapest"`
}

// CompareResponse is the payload of GET /v1/compare.
type CompareResponse struct {
	CanonicalKey string        `json:"canonical_key"`
	Items        []CompareItem `json:"items"`
}

// HandleCompare serves GET /v1/compare?canonical_key=... (FR-PRICE-004): the
// current price of the same physical product across Shopee/TikTok/Lazada, with a
// server-computed cheapest flag. Auth is centralized at the gateway; this handler
// does not verify tokens itself (§1 #1).
func (h *Handler) HandleCompare(w http.ResponseWriter, req *http.Request) {
	key := strings.TrimSpace(req.URL.Query().Get("canonical_key"))
	if key == "" {
		writeErr(w, http.StatusBadRequest, "canonical_key là bắt buộc") // §1 #2, no DB query
		return
	}

	rows, err := h.repo.CompareByCanonicalKey(req.Context(), key)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "lỗi truy vấn")
		return
	}
	if len(rows) == 0 {
		// §1 #10: unknown key or no listing -> 404, not an empty 200 array,
		// so the client can tell "no such product" from "exists but empty".
		writeErr(w, http.StatusNotFound, "không có sản phẩm cho canonical_key này")
		return
	}

	items := toCompareItems(rows)
	markCheapest(items) // set is_cheapest server-side (DEC-PRICE-42)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(CompareResponse{CanonicalKey: key, Items: items})
}

func toCompareItems(rows []price.CompareRow) []CompareItem {
	items := make([]CompareItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, CompareItem{
			PlatformCode: r.PlatformCode,
			PlatformName: price.DisplayName(r.PlatformCode),
			ProductID:    r.ProductID,
			Price:        r.Price,
			Currency:     "VND",
			TS:           r.TS.UTC().Format(time.RFC3339),
			ItemURL:      r.PlatformItem,
		})
	}
	return items
}

// markCheapest flags every row whose price equals the minimum, so ties all win
// the badge (§1 #6). A single-platform key gets is_cheapest=true (DEC-PRICE-44).
func markCheapest(items []CompareItem) {
	if len(items) == 0 {
		return
	}
	min := items[0].Price
	for _, it := range items {
		if it.Price < min {
			min = it.Price
		}
	}
	for i := range items {
		items[i].IsCheapest = items[i].Price == min
	}
}
