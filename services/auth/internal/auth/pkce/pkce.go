// Package pkce implements the PKCE (RFC 7636) code_verifier / code_challenge
// pair used by the OAuth Authorization Code flow (DEC-AUTH-17). Only the S256
// method is offered; the plain method is deliberately not supported.
package pkce

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// Method is the only challenge method offered (S256; plain is disallowed).
const Method = "S256"

// NewVerifier returns a high-entropy code_verifier: 32 random bytes,
// base64url-encoded without padding (43 chars, inside the RFC 7636 43-128 range).
func NewVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Challenge returns the S256 code_challenge for a verifier:
// base64url(SHA-256(verifier)), without padding.
func Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
