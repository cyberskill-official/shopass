package fraud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGraph_DenseClusterTriggers(t *testing.T) {
	g := Graph{Cfg: DefaultConfig(), Edges: stubCluster{size: 20}}
	res, err := g.Evaluate(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, res.Triggered)
	require.Contains(t, res.Reason.Detail, "cluster size 20")
	require.Contains(t, res.Reason.Detail, "threshold 5")
}

func TestGraph_SingletonSafe(t *testing.T) {
	g := Graph{Cfg: DefaultConfig(), Edges: stubCluster{size: 1}}
	res, err := g.Evaluate(context.Background(), 1)
	require.NoError(t, err)
	require.False(t, res.Triggered)
}
