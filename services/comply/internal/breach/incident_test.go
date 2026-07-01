package breach

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	m map[int64]*BreachIncident
}

func (r *mockRepo) create(ctx context.Context, in BreachInput, ack time.Time) (int64, error) {
	id := int64(len(r.m) + 1)
	r.m[id] = &BreachIncident{
		ID:             id,
		Summary:        in.Summary,
		Severity:       in.Severity,
		Status:         "detected",
		AcknowledgedAt: ack,
	}
	return id, nil
}

func (r *mockRepo) get(ctx context.Context, id int64) (BreachIncident, error) {
	if b, ok := r.m[id]; ok {
		return *b, nil
	}
	return BreachIncident{}, nil
}

func (r *mockRepo) transition(ctx context.Context, id int64, to Status, t time.Time) error {
	b := r.m[id]
	b.Status = to
	switch to {
	case "triaged":
		b.TriagedAt = &t
	case "notified_authority":
		b.NotifiedAuthorityAt = &t
	case "notified_subjects":
		b.NotifiedSubjectsAt = &t
	case "closed":
		b.ClosedAt = &t
	}
	return nil
}

func (r *mockRepo) overdue(ctx context.Context) ([]BreachIncident, error) {
	return nil, nil
}

func setupMock(t *testing.T) *Service {
	return NewService(&mockRepo{m: make(map[int64]*BreachIncident)})
}

func TestAdvance_SequentialOnly(t *testing.T) {
	s := setupMock(t)
	ctx := context.Background()
	id, _ := s.Open(ctx, BreachInput{Summary: "log leak", Severity: "high"})
	require.NoError(t, s.Advance(ctx, id, "triaged"))
	err := s.Advance(ctx, id, "notified_subjects") // nhay buoc (chua notified_authority)
	require.ErrorIs(t, err, ErrInvalidTransition)
}

func TestAdvance_NoBackward(t *testing.T) {
	s := setupMock(t)
	ctx := context.Background()
	id, _ := s.Open(ctx, BreachInput{Summary: "x", Severity: "low"})
	s.Advance(ctx, id, "triaged")
	err := s.Advance(ctx, id, "detected") // lui
	require.ErrorIs(t, err, ErrInvalidTransition)
}

func TestClose_CriticalNeedsSubjectNotice(t *testing.T) {
	s := setupMock(t)
	ctx := context.Background()
	id, _ := s.Open(ctx, BreachInput{Summary: "PII leak", Severity: "critical"})
	s.Advance(ctx, id, "triaged")
	s.Advance(ctx, id, "notified_authority")
	err := s.Close(ctx, id) // chua notified_subjects
	require.ErrorIs(t, err, ErrSubjectsNotNotified)
}

func TestClose_LowSeverityNoSubjectNotice(t *testing.T) {
	s := setupMock(t)
	ctx := context.Background()
	id, _ := s.Open(ctx, BreachInput{Summary: "minor", Severity: "low"})
	s.Advance(ctx, id, "triaged")
	s.Advance(ctx, id, "notified_authority")
	require.NoError(t, s.Close(ctx, id)) // low khong can notified_subjects
}
