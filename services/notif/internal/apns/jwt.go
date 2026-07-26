package apns

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// NoopTokenSource never talks to Apple; used when APNs creds are unset.
type NoopTokenSource struct {
	log *slog.Logger
}

func NewNoopTokenSource(log *slog.Logger) NoopTokenSource {
	if log == nil {
		log = slog.Default()
	}
	return NoopTokenSource{log: log}
}

func (n NoopTokenSource) Bearer(_ context.Context) (string, error) {
	n.log.Debug("apns noop token issued")
	return "noop", nil
}

// P8TokenSource signs short-lived ES256 JWTs from an Apple Auth Key (.p8).
type P8TokenSource struct {
	keyID  string
	teamID string
	key    *ecdsa.PrivateKey

	mu        sync.Mutex
	cached    string
	expiresAt time.Time
}

func NewP8TokenSource(keyPEM, keyID, teamID string) (*P8TokenSource, error) {
	if keyPEM == "" || keyID == "" || teamID == "" {
		return nil, errors.New("apns: key, key_id, and team_id required")
	}
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, errors.New("apns: invalid .p8 PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("apns: parse key: %w", err)
	}
	ecKey, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("apns: .p8 is not ECDSA")
	}
	return &P8TokenSource{keyID: keyID, teamID: teamID, key: ecKey}, nil
}

func NewP8TokenSourceFromEnv() (*P8TokenSource, error) {
	return NewP8TokenSource(os.Getenv("APNS_KEY_P8"), os.Getenv("APNS_KEY_ID"), os.Getenv("APNS_TEAM_ID"))
}

func (s *P8TokenSource) Bearer(_ context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached != "" && time.Now().Before(s.expiresAt) {
		return s.cached, nil
	}
	now := time.Now()
	// Apple allows up to 60m; cache for 50m.
	claims := jwt.RegisteredClaims{
		Issuer:    s.teamID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(50 * time.Minute)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = s.keyID
	signed, err := tok.SignedString(s.key)
	if err != nil {
		return "", err
	}
	s.cached = signed
	s.expiresAt = now.Add(50 * time.Minute)
	return signed, nil
}

// TokenSourceFromEnv returns a real P8 source when creds are set, else noop.
func TokenSourceFromEnv(log *slog.Logger) TokenSource {
	src, err := NewP8TokenSourceFromEnv()
	if err != nil || src == nil {
		return NewNoopTokenSource(log)
	}
	return src
}
