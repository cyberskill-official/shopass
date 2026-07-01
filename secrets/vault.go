package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type vaultProvider struct {
	addr      string
	token     string
	mountPath string
	client    *http.Client
}

func NewVaultProvider(addr, token string) SecretProvider {
	return &vaultProvider{
		addr:      strings.TrimRight(addr, "/"),
		token:     token,
		mountPath: "secret",
		client:    http.DefaultClient,
	}
}

func (v *vaultProvider) Get(ctx context.Context, path string) (Secret, error) {
	if err := validatePath(path); err != nil {
		return Secret{}, err
	}
	reqURL := v.addr + "/v1/" + url.PathEscape(v.mountPath) + "/data/sandeal/" + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return Secret{}, err
	}
	if v.token != "" {
		req.Header.Set("X-Vault-Token", v.token)
	}
	res, err := v.client.Do(req)
	if err != nil {
		return Secret{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Secret{}, fmt.Errorf("vault: status %d for path %s", res.StatusCode, path)
	}
	var body struct {
		Data struct {
			Data map[string]string `json:"data"`
			Meta struct {
				Version int `json:"version"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return Secret{}, err
	}
	value := body.Data.Data["value"]
	if value == "" {
		return Secret{}, fmt.Errorf("vault: missing value for path %s", path)
	}
	return NewSecret(value, fmt.Sprintf("%d", body.Data.Meta.Version)), nil
}
