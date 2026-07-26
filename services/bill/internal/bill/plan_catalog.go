package bill

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// CatalogPlan is a row from plan_catalog used by checkout.
type CatalogPlan struct {
	ID    int16
	Tier  string
	Price int64
}

func (r *Repo) PlanByTier(ctx context.Context, tier string) (CatalogPlan, bool) {
	var p CatalogPlan
	err := r.pool.QueryRow(ctx, `
		SELECT id, tier, price FROM plan_catalog
		WHERE tier = $1 AND active = true`, tier).Scan(&p.ID, &p.Tier, &p.Price)
	if errors.Is(err, pgx.ErrNoRows) || err != nil {
		return CatalogPlan{}, false
	}
	return p, true
}

// EnsurePaidSubscription creates or renews an active subscription after IPN paid.
func (r *Repo) EnsurePaidSubscription(ctx context.Context, paymentID, userID int64, planTier string, duration time.Duration) (int64, error) {
	plan, ok := r.PlanByTier(ctx, planTier)
	if !ok {
		return 0, errors.New("unknown plan tier")
	}
	now := time.Now()
	subID, err := r.CreateSubscription(ctx, userID, plan.ID, now.Add(duration))
	if err != nil {
		existing, found, gerr := r.GetActive(ctx, userID)
		if gerr != nil {
			return 0, gerr
		}
		if found {
			if err := r.ActivateSubscription(ctx, existing.ID, duration); err != nil {
				return 0, err
			}
			_, _ = r.pool.Exec(ctx, `UPDATE payment SET subscription_id=$1 WHERE id=$2`, existing.ID, paymentID)
			return existing.ID, nil
		}
		return 0, err
	}
	_, _ = r.pool.Exec(ctx, `UPDATE payment SET subscription_id=$1 WHERE id=$2`, subID, paymentID)
	return subID, nil
}
