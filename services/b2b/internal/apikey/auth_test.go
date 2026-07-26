package apikey

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKey_RoundTrip(t *testing.T) {
	prefix, secret, hash, err := NewKey()
	require.NoError(t, err)
	raw := Format(prefix, secret)
	p, s, err := ParsePresented(raw)
	require.NoError(t, err)
	require.Equal(t, prefix, p)
	require.True(t, Verify(s, hash))
	require.False(t, Verify("wrong", hash))
}
