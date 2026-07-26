package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestWaitlist_PersistsLead(t *testing.T) {
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS marketing_lead (
		  id BIGSERIAL PRIMARY KEY,
		  email TEXT NOT NULL,
		  zalo TEXT,
		  source TEXT NOT NULL DEFAULT 'pricing',
		  tier_interest TEXT NOT NULL DEFAULT 'premium_basic',
		  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		  UNIQUE (email, source)
		);
		TRUNCATE marketing_lead;
	`)
	require.NoError(t, err)

	h := NewWaitlistHandler(pool)
	body := []byte(`{"email":"Buyer@Example.com","zalo":"0901234567","source":"pricing","tier_interest":"premium_pro"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/leads/waitlist", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.HandleWaitlist(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	var got struct {
		OK bool  `json:"ok"`
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
	require.True(t, got.OK)
	require.Greater(t, got.ID, int64(0))

	var email, tier string
	err = pool.QueryRow(ctx, `SELECT email, tier_interest FROM marketing_lead WHERE id=$1`, got.ID).Scan(&email, &tier)
	require.NoError(t, err)
	require.Equal(t, "buyer@example.com", email)
	require.Equal(t, "premium_pro", tier)
}

func TestWaitlist_RejectsBadEmail(t *testing.T) {
	h := NewWaitlistHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/leads/waitlist", bytes.NewReader([]byte(`{"email":"nope"}`)))
	rr := httptest.NewRecorder()
	h.HandleWaitlist(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}
