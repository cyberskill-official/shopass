package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"shopass/services/bill/internal/referral"
)

type stubReferralRepo struct {
	code string
	uses int
}

func (s *stubReferralRepo) FindByCode(ctx context.Context, code string) (referral.ReferralCode, bool, error) {
	if code == s.code {
		return referral.ReferralCode{ID: 1, UserID: 100, Code: s.code, Uses: s.uses}, true, nil
	}
	return referral.ReferralCode{}, false, nil
}
func (s *stubReferralRepo) HasReferrer(ctx context.Context, userID int64) (bool, error) {
	return false, nil
}
func (s *stubReferralRepo) SetReferrer(ctx context.Context, userID int64, codeID int64) error {
	return nil
}
func (s *stubReferralRepo) IncrementUses(ctx context.Context, codeID int64) error {
	s.uses++
	return nil
}
func (s *stubReferralRepo) CreateCodeForUser(ctx context.Context, userID int64) (string, error) {
	return "SDTEST1", nil
}

func TestReferralAttribute_SelfBlocked(t *testing.T) {
	repo := &stubReferralRepo{code: "SDABC12"}
	// Attribute via service directly for unit clarity
	svc := referral.NewService(repo, referralEventBus{})
	err := svc.Attribute(context.Background(), 100, "SDABC12")
	require.ErrorIs(t, err, referral.ErrSelfReferral)
}

func TestReferralMe_Unauthorized(t *testing.T) {
	// Handler needs PGRepo — unauthorized path only
	h := &ReferralHandler{}
	req := httptest.NewRequest(http.MethodGet, "/v1/referral/me", nil)
	rr := httptest.NewRecorder()
	h.HandleMe(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestReferralAttribute_BadJSON(t *testing.T) {
	h := &ReferralHandler{svc: referral.NewService(&stubReferralRepo{code: "X"}, referralEventBus{})}
	req := httptest.NewRequest(http.MethodPost, "/v1/referral/attribute", bytes.NewReader([]byte(`{`)))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", int64(2)))
	rr := httptest.NewRecorder()
	h.HandleAttribute(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestReferralAttribute_Unknown(t *testing.T) {
	h := &ReferralHandler{svc: referral.NewService(&stubReferralRepo{code: "SDOK"}, referralEventBus{})}
	body, _ := json.Marshal(map[string]string{"code": "SDNOPE"})
	req := httptest.NewRequest(http.MethodPost, "/v1/referral/attribute", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", int64(2)))
	rr := httptest.NewRecorder()
	h.HandleAttribute(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)
}
