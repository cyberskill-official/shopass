package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"shopass/services/bill/internal/gating"
)

type mockGate struct {
	ok  bool
	err error
}

func (m mockGate) AllowWithUsage(ctx context.Context, userID int64, featureKey string, usage *int64) (bool, error) {
	return m.ok, m.err
}

func TestGatingCheck_RequiresToken(t *testing.T) {
	h := NewGatingHandler(mockGate{ok: true}, "secret")
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/gating/check", bytes.NewBufferString(`{"user_id":1,"feature_key":"bottom_predict"}`))
	rr := httptest.NewRecorder()
	h.HandleCheck(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGatingCheck_AllowsWhenGateSaysYes(t *testing.T) {
	h := NewGatingHandler(mockGate{ok: true}, "secret")
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/gating/check", bytes.NewBufferString(`{"user_id":1,"feature_key":"bottom_predict"}`))
	req.Header.Set("X-Service-Token", "secret")
	rr := httptest.NewRecorder()
	h.HandleCheck(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	require.Equal(t, true, body["allowed"])
}

func TestGatingCheck_LimitReached(t *testing.T) {
	h := NewGatingHandler(mockGate{ok: false, err: gating.ErrLimitReached}, "secret")
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/gating/check", bytes.NewBufferString(`{"user_id":1,"feature_key":"wishlist_items","usage":20}`))
	req.Header.Set("X-Service-Token", "secret")
	rr := httptest.NewRecorder()
	h.HandleCheck(rr, req)
	var body map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	require.Equal(t, false, body["allowed"])
	require.Equal(t, true, body["limit_reached"])
}
