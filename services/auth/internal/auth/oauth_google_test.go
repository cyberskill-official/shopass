package auth

import (
	"context"
	"crypto/rsa"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func fakeDoer(status int, body string, capture *http.Request) httpDoer {
	return func(req *http.Request) (*http.Response, error) {
		if capture != nil {
			*capture = *req
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}
}

func TestGoogle_AuthCodeURL(t *testing.T) {
	g := NewGoogleProvider("client-abc", "secret", "https://app/callback", StaticKeySet{})
	u := g.AuthCodeURL("state-1", "chal-1", "nonce-1")
	for _, want := range []string{
		"client_id=client-abc",
		"code_challenge=chal-1",
		"code_challenge_method=S256",
		"state=state-1",
		"nonce=nonce-1",
		"scope=openid",
		"response_type=code",
	} {
		if !strings.Contains(u, want) {
			t.Errorf("authorize url missing %q in %q", want, u)
		}
	}
}

func TestGoogle_ExchangeAndVerify_HappyPath(t *testing.T) {
	priv := genRSA(t)
	// Google's verifier: audience must equal this client id, issuer Google's.
	keys := StaticKeySet{Keys: map[string]*rsa.PublicKey{oidcKID: &priv.PublicKey}}
	g := NewGoogleProvider(oidcAud, "secret", "https://app/callback", keys)
	g.verifier.now = func() time.Time { return oidcBase }

	c := validClaims() // iss=Google, aud=oidcAud, sub set, verified email
	idToken := signRS256(t, priv, oidcKID, c)
	var captured http.Request
	g.do = fakeDoer(http.StatusOK, `{"id_token":"`+idToken+`","access_token":"a"}`, &captured)

	claims, err := g.ExchangeAndVerify(context.Background(), "auth-code", "verifier-xyz", oidcNonce)
	if err != nil {
		t.Fatalf("exchange should succeed: %v", err)
	}
	if claims.Subject != "google-sub-1" || !claims.EmailVerified {
		t.Fatalf("claims wrong: %+v", claims)
	}
	// The exchange must send the PKCE code_verifier and the auth code.
	sent, _ := io.ReadAll(captured.Body)
	form := string(sent)
	if !strings.Contains(form, "code_verifier=verifier-xyz") || !strings.Contains(form, "code=auth-code") {
		t.Fatalf("token request missing code_verifier/code: %q", form)
	}
}

func TestGoogle_ExchangeAndVerify_TokenEndpointError(t *testing.T) {
	g := NewGoogleProvider(oidcAud, "secret", "https://app/callback", StaticKeySet{})
	g.do = fakeDoer(http.StatusBadRequest, `{"error":"invalid_grant"}`, nil)
	if _, err := g.ExchangeAndVerify(context.Background(), "bad", "v", "n"); err == nil {
		t.Fatal("a 400 from the token endpoint must be an error")
	}
}

func TestGoogle_ExchangeAndVerify_MissingIDToken(t *testing.T) {
	g := NewGoogleProvider(oidcAud, "secret", "https://app/callback", StaticKeySet{})
	g.do = fakeDoer(http.StatusOK, `{"access_token":"a"}`, nil)
	if _, err := g.ExchangeAndVerify(context.Background(), "c", "v", "n"); err == nil {
		t.Fatal("a token response without id_token must be an error")
	}
}
