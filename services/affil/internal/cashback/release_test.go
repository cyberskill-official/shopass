package cashback

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReleaseDue_EmptyStore(t *testing.T) {
	n, err := NewReleaser(&Ledger{Store: NewMemStore()}).ReleaseDue(context.Background(), time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, 0, n)
}

func TestReleaseDue_NilSafe(t *testing.T) {
	n, err := (*Releaser)(nil).ReleaseDue(context.Background(), time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, 0, n)
}
