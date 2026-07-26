package seller

import (
	"context"
	"fmt"
	"sync"
)

// ErrNotVerifiedOwner maps to HTTP 403 (DEC-B2B-24).
type ErrNotVerifiedOwner struct {
	ShopID string
}

func (e ErrNotVerifiedOwner) Error() string {
	return fmt.Sprintf("seller: shop %s not verified for org", e.ShopID)
}

type OwnedSKU struct {
	SellerOrgID int64
	ShopID      string
	ProductID   int64
	Verified    bool
}

// OwnershipStore checks verified shop ownership.
type OwnershipStore interface {
	CountVerified(ctx context.Context, sellerOrgID int64, shopID string) (int, error)
	OwnsProduct(ctx context.Context, sellerOrgID int64, shopID string, productID int64) (bool, error)
}

type Ownership struct {
	Store OwnershipStore
}

func (o *Ownership) AssertOwned(ctx context.Context, sellerOrgID int64, shopID string) error {
	n, err := o.Store.CountVerified(ctx, sellerOrgID, shopID)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotVerifiedOwner{ShopID: shopID}
	}
	return nil
}

// MemoryOwnership is an in-process store for tests / noop deploys.
type MemoryOwnership struct {
	mu   sync.Mutex
	rows []OwnedSKU
}

func NewMemoryOwnership(rows ...OwnedSKU) *MemoryOwnership {
	return &MemoryOwnership{rows: append([]OwnedSKU{}, rows...)}
}

func (m *MemoryOwnership) CountVerified(_ context.Context, sellerOrgID int64, shopID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, r := range m.rows {
		if r.SellerOrgID == sellerOrgID && r.ShopID == shopID && r.Verified {
			n++
		}
	}
	return n, nil
}

func (m *MemoryOwnership) OwnsProduct(_ context.Context, sellerOrgID int64, shopID string, productID int64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rows {
		if r.SellerOrgID == sellerOrgID && r.ShopID == shopID && r.ProductID == productID && r.Verified {
			return true, nil
		}
	}
	return false, nil
}
