package fcm

import (
	"context"
	"fmt"
	"os"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

// GoogleOAuth is an OAuthTokenSource backed by a GCP service-account JSON
// (file path via GOOGLE_APPLICATION_CREDENTIALS, or raw JSON in
// FCM_SERVICE_ACCOUNT_JSON).
type GoogleOAuth struct {
	ts oauth2.TokenSource
	mu sync.Mutex
}

// NewGoogleOAuthFromEnv loads credentials from the environment.
// Returns (nil, nil) when neither credential source is set (FCM disabled).
func NewGoogleOAuthFromEnv(ctx context.Context) (*GoogleOAuth, error) {
	if raw := os.Getenv("FCM_SERVICE_ACCOUNT_JSON"); raw != "" {
		creds, err := google.CredentialsFromJSON(ctx, []byte(raw), fcmScope)
		if err != nil {
			return nil, fmt.Errorf("fcm: parse FCM_SERVICE_ACCOUNT_JSON: %w", err)
		}
		return &GoogleOAuth{ts: creds.TokenSource}, nil
	}
	if path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("fcm: read GOOGLE_APPLICATION_CREDENTIALS: %w", err)
		}
		creds, err := google.CredentialsFromJSON(ctx, b, fcmScope)
		if err != nil {
			return nil, fmt.Errorf("fcm: parse service account file: %w", err)
		}
		return &GoogleOAuth{ts: creds.TokenSource}, nil
	}
	return nil, nil
}

// Token returns a fresh Bearer access token for FCM HTTP v1.
func (g *GoogleOAuth) Token(ctx context.Context) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	tok, err := g.ts.Token()
	if err != nil {
		return "", fmt.Errorf("fcm: oauth token: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("fcm: empty oauth access token")
	}
	return tok.AccessToken, nil
}
