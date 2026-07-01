package region

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func regWithFlag(t *testing.T, flag string, values map[string]bool) *Registry {
	t.Helper()
	reg := &Registry{}
	reg.SetFlagForTest(flag, values)
	return reg
}

func TestLookup_PerCountry(t *testing.T) {
	reg := regWithFlag(t, "real_sale_v2", map[string]bool{"VN": true, "MY": false})
	require.True(t, reg.Lookup("real_sale_v2", "VN"))
	require.False(t, reg.Lookup("real_sale_v2", "MY"))
}

func TestLookup_MissingCountry_False(t *testing.T) {
	reg := mustLoad(t, "config/countries.yaml")
	require.False(t, reg.Lookup("any_flag", "")) // thiếu country = false
}
