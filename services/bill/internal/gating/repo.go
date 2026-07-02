package gating

import "context"

type Repo interface {
	LimitFor(ctx context.Context, tier string, featureKey string) (int64, error)
	CountUsage(ctx context.Context, userID int64, featureKey string) (int64, error)
}
