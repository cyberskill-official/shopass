package referral

import (
	"context"
	"errors"
	"log/slog"
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
	repo       Repo
	bus        EventBus
	metrics    AttributeMetrics
	edgeWriter ReferralEdgeWriter
	assessor   FraudAssessor
	log        *slog.Logger
}

type ReferralEdgeWriter interface {
	UpsertReferralEdge(ctx context.Context, referrerID, refereeID int64) error
}

type FraudAssessor interface {
	Assess(ctx context.Context, userID int64, extras map[string]any) error
}

type AssessorFunc func(ctx context.Context, userID int64, extras map[string]any) error

func (f AssessorFunc) Assess(ctx context.Context, userID int64, extras map[string]any) error {
	return f(ctx, userID, extras)
}

type Option func(*Service)

func WithReferralFraud(edgeWriter ReferralEdgeWriter, assessor FraudAssessor) Option {
	return func(s *Service) {
		s.edgeWriter = edgeWriter
		s.assessor = assessor
	}
}

func WithLogger(log *slog.Logger) Option {
	return func(s *Service) {
		s.log = log
	}
}

func NewService(repo Repo, bus EventBus, opts ...Option) *Service {
	s := &Service{repo: repo, bus: bus, metrics: dummyMetrics{}, log: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	if s.log == nil {
		s.log = slog.Default()
	}
	return s
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
	s.recordFraudSignals(ctx, rc.UserID, refereeID)
	s.bus.Publish(ctx, ReferralAttributed{
		ReferrerID: rc.UserID,
		RefereeID:  refereeID,
		At:         time.Now(),
	})
	s.metrics.Attributed()
	return nil
}

func (s *Service) recordFraudSignals(ctx context.Context, referrerID, refereeID int64) {
	if s.edgeWriter != nil {
		if err := s.edgeWriter.UpsertReferralEdge(ctx, referrerID, refereeID); err != nil {
			s.log.Warn("referral.fraud_edge_upsert_failed", "referrer_id", referrerID, "referee_id", refereeID, "err", err)
		}
	}
	if s.assessor != nil {
		err := s.assessor.Assess(ctx, referrerID, map[string]any{
			"event":      "referral_attributed",
			"referee_id": refereeID,
		})
		if err != nil {
			s.log.Warn("referral.fraud_assess_failed", "referrer_id", referrerID, "referee_id", refereeID, "err", err)
		}
	}
}
