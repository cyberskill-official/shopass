package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrInvalidResetToken = errors.New("invalid or expired reset token")
)

type Notifier interface {
	SendReset(ctx context.Context, u AppUser, token string) error
}

type LifecycleService struct {
	repo   Repo
	notif  Notifier
	params Argon2Params
}

func NewLifecycleService(repo Repo, notif Notifier, params Argon2Params) *LifecycleService {
	return &LifecycleService{repo: repo, notif: notif, params: params}
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *LifecycleService) RequestReset(ctx context.Context, identifier string) error {
	u, ok := s.repo.FindByIdentifier(ctx, identifier)
	if ok && u.Status == "active" {
		token := randToken()
		// We ignore error here to ensure we don't leak existence of the user
		_ = s.repo.SaveReset(ctx, u.ID, hashToken(token), time.Now().Add(30*time.Minute))
		if s.notif != nil {
			_ = s.notif.SendReset(ctx, u, token)
		}
	}
	return nil // LUÔN trả nil — phản hồi đồng nhất (§1 #3), không lộ tồn tại
}

func (s *LifecycleService) ConfirmReset(ctx context.Context, token, newPassword string) error {
	pr, ok := s.repo.FindResetByHash(ctx, hashToken(token))
	if !ok || pr.UsedAt != nil || time.Now().After(pr.ExpiresAt) {
		return ErrInvalidResetToken
	}

	// Assuming simple password check for this test
	if len(newPassword) < 8 {
		return errors.New("password too short")
	}

	hashedPwd, err := Hash(newPassword, s.params)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(ctx, pr.UserID, hashedPwd); err != nil {
		return err
	}
	if err := s.repo.MarkResetUsed(ctx, pr.ID); err != nil {
		return err
	}

	// đổi mật khẩu = đăng xuất mọi phiên (§1 #4)
	return s.repo.RevokeAllRefresh(ctx, pr.UserID)
}
