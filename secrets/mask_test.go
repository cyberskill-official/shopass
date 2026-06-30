package secrets

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecret_NeverLeaksRaw(t *testing.T) {
	s := Secret{value: "super-secret-token-123", Version: "v3"}

	// Print formatting
	require.NotContains(t, fmt.Sprintf("%v", s), "super-secret-token-123")
	require.NotContains(t, fmt.Sprintf("%+v", s), "super-secret-token-123")
	require.Equal(t, "****-123", fmt.Sprintf("%v", s))

	// JSON serialization
	b, err := json.Marshal(s)
	require.NoError(t, err)
	require.NotContains(t, string(b), "super-secret-token-123")

	// Reveal method
	require.Equal(t, "super-secret-token-123", s.Reveal())
}

func TestMask(t *testing.T) {
	require.Equal(t, "****", mask("abc"))
	require.Equal(t, "****", mask("abcd"))
	require.Equal(t, "****bcde", mask("abcde"))
}
