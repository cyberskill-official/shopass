package trend

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildCell_SuppressesBelowK(t *testing.T) {
	c := BuildCell(1, 1, time.Now(), 49, 1, 2, 3)
	require.True(t, c.Suppressed)
	require.Nil(t, c.Median)
}

func TestBuildCell_PublishesAtK(t *testing.T) {
	c := BuildCell(1, 1, time.Now(), 50, 10, 20, 30)
	require.False(t, c.Suppressed)
	require.Equal(t, int64(20), *c.Median)
}
