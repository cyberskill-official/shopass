package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrBadIDToken is returned when a provider id_token fails verification.
var ErrBadIDToken = errors.New("id_token verification failed")

// OIDCClaims are the claims we trust only AFTER verifying an id_token (§1 #4).
type OIDCClaims struct {
	Subject       string
	Email         string
	EmailVerified bool
}

// KeySet resolves a provider signing key by its kid (the provider's JWKS).
type KeySet interface {
	KeyByKID(ctx context.Context, kid string) (*rsa.PublicKey, error)
}

// StaticKeySet is an in-memory KeySet (a pinned set, or a test key).
type StaticKeySet struct{ Keys map[string]*rsa.PublicKey }

func (s StaticKeySet) KeyByKID(_ context.Context, kid string) (*rsa.PublicKey, error) {
	k, ok := s.Keys[kid]
	if !ok || k == nil {
		return nil, fmt.Errorf("unknown kid %q", kid)
	}
	return k, nil
}

// rsaFromJWK builds an rsa.PublicKey from a JWK's base64url n/e (reusing the JWK
// shape defined for this service's own JWKS in jwks.go).
func rsaFromJWK(j JWK) (*rsa.PublicKey, error) {
	if j.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported kty %q", j.Kty)
	}
	nb, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, fmt.Errorf("bad modulus: %w", err)
	}
	eb, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, fmt.Errorf("bad exponent: %w", err)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: int(new(big.Int).SetBytes(eb).Int64()),
	}, nil
}

// KeySetFromJWKS builds a StaticKeySet from a fetched JWKS document.
func KeySetFromJWKS(j JWKS) (StaticKeySet, error) {
	m := make(map[string]*rsa.PublicKey, len(j.Keys))
	for _, k := range j.Keys {
		pk, err := rsaFromJWK(k)
		if err != nil {
			return StaticKeySet{}, err
		}
		m[k.Kid] = pk
	}
	return StaticKeySet{Keys: m}, nil
}

// IDTokenVerifier verifies OIDC id_tokens against a provider's JWKS.
type IDTokenVerifier struct {
	keys     KeySet
	issuer   string
	audience string
	now      func() time.Time
}

// NewIDTokenVerifier builds a verifier pinned to one issuer and audience (client id).
func NewIDTokenVerifier(keys KeySet, issuer, audience string) *IDTokenVerifier {
	return &IDTokenVerifier{keys: keys, issuer: issuer, audience: audience, now: time.Now}
}

type oidcTokenClaims struct {
	Nonce         string `json:"nonce"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	jwt.RegisteredClaims
}

// Verify checks the id_token: RS256 signature via the provider JWKS, issuer,
// audience (client id), expiry, and nonce, then returns the trusted claims
// (§1 #4). It never trusts claims from an unverified token: any failure -
// bad signature, wrong aud/iss, expired, nonce mismatch, disallowed alg
// (e.g. none/HS256), or unknown kid - returns ErrBadIDToken.
func (v *IDTokenVerifier) Verify(ctx context.Context, rawIDToken, wantNonce string) (OIDCClaims, error) {
	keyfunc := func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		return v.keys.KeyByKID(ctx, kid)
	}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}), // reject alg=none / HS256 confusion
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(v.now),
	)
	var c oidcTokenClaims
	tok, err := parser.ParseWithClaims(rawIDToken, &c, keyfunc)
	if err != nil || tok == nil || !tok.Valid {
		return OIDCClaims{}, fmt.Errorf("%w: %v", ErrBadIDToken, err)
	}
	if c.Subject == "" {
		return OIDCClaims{}, fmt.Errorf("%w: empty subject", ErrBadIDToken)
	}
	if wantNonce != "" && c.Nonce != wantNonce {
		return OIDCClaims{}, fmt.Errorf("%w: nonce mismatch", ErrBadIDToken)
	}
	return OIDCClaims{Subject: c.Subject, Email: c.Email, EmailVerified: c.EmailVerified}, nil
}
