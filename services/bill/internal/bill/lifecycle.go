package bill

import (
	"context"
	"errors"
	"fmt"
)

var ErrInvalidTransition = errors.New("invalid status transition")

var valid = map[string]map[string]bool{
	"active":   {"past_due": true, "canceled": true},
	"past_due": {"active": true, "canceled": true, "expired": true},
	"canceled": {}, // cuối
	"expired":  {}, // cuối
}

func CanTransition(from, to string) bool {
	if valid[from] == nil {
		return false
	}
	return valid[from][to]
}

func (r *Repo) UpdateStatus(ctx context.Context, subID int64, to string) error {
	cur, err := r.statusOf(ctx, subID)
	if err != nil {
		return err
	}
	if !CanTransition(cur, to) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, cur, to)
	}
	_, err = r.pool.Exec(ctx, `UPDATE subscription SET status=$1 WHERE id=$2`, to, subID)
	if err == nil && r.metrics != nil {
		r.metrics.StatusChange(cur, to)
	}
	return err
}
