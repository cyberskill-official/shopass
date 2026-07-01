package orchestrator

import (
	"context"
	"time"
)

type PriceSnapshot struct {
	ProductID int64
	TS        time.Time
	Price     int64
	ListPrice *int64
	Stock     *int32
	Sold      *int32
	FlashSale bool
}

type ScrapeJob struct {
	ProductID      int64
	PlatformID     int16
	PlatformItemID string // platform-specific ref, e.g. "itemID:shopID" for Shopee
	Tier           Tier
	Attempts       int
}

type PlatformAdapter interface {
	// Fetch lấy snapshot giá cho một job. Lỗi mạng/parse trả error;
	// orchestrator quyết định retry/backoff dựa trên error.
	Fetch(ctx context.Context, job ScrapeJob) (PriceSnapshot, error)
	PlatformID() int16
}

// PriceRepo is a minimal interface to the Price service to avoid circular dependencies
// In real app, this might be a gRPC client or a direct DB repo if in a monolith
type PriceRepo interface {
	InsertSnapshot(ctx context.Context, snap PriceSnapshot) (bool, error)
}
