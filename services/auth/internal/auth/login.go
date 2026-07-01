package auth

import (
	"context"
)

// In a real application, Login might live in a higher-level Service struct, 
// but based on FR-AUTH-002, we can implement it here.
func (s *TokenService) Login(ctx context.Context, email, password string) (TokenPair, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return TokenPair{}, err
	}
	if u.ID == 0 {
		return TokenPair{}, ErrInvalidCredentials
	}
	if u.Status != "active" {
		return TokenPair{}, ErrAccountNotActive
	}

	// Verify password
	// Assume password hashing uses specific params, but for this layer we'll just check if Verify returns true.
	// Since password verification is part of FR-AUTH-001, we mock/use a simple check here or reuse the logic.
	// Wait, we can use the Verify method from password.go if it exists.
	// We'll just verify the hash.
	ok, err := Verify(password, u.PwdHash)
	if err != nil || !ok {
		return TokenPair{}, ErrInvalidCredentials
	}

	return s.IssueTokenPair(ctx, u.ID)
}
