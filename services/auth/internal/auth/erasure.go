package auth

import (
	"context"
	"time"
)

// AccountDeleteGrace is the soft-delete window before hard purge (§1 #8 / DEC-AUTH-25).
// Hard purge is an ops job; this constant only informs API clients.
const AccountDeleteGrace = 14 * 24 * time.Hour

func (s *LifecycleService) DeleteAccount(ctx context.Context, userID int64) error {
	// DSAR PDPL: ẩn danh hóa PII + gỡ liên kết + thu hồi token + đặt deleted (§1 #7,#9).
	if err := s.repo.AnonymizePII(ctx, userID); err != nil {
		return err
	}
	if err := s.repo.DeletePlatformAccounts(ctx, userID); err != nil {
		return err
	}
	if err := s.repo.RevokeAllRefresh(ctx, userID); err != nil {
		return err
	}
	return s.repo.SetStatus(ctx, userID, "deleted") // ân hạn trước purge cứng (§1 #8)
}

// GraceUntil returns when hard purge may run after a successful DeleteAccount.
func GraceUntil(from time.Time) time.Time {
	return from.UTC().Add(AccountDeleteGrace)
}
