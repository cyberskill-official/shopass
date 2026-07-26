package api

import "context"

// FeatureGate checks subscription entitlements via billsvc.
type FeatureGate interface {
	Check(ctx context.Context, userID int64, featureKey string, usage *int64) (allowed bool, limitReached bool, err error)
}

// BillGate adapts a billclient-like checker to FeatureGate.
type BillGate struct {
	CheckFn func(ctx context.Context, userID int64, featureKey string, usage *int64) (allowed bool, limitReached bool, err error)
}

func (g BillGate) Check(ctx context.Context, userID int64, featureKey string, usage *int64) (bool, bool, error) {
	return g.CheckFn(ctx, userID, featureKey, usage)
}
