package apikey

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func setupKeyWith(t *testing.T, tier string, revoked bool) (*Auth, string) {
	t.Helper()
	store := NewMemoryKeyStore()
	prefix, secret, hash, err := NewKey()
	require.NoError(t, err)
	store.Put(&APIKey{
		ID: 7, Prefix: prefix, SecretHash: hash, OrgName: "acme",
		Tier: tier, RatePerMin: 60, MonthlyQuota: 1000, Revoked: revoked,
	})
	return NewAuth(store), Format(prefix, secret)
}

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

func TestAuth_ValidKey_OK(t *testing.T) {
	a, raw := setupKeyWith(t, "pro", false)
	k, err := a.Authenticate(context.Background(), raw)
	require.NoError(t, err)
	require.Equal(t, "pro", k.Tier)
}

func TestAuth_WrongSecret_401(t *testing.T) {
	a, raw := setupKeyWith(t, "pro", false)
	parts := strings.Split(raw, ".")
	bad := parts[0] + ".deadbeef"
	_, err := a.Authenticate(context.Background(), bad)
	require.ErrorIs(t, err, ErrUnauthorized)
}

func TestAuth_Revoked_401(t *testing.T) {
	a, raw := setupKeyWith(t, "pro", true)
	_, err := a.Authenticate(context.Background(), raw)
	require.ErrorIs(t, err, ErrUnauthorized)
}

func TestAuth_RevocationTakesEffect(t *testing.T) {
	a, raw := setupKeyWith(t, "pro", false)
	_, err := a.Authenticate(context.Background(), raw)
	require.NoError(t, err)
	prefix, _, err := ParsePresented(raw)
	require.NoError(t, err)
	require.NoError(t, a.Revoke(context.Background(), prefix))
	a.AdvanceCache(61 * time.Second)
	_, err = a.Authenticate(context.Background(), raw)
	require.ErrorIs(t, err, ErrUnauthorized)
}
