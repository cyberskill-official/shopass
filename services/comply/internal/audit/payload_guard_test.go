package audit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGuard_RejectsCookie(t *testing.T) {
	err := GuardPayload(map[string]any{"productId": 1, "cookie": "abc"})
	require.ErrorIs(t, err, ErrForbiddenField)
}

func TestGuard_RejectsAuthorizationAnyCase(t *testing.T) {
	err := GuardPayload(map[string]any{"Authorization": "Bearer x"})
	require.ErrorIs(t, err, ErrForbiddenField)
}

func TestGuard_MinimalPayloadPasses(t *testing.T) {
	err := GuardPayload(map[string]any{"productId": 1, "price": 89000, "qty": 2})
	require.NoError(t, err)
}
