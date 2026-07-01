package region

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func mustLoad(t *testing.T, path string) *Registry {
	t.Helper()
	reg, err := Load(path)
	require.NoError(t, err)
	return reg
}

func TestPolicy_VNStacks_MYPHDoNot(t *testing.T) {
	reg := mustLoad(t, "config/countries.yaml")
	require.True(t, reg.Policy("VN").VoucherStackingAllowed)
	require.False(t, reg.Policy("MY").VoucherStackingAllowed)
	require.False(t, reg.Policy("PH").VoucherStackingAllowed)
}

func TestPolicy_UnknownCountry_Restrictive(t *testing.T) {
	reg := mustLoad(t, "config/countries.yaml")
	p := reg.Policy("XX")
	require.False(t, p.VoucherStackingAllowed)
	require.Empty(t, p.AffiliateChannelsAllowed) // không bật kênh nào
}

func TestPolicy_AffiliateChannel_PerCountry(t *testing.T) {
	reg := mustLoad(t, "config/countries.yaml")
	require.Contains(t, reg.Policy("VN").AffiliateChannelsAllowed, ChannelExtension)
	require.NotContains(t, reg.Policy("ID").AffiliateChannelsAllowed, ChannelExtension)
}
