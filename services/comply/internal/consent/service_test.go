package consent

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	records []ConsentRecord
	version map[string]int32
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		records: make([]ConsentRecord, 0),
		version: map[string]int32{
			"cart_read":              1,
			"price_tracking":         1,
			"marketing_notification": 1,
			"analytics_b2b":          1,
		},
	}
}

func (m *mockRepo) latest(ctx context.Context, userID int64, purposeKey string) (ConsentRecord, error) {
	for i := len(m.records) - 1; i >= 0; i-- {
		r := m.records[i]
		if r.UserID == userID && r.PurposeKey == purposeKey {
			return r, nil
		}
	}
	return ConsentRecord{}, sql.ErrNoRows
}

func (m *mockRepo) append(ctx context.Context, rec ConsentRecord) error {
	if rec.TS.IsZero() {
		rec.TS = time.Now()
	}
	m.records = append(m.records, rec)
	return nil
}

func (m *mockRepo) effectiveVersion(ctx context.Context, purposeKey string) (int32, error) {
	if v, ok := m.version[purposeKey]; ok {
		return v, nil
	}
	return 0, sql.ErrNoRows
}

func (m *mockRepo) history(ctx context.Context, userID int64, purposeKey string) ([]ConsentRecord, error) {
	var res []ConsentRecord
	for _, r := range m.records {
		if r.UserID == userID && r.PurposeKey == purposeKey {
			res = append(res, r)
		}
	}
	return res, nil
}

func setupWithUser(t *testing.T) (*Service, int64) {
	return NewService(newMockRepo()), 123
}

func TestConsent_SilenceIsNotConsent(t *testing.T) {
	ctx := context.Background()
	s, uid := setupWithUser(t)
	ok, err := s.IsAllowed(ctx, uid, PurposeMarketing)
	require.NoError(t, err)
	require.False(t, ok) // chua co ban ghi -> false, khong mac dinh dong thuan
}

func TestConsent_GrantThenAllowed(t *testing.T) {
	ctx := context.Background()
	s, uid := setupWithUser(t)
	require.NoError(t, s.Grant(ctx, uid, PurposeCartRead, "extension", ReqMeta{}))
	ok, err := s.IsAllowed(ctx, uid, PurposeCartRead)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestConsent_WithdrawAppendsRow_KeepsHistory(t *testing.T) {
	ctx := context.Background()
	s, uid := setupWithUser(t)
	s.Grant(ctx, uid, PurposeCartRead, "web", ReqMeta{})
	require.NoError(t, s.Withdraw(ctx, uid, PurposeCartRead, "web", ReqMeta{}))

	ok, _ := s.IsAllowed(ctx, uid, PurposeCartRead)
	require.False(t, ok) // dong moi nhat granted=false

	h, _ := s.History(ctx, uid, PurposeCartRead)
	require.Len(t, h, 2) // grant + withdraw, khong xoa lich su
}

func TestConsent_UnknownPurpose_Rejected(t *testing.T) {
	ctx := context.Background()
	s, uid := setupWithUser(t)
	err := s.Grant(ctx, uid, Purpose("bad_purpose"), "web", ReqMeta{})
	require.ErrorIs(t, err, ErrUnknownPurpose)
}

func TestConsent_OldConsentKeepsOldVersion(t *testing.T) {
	ctx := context.Background()
	s, uid := setupWithUser(t)
	s.Grant(ctx, uid, PurposeCartRead, "web", ReqMeta{}) // version 1

	// simulate policy update
	mock := s.repo.(*mockRepo)
	mock.version["cart_read"] = 2

	s.Grant(ctx, uid, PurposeCartRead, "web", ReqMeta{}) // version 2

	h, _ := s.History(ctx, uid, PurposeCartRead)
	require.Len(t, h, 2)
	require.Equal(t, int32(1), h[0].PolicyVersion)
	require.Equal(t, int32(2), h[1].PolicyVersion) // khong tu nang cap dong cu
}
