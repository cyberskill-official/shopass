package gating

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type memSrc struct{ rows []Rule }

func (m memSrc) Load(ctx context.Context) ([]Rule, error) { return m.rows, nil }

func TestRegistry_DenyByDefault(t *testing.T) {
	reg := NewRegistry(memSrc{rows: []Rule{
		{Country: "VN", GateKey: GateVoucherStacking, Allowed: true, Value: "allowed"},
		{Country: "MY", GateKey: GateVoucherStacking, Allowed: false},
	}})
	require.NoError(t, reg.Reload(context.Background()))
	require.True(t, reg.Allow("VN", GateVoucherStacking))
	require.False(t, reg.Allow("MY", GateVoucherStacking))
	require.False(t, reg.Allow("XX", GateVoucherStacking))
}

func TestRegistry_RegimeValue(t *testing.T) {
	reg := NewRegistry(memSrc{rows: []Rule{
		{Country: "ID", GateKey: GateDataProtectionRegime, Allowed: true, Value: "ID_PDP"},
	}})
	require.NoError(t, reg.Reload(context.Background()))
	v, ok := reg.Value("ID", GateDataProtectionRegime)
	require.True(t, ok)
	require.Equal(t, "ID_PDP", v)
}
