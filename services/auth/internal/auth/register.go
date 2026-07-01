package auth

import (
	"context"
	"strings"
)

type Service struct {
	repo   Repo
	params Argon2Params
}

func NewService(repo Repo, params Argon2Params) *Service {
	return &Service{
		repo:   repo,
		params: params,
	}
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (int64, error) {
	if in.Email == "" && in.Phone == "" {
		return 0, ErrNoIdentifier
	}
	if err := checkPasswordStrength(in.Password); err != nil {
		return 0, err
	}
	email := normalizeEmail(in.Email)
	hash, err := Hash(in.Password, s.params)
	if err != nil {
		return 0, err
	}
	u := AppUser{
		Email:   email,
		Phone:   in.Phone,
		Locale:  "vi-VN",
		Status:  "active",
		PwdHash: hash,
	}
	id, err := s.repo.InsertUser(ctx, u)
	if isUniqueViolation(err, "app_user_email_key") {
		return 0, ErrEmailTaken
	}
	return id, err
}

func checkPasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(email)
}
