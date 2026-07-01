package breach

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrSubjectsNotNotified = errors.New("cannot close severe breach without notifying subjects")
)

var order = map[Status]int{
	"detected":           0,
	"triaged":            1,
	"notified_authority": 2,
	"notified_subjects":  3,
	"closed":             4,
}

type repo interface {
	create(ctx context.Context, in BreachInput, ack time.Time) (int64, error)
	get(ctx context.Context, id int64) (BreachIncident, error)
	transition(ctx context.Context, id int64, to Status, t time.Time) error
	overdue(ctx context.Context) ([]BreachIncident, error)
}

type Service struct {
	repo repo
}

func NewService(r repo) *Service {
	return &Service{repo: r}
}

func (s *Service) Open(ctx context.Context, in BreachInput) (int64, error) {
	return s.repo.create(ctx, in, time.Now())
}

// Advance chi cho phep tien dung mot buoc tuan tu.
func (s *Service) Advance(ctx context.Context, id int64, to Status) error {
	b, err := s.repo.get(ctx, id)
	if err != nil {
		return err
	}
	if order[to] != order[b.Status]+1 {
		return ErrInvalidTransition // cam lui hoac nhay buoc
	}
	return s.repo.transition(ctx, id, to, time.Now())
}

// Close tu choi dong neu nghiem trong ma chua thong bao chu the.
func (s *Service) Close(ctx context.Context, id int64) error {
	b, err := s.repo.get(ctx, id)
	if err != nil {
		return err
	}
	if (b.Severity == "high" || b.Severity == "critical") && b.NotifiedSubjectsAt == nil {
		return ErrSubjectsNotNotified
	}
	// Note: Close is equivalent to Advance to "closed". If the order logic requires it,
	// we might need to check if the previous state is valid. But the spec says:
	// "Tiến đủ chuỗi high qua notified_subjects rồi Close -> đóng được."
	// And "Close sự cố low (không cần thông báo chủ thể) -> đóng được, closed_at set."
	// Let's ensure it handles order properly by calling transition directly since order for Close might bypass some steps for "low"?
	// Wait, the spec says: "Advance(detected->triaged)... Advance(triaged->notified_authority)... Advance(notified_authority->notified_subjects)... Close"
	// For low, it bypasses notified_subjects: "notified_authority -> closed".
	// If it's low, and current is notified_authority (2), closed is (4). So order is not exactly +1.
	// So Close is a special action that checks if it's safe to jump to closed.
	
	// Valid states to close from:
	if b.Status == "closed" {
		return ErrInvalidTransition // Already closed
	}
	if order[b.Status] < order["notified_authority"] {
		return ErrInvalidTransition // Must at least notify authority
	}
	
	return s.repo.transition(ctx, id, "closed", time.Now())
}

func (s *Service) Overdue(ctx context.Context) ([]BreachIncident, error) {
	// To comply with "metric breach_overdue_total phan anh" and "Overdue() tra dung"
	return s.repo.overdue(ctx)
}
