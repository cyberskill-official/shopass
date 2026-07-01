package dpia

import (
	"context"
	"errors"
)

var ErrTIARequired = errors.New("TIA is required for cross-border processing")

type repo interface {
	createWithDPIA(ctx context.Context, a ProcessingActivity, in DPIAInput) (int64, error)
	createDPIAVersion(ctx context.Context, activityID int64, in DPIAInput) error
	markFiled(ctx context.Context, dpiaID int64) error
	overdue(ctx context.Context) ([]ActivityStatus, error)
	report(ctx context.Context) ([]ActivityStatus, error) // Can expand to full export later
}

type Service struct {
	repo repo
}

func NewService(r repo) *Service {
	return &Service{repo: r}
}

// RegisterActivity tao activity + DPIA v1; bat buoc TIA khi cross_border.
func (s *Service) RegisterActivity(ctx context.Context, a ProcessingActivity, in DPIAInput) (int64, error) {
	if a.CrossBorder && (in.TIA == nil || in.TIA.Safeguard == "") {
		return 0, ErrTIARequired
	}
	return s.repo.createWithDPIA(ctx, a, in)
}

func (s *Service) ReviewDPIA(ctx context.Context, activityID int64, in DPIAInput) error {
	return s.repo.createDPIAVersion(ctx, activityID, in)
}

func (s *Service) MarkFiled(ctx context.Context, dpiaID int64) error {
	return s.repo.markFiled(ctx, dpiaID)
}

func (s *Service) Overdue(ctx context.Context) ([]ActivityStatus, error) {
	return s.repo.overdue(ctx)
}

func (s *Service) Report(ctx context.Context) ([]ActivityStatus, error) {
	return s.repo.report(ctx)
}
