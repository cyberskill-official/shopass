package fraud

import (
	"context"
	"fmt"
)

// ClusterSizer reports linked-account cluster size for a user.
type ClusterSizer interface {
	ClusterSize(ctx context.Context, userID int64) (int, error)
}

type Graph struct {
	Cfg   Config
	Edges ClusterSizer
}

func (g Graph) Evaluate(ctx context.Context, userID int64) (signalResult, error) {
	if g.Edges == nil {
		return signalResult{}, nil
	}
	size, err := g.Edges.ClusterSize(ctx, userID)
	if err != nil {
		return signalResult{}, err
	}
	if size < g.Cfg.GraphClusterMinSize {
		return signalResult{}, nil
	}
	return signalResult{
		Triggered: true,
		Weight:    g.Cfg.GraphWeight,
		Reason: Reason{
			Signal:       "graph",
			Detail:       fmt.Sprintf("linked-account cluster size %d met threshold %d", size, g.Cfg.GraphClusterMinSize),
			Contribution: g.Cfg.GraphWeight,
		},
	}, nil
}
