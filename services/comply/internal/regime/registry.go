package regime

import (
	"context"

	"shopass/services/comply/internal/gating"
)

type CountryGater interface {
	Allow(country, gate string) bool
	Value(country, gate string) (string, bool)
}

type Registry struct {
	gating CountryGater
	byCode map[string]RegimeAdapter
}

func NewRegistry(g CountryGater, adapters ...RegimeAdapter) *Registry {
	m := map[string]RegimeAdapter{}
	for _, a := range adapters {
		m[a.Code()] = a
	}
	return &Registry{gating: g, byCode: m}
}

// ForCountry selects a regime via COMPLY-006 — never hardcodes country→regime.
func (r *Registry) ForCountry(country string) (RegimeAdapter, error) {
	if r.gating == nil || !r.gating.Allow(country, gating.GateDataProtectionRegime) {
		return nil, ErrCountryNotOpen
	}
	code, ok := r.gating.Value(country, gating.GateDataProtectionRegime)
	if !ok || code == "" {
		return nil, ErrCountryNotOpen
	}
	a, ok := r.byCode[code]
	if !ok {
		return nil, ErrUnsupportedRegime
	}
	return a, nil
}

// For is an alias of ForCountry matching the COMPLY-007 API name.
func (r *Registry) For(country string) (RegimeAdapter, error) {
	return r.ForCountry(country)
}

// Profile resolves the regime profile for a country (deny-by-default).
func (r *Registry) Profile(_ context.Context, country string) (RegimeProfile, error) {
	a, err := r.ForCountry(country)
	if err != nil {
		return RegimeProfile{}, err
	}
	return a.Profile(), nil
}

// ResolveProfile is the package-level helper used by complysvc wiring.
func ResolveProfile(ctx context.Context, r *Registry, country string) (RegimeProfile, error) {
	if r == nil {
		return RegimeProfile{}, ErrCountryNotOpen
	}
	return r.Profile(ctx, country)
}
