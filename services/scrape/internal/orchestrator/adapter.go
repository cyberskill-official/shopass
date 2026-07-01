package orchestrator

import (
	"context"
)

type PriceSnapshot struct {
	ProductID int64
	Price     int64
	FlashSale bool
}

type ScrapeJob struct {
	ProductID  int64
	PlatformID int16
	Tier       Tier
	Attempts   int
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
