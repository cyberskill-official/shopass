package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"shopass/services/auth/internal/auth/pkce"
)

// outboundHTTP is used for Google token exchange when no custom doer is injected.
var outboundHTTP = &http.Client{Timeout: 10 * time.Second}

const (
	googleAuthEndpoint  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenEndpoint = "https://oauth2.googleapis.com/token"
	// GoogleIssuer is Google's OIDC issuer, used to build the id_token verifier.
	GoogleIssuer = "https://accounts.google.com"
)

type httpDoer func(*http.Request) (*http.Response, error)

// GoogleProvider implements OAuthProvider for Google (OIDC). The client secret
// and config come from the secrets manager (TASK-INFRA-003), never from code/env
// literals (§1 #9). The HTTP doer is injectable so the exchange is testable; a
// real token exchange is the only part that needs live Google credentials.
type GoogleProvider struct {
	clientID      string
	clientSecret  string
	redirectURI   string
	authEndpoint  string
	tokenEndpoint string
	verifier      *IDTokenVerifier
	do            httpDoer
}

// NewGoogleProvider builds the provider with an id_token verifier pinned to
// Google's issuer and this client id as the audience.
func NewGoogleProvider(clientID, clientSecret, redirectURI string, keys KeySet) *GoogleProvider {
	return &GoogleProvider{
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   redirectURI,
		authEndpoint:  googleAuthEndpoint,
		tokenEndpoint: googleTokenEndpoint,
		verifier:      NewIDTokenVerifier(keys, GoogleIssuer, clientID),
		do:            outboundHTTP.Do,
	}
}

// AuthCodeURL builds Google's authorize URL with PKCE S256, state, and nonce.
func (g *GoogleProvider) AuthCodeURL(state, challenge, nonce string) string {
	q := url.Values{}
	q.Set("client_id", g.clientID)
	q.Set("redirect_uri", g.redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", pkce.Method)
	q.Set("nonce", nonce)
	return g.authEndpoint + "?" + q.Encode()
}

// ExchangeAndVerify swaps the auth code for tokens at Google's token endpoint
// (sending the PKCE code_verifier), then verifies the returned id_token before
// trusting any claim (§1 #4).
func (g *GoogleProvider) ExchangeAndVerify(ctx context.Context, code, verifier, nonce string) (OIDCClaims, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", g.redirectURI)
	form.Set("client_id", g.clientID)
	form.Set("client_secret", g.clientSecret)
	form.Set("code_verifier", verifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return OIDCClaims{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := g.do(req)
	if err != nil {
		return OIDCClaims{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return OIDCClaims{}, fmt.Errorf("google token endpoint status %d: %s", resp.StatusCode, string(body))
	}

	var tr struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return OIDCClaims{}, fmt.Errorf("decode token response: %w", err)
	}
	if tr.IDToken == "" {
		return OIDCClaims{}, fmt.Errorf("token response missing id_token")
	}
	return g.verifier.Verify(ctx, tr.IDToken, nonce)
}

var _ OAuthProvider = (*GoogleProvider)(nil)
