package fraud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVelocity_BurstTriggers(t *testing.T) {
	v := Velocity{Cfg: DefaultConfig(), Counter: stubCounter{n: 99}}
	res, err := v.Evaluate(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, res.Triggered)
	require.Equal(t, "velocity", res.Reason.Signal)
}

func TestVelocity_NormalDoesNotTrigger(t *testing.T) {
	v := Velocity{Cfg: DefaultConfig(), Counter: stubCounter{n: 2}}
	res, err := v.Evaluate(context.Background(), 1)
	require.NoError(t, err)
	require.False(t, res.Triggered)
}
