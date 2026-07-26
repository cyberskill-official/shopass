package regime

import (
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
