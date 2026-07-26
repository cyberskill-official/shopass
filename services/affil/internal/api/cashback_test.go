package api

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"shopass/services/affil/internal/auth"
	"shopass/services/affil/internal/cashback"
)

func TestCashbackSummary_Unauthorized(t *testing.T) {
	h := NewCashbackHandler(&cashback.Ledger{Cfg: cashback.DefaultConfig(), Store: cashback.NewMemStore()})
	req := httptest.NewRequest("GET", "/v1/cashback/summary", nil)
	rec := httptest.NewRecorder()
	h.HandleSummary(rec, req)
	require.Equal(t, 401, rec.Code)
}

func TestCashbackSummary_Disclosure(t *testing.T) {
	st := cashback.NewMemStore()
	l := &cashback.Ledger{Cfg: cashback.DefaultConfig(), Store: st}
	t0 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	_, err := l.OnConfirmed(t.Context(), cashback.Conversion{
		ID: 1, UserID: 42, Commission: 100_000, UserTier: cashback.TierPremium, ConfirmedAt: t0,
	})
	require.NoError(t, err)

	h := NewCashbackHandler(l)
	req := httptest.NewRequest("GET", "/v1/cashback/summary", nil)
	req = req.WithContext(auth.WithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()
	h.HandleSummary(rec, req)
	require.Equal(t, 200, rec.Code)

	var resp cashback.SummaryResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.Pending.Count)
	require.Equal(t, int64(50_000), resp.Pending.AmountVND)
	require.NotNil(t, resp.Pending.NextAvailableAt)
	require.Equal(t, cashback.DisclosureNote, resp.Note)
}
