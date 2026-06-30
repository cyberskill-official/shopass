package secrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

type vaultProvider struct {
	client *api.Client
}

// NewVaultProvider creates a provider using HashiCorp Vault KV v2.
func NewVaultProvider(client *api.Client) SecretProvider {
	return &vaultProvider{client: client}
}

func (p *vaultProvider) Get(ctx context.Context, path string) (Secret, error) {
	secret, err := p.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return Secret{}, err
	}
	if secret == nil {
		return Secret{}, fmt.Errorf("secret not found at path %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		data = secret.Data
	}

	valRaw, ok := data["value"]
	if !ok {
		return Secret{}, fmt.Errorf("no 'value' key found in secret data")
	}
	val, ok := valRaw.(string)
	if !ok {
		return Secret{}, fmt.Errorf("'value' is not a string")
	}

	version := "unknown"
	if meta, ok := secret.Data["metadata"].(map[string]interface{}); ok {
		if v, ok := meta["version"].(float64); ok {
			version = fmt.Sprintf("%.0f", v)
		} else if v, ok := meta["version"].(string); ok {
			version = v
		}
	}

	return Secret{value: val, Version: version}, nil
}
