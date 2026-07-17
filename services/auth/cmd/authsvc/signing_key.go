package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
)

// loadSigningKey requires a durable key in production. Generating one at boot
// is convenient for an isolated developer run, but it invalidates every access
// token at restart and makes a gateway JWKS cache unreliable.
func loadSigningKey(appEnv string) (*rsa.PrivateKey, error) {
	path := strings.TrimSpace(os.Getenv("AUTH_SIGNING_KEY_FILE"))
	if path == "" {
		if appEnv == "production" {
			return nil, errors.New("AUTH_SIGNING_KEY_FILE is required in production")
		}
		return rsa.GenerateKey(rand.Reader, 2048)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read signing key: %w", err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("signing key is not PEM")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse signing key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("signing key must be RSA")
	}
	return rsaKey, nil
}
