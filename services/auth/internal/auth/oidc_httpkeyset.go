package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HTTPKeySet fetches a provider's JWKS over HTTP and caches it. On a cache miss
// (an unknown kid, e.g. after key rotation) it refetches once, so rotation is
// picked up without a restart. Used to build the id_token verifier for a live
// provider like Google.
type HTTPKeySet struct {
	url string
	do  func(*http.Request) (*http.Response, error)

	mu      sync.Mutex
	cached  StaticKeySet
	fetched time.Time
	ttl     time.Duration
}

// NewHTTPKeySet targets a JWKS URL (e.g. https://www.googleapis.com/oauth2/v3/certs).
func NewHTTPKeySet(url string) *HTTPKeySet {
	return &HTTPKeySet{url: url, do: http.DefaultClient.Do, ttl: time.Hour, cached: StaticKeySet{Keys: map[string]*rsa.PublicKey{}}}
}

func (h *HTTPKeySet) KeyByKID(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if k := h.lookup(kid); k != nil {
		return k, nil
	}
	if err := h.refresh(ctx); err != nil {
		return nil, err
	}
	if k := h.lookup(kid); k != nil {
		return k, nil
	}
	return nil, fmt.Errorf("kid %q not found in JWKS", kid)
}

func (h *HTTPKeySet) lookup(kid string) *rsa.PublicKey {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.cached.Keys[kid]
}

func (h *HTTPKeySet) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, nil)
	if err != nil {
		return err
	}
	resp, err := h.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks endpoint status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("decode jwks: %w", err)
	}
	ks, err := KeySetFromJWKS(jwks)
	if err != nil {
		return err
	}
	h.mu.Lock()
	h.cached = ks
	h.fetched = time.Now()
	h.mu.Unlock()
	return nil
}
