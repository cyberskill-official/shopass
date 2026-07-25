package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestReset_NoEnumeration(t *testing.T) {
	suite := newTestLifecycleSuite(t)
	suite.seedUser(t, "exists@x.com")

	ctx := context.Background()
	err1 := suite.ls.RequestReset(ctx, "exists@x.com")
	err2 := suite.ls.RequestReset(ctx, "nope@x.com")

	require.NoError(t, err1)
	require.NoError(t, err2) // cùng phản hồi, không lộ tồn tại
}

func TestConfirmReset_OneTime(t *testing.T) {
	suite := newTestLifecycleSuite(t)
	suite.seedUser(t, "a@x.com")
	tok := suite.issueResetFor(t, "a@x.com")

	ctx := context.Background()
	require.NoError(t, suite.ls.ConfirmReset(ctx, tok, "newp@ss12345"))
	require.ErrorIs(t, suite.ls.ConfirmReset(ctx, tok, "againp@ss12"), ErrInvalidResetToken)
}

func TestConfirmReset_RevokesSessions(t *testing.T) {
	suite := newTestLifecycleSuite(t)
	uid := suite.seedUser(t, "a@x.com")

	ctx := context.Background()
	pair, err := suite.ts.IssueTokenPair(ctx, uid)
	require.NoError(t, err)

	tok := suite.issueResetFor(t, "a@x.com")
	require.NoError(t, suite.ls.ConfirmReset(ctx, tok, "newp@ss12345"))

	_, err = suite.ts.Refresh(ctx, pair.Refresh) // refresh cũ vô hiệu
	require.Error(t, err)
}
