package regime

import "shopass/services/comply/internal/gating"

type CountryGater interface {
	Allow(country, gate string) bool
	Value(country, gate string) (string, bool)
}

type Registry struct {
	gating CountryGater
	byCode map[string]Adapter
}

func NewRegistry(g CountryGater, adapters ...Adapter) *Registry {
	m := map[string]Adapter{}
	for _, a := range adapters {
		m[a.Code()] = a
	}
	return &Registry{gating: g, byCode: m}
}

func DefaultAdapters() []Adapter {
	return []Adapter{VNPDPL{}, IDPDP{}, THPDPA{}}
}

// ForCountry selects a regime via COMPLY-006 — never hardcodes country→regime.
func (r *Registry) ForCountry(country string) (Adapter, error) {
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
