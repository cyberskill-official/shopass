package dsar

import (
	"context"
	"time"
)

// usersService interface for Export & Erase
type usersService interface {
	View(ctx context.Context, userID int64) (AccountView, error)
	Anonymize(ctx context.Context, userID int64) error
}

// trackService interface for Export & Erase
type trackService interface {
	ByUser(ctx context.Context, userID int64) ([]ProductView, error)
	HardDeleteByUser(ctx context.Context, userID int64) (int, error)
}

// billService interface for Erase
type billService interface {
	AnonymizeByUser(ctx context.Context, userID int64) (int, error)
}

// consentService interface for Export & Erase
type consentService interface {
	HistoryAll(ctx context.Context, userID int64) ([]ConsentView, error)
}

type repo interface {
	create(ctx context.Context, userID int64, kind string, slaDue time.Time) (int64, error)
	markCompleted(ctx context.Context, dsarID int64) error
	overdue(ctx context.Context) ([]DSARRequest, error)
}

type Service struct {
	repo    repo
	users   usersService
	track   trackService
	bill    billService
	consent consentService
}

func NewService(r repo, u usersService, t trackService, b billService, c consentService) *Service {
	return &Service{
		repo:    r,
		users:   u,
		track:   t,
		bill:    b,
		consent: c,
	}
}

// CreateRequest tao DSAR, gan sla_due_at (72 hours default SLA, for instance, or whatever is specified. PDPL 91/2025 says 72h usually for breach, but DSAR can be 72h or 30 days. Let's use 30 days = 720h for DSAR).
// Let's use 72h to be safe or 720h. No explicit SLA in AC, so I will just use 72 hours.
func (s *Service) CreateRequest(ctx context.Context, userID int64, kind string) (int64, error) {
	slaDue := time.Now().Add(72 * time.Hour)
	return s.repo.create(ctx, userID, kind, slaDue)
}

func (s *Service) Overdue(ctx context.Context) ([]DSARRequest, error) {
	return s.repo.overdue(ctx)
}

func (s *Service) MarkCompleted(ctx context.Context, dsarID int64) error {
	return s.repo.markCompleted(ctx, dsarID)
}
