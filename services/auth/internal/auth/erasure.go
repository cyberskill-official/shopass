package auth

import (
	"context"
)

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
