package zalo

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogProvider_Refuses(t *testing.T) {
	p := NewLogProvider(slog.Default(), "noop")
	out, err := p.Send(context.Background(), Message{UserOAID: "u1", Template: "price_drop"})
	require.NoError(t, err)
	require.Equal(t, ResultFailed, out.Result)
	require.Equal(t, "noop", out.ProviderMessageID)
}

func TestConfigured_FailClosed(t *testing.T) {
	t.Setenv("ZALO_OA_ID", "")
	t.Setenv("ZALO_OA_SECRET", "")
	t.Setenv("ZALO_OA_ACCESS_TOKEN", "")
	require.False(t, Configured())

	t.Setenv("ZALO_OA_ID", "oa-1")
	require.False(t, Configured())

	t.Setenv("ZALO_OA_ACCESS_TOKEN", "tok")
	require.True(t, Configured())
}
