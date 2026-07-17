package gw

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jwksClient(t *testing.T, private *rsa.PrivateKey, kid string) *http.Client {
	t.Helper()
	doc := remoteJWKS{Keys: []remoteJWK{{
		Kty: "RSA",
		Kid: kid,
		Use: "sig",
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(private.PublicKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(private.PublicKey.E)).Bytes()),
	}}}
	body, err := json.Marshal(doc)
	require.NoError(t, err)
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	})}
}

func TestHTTPJWKSVerifiesRS256Claims(t *testing.T) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	const kid = "key-1"
	verifier := NewHTTPJWKS("http://auth/.well-known/jwks.json", "shopass-auth", "shopass-gateway", time.Minute, jwksClient(t, private, kid))

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
		UserID: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "shopass-auth",
			Audience:  jwt.ClaimStrings{"shopass-gateway"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		},
	})
	token.Header["kid"] = kid
	raw, err := token.SignedString(private)
	require.NoError(t, err)

	claims, err := verifier.Verify(context.Background(), raw)
	require.NoError(t, err)
	require.Equal(t, int64(42), claims.UserID)
}

func TestHTTPJWKSRejectsWrongAudience(t *testing.T) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	const kid = "key-1"
	verifier := NewHTTPJWKS("http://auth/.well-known/jwks.json", "shopass-auth", "shopass-gateway", time.Minute, jwksClient(t, private, kid))

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
		UserID: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "shopass-auth",
			Audience:  jwt.ClaimStrings{"another-service"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		},
	})
	token.Header["kid"] = kid
	raw, err := token.SignedString(private)
	require.NoError(t, err)

	_, err = verifier.Verify(context.Background(), raw)
	require.Error(t, err)
}
