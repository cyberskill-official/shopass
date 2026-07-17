package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// refreshRotationStub embeds the general repository test double and overrides
// only the refresh boundary. It lets the service-level atomic-rotation contract
// run without requiring a local PostgreSQL instance.
type refreshRotationStub struct {
	*mockRepo
	row            RefreshTokenRow
	rotationStatus RefreshRotationStatus
	rotationErr    error
	rotationCalls  int
	markUsedCalls  int
	gotOldHash     string
	gotNewHash     string
	gotReplacement time.Time
}

func newRefreshRotationStub() *refreshRotationStub {
	return &refreshRotationStub{
		mockRepo: newMockRepo(),
		row: RefreshTokenRow{
			ID:        1,
			UserID:    90112,
			FamilyID:  "family-1",
			ExpiresAt: time.Now().Add(time.Hour),
		},
		rotationStatus: RefreshRotationSucceeded,
	}
}

func (r *refreshRotationStub) FindRefreshByHash(_ context.Context, _ string) (RefreshTokenRow, error) {
	return r.row, nil
}

func (r *refreshRotationStub) MarkUsed(_ context.Context, _ int64) error {
	r.markUsedCalls++
	return errors.New("MarkUsed must not be called by Refresh")
}

func (r *refreshRotationStub) RotateRefreshToken(_ context.Context, oldHash, replacementHash string, replacementExpiresAt time.Time) (RefreshRotationStatus, error) {
	r.rotationCalls++
	r.gotOldHash = oldHash
	r.gotNewHash = replacementHash
	r.gotReplacement = replacementExpiresAt
	return r.rotationStatus, r.rotationErr
}

func newServiceForRefreshStub(t *testing.T, repo Repo) *TokenService {
	t.Helper()
	s := NewTokenService(repo, "shopass-auth", "shopass-gateway", 15*time.Minute)
	s.AddSigningKey("key-1", genKey(t))
	return s
}

func TestRefresh_DelegatesToAtomicRotationInsteadOfMarkUsed(t *testing.T) {
	repo := newRefreshRotationStub()
	s := newServiceForRefreshStub(t, repo)

	pair, err := s.Refresh(context.Background(), "raw-refresh-token")
	require.NoError(t, err)
	require.NotEmpty(t, pair.Access)
	require.NotEmpty(t, pair.Refresh)
	require.Equal(t, 1, repo.rotationCalls)
	require.Zero(t, repo.markUsedCalls)
	require.Equal(t, hashStr("raw-refresh-token"), repo.gotOldHash)
	require.NotEmpty(t, repo.gotNewHash)
	require.NotEqual(t, repo.gotOldHash, repo.gotNewHash)
	require.True(t, repo.gotReplacement.After(time.Now()))
}

func TestRefresh_PropagatesAtomicRotationFailure(t *testing.T) {
	repo := newRefreshRotationStub()
	want := errors.New("transaction failed")
	repo.rotationErr = want
	s := newServiceForRefreshStub(t, repo)

	_, err := s.Refresh(context.Background(), "raw-refresh-token")
	require.ErrorIs(t, err, want)
	require.Equal(t, 1, repo.rotationCalls)
	require.Zero(t, repo.markUsedCalls)
}

func TestRefresh_AtomicReuseOutcomePreservesFamilyTheftSignal(t *testing.T) {
	repo := newRefreshRotationStub()
	repo.rotationStatus = RefreshRotationReuseDetected
	s := newServiceForRefreshStub(t, repo)

	_, err := s.Refresh(context.Background(), "raw-refresh-token")
	require.ErrorIs(t, err, ErrRefreshReuseDetected)
	require.Equal(t, 1, repo.rotationCalls)
	require.Zero(t, repo.markUsedCalls)
}

func TestRefresh_OneTimeUse(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()

	pair, _ := s.IssueTokenPair(ctx, 1)

	p2, err := s.Refresh(ctx, pair.Refresh)
	require.NoError(t, err)
	require.NotEqual(t, pair.Refresh, p2.Refresh) // rotation cấp mới

	_, err2 := s.Refresh(ctx, pair.Refresh) // dùng lại token cũ
	require.ErrorIs(t, err2, ErrRefreshReuseDetected)
}

func TestRefresh_StoredAsHash(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()

	pair, _ := s.IssueTokenPair(ctx, 1)

	row, err := s.repo.(*pgRepo).FindRefreshByHash(ctx, hashStr(pair.Refresh))
	require.NoError(t, err)
	require.NotEqual(t, pair.Refresh, row.TokenHash) // lưu hash, không cleartext
}

func TestRefresh_ReuseRevokesFamily(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()

	pair, _ := s.IssueTokenPair(ctx, 1)
	p2, _ := s.Refresh(ctx, pair.Refresh)

	_, err := s.Refresh(ctx, pair.Refresh) // tái sử dụng -> thu hồi family
	require.ErrorIs(t, err, ErrRefreshReuseDetected)

	_, err = s.Refresh(ctx, p2.Refresh) // token mới cùng family cũng bị chặn
	require.ErrorIs(t, err, ErrInvalidRefresh)
}

func TestRefresh_ExpiredToken(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()

	// Create an expired token manually
	raw, _ := generateRandomToken()
	hash := hashStr(raw)
	err := s.repo.InsertRefreshToken(ctx, 1, hash, "fam1", time.Now().Add(-time.Hour))
	require.NoError(t, err)

	_, err = s.Refresh(ctx, raw)
	require.ErrorIs(t, err, ErrInvalidRefresh)
}

func TestRefresh_RevokedToken(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()

	pair, _ := s.IssueTokenPair(ctx, 1)

	// Revoke the family
	row, _ := s.repo.(*pgRepo).FindRefreshByHash(ctx, hashStr(pair.Refresh))
	err := s.repo.RevokeFamily(ctx, row.FamilyID)
	require.NoError(t, err)

	_, err = s.Refresh(ctx, pair.Refresh)
	require.ErrorIs(t, err, ErrInvalidRefresh)
}

func TestRefresh_ConcurrentUseHasOneWinnerAndRevokesFamily(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()
	pair, err := s.IssueTokenPair(ctx, 1)
	require.NoError(t, err)

	type result struct {
		pair TokenPair
		err  error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for range 2 {
		go func() {
			<-start
			rotated, refreshErr := s.Refresh(ctx, pair.Refresh)
			results <- result{pair: rotated, err: refreshErr}
		}()
	}
	close(start)

	var winner TokenPair
	successes := 0
	reuses := 0
	for range 2 {
		got := <-results
		switch {
		case got.err == nil:
			successes++
			winner = got.pair
		case errors.Is(got.err, ErrRefreshReuseDetected):
			reuses++
		default:
			t.Fatalf("unexpected concurrent refresh error: %v", got.err)
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, reuses)

	// The second use is a theft signal, so it must revoke the replacement that
	// was issued to the first caller as well.
	_, err = s.Refresh(ctx, winner.Refresh)
	require.ErrorIs(t, err, ErrInvalidRefresh)
}
