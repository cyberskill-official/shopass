package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"shopass/services/auth/internal/auth/pkce"
)

var (
	// ErrBadState is returned when the callback state is missing, unknown, or
	// does not match the provider it claims (CSRF / replay defense, §1 #3).
	ErrBadState = errors.New("invalid or missing oauth state")
	// ErrUnknownProvider is returned for a provider with no configured adapter.
	ErrUnknownProvider = errors.New("unknown oauth provider")
)

// OAuthProvider is one identity provider (Google today; Facebook/Zalo behind the
// same interface, DEC-AUTH-16). It builds the authorize URL and, on callback,
// exchanges the code and verifies the returned id_token.
type OAuthProvider interface {
	AuthCodeURL(state, challenge, nonce string) string
	ExchangeAndVerify(ctx context.Context, code, verifier, nonce string) (OIDCClaims, error)
}

// SocialRepo is the persistence the OAuth flow needs. It is deliberately narrow
// and separate from the main Repo interface so adding it does not disturb other
// implementations. pgRepo satisfies it (social_repo.go).
type SocialRepo interface {
	FindBySocial(ctx context.Context, provider, subject string) (userID int64, found bool, err error)
	FindByEmail(ctx context.Context, email string) (AppUser, error)
	LinkSocial(ctx context.Context, userID int64, provider, subject string) error
	CreateSocialUser(ctx context.Context, email, locale string) (int64, error)
}

// TokenIssuer issues the shared TokenPair of TASK-AUTH-002 (DEC-AUTH-20): social
// login goes down the same session path as password login. *TokenService fits.
type TokenIssuer interface {
	IssueTokenPair(ctx context.Context, userID int64) (TokenPair, error)
}

// OAuthService drives the Authorization Code + PKCE flow (DEC-AUTH-17).
type OAuthService struct {
	providers map[string]OAuthProvider
	tmp       TmpStore
	repo      SocialRepo
	tokens    TokenIssuer
	ttl       time.Duration
}

// NewOAuthService wires the providers, the short-lived state store, the social
// repo, and the token issuer.
func NewOAuthService(providers map[string]OAuthProvider, tmp TmpStore, repo SocialRepo, tokens TokenIssuer) *OAuthService {
	return &OAuthService{providers: providers, tmp: tmp, repo: repo, tokens: tokens, ttl: 5 * time.Minute}
}

// StartOAuth begins a flow: it generates a PKCE verifier, a random state (CSRF)
// and a nonce (replay), stores them for the short TTL, and returns the provider
// authorize URL (§1 #2).
func (s *OAuthService) StartOAuth(ctx context.Context, provider string) (string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", ErrUnknownProvider
	}
	verifier, err := pkce.NewVerifier()
	if err != nil {
		return "", err
	}
	state, err := oauthRandToken()
	if err != nil {
		return "", err
	}
	nonce, err := oauthRandToken()
	if err != nil {
		return "", err
	}
	s.tmp.Put(ctx, "oauth:"+state, oauthTmp{Verifier: verifier, Provider: provider, Nonce: nonce}, s.ttl)
	return p.AuthCodeURL(state, pkce.Challenge(verifier), nonce), nil
}

// OAuthCallback completes a flow: it validates state (one-time), exchanges the
// code and verifies the id_token, resolves the app_user, and issues the shared
// TokenPair (§1 #3, #4, #8).
func (s *OAuthService) OAuthCallback(ctx context.Context, provider, code, state string) (TokenPair, error) {
	tmp, ok := s.tmp.Take(ctx, "oauth:"+state)
	if !ok || tmp.Provider != provider {
		return TokenPair{}, ErrBadState
	}
	p, ok := s.providers[provider]
	if !ok {
		return TokenPair{}, ErrUnknownProvider
	}
	claims, err := p.ExchangeAndVerify(ctx, code, tmp.Verifier, tmp.Nonce)
	if err != nil {
		return TokenPair{}, err
	}
	uid, err := s.resolveUser(ctx, provider, claims)
	if err != nil {
		return TokenPair{}, err
	}
	return s.tokens.IssueTokenPair(ctx, uid)
}

// resolveUser maps verified provider claims to an app_user (§1 #6, #7):
//  1. an existing social_identity wins outright;
//  2. otherwise, only if the provider marked the email verified AND an account
//     with that email exists, link the new identity to it;
//  3. otherwise create a new social-only account. An unverified email is never
//     used to merge into or create against an existing email - that is the
//     classic account-takeover hole (DEC-AUTH-19). The new account is created
//     with a NULL email in that case, so it cannot collide with the victim.
func (s *OAuthService) resolveUser(ctx context.Context, provider string, c OIDCClaims) (int64, error) {
	if uid, found, err := s.repo.FindBySocial(ctx, provider, c.Subject); err != nil {
		return 0, err
	} else if found {
		return uid, nil
	}

	if c.EmailVerified && c.Email != "" {
		if u, err := s.repo.FindByEmail(ctx, c.Email); err != nil {
			return 0, err
		} else if u.ID != 0 {
			if err := s.repo.LinkSocial(ctx, u.ID, provider, c.Subject); err != nil {
				return 0, err
			}
			return u.ID, nil
		}
	}

	emailToStore := ""
	if c.EmailVerified {
		emailToStore = c.Email // only a verified email is attached to a new account
	}
	uid, err := s.repo.CreateSocialUser(ctx, emailToStore, "vi-VN")
	if err != nil {
		return 0, err
	}
	if err := s.repo.LinkSocial(ctx, uid, provider, c.Subject); err != nil {
		return 0, err
	}
	return uid, nil
}

// oauthRandToken returns a URL-safe 256-bit random string for state and nonce.
func oauthRandToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
