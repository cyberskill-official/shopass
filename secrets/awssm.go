package secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type awsSMProvider struct {
	region   string
	endpoint string
	client   *http.Client
}

func NewAWSSecretsManagerProvider(region string) SecretProvider {
	return &awsSMProvider{
		region:   region,
		endpoint: "https://secretsmanager." + region + ".amazonaws.com/",
		client:   http.DefaultClient,
	}
}

func NewAWSSecretsManagerProviderWithEndpoint(region, endpoint string) SecretProvider {
	return &awsSMProvider{
		region:   region,
		endpoint: endpoint,
		client:   http.DefaultClient,
	}
}

func (a *awsSMProvider) Get(ctx context.Context, path string) (Secret, error) {
	if err := validatePath(path); err != nil {
		return Secret{}, err
	}
	payload, _ := json.Marshal(map[string]any{
		"SecretId":     "sandeal/" + path,
		"VersionStage": "AWSCURRENT",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(payload))
	if err != nil {
		return Secret{}, err
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", "secretsmanager.GetSecretValue")
	res, err := a.client.Do(req)
	if err != nil {
		return Secret{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Secret{}, fmt.Errorf("awssm: status %d for path %s", res.StatusCode, path)
	}
	var body struct {
		SecretString string `json:"SecretString"`
		VersionID    string `json:"VersionId"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return Secret{}, err
	}
	if strings.TrimSpace(body.SecretString) == "" {
		return Secret{}, fmt.Errorf("awssm: missing SecretString for path %s", path)
	}
	return NewSecret(body.SecretString, body.VersionID), nil
}
