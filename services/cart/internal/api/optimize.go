package api

import (
	"encoding/json"
	"net/http"

	"shopass/region"
	"shopass/services/cart/internal/optimizer"
	"shopass/services/cart/internal/optimizer/stacking"
)

const gateVoucherStacking = "voucher_stacking"

type CountryGater interface {
	Allow(country, gate string) bool
}

type denyAllGater struct{}

func (denyAllGater) Allow(country, gate string) bool {
	return false
}

type OptimizeRequest struct {
	PlatformID int16                `json:"platform"`
	Country    string               `json:"country"` // Usually derived from platform, passed via payload or context for now
	Items      []optimizer.CartItem `json:"items"`
	Vouchers   optimizer.Vouchers   `json:"vouchers"`
}

type OptimizeHandler struct {
	gater CountryGater
}

func NewOptimizeHandler(gaters ...CountryGater) *OptimizeHandler {
	gater := CountryGater(denyAllGater{})
	if len(gaters) > 0 && gaters[0] != nil {
		gater = gaters[0]
	}
	return &OptimizeHandler{gater: gater}
}

func (h *OptimizeHandler) OptimizeCart(w http.ResponseWriter, r *http.Request) {
	var req OptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	policy := region.CountryPolicy{
		Country:                req.Country,
		VoucherStackingAllowed: h.gater.Allow(req.Country, gateVoucherStacking),
	}
	policy.FreeshipGroupedWithPlatform = !policy.VoucherStackingAllowed

	rules := stacking.RulesForCountry(req.Country, policy)

	result := optimizer.OptimizeCart(req.Items, req.Vouchers, rules)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
