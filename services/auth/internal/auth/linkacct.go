package auth

import (
	"context"
	"regexp"
)

// looksLikeRawCredential checks if the string looks like an email, token, cookie etc.
// In reality, this could be more sophisticated.
func looksLikeRawCredential(s string) bool {
	// check for @ symbol (email)
	if match, _ := regexp.MatchString(`@`, s); match {
		return true
	}
	// check for things that look like session tokens (jwt, cookie strings with '=')
	if match, _ := regexp.MatchString(`=`, s); match {
		return true
	}
	if match, _ := regexp.MatchString(`^eyJ`, s); match { // JWT start
		return true
	}
	return false
}

func (s *Service) LinkAccount(ctx context.Context, userID int64, platformID int16, extUserRef string) error {
	if extUserRef == "" || len(extUserRef) > 128 {
		return ErrInvalidExtRef
	}
	if looksLikeRawCredential(extUserRef) {
		return ErrExtRefNotAnonymized
	}

	// Delegate to repo to upsert
	return s.repo.UpsertPlatformAccount(ctx, PlatformAccount{
		UserID:     userID,
		PlatformID: platformID,
		ExtUserRef: extUserRef,
	})
}

func (s *Service) UnlinkAccount(ctx context.Context, userID int64, platformID int16) error {
	return s.repo.DeletePlatformAccount(ctx, userID, platformID)
}

func (s *Service) ListLinks(ctx context.Context, userID int64) ([]PlatformAccount, error) {
	return s.repo.ListPlatformAccountsByUser(ctx, userID)
}
