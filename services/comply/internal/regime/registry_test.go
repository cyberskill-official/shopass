package regime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"shopass/services/comply/internal/gating"
)

type stubGate struct {
	allow map[string]bool
	val   map[string]string
}

func (s stubGate) Allow(country, gate string) bool {
	return s.allow[country+"|"+gate]
}
func (s stubGate) Value(country, gate string) (string, bool) {
	v, ok := s.val[country+"|"+gate]
	return v, ok
}

func TestForCountry_OpenVN(t *testing.T) {
	g := stubGate{
		allow: map[string]bool{"VN|" + gating.GateDataProtectionRegime: true},
		val:   map[string]string{"VN|" + gating.GateDataProtectionRegime: "VN_PDPL"},
	}
	reg := NewRegistry(g, DefaultAdapters()...)
	a, err := reg.ForCountry("VN")
	require.NoError(t, err)
	require.Equal(t, "VN_PDPL", a.Code())
	p := a.Profile()
	require.Equal(t, 72, p.BreachWindowHours)
	require.Equal(t, 60, p.DPIAFilingDays)
	require.Equal(t, 30, p.DSARDays)
	require.Contains(t, p.ConsentLanguages, "vi")
	require.NotEmpty(t, p.Notes)
}

func TestForCountry_ClosedMY(t *testing.T) {
	reg := NewRegistry(stubGate{}, DefaultAdapters()...)
	_, err := reg.ForCountry("MY")
	require.ErrorIs(t, err, ErrCountryNotOpen)
}

func TestForCountry_UnsupportedExplicit(t *testing.T) {
	g := stubGate{
		allow: map[string]bool{"XX|" + gating.GateDataProtectionRegime: true},
		val:   map[string]string{"XX|" + gating.GateDataProtectionRegime: "UNKNOWN"},
	}
	reg := NewRegistry(g, DefaultAdapters()...)
	_, err := reg.ForCountry("XX")
	require.ErrorIs(t, err, ErrUnsupportedRegime)
}

func TestTH_ConsentLanguages(t *testing.T) {
	p := THPDPA{}.Profile()
	require.Equal(t, []string{"th", "en"}, p.ConsentLanguages)
	require.Equal(t, 30, p.DSARDays) // baseline, not invented foreign SLA
}

func TestID_InheritsBaselineSLAs(t *testing.T) {
	p := IDPDP{}.Profile()
	b := baseline()
	require.Equal(t, b.BreachWindowHours, p.BreachWindowHours)
	require.Equal(t, b.DSARDays, p.DSARDays)
	require.Equal(t, []string{"id", "en"}, p.ConsentLanguages)
}

func TestResolveProfile_Helper(t *testing.T) {
	g := stubGate{
		allow: map[string]bool{"VN|" + gating.GateDataProtectionRegime: true},
		val:   map[string]string{"VN|" + gating.GateDataProtectionRegime: "VN_PDPL"},
	}
	reg := NewRegistry(g, DefaultAdapters()...)
	p, err := ResolveProfile(context.Background(), reg, "VN")
	require.NoError(t, err)
	require.Equal(t, "VN_PDPL", p.Code)
}
