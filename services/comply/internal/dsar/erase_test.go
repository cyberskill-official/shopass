package dsar

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErase_KeepsConsentLog(t *testing.T) {
	s, mu, mt, mb, mc := setupMocks()

	mu.view = AccountView{UserID: 1, Email: "a@a.com"}
	mt.prods[1] = []ProductView{{ID: 10}}
	mc.c = []ConsentView{{UserID: 1, PurposeKey: "cart"}}

	res, err := s.Erase(context.Background(), 1)
	require.NoError(t, err)

	require.Equal(t, 1, res.WishlistDeleted)
	require.Equal(t, 1, mt.delCount)
	require.Equal(t, "anon@example.com", mu.view.Email)
	require.Equal(t, 1, mb.anonCount)
	require.True(t, res.ConsentLogRetained)

	// Consent still there
	require.Len(t, mc.c, 1)
}

func TestErase_Idempotent(t *testing.T) {
	s, _, mt, mb, _ := setupMocks()
	mt.prods[1] = []ProductView{{ID: 10}}

	_, err1 := s.Erase(context.Background(), 1)
	_, err2 := s.Erase(context.Background(), 1)

	require.NoError(t, err1)
	require.NoError(t, err2)

	require.Equal(t, 1, mt.delCount)  // Should be 1 because prods were cleared
	require.Equal(t, 2, mb.anonCount) // Mock blindly increments, but ok for test
}
