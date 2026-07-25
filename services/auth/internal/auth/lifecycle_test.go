package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuspended_CannotLogin_TokensRevoked(t *testing.T) {
	suite := newTestLifecycleSuite(t)
	uid := suite.seedActiveUser(t, "a@x.com", "p@ss12345")

	ctx := context.Background()
	pair, err := suite.ts.IssueTokenPair(ctx, uid)
	require.NoError(t, err)

	require.NoError(t, suite.ls.SetStatus(ctx, uid, "suspended"))

	_, err = suite.ts.Login(ctx, "a@x.com", "p@ss12345")
	require.ErrorIs(t, err, ErrAccountNotActive)

	_, rerr := suite.ts.Refresh(ctx, pair.Refresh)
	require.Error(t, rerr) // refresh thu hồi
}
