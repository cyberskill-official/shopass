package ecom

import "context"

// store abstracts the ecom persistence layer so the DB-backed *Repo is used in
// production and an in-memory fake is used in unit tests (FR-COMPLY-008).
type store interface {
	txCount(ctx context.Context, year int) (int64, error)
	threshold(ctx context.Context, key string) (int64, error)
	Obligations(ctx context.Context) ([]EcommerceObligation, error)
	MarkObligation(ctx context.Context, key, status string) error
}

type Service struct {
	repo store
}

func NewService(repo store) *Service {
	return &Service{repo: repo}
}

// Threshold suy co tu bo dem, KHONG nhap tay.
func (s *Service) Threshold(ctx context.Context, year int) (ThresholdState, error) {
	cnt, err := s.repo.txCount(ctx, year)
	if err != nil {
		return ThresholdState{}, err
	}
	th, err := s.repo.threshold(ctx, "foreign_platform_yearly_tx")
	if err != nil {
		return ThresholdState{}, err
	}
	return ThresholdState{
		Year:         year,
		Count:        cnt,
		Threshold:    th,
		MustRegister: cnt > th, // vuot nguong -> phai dang ky
	}, nil
}

func (s *Service) Reload(ctx context.Context) {
	// Currently unused. If thresholds are cached in-memory in the future,
	// this method should clear the cache and trigger a fresh DB read.
}
