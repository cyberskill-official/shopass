package fraud

import (
	"context"
)

type Reason struct {
	Signal       string `json:"signal"`
	Detail       string `json:"detail"`
	Contribution int    `json:"contribution"`
}

type Assessment struct {
	UserID     int64
	RiskScore  int
	Reasons    []Reason
	HoldReward bool
}

// SignalStore persists fraud_signal rows (idempotent upsert). Engine never auto-bans.
type SignalStore interface {
	UpsertOpen(ctx context.Context, userID int64, kind string, score int, reasons []Reason) error
}

// RewardHolder places rewards on HOLD when score exceeds threshold (TRUST-005).
type RewardHolder interface {
	Hold(ctx context.Context, userID int64) error
}

type Engine struct {
	Cfg      Config
	Velocity Velocity
	Graph    Graph
	Rules    Rules
	Store    SignalStore
	Holder   RewardHolder
}

func NewEngine(cfg Config, counter EventCounter, edges ClusterSizer, store SignalStore, holder RewardHolder) *Engine {
	return &Engine{
		Cfg:      cfg,
		Velocity: Velocity{Cfg: cfg, Counter: counter},
		Graph:    Graph{Cfg: cfg, Edges: edges},
		Rules:    Rules{Cfg: cfg},
		Store:    store,
		Holder:   holder,
	}
}

// Assess combines velocity + graph + rules into an explainable risk_score.
// MUST NOT ban or seize funds — only score, flag, and optionally HOLD.
func (e *Engine) Assess(ctx context.Context, userID int64, extras map[string]any) (Assessment, error) {
	var reasons []Reason
	score := 0

	if v, err := e.Velocity.Evaluate(ctx, userID); err != nil {
		return Assessment{}, err
	} else if v.Triggered {
		score += v.Weight
		reasons = append(reasons, v.Reason)
	}
	if g, err := e.Graph.Evaluate(ctx, userID); err != nil {
		return Assessment{}, err
	} else if g.Triggered {
		score += g.Weight
		reasons = append(reasons, g.Reason)
	}
	if r := e.Rules.Evaluate(ctx, userID, extras); r.Triggered {
		score += r.Weight
		reasons = append(reasons, r.Reason)
	}
	if score > 100 {
		score = 100
	}

	a := Assessment{UserID: userID, RiskScore: score, Reasons: reasons}
	if score >= e.Cfg.HoldThreshold {
		a.HoldReward = true
		if e.Holder != nil {
			_ = e.Holder.Hold(ctx, userID)
		}
	}
	if e.Store != nil && score > 0 {
		kind := "combined"
		if err := e.Store.UpsertOpen(ctx, userID, kind, score, reasons); err != nil {
			return a, err
		}
	}
	return a, nil
}
