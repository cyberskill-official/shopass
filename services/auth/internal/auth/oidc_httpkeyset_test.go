package auth

import (
	"context"
	"encoding/base64"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
)

func jwksJSON(t *testing.T, priv *rsaPub) string {
	// Build a one-key JWKS document from a public key.
	n := base64.RawURLEncoding.EncodeToString(priv.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(priv.E)).Bytes())
	return `{"keys":[{"kty":"RSA","kid":"` + priv.kid + `","use":"sig","alg":"RS256","n":"` + n + `","e":"` + e + `"}]}`
}

type rsaPub struct {
	N   *big.Int
	E   int
	kid string
}

func TestHTTPKeySet_FetchesAndResolves(t *testing.T) {
	priv := genRSA(t)
	pub := &rsaPub{N: priv.PublicKey.N, E: priv.PublicKey.E, kid: "kid-http-1"}
	calls := 0
	ks := NewHTTPKeySet("https://issuer/certs")
	ks.do = func(req *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(jwksJSON(t, pub))),
			Header:     make(http.Header),
		}, nil
	}

	got, err := ks.KeyByKID(context.Background(), "kid-http-1")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got.N.Cmp(priv.PublicKey.N) != 0 || got.E != priv.PublicKey.E {
		t.Fatal("resolved key does not match the source key")
	}
	// A second lookup of the same kid is served from cache (no refetch).
	if _, err := ks.KeyByKID(context.Background(), "kid-http-1"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 fetch (then cache), got %d", calls)
	}
	// An unknown kid triggers one refresh, then errors.
	if _, err := ks.KeyByKID(context.Background(), "nope"); err == nil {
		t.Fatal("unknown kid must error")
	}
}
