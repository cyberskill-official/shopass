package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDelete_ErasesPII_AndLinks(t *testing.T) {
	suite := newTestLifecycleSuite(t)
	uid := suite.seedUserWithLinks(t, "chi@x.com")
	
	ctx := context.Background()
	require.NoError(t, suite.ls.DeleteAccount(ctx, uid))
	
	// manually fetch from repo
	_, ok := suite.repo.FindByIdentifier(ctx, "chi@x.com")
	require.False(t, ok) // "chi@x.com" no longer exists
	
	// find by ID to check tombstone
	var actualEmail, status string
	err := suite.repo.db.QueryRow(`SELECT email, status FROM app_user WHERE id = $1`, uid).Scan(&actualEmail, &status)
	require.NoError(t, err)
	
	expectedEmail := fmt.Sprintf("deleted_%d@tombstone.local", uid)
	require.Equal(t, expectedEmail, actualEmail)
	require.Equal(t, "deleted", status)
	
	links, err := suite.repo.ListPlatformAccountsByUser(ctx, uid)
	require.NoError(t, err)
	require.Empty(t, links)
}

func TestDelete_Idempotent(t *testing.T) {
	suite := newTestLifecycleSuite(t)
	uid := suite.seedUser(t, "a@x.com")
	
	ctx := context.Background()
	require.NoError(t, suite.ls.DeleteAccount(ctx, uid))
	require.NoError(t, suite.ls.DeleteAccount(ctx, uid)) // lần hai không lỗi
}
