package auth

import (
	"context"
	"errors"
)

var (
	ErrBadStatus = errors.New("invalid status")
)

func validStatus(status string) bool {
	return status == "active" || status == "suspended" || status == "deleted"
}

func (s *LifecycleService) SetStatus(ctx context.Context, userID int64, status string) error {
	if !validStatus(status) {
		return ErrBadStatus
	}
	if err := s.repo.SetStatus(ctx, userID, status); err != nil {
		return err
	}
	if status == "suspended" || status == "deleted" {
		return s.repo.RevokeAllRefresh(ctx, userID) // thu hồi token (§1 #5)
	}
	return nil
}
