package api

import (
	"encoding/json"
	"net/http"

	"shopass/region"
	"shopass/services/cart/internal/optimizer"
	"shopass/services/cart/internal/optimizer/stacking"
)

type OptimizeRequest struct {
	PlatformID int16                `json:"platform"`
	Country    string               `json:"country"` // Usually derived from platform, passed via payload or context for now
	Items      []optimizer.CartItem `json:"items"`
	Vouchers   optimizer.Vouchers   `json:"vouchers"`
}

type OptimizeHandler struct {
}

func NewOptimizeHandler() *OptimizeHandler {
	return &OptimizeHandler{}
}

func (h *OptimizeHandler) OptimizeCart(w http.ResponseWriter, r *http.Request) {
	var req OptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Load policy based on country
	policy := region.CountryPolicy{
		Country:                req.Country,
		VoucherStackingAllowed: req.Country == "VN" || req.Country == "", // Default to true if not specified, else VN only
	}
	if req.Country == "MY" || req.Country == "PH" {
		policy.FreeshipGroupedWithPlatform = true
		policy.VoucherStackingAllowed = false // MY/PH no-stack
	}

	rules := stacking.RulesForCountry(req.Country, policy)

	result := optimizer.OptimizeCart(req.Items, req.Vouchers, rules)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
