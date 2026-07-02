package ecom

import "context"

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
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
