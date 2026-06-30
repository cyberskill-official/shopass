package secrets

import (
	"context"
)

// Secret represents a fetched secret from a provider.
type Secret struct {
	value   string
	Version string
}

// Reveal returns the raw secret value. It should only be called at the point of use.
func (s Secret) Reveal() string {
	return s.value
}

// SecretProvider abstracts reading secrets from a backend (Vault, AWS SM).
type SecretProvider interface {
	Get(ctx context.Context, path string) (Secret, error)
}
