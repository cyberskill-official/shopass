package engine

import (
	"context"

	"shopass/services/track/internal/track"
)

type StateRepo interface {
	LastConditionMet(ctx context.Context, ruleID int64) (bool, error)
	Set(ctx context.Context, ruleID int64, met bool) error
}

// fireIfRisingEdge chỉ bắn khi điều kiện chuyển false->true (DEC-TRACK-32).
func (e *Engine) fireIfRisingEdge(ctx context.Context, r track.AlertRule, met bool, payload map[string]any) error {
	prev, err := e.state.LastConditionMet(ctx, r.ID)
	if err != nil {
		return err
	}
	if met && !prev {
		alertID, err := e.handoff.CreateAndEnqueue(ctx, r, payload)
		if err != nil {
			return err
		}
		_ = alertID
		return e.state.Set(ctx, r.ID, true)
	}
	if !met && prev {
		return e.state.Set(ctx, r.ID, false)
	}
	return nil
}
