package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	oidcIssuer = "https://accounts.google.com"
	oidcAud    = "client-123.apps.googleusercontent.com"
	oidcKID    = "kid-1"
	oidcNonce  = "nonce-xyz"
)

var oidcBase = time.Unix(1_700_000_000, 0)

func genRSA(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func validClaims() oidcTokenClaims {
	return oidcTokenClaims{
		Nonce:         oidcNonce,
		Email:         "chi@example.com",
		EmailVerified: true,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    oidcIssuer,
			Subject:   "google-sub-1",
			Audience:  jwt.ClaimStrings{oidcAud},
			ExpiresAt: jwt.NewNumericDate(oidcBase.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(oidcBase),
		},
	}
}

func signRS256(t *testing.T, priv *rsa.PrivateKey, kid string, c oidcTokenClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	if kid != "" {
		tok.Header["kid"] = kid
	}
	s, err := tok.SignedString(priv)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func newVerifier(priv *rsa.PrivateKey) *IDTokenVerifier {
	v := NewIDTokenVerifier(StaticKeySet{Keys: map[string]*rsa.PublicKey{oidcKID: &priv.PublicKey}}, oidcIssuer, oidcAud)
	v.now = func() time.Time { return oidcBase }
	return v
}

func TestIDTokenVerify_ValidToken(t *testing.T) {
	priv := genRSA(t)
	raw := signRS256(t, priv, oidcKID, validClaims())
	claims, err := newVerifier(priv).Verify(context.Background(), raw, oidcNonce)
	if err != nil {
		t.Fatalf("valid token should verify: %v", err)
	}
	if claims.Subject != "google-sub-1" || claims.Email != "chi@example.com" || !claims.EmailVerified {
		t.Fatalf("claims wrong: %+v", claims)
	}
}

func TestIDTokenVerify_Rejections(t *testing.T) {
	priv := genRSA(t)
	other := genRSA(t)
	ctx := context.Background()

	cases := []struct {
		name  string
		raw   func() string
		nonce string
	}{
		{"nonce mismatch", func() string { return signRS256(t, priv, oidcKID, validClaims()) }, "wrong-nonce"},
		{"wrong audience", func() string {
			c := validClaims()
			c.Audience = jwt.ClaimStrings{"someone-else"}
			return signRS256(t, priv, oidcKID, c)
		}, oidcNonce},
		{"wrong issuer", func() string {
			c := validClaims()
			c.Issuer = "https://evil.example.com"
			return signRS256(t, priv, oidcKID, c)
		}, oidcNonce},
		{"expired", func() string {
			c := validClaims()
			c.ExpiresAt = jwt.NewNumericDate(oidcBase.Add(-time.Hour))
			return signRS256(t, priv, oidcKID, c)
		}, oidcNonce},
		{"signed by wrong key", func() string { return signRS256(t, other, oidcKID, validClaims()) }, oidcNonce},
		{"unknown kid", func() string { return signRS256(t, priv, "kid-unknown", validClaims()) }, oidcNonce},
		{"alg none", func() string {
			tok := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims())
			tok.Header["kid"] = oidcKID
			s, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
			if err != nil {
				t.Fatal(err)
			}
			return s
		}, oidcNonce},
		{"alg HS256 confusion", func() string {
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims())
			tok.Header["kid"] = oidcKID
			s, err := tok.SignedString([]byte("shared-secret"))
			if err != nil {
				t.Fatal(err)
			}
			return s
		}, oidcNonce},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := newVerifier(priv).Verify(ctx, tc.raw(), tc.nonce); err == nil {
				t.Fatalf("%s: expected verification to fail, got nil", tc.name)
			}
		})
	}
}

func TestKeySetFromJWKS_RoundTrip(t *testing.T) {
	// A JWKS built from a real key must reconstruct the same modulus/exponent so
	// a token signed by that key verifies through the JWKS path.
	priv := genRSA(t)
	ts := &TokenService{keys: map[string]*rsa.PublicKey{oidcKID: &priv.PublicKey}}
	ks, err := KeySetFromJWKS(ts.GetJWKS())
	if err != nil {
		t.Fatal(err)
	}
	v := NewIDTokenVerifier(ks, oidcIssuer, oidcAud)
	v.now = func() time.Time { return oidcBase }
	if _, err := v.Verify(context.Background(), signRS256(t, priv, oidcKID, validClaims()), oidcNonce); err != nil {
		t.Fatalf("JWKS-derived key should verify: %v", err)
	}
}
