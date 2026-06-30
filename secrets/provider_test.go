package secrets

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A simple mock for testing the SecretProvider contract
type mockProvider struct {
	data map[string]Secret
}

func (m *mockProvider) Get(ctx context.Context, path string) (Secret, error) {
	if s, ok := m.data[path]; ok {
		return s, nil
	}
	return Secret{}, errors.New("not found")
}

func TestProviderContract(t *testing.T) {
	ctx := context.Background()
	mock := &mockProvider{
		data: map[string]Secret{
			"auth/jwt-signing": {value: "secret-key", Version: "v1"},
		},
	}

	// Wrap in cache as normal service usage would
	provider := NewCachedProvider(mock, time.Minute)

	// Valid path
	sec, err := provider.Get(ctx, "auth/jwt-signing")
	require.NoError(t, err)
	require.Equal(t, "secret-key", sec.Reveal())
	require.Equal(t, "v1", sec.Version)

	// Invalid path (Least privilege simulation - if they try a path they don't have access to, they get an error)
	_, err = provider.Get(ctx, "db/main")
	require.Error(t, err)
}
