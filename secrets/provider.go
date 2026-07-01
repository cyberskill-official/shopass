package secrets

import (
	"context"
	"encoding/json"
	"errors"
)

type Secret struct {
	value   string
	Version string
}

func NewSecret(value, version string) Secret {
	return Secret{value: value, Version: version}
}

func (s Secret) Reveal() string { return s.value }

func (s Secret) String() string { return mask(s.value) }

func (s Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal(mask(s.value))
}

type SecretProvider interface {
	Get(ctx context.Context, path string) (Secret, error)
}

var ErrInvalidPath = errors.New("invalid secret path")

func validatePath(path string) error {
	if path == "" {
		return ErrInvalidPath
	}
	for _, r := range path {
		if r == '/' || r == '-' || r == '_' || r == '.' ||
			(r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') {
			continue
		}
		return ErrInvalidPath
	}
	return nil
}
