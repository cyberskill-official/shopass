package seller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOwnership_Verified_OK(t *testing.T) {
	o := &Ownership{Store: NewMemoryOwnership(OwnedSKU{1, "shopA", 100, true})}
	require.NoError(t, o.AssertOwned(context.Background(), 1, "shopA"))
}

func TestOwnership_NotVerified_403(t *testing.T) {
	o := &Ownership{Store: NewMemoryOwnership(OwnedSKU{1, "shopA", 100, false})}
	err := o.AssertOwned(context.Background(), 1, "shopA")
	var e ErrNotVerifiedOwner
	require.ErrorAs(t, err, &e)
}

func TestOwnership_OtherShop_403(t *testing.T) {
	o := &Ownership{Store: NewMemoryOwnership(OwnedSKU{1, "shopA", 100, true})}
	err := o.AssertOwned(context.Background(), 1, "shopB")
	var e ErrNotVerifiedOwner
	require.ErrorAs(t, err, &e)
}
