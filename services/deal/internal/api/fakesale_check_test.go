package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/deal/internal/chart"
)

type mockLookup struct {
	id    int64
	found bool
	err   error
}

func (m *mockLookup) FindProductID(_ context.Context, _, _ string) (int64, bool, error) {
	return m.id, m.found, m.err
}

func TestFakeSaleCheck_Untracked(t *testing.T) {
	h := NewFakeSaleCheckHandler(&mockLookup{found: false}, &mockRepo{}, &mockDeal{})
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/tools/fake-sale-check", h.HandleFakeSaleCheck)

	body := []byte(`{"item_url":"https://shopee.vn/x-i.1.2"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/tools/fake-sale-check", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var got FakeSaleCheckResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
	require.False(t, got.Tracked)
	require.Equal(t, "shopee", got.Platform)
	require.Equal(t, "not_tracked", got.Message)
}

func TestFakeSaleCheck_TrackedVerdict(t *testing.T) {
	repo := &mockRepo{
		daily: []chart.DailyPoint{{
			Day: time.Now().AddDate(0, 0, -1), MinP: 90_000, MaxP: 120_000, CloseP: 100_000,
		}},
	}
	h := NewFakeSaleCheckHandler(&mockLookup{id: 42, found: true}, repo, &mockDeal{mat: "MATURE", verdict: "SALE_AO"})
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/tools/fake-sale-check", h.HandleFakeSaleCheck)

	body := []byte(`{"item_url":"https://shopee.vn/x-i.88123.20114455667"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/tools/fake-sale-check", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var got FakeSaleCheckResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
	require.True(t, got.Tracked)
	require.Equal(t, int64(42), got.ProductID)
	require.Equal(t, "SALE_AO", got.Verdict)
	require.Equal(t, int64(100_000), got.CurrentPrice)
}

func TestFakeSaleCheck_RejectsBadURL(t *testing.T) {
	h := NewFakeSaleCheckHandler(&mockLookup{}, &mockRepo{}, &mockDeal{})
	req := httptest.NewRequest(http.MethodPost, "/v1/tools/fake-sale-check", bytes.NewReader([]byte(`{"item_url":"https://example.com/x"}`)))
	rr := httptest.NewRecorder()
	h.HandleFakeSaleCheck(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}
