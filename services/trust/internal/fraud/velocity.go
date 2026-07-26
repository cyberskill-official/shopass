package fraud

import (
	"context"
	"fmt"
)

// EventCounter counts subject events in a sliding window (injected for tests / SQL).
type EventCounter interface {
	CountRedeems(ctx context.Context, userID int64, windowMinutes int) (int, error)
}

type Velocity struct {
	Cfg     Config
	Counter EventCounter
}

func (v Velocity) Evaluate(ctx context.Context, userID int64) (signalResult, error) {
	if v.Counter == nil {
		return signalResult{}, nil
	}
	n, err := v.Counter.CountRedeems(ctx, userID, v.Cfg.VelocityWindowMinutes)
	if err != nil {
		return signalResult{}, err
	}
	if n <= v.Cfg.VelocityRedeemMax {
		return signalResult{}, nil
	}
	return signalResult{
		Triggered: true,
		Weight:    v.Cfg.VelocityWeight,
		Reason: Reason{
			Signal:       "velocity",
			Detail:       fmt.Sprintf("%d referral redeems in %d minutes exceeded threshold %d", n, v.Cfg.VelocityWindowMinutes, v.Cfg.VelocityRedeemMax),
			Contribution: v.Cfg.VelocityWeight,
		},
	}, nil
}
