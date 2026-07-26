package fraud

import "context"

type signalResult struct {
	Triggered bool
	Weight    int
	Reason    Reason
}

// Rules evaluates configurable heuristics beyond velocity/graph.
type Rules struct {
	Cfg Config
}

func (r Rules) Evaluate(_ context.Context, userID int64, extras map[string]any) signalResult {
	if extras == nil {
		return signalResult{}
	}
	if v, ok := extras["self_referral"].(bool); ok && v {
		return signalResult{
			Triggered: true,
			Weight:    r.Cfg.RuleWeight,
			Reason: Reason{
				Signal:       "rule",
				Detail:       "self_referral_attempt",
				Contribution: r.Cfg.RuleWeight,
			},
		}
	}
	_ = userID
	return signalResult{}
}
