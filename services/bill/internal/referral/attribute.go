package referral

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUnknownCode       = errors.New("unknown referral code")
	ErrSelfReferral      = errors.New("cannot refer yourself")
	ErrAlreadyAttributed = errors.New("user already has a referrer")
)

type EventBus interface {
	Publish(ctx context.Context, event interface{})
}

type ReferralAttributed struct {
	ReferrerID int64
	RefereeID  int64
	At         time.Time
}

type AttributeMetrics interface {
	Attributed()
}

type dummyMetrics struct{}
func (dummyMetrics) Attributed() {}

type Service struct {
	repo    Repo
	bus     EventBus
	metrics AttributeMetrics
}

func NewService(repo Repo, bus EventBus) *Service {
	return &Service{repo: repo, bus: bus, metrics: dummyMetrics{}}
}

func (s *Service) Attribute(ctx context.Context, refereeID int64, code string) error {
	rc, ok, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnknownCode
	}
	if rc.UserID == refereeID {
		return ErrSelfReferral
	}

	already, err := s.repo.HasReferrer(ctx, refereeID)
	if err != nil {
		return err
	}
	if already {
		return ErrAlreadyAttributed
	}

	if err := s.repo.SetReferrer(ctx, refereeID, rc.ID); err != nil {
		return err
	}
	if err := s.repo.IncrementUses(ctx, rc.ID); err != nil {
		return err
	}
	s.bus.Publish(ctx, ReferralAttributed{
		ReferrerID: rc.UserID,
		RefereeID:  refereeID,
		At:         time.Now(),
	})
	s.metrics.Attributed()
	return nil
}
