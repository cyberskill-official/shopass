package api

import (
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

// setupCompare builds the minimal schema the compare query joins (platform,
// tracked_product, price_snapshot) and returns a live server + pool.
func setupCompare(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set; skipping compare integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	// Fresh tables (drop dependents first for FK order).
	for _, ddl := range []string{
		`DROP TABLE IF EXISTS price_snapshot`,
		`DROP TABLE IF EXISTS tracked_product`,
		`DROP TABLE IF EXISTS platform`,
		`CREATE TABLE platform (
			id       SMALLINT PRIMARY KEY,
			code     TEXT UNIQUE NOT NULL CHECK (code IN ('shopee','tiktok','lazada')),
			country  TEXT NOT NULL DEFAULT 'VN',
			base_url TEXT
		)`,
		`CREATE TABLE tracked_product (
			id               BIGSERIAL PRIMARY KEY,
			platform_id      SMALLINT NOT NULL REFERENCES platform(id),
			platform_item_id TEXT NOT NULL,
			canonical_key    TEXT,
			UNIQUE (platform_id, platform_item_id)
		)`,
		// No FK on product_id: keep this table schema-compatible with the ingest
		// test's price_snapshot (which inserts product_ids without a tracked_product),
		// since both suites share one database. The compare query only joins.
		`CREATE TABLE price_snapshot (
			product_id BIGINT NOT NULL,
			ts         TIMESTAMPTZ NOT NULL,
			price      BIGINT NOT NULL CHECK (price > 0),
			list_price BIGINT,
			stock      INTEGER,
			sold       INTEGER,
			flash_sale BOOLEAN NOT NULL DEFAULT false,
			PRIMARY KEY (product_id, ts)
		)`,
		`INSERT INTO platform (id, code) VALUES (1,'shopee'),(2,'tiktok'),(3,'lazada')`,
	} {
		_, err := pool.Exec(ctx, ddl)
		require.NoError(t, err, ddl)
	}

	mux := http.NewServeMux()
	NewHandler(price.NewRepo(pool)).RegisterRoutes(mux)
	return httptest.NewServer(mux), pool
}

// seedListing inserts a tracked_product on a platform for a canonical_key and
// one or more (ts, price) snapshots; the newest ts is the "current" price.
func seedListing(t *testing.T, pool *pgxpool.Pool, platformID int16, itemID, key string, snaps [][2]any) int64 {
	ctx := context.Background()
	var pid int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO tracked_product (platform_id, platform_item_id, canonical_key)
		 VALUES ($1,$2,$3) RETURNING id`, platformID, itemID, key).Scan(&pid))
	for _, s := range snaps {
		_, err := pool.Exec(ctx,
			`INSERT INTO price_snapshot (product_id, ts, price) VALUES ($1,$2,$3)`,
			pid, s[0], s[1])
		require.NoError(t, err)
	}
	return pid
}

func getCompare(t *testing.T, srv *httptest.Server, key string) *http.Response {
	resp, err := http.Get(srv.URL + "/v1/compare?canonical_key=" + key)
	require.NoError(t, err)
	return resp
}

func TestCompare_MissingKeyIs400(t *testing.T) {
	srv, pool := setupCompare(t)
	defer srv.Close()
	defer pool.Close()
	resp, err := http.Get(srv.URL + "/v1/compare")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCompare_UnknownKeyIs404(t *testing.T) {
	srv, pool := setupCompare(t)
	defer srv.Close()
	defer pool.Close()
	require.Equal(t, http.StatusNotFound, getCompare(t, srv, "nope-nothing").StatusCode)
}

func TestCompare_MarksCheapestAcrossPlatforms(t *testing.T) {
	srv, pool := setupCompare(t)
	defer srv.Close()
	defer pool.Close()

	key := "phone-x"
	// Shopee: two snapshots, latest (later ts) is 199000 - proves latest-per-product.
	seedListing(t, pool, 1, "s-1", key, [][2]any{
		{"2026-07-01T00:00:00Z", 250000},
		{"2026-07-02T00:00:00Z", 199000},
	})
	// Lazada: current 210000 (dearer).
	seedListing(t, pool, 3, "l-1", key, [][2]any{{"2026-07-02T00:00:00Z", 210000}})

	resp := getCompare(t, srv, key)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out CompareResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	require.Equal(t, key, out.CanonicalKey)
	require.Len(t, out.Items, 2)

	// Ordered cheapest first: Shopee 199000 (latest), then Lazada 210000.
	require.Equal(t, "shopee", out.Items[0].PlatformCode)
	require.Equal(t, "Shopee", out.Items[0].PlatformName)
	require.EqualValues(t, 199000, out.Items[0].Price)
	require.True(t, out.Items[0].IsCheapest, "cheapest row must be flagged")
	require.Equal(t, "VND", out.Items[0].Currency)
	require.NotEmpty(t, out.Items[0].TS, "per-platform freshness ts must be present")

	require.Equal(t, "lazada", out.Items[1].PlatformCode)
	require.Equal(t, "Lazada", out.Items[1].PlatformName)
	require.EqualValues(t, 210000, out.Items[1].Price)
	require.False(t, out.Items[1].IsCheapest)
}

func TestCompare_SinglePlatformIsCheapest(t *testing.T) {
	srv, pool := setupCompare(t)
	defer srv.Close()
	defer pool.Close()

	key := "solo-item"
	seedListing(t, pool, 3, "l-solo", key, [][2]any{{"2026-07-02T00:00:00Z", 88000}})

	resp := getCompare(t, srv, key)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out CompareResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	require.Len(t, out.Items, 1) // DEC-PRICE-44: do not force all 3 platforms
	require.True(t, out.Items[0].IsCheapest, "the only listing is cheapest by definition")
}

func TestCompare_TiesAllMarkedCheapest(t *testing.T) {
	srv, pool := setupCompare(t)
	defer srv.Close()
	defer pool.Close()

	key := "tie-item"
	seedListing(t, pool, 1, "s-tie", key, [][2]any{{"2026-07-02T00:00:00Z", 120000}})
	seedListing(t, pool, 3, "l-tie", key, [][2]any{{"2026-07-02T00:00:00Z", 120000}})

	resp := getCompare(t, srv, key)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out CompareResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	require.Len(t, out.Items, 2)
	require.True(t, out.Items[0].IsCheapest && out.Items[1].IsCheapest, "equal-min rows all flagged (§1 #6)")
}
