package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type logoutRepoStub struct {
	*mockRepo
	row           RefreshTokenRow
	revoked       bool
	revokeCalls   int
	revokedFamily string
}

func newLogoutRepoStub() *logoutRepoStub {
	return &logoutRepoStub{
		mockRepo: newMockRepo(),
		row: RefreshTokenRow{
			ID:        1,
			UserID:    90112,
			FamilyID:  "family-logout",
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
}

func (r *logoutRepoStub) FindRefreshByHash(_ context.Context, _ string) (RefreshTokenRow, error) {
	row := r.row
	if r.revoked {
		now := time.Now()
		row.RevokedAt = &now
	}
	return row, nil
}

func (r *logoutRepoStub) RevokeFamily(_ context.Context, familyID string) error {
	r.revokeCalls++
	r.revokedFamily = familyID
	r.revoked = true
	return nil
}

func (r *logoutRepoStub) RotateRefreshToken(_ context.Context, _, _ string, _ time.Time) (RefreshRotationStatus, error) {
	if r.revoked {
		return RefreshRotationInvalid, nil
	}
	return RefreshRotationSucceeded, nil
}

func TestLogoutRevokesRefreshFamily(t *testing.T) {
	repo := newLogoutRepoStub()
	ts := newServiceForRefreshStub(t, repo)
	pair, err := ts.IssueTokenPair(context.Background(), 90112)
	require.NoError(t, err)

	require.NoError(t, ts.Logout(context.Background(), pair.Refresh))
	require.Equal(t, 1, repo.revokeCalls)
	require.Equal(t, repo.row.FamilyID, repo.revokedFamily)
	_, err = ts.Refresh(context.Background(), pair.Refresh)
	require.ErrorIs(t, err, ErrInvalidRefresh)
}
