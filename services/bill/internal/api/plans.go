package api

import (
	"context"

	"shopass/services/bill/internal/bill"
)

type sqlPlanCatalog struct {
	repo *bill.Repo
}

func NewSQLPlanCatalog(repo *bill.Repo) PlanCatalog {
	return sqlPlanCatalog{repo: repo}
}

func (c sqlPlanCatalog) ByTier(ctx context.Context, tier string) (Plan, bool) {
	p, ok := c.repo.PlanByTier(ctx, tier)
	if !ok {
		return Plan{}, false
	}
	return Plan{Tier: p.Tier, Price: p.Price}, true
}
