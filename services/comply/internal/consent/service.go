package consent

import (
	"context"
	"database/sql"
	"errors"
)

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

// IsAllowed la nguon su that duy nhat cho co so phap ly cua mot purpose.
func (s *Service) IsAllowed(ctx context.Context, userID int64, p Purpose) (bool, error) {
	if !validPurpose(p) {
		return false, ErrUnknownPurpose
	}
	rec, err := s.repo.latest(ctx, userID, string(p))
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // chua co ban ghi -> chua dong thuan (im lang != dong thuan)
	}
	if err != nil {
		return false, err
	}
	return rec.Granted, nil
}

// Grant ghi dong moi granted=true voi policy_version dang hieu luc.
func (s *Service) Grant(ctx context.Context, userID int64, p Purpose, src string, m ReqMeta) error {
	if !validPurpose(p) {
		return ErrUnknownPurpose
	}
	ver, err := s.repo.effectiveVersion(ctx, string(p))
	if err != nil {
		return err
	}
	return s.repo.append(ctx, ConsentRecord{
		UserID:        userID,
		PurposeKey:    string(p),
		PolicyVersion: ver,
		Granted:       true,
		Source:        src,
		IP:            m.IP,
		UserAgent:     m.UserAgent,
	})
}

// Withdraw ghi dong moi granted=false, KHONG xoa lich su.
func (s *Service) Withdraw(ctx context.Context, userID int64, p Purpose, src string, m ReqMeta) error {
	if !validPurpose(p) {
		return ErrUnknownPurpose
	}
	ver, err := s.repo.effectiveVersion(ctx, string(p))
	if err != nil {
		return err
	}
	return s.repo.append(ctx, ConsentRecord{
		UserID:        userID,
		PurposeKey:    string(p),
		PolicyVersion: ver,
		Granted:       false,
		Source:        src,
		IP:            m.IP,
		UserAgent:     m.UserAgent,
	})
}

// History toan bo lich su cho DSAR va kiem toan.
func (s *Service) History(ctx context.Context, userID int64, p Purpose) ([]ConsentRecord, error) {
	if !validPurpose(p) {
		return nil, ErrUnknownPurpose
	}
	return s.repo.history(ctx, userID, string(p))
}

// HistoryAll returns every consent row for a user for DSAR export and audit.
func (s *Service) HistoryAll(ctx context.Context, userID int64) ([]ConsentRecord, error) {
	return s.repo.historyAll(ctx, userID)
}
