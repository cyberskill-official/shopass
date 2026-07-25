package dsar

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockUsers struct{ view AccountView }

func (m *mockUsers) View(ctx context.Context, userID int64) (AccountView, error) {
	if m.view.UserID == userID {
		return m.view, nil
	}
	return AccountView{}, nil
}
func (m *mockUsers) Anonymize(ctx context.Context, userID int64) error {
	if m.view.UserID == userID {
		m.view.Email = "anon@example.com"
	}
	return nil
}

type mockTrack struct {
	prods    map[int64][]ProductView
	delCount int
}

func (m *mockTrack) ByUser(ctx context.Context, userID int64) ([]ProductView, error) {
	return m.prods[userID], nil
}
func (m *mockTrack) HardDeleteByUser(ctx context.Context, userID int64) (int, error) {
	c := len(m.prods[userID])
	m.delCount += c
	m.prods[userID] = nil
	return c, nil
}

type mockBill struct{ anonCount int }

func (m *mockBill) AnonymizeByUser(ctx context.Context, userID int64) (int, error) {
	m.anonCount++
	return 1, nil
}

type mockConsent struct {
	c []ConsentView
}

func (m *mockConsent) HistoryAll(ctx context.Context, userID int64) ([]ConsentView, error) {
	return m.c, nil
}

type mockRepo struct {
	dsars []DSARRequest
}

func (m *mockRepo) create(ctx context.Context, userID int64, kind string, slaDue time.Time) (int64, error) {
	id := int64(len(m.dsars) + 1)
	m.dsars = append(m.dsars, DSARRequest{
		ID: id, UserID: userID, Kind: kind, SLADueAt: slaDue, Status: "open",
	})
	return id, nil
}
func (m *mockRepo) markCompleted(ctx context.Context, dsarID int64) error {
	for i, d := range m.dsars {
		if d.ID == dsarID {
			m.dsars[i].Status = "completed"
		}
	}
	return nil
}
func (m *mockRepo) overdue(ctx context.Context) ([]DSARRequest, error) {
	var r []DSARRequest
	for _, d := range m.dsars {
		if d.Status != "completed" && time.Now().After(d.SLADueAt) {
			r = append(r, d)
		}
	}
	return r, nil
}

func setupMocks() (*Service, *mockUsers, *mockTrack, *mockBill, *mockConsent) {
	mu := &mockUsers{}
	mt := &mockTrack{prods: make(map[int64][]ProductView)}
	mb := &mockBill{}
	mc := &mockConsent{}
	mr := &mockRepo{}
	return NewService(mr, mu, mt, mb, mc), mu, mt, mb, mc
}

func TestExport_OnlyOwnData(t *testing.T) {
	s, mu, mt, _, mc := setupMocks()

	// userA (ID: 1)
	mu.view = AccountView{UserID: 1, Email: "a@a.com"}
	mt.prods[1] = []ProductView{{ID: 10, Platform: "shopee", Name: "A"}}
	mt.prods[2] = []ProductView{{ID: 20, Platform: "shopee", Name: "B"}} // userB data
	mc.c = []ConsentView{{UserID: 1, PurposeKey: "cart_read"}}

	bundle, err := s.Export(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), bundle.Account.UserID)

	for _, p := range bundle.TrackedProducts {
		require.NotEqual(t, int64(20), p.ID)
	}
}

func TestExport_PortabilityIsJSON(t *testing.T) {
	s, mu, _, _, mc := setupMocks()
	mu.view = AccountView{UserID: 1, Email: "a@a.com"}
	mc.c = []ConsentView{{UserID: 1, PurposeKey: "cart_read"}}

	bundle, _ := s.Export(context.Background(), 1)
	raw, err := json.Marshal(bundle)
	require.NoError(t, err)
	require.Contains(t, string(raw), "consent_history")
}
