package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRefresh_OneTimeUse(t *testing.T) {
	s := newServiceWithKeys(t)
	ctx := context.Background()
	
	pair, _ := s.IssueTokenPair(ctx, 1)
	
	p2, err := s.Refresh(ctx, pair.Refresh)
	require.NoError(t, err)
	require.NotEqual(t, pair.Refresh, p2.Refresh) // rotation cấp mới
	
	_, err2 := s.Refresh(ctx, pair.Refresh)       // dùng lại token cũ
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
	
	_, _ = s.Refresh(ctx, pair.Refresh)     // tái sử dụng → thu hồi family
	
	_, err := s.Refresh(ctx, p2.Refresh)     // token mới cùng family cũng bị chặn
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
