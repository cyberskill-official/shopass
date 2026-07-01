package obs_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"shopass/obs"
)

func TestLog_RedactsPII(t *testing.T) {
	var buf bytes.Buffer
	obs.SetOutput(&buf)

	ctx := context.Background()
	obs.Info(ctx, "login", slog.String("email", "chi@example.com"),
		slog.String("token", "secret-jwt-xyz"))

	out := buf.String()
	require.NotContains(t, out, "chi@example.com")
	require.NotContains(t, out, "secret-jwt-xyz")
	require.Contains(t, out, "c***m") // or whatever the mask is
}
