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

func (s *TokenService) issuePairInFamily(ctx context.Context, userID int64, familyID string) (TokenPair, error) {
	// Re-fetch user to get current tier and locale (simplification here)
	// Alternatively, we could pass it or fetch from Repo.
	// For now, let's create a minimal AppUser with just ID since we only strictly need ID for refresh tests.
	// A robust impl would `FindByID`.
	u := AppUser{ID: userID} 

	access, err := s.IssueAccess(u)
	if err != nil {
		return TokenPair{}, err
	}

	rawRefresh, err := generateRandomToken()
	if err != nil {
		return TokenPair{}, err
	}
	
	// Refresh token lives for 30 days
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	
	err = s.repo.InsertRefreshToken(ctx, userID, hashStr(rawRefresh), familyID, expiresAt)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		Access:  access,
		Refresh: rawRefresh,
		Type:    "Bearer",
		Expires: int(s.accessTTL.Seconds()),
	}, nil
}

func (s *TokenService) IssueTokenPair(ctx context.Context, userID int64) (TokenPair, error) {
	familyID := uuid.New().String()
	return s.issuePairInFamily(ctx, userID, familyID)
}

func (s *TokenService) Refresh(ctx context.Context, raw string) (TokenPair, error) {
	rt, err := s.repo.FindRefreshByHash(ctx, hashStr(raw))
	if err != nil {
		return TokenPair{}, ErrInvalidRefresh
	}
	
	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		return TokenPair{}, ErrInvalidRefresh
	}
	
	if rt.UsedAt != nil { // đã xoay rồi mà dùng lại = nghi đánh cắp (§1 #10)
		_ = s.repo.RevokeFamily(ctx, rt.FamilyID)
		return TokenPair{}, ErrRefreshReuseDetected
	}
	
	_ = s.repo.MarkUsed(ctx, rt.ID) // dùng một lần (§1 #7)
	
	return s.issuePairInFamily(ctx, rt.UserID, rt.FamilyID)
}
