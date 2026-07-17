package gw

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrJWKSUnavailable means the gateway cannot obtain the key material needed
// to decide whether a token is valid. This must be surfaced as 503, never as a
// fail-open authentication path.
var ErrJWKSUnavailable = errors.New("jwks unavailable")

type remoteJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type remoteJWKS struct {
	Keys []remoteJWK `json:"keys"`
}

// HTTPJWKS verifies the auth service's RS256 access tokens. Key material is
// cached for a short period and reloaded immediately on an unknown kid, which
// permits deliberate key rotation without making the gateway trust client
// supplied identity headers.
type HTTPJWKS struct {
	url      string
	issuer   string
	audience string
	ttl      time.Duration
	client   *http.Client

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
}

func NewHTTPJWKS(url, issuer, audience string, ttl time.Duration, client *http.Client) *HTTPJWKS {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPJWKS{
		url:      url,
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
		client:   client,
		keys:     make(map[string]*rsa.PublicKey),
	}
}

func (c *HTTPJWKS) Verify(ctx context.Context, tokenString string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	claims := &Claims{}
	unverified, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	kid, _ := unverified.Header["kid"].(string)
	if kid == "" {
		return nil, errors.New("missing kid")
	}

	key, err := c.keyFor(ctx, kid)
	if err != nil {
		return nil, err
	}

	verified := &Claims{}
	parsed, err := jwt.ParseWithClaims(tokenString, verified, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return key, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithIssuer(c.issuer), jwt.WithAudience(c.audience))
	if err != nil || !parsed.Valid {
		if err == nil {
			err = errors.New("invalid token")
		}
		return nil, err
	}
	if verified.UserID <= 0 || verified.IssuedAt == nil {
		return nil, errors.New("invalid token claims")
	}
	return verified, nil
}

func (c *HTTPJWKS) keyFor(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	fresh := time.Now().Before(c.expiresAt)
	c.mu.RUnlock()
	if ok && fresh {
		return key, nil
	}

	if err := c.refresh(ctx); err != nil {
		// A still-fresh cached key is safe to use during a short upstream outage.
		c.mu.RLock()
		key, ok = c.keys[kid]
		fresh = time.Now().Before(c.expiresAt)
		c.mu.RUnlock()
		if ok && fresh {
			return key, nil
		}
		return nil, fmt.Errorf("%w: %v", ErrJWKSUnavailable, err)
	}

	c.mu.RLock()
	key, ok = c.keys[kid]
	c.mu.RUnlock()
	if !ok {
		return nil, errors.New("unknown signing key")
	}
	return key, nil
}

func (c *HTTPJWKS) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	var doc remoteJWKS
	if err := json.Unmarshal(body, &doc); err != nil {
		return err
	}
	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, jwk := range doc.Keys {
		if jwk.Kty != "RSA" || jwk.Alg != "RS256" || jwk.Use != "sig" || jwk.Kid == "" {
			continue
		}
		pub, err := rsaPublicKey(jwk)
		if err != nil {
			return fmt.Errorf("invalid jwk %q: %w", jwk.Kid, err)
		}
		keys[jwk.Kid] = pub
	}
	if len(keys) == 0 {
		return errors.New("jwks has no usable rs256 keys")
	}
	c.mu.Lock()
	c.keys = keys
	c.expiresAt = time.Now().Add(c.ttl)
	c.mu.Unlock()
	return nil
}

func rsaPublicKey(jwk remoteJWK) (*rsa.PublicKey, error) {
	n, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil || len(n) == 0 {
		return nil, errors.New("invalid modulus")
	}
	e, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil || len(e) == 0 {
		return nil, errors.New("invalid exponent")
	}
	exponent := new(big.Int).SetBytes(e)
	if !exponent.IsInt64() || exponent.Int64() < 3 {
		return nil, errors.New("invalid exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(exponent.Int64())}, nil
}
