package dsar

import (
	"context"
	"time"
)

// Export gom du lieu cua DUNG user_id (chong ro cheo user).
func (s *Service) Export(ctx context.Context, userID int64) (ExportBundle, error) {
	acc, err := s.users.View(ctx, userID)
	if err != nil {
		return ExportBundle{}, err
	}
	prods, err := s.track.ByUser(ctx, userID) // moi truy van rang buoc user_id
	if err != nil {
		return ExportBundle{}, err
	}
	consent, err := s.consent.HistoryAll(ctx, userID)
	if err != nil {
		return ExportBundle{}, err
	}
	return ExportBundle{
		Account:         acc,
		TrackedProducts: prods,
		ConsentHistory:  consent,
		GeneratedAt:     time.Now(),
	}, nil
}
