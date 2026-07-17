package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// hash computes the sha256 hash of a raw token string
func hashStr(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

const refreshTTL = 30 * 24 * time.Hour

// prepareTokenPair creates a pair without persisting it. Refresh rotation uses
// this so the old-token state change and replacement-token insert can happen in
// one repository transaction.
func (s *TokenService) prepareTokenPair(userID int64) (TokenPair, string, time.Time, error) {
	access, err := s.IssueAccess(AppUser{ID: userID})
	if err != nil {
		return TokenPair{}, "", time.Time{}, err
	}

	rawRefresh, err := generateRandomToken()
	if err != nil {
		return TokenPair{}, "", time.Time{}, err
	}

	pair := TokenPair{
		Access:  access,
		Refresh: rawRefresh,
		Type:    "Bearer",
		Expires: int(s.accessTTL.Seconds()),
	}
	return pair, hashStr(rawRefresh), time.Now().Add(refreshTTL), nil
}

func (s *TokenService) issuePairInFamily(ctx context.Context, userID int64, familyID string) (TokenPair, error) {
	pair, refreshHash, expiresAt, err := s.prepareTokenPair(userID)
	if err != nil {
		return TokenPair{}, err
	}
	if err := s.repo.InsertRefreshToken(ctx, userID, refreshHash, familyID, expiresAt); err != nil {
		return TokenPair{}, err
	}
	return pair, nil
}

func (s *TokenService) IssueTokenPair(ctx context.Context, userID int64) (TokenPair, error) {
	familyID := uuid.New().String()
	return s.issuePairInFamily(ctx, userID, familyID)
}

func (s *TokenService) Refresh(ctx context.Context, raw string) (TokenPair, error) {
	if raw == "" {
		return TokenPair{}, ErrInvalidRefresh
	}
	oldHash := hashStr(raw)

	// This read is only used to prepare an access token with the right user ID.
	// It is not authorization: RotateRefreshToken below makes the authoritative
	// conditional state transition inside one database transaction.
	rt, err := s.repo.FindRefreshByHash(ctx, oldHash)
	if err != nil {
		return TokenPair{}, ErrInvalidRefresh
	}

	pair, replacementHash, replacementExpiresAt, err := s.prepareTokenPair(rt.UserID)
	if err != nil {
		return TokenPair{}, err
	}

	status, err := s.repo.RotateRefreshToken(ctx, oldHash, replacementHash, replacementExpiresAt)
	if err != nil {
		return TokenPair{}, err
	}
	switch status {
	case RefreshRotationSucceeded:
		return pair, nil
	case RefreshRotationReuseDetected:
		return TokenPair{}, ErrRefreshReuseDetected
	default:
		return TokenPair{}, ErrInvalidRefresh
	}
}

// Logout revokes the entire refresh-token family for this browser session. It
// accepts a refresh token because the web keeps it in an HttpOnly cookie and
// cannot read it from client JavaScript.
func (s *TokenService) Logout(ctx context.Context, raw string) error {
	rt, err := s.repo.FindRefreshByHash(ctx, hashStr(raw))
	if err != nil {
		return ErrInvalidRefresh
	}
	return s.repo.RevokeFamily(ctx, rt.FamilyID)
}
