package referral

import (
	"context"
	"testing"
)

type mockRepo struct {
	codes      map[string]*ReferralCode
	attributed map[int64]int64 // referee -> codeID
}

func (m *mockRepo) FindByCode(ctx context.Context, code string) (ReferralCode, bool, error) {
	if c, ok := m.codes[code]; ok {
		return *c, true, nil
	}
	return ReferralCode{}, false, nil
}
func (m *mockRepo) HasReferrer(ctx context.Context, userID int64) (bool, error) {
	_, ok := m.attributed[userID]
	return ok, nil
}
func (m *mockRepo) SetReferrer(ctx context.Context, userID int64, codeID int64) error {
	m.attributed[userID] = codeID
	return nil
}
func (m *mockRepo) IncrementUses(ctx context.Context, codeID int64) error {
	for _, c := range m.codes {
		if c.ID == codeID {
			c.Uses++
		}
	}
	return nil
}
func (m *mockRepo) CreateCodeForUser(ctx context.Context, userID int64) (string, error) { return "", nil }

type mockEventBus struct {
	events []interface{}
}

func (m *mockEventBus) Publish(ctx context.Context, event interface{}) {
	m.events = append(m.events, event)
}
func (m *mockEventBus) CountOf(eventType string) int {
	return len(m.events) // naive check
}

func setup() (*Service, *mockRepo, *mockEventBus) {
	repo := &mockRepo{
		codes:      make(map[string]*ReferralCode),
		attributed: make(map[int64]int64),
	}
	bus := &mockEventBus{}
	s := NewService(repo, bus)
	return s, repo, bus
}

func TestAttribute_Valid(t *testing.T) {
	s, repo, _ := setup()
	repo.codes["SD123"] = &ReferralCode{ID: 1, UserID: 100, Code: "SD123", Uses: 0}
	
	err := s.Attribute(context.Background(), 200, "SD123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.codes["SD123"].Uses != 1 {
		t.Fatalf("expected uses to increment to 1")
	}
	if repo.attributed[200] != 1 {
		t.Fatalf("expected user to be attributed to code ID 1")
	}
}

func TestAttribute_SelfReferral_Blocked(t *testing.T) {
	s, repo, _ := setup()
	repo.codes["SD123"] = &ReferralCode{ID: 1, UserID: 100, Code: "SD123", Uses: 0}
	
	err := s.Attribute(context.Background(), 100, "SD123")
	if err != ErrSelfReferral {
		t.Fatalf("expected ErrSelfReferral, got %v", err)
	}
	if repo.codes["SD123"].Uses != 0 {
		t.Fatalf("expected uses to remain 0")
	}
}

func TestAttribute_AlreadyAttributed_Blocked(t *testing.T) {
	s, repo, _ := setup()
	repo.codes["SD123"] = &ReferralCode{ID: 1, UserID: 100, Code: "SD123", Uses: 0}
	repo.attributed[200] = 99 // already attributed
	
	err := s.Attribute(context.Background(), 200, "SD123")
	if err != ErrAlreadyAttributed {
		t.Fatalf("expected ErrAlreadyAttributed, got %v", err)
	}
	if repo.codes["SD123"].Uses != 0 {
		t.Fatalf("expected uses to remain 0")
	}
}

func TestAttribute_UnknownCode(t *testing.T) {
	s, _, _ := setup()
	err := s.Attribute(context.Background(), 200, "SDNOPE")
	if err != ErrUnknownCode {
		t.Fatalf("expected ErrUnknownCode, got %v", err)
	}
}

func TestAttribute_PublishesEvent_NoDirectReward(t *testing.T) {
	s, repo, bus := setup()
	repo.codes["SD123"] = &ReferralCode{ID: 1, UserID: 100, Code: "SD123", Uses: 0}
	s.Attribute(context.Background(), 200, "SD123")
	if bus.CountOf("referral.attributed") != 1 {
		t.Fatalf("expected 1 event published, got %d", bus.CountOf("referral.attributed"))
	}
}
