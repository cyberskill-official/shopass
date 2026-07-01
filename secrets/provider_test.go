package secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type fakeProvider struct {
	val     string
	version string
}

func (f *fakeProvider) Get(_ context.Context, _ string) (Secret, error) {
	return Secret{value: f.val, Version: f.version}, nil
}

func TestProvider_Get_ReturnsValueAndVersion(t *testing.T) {
	p := &fakeProvider{val: "secret-value", version: "v3"}
	s, err := p.Get(context.Background(), "auth/jwt-signing")
	if err != nil {
		t.Fatal(err)
	}
	if s.Reveal() != "secret-value" {
		t.Errorf("expected 'secret-value', got %s", s.Reveal())
	}
	if s.Version != "v3" {
		t.Errorf("expected 'v3', got %s", s.Version)
	}
}

func TestVaultProvider_ReadsKVv2ValueAndVersion(t *testing.T) {
	p := &vaultProvider{
		addr:      "https://vault.test",
		token:     "token",
		mountPath: "secret",
		client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("X-Vault-Token") != "token" {
				t.Fatal("missing vault token")
			}
			return jsonResponse(200, map[string]any{
				"data": map[string]any{
					"data":     map[string]string{"value": "vault-secret"},
					"metadata": map[string]int{"version": 7},
				},
			}), nil
		})},
	}
	s, err := p.Get(context.Background(), "auth/jwt-signing")
	if err != nil {
		t.Fatal(err)
	}
	if s.Reveal() != "vault-secret" || s.Version != "7" {
		t.Fatalf("unexpected secret %q version %s", s.Reveal(), s.Version)
	}
}

func TestAWSSecretsManagerProvider_ReadsCurrentSecret(t *testing.T) {
	p := &awsSMProvider{
		region:   "ap-southeast-1",
		endpoint: "https://secretsmanager.test",
		client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("X-Amz-Target") != "secretsmanager.GetSecretValue" {
				t.Fatal("missing aws target")
			}
			return jsonResponse(200, map[string]string{
				"SecretString": "aws-secret",
				"VersionId":    "v9",
			}), nil
		})},
	}
	s, err := p.Get(context.Background(), "db/main")
	if err != nil {
		t.Fatal(err)
	}
	if s.Reveal() != "aws-secret" || s.Version != "v9" {
		t.Fatalf("unexpected secret %q version %s", s.Reveal(), s.Version)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func jsonResponse(status int, v any) *http.Response {
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(v)
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(&buf),
		Header:     make(http.Header),
	}
}
