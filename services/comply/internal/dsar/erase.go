package dsar

import (
	"context"
)

// Erase: hard-delete du lieu thuan, soft-anonymize du lieu rang buoc, GIU consent log.
func (s *Service) Erase(ctx context.Context, userID int64) (EraseResult, error) {
	var r EraseResult
	var err error
	
	r.WishlistDeleted, err = s.track.HardDeleteByUser(ctx, userID) // du lieu thuan
	if err != nil {
		return r, err
	}
	
	r.PaymentsAnonymized, err = s.bill.AnonymizeByUser(ctx, userID) // rang buoc ke toan
	if err != nil {
		return r, err
	}
	
	err = s.users.Anonymize(ctx, userID) // thay PII app_user bang gia tri an danh
	if err != nil {
		return r, err
	}
	
	// KHONG dong toi consent_record (chung cu phap ly).
	r.ConsentLogRetained = true
	r.Status = "completed"
	
	// Tu dong ghi DSAR de luu dau vet erase
	dsarID, err := s.CreateRequest(ctx, userID, "erase")
	if err == nil {
		s.MarkCompleted(ctx, dsarID)
	}

	return r, nil
}
