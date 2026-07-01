package ecom

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func findKey(obs []EcommerceObligation, key string) *EcommerceObligation {
	for _, o := range obs {
		if o.ObligationKey == key {
			return &o
		}
	}
	return nil
}

func TestObligation_DraftMarkedProvisional(t *testing.T) {
	s := setupSeeded()
	ctx := context.Background()
	obs, _ := s.Obligations(ctx)
	aff := findKey(obs, "affiliate_disclosure")
	require.NotNil(t, aff)
	require.Equal(t, "DRAFT_2025", aff.SourceLaw)
	require.Contains(t, aff.DescriptionVi, "cho luat chot")
}

func TestObligation_MarkApproved(t *testing.T) {
	s := setupSeeded()
	ctx := context.Background()
	require.NoError(t, s.MarkObligation(ctx, "moit_registration", "approved"))
	obs, _ := s.Obligations(ctx)
	require.Equal(t, "approved", findKey(obs, "moit_registration").Status)
}

func TestObligation_InvalidStatusRejected(t *testing.T) {
	s := setupSeeded()
	ctx := context.Background()
	err := s.MarkObligation(ctx, "moit_registration", "rejected")
	require.Error(t, err) // CHECK status
}
