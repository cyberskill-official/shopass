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

	"shopass/services/price/internal/price"
)

func setupIngest(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set; skipping price ingest integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS price_snapshot (
			product_id BIGINT      NOT NULL,
			ts         TIMESTAMPTZ NOT NULL,
			price      BIGINT      NOT NULL CHECK (price > 0),
			list_price BIGINT      CHECK (list_price IS NULL OR list_price >= price),
			stock      INTEGER,
			sold       INTEGER,
			flash_sale BOOLEAN     NOT NULL DEFAULT false,
			PRIMARY KEY (product_id, ts)
		)`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `TRUNCATE price_snapshot`)
	require.NoError(t, err)

	mux := http.NewServeMux()
	NewIngestHandler(price.NewSnapshotRepo(pool)).RegisterRoutes(mux)
	return httptest.NewServer(mux), pool
}

func postSnap(t *testing.T, srv *httptest.Server, body string) *http.Response {
	resp, err := http.Post(srv.URL+"/v1/price/snapshots", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	return resp
}

func rowCount(t *testing.T, pool *pgxpool.Pool, pid int64) int {
	var n int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT count(*) FROM price_snapshot WHERE product_id=$1`, pid).Scan(&n))
	return n
}

func TestIngest_WritesSnapshot(t *testing.T) {
	srv, pool := setupIngest(t)
	defer srv.Close()
	defer pool.Close()

	resp := postSnap(t, srv, `{"product_id":100,"price":199000,"list_price":250000,"flash_sale":true}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var out ingestResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	require.True(t, out.Written)
	require.Equal(t, 1, rowCount(t, pool, 100))
}

func TestIngest_DeltaOnlySkipsUnchanged(t *testing.T) {
	srv, pool := setupIngest(t)
	defer srv.Close()
	defer pool.Close()

	require.Equal(t, http.StatusCreated,
		postSnap(t, srv, `{"product_id":101,"ts":"2026-07-01T00:00:00Z","price":50000}`).StatusCode)

	// same price at a later ts -> delta-only should skip
	r2 := postSnap(t, srv, `{"product_id":101,"ts":"2026-07-01T01:00:00Z","price":50000}`)
	require.Equal(t, http.StatusOK, r2.StatusCode)
	var out ingestResponse
	require.NoError(t, json.NewDecoder(r2.Body).Decode(&out))
	require.False(t, out.Written)
	require.Equal(t, 1, rowCount(t, pool, 101))

	// changed price -> writes
	require.Equal(t, http.StatusCreated,
		postSnap(t, srv, `{"product_id":101,"ts":"2026-07-01T02:00:00Z","price":45000}`).StatusCode)
	require.Equal(t, 2, rowCount(t, pool, 101))
}

func TestIngest_RejectsNonPositivePrice(t *testing.T) {
	srv, pool := setupIngest(t)
	defer srv.Close()
	defer pool.Close()
	require.Equal(t, http.StatusBadRequest, postSnap(t, srv, `{"product_id":102,"price":0}`).StatusCode)
}
