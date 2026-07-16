package gating

import (
	"context"
	"errors"

	"shopass/services/bill/internal/bill"
)

var ErrLimitReached = errors.New("feature limit reached")

// We mock SubscriptionService interface for now to rely on TASK-BILL-001 concepts
type SubscriptionService interface {
	GetActive(ctx context.Context, userID int64) (bill.Subscription, bool, error)
}
type PlanCatalog interface {
	TierOf(planID int16) string
}

type Gate struct {
	repo  Repo
	subs  SubscriptionService
	plans PlanCatalog
}

func NewGate(repo Repo, subs SubscriptionService, plans PlanCatalog) *Gate {
	return &Gate{repo: repo, subs: subs, plans: plans}
}

func (g *Gate) Allow(ctx context.Context, userID int64, featureKey string) (bool, error) {
	tier := "free"
	if sub, ok, err := g.subs.GetActive(ctx, userID); err == nil && ok {
		tier = g.plans.TierOf(sub.PlanID)
	}

	limit, err := g.repo.LimitFor(ctx, tier, featureKey)
	if err != nil {
		return false, nil // fail-safe to free/no-access
	}
	switch {
	case limit == 0:
		return false, nil // no access
	case limit < 0:
		return true, nil // unlimited
	default:
		used, err := g.repo.CountUsage(ctx, userID, featureKey)
		if err != nil {
			return false, nil
		}
		if used < limit {
			return true, nil
		}
		return false, ErrLimitReached
	}
}
