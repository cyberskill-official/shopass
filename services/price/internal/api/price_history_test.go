package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"shopass/services/price/internal/price"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	dsn := "postgres://postgres:postgres@localhost:5432/shopass_test?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skip("Postgres not available")
	}

	if err := pool.Ping(ctx); err != nil {
		t.Skip("Postgres ping failed")
	}

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tracked_product (
			id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS price_daily (
			product_id BIGINT,
			day TIMESTAMPTZ,
			min_p BIGINT,
			max_p BIGINT,
			close_p BIGINT
		);
		CREATE TABLE IF NOT EXISTS price_snapshot (
			product_id BIGINT,
			ts TIMESTAMPTZ,
			price BIGINT,
			flash_sale BOOLEAN
		);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Exec(ctx, `DROP TABLE IF EXISTS price_daily CASCADE; DROP TABLE IF EXISTS price_snapshot CASCADE; DROP TABLE IF EXISTS tracked_product CASCADE;`)
		pool.Close()
	})

	return pool
}

func setupWithHistory(t *testing.T, days int) (*Handler, int64) {
	pool := setupTestDB(t)
	repo := price.NewRepo(pool)
	h := NewHandler(repo)

	var pid int64
	err := pool.QueryRow(context.Background(), `INSERT INTO tracked_product (title) VALUES ('test') RETURNING id`).Scan(&pid)
	require.NoError(t, err)

	now := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < days; i++ {
		day := now.Add(-time.Duration(i) * 24 * time.Hour)
		_, err := pool.Exec(context.Background(), `INSERT INTO price_daily (product_id, day, min_p, max_p, close_p) VALUES ($1, $2, 100000, 150000, 120000)`, pid, day)
		require.NoError(t, err)
	}

	return h, pid
}

func insertRaw(t *testing.T, h *Handler, pid int64, ts time.Time, p int64, fs bool) {
	_, err := h.repo.UnexportedPoolForTest().Exec(context.Background(), `INSERT INTO price_snapshot (product_id, ts, price, flash_sale) VALUES ($1, $2, $3, $4)`, pid, ts, p, fs)
	require.NoError(t, err)
}

func todayAt(hour, min int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func doGET(t *testing.T, h *Handler, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", path, nil)
	require.NoError(t, err)
	
	// Create a new ServeMux with 1.22 pattern
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func decode(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	err := json.NewDecoder(rr.Body).Decode(v)
	require.NoError(t, err)
}

func TestPriceHistory_Default90d(t *testing.T) {
	h, pid := setupWithHistory(t, 120) // 120 ngày dữ liệu
	rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/price-history") // không range
	require.Equal(t, 200, rec.Code)
	var body HistoryResponse
	decode(t, rec, &body)
	require.Equal(t, "90d", body.Range)
	require.LessOrEqual(t, len(body.Daily), 91) // ~90 ngày, không phải 120
}

func TestPriceHistory_BadRange_400(t *testing.T) {
	h, pid := setupWithHistory(t, 30)
	rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/price-history?range=5d")
	require.Equal(t, 400, rec.Code)
	require.Contains(t, rec.Body.String(), "invalid range")
}

func TestPriceHistory_UnknownProduct_404(t *testing.T) {
	h, _ := setupWithHistory(t, 30)
	rec := doGET(t, h, "/v1/products/999999/price-history?range=30d")
	require.Equal(t, 404, rec.Code) // DEC-PRICE-34
}

func TestPriceHistory_StitchesRawTail(t *testing.T) {
	h, pid := setupWithHistory(t, 30)
	// ghi một snapshot raw lúc trưa nay, sau bucket cagg gần nhất
	insertRaw(t, h, pid, todayAt(12, 0), 79_000, true)
	rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/price-history?range=30d")
	var body HistoryResponse
	decode(t, rec, &body)
	require.NotEmpty(t, body.Tail) // đuôi raw có mặt, không chờ cagg refresh
	require.Equal(t, int64(79_000), body.Tail[len(body.Tail)-1].Price)
	require.True(t, body.Tail[len(body.Tail)-1].FlashSale)
}

func TestPriceHistory_PayloadShape(t *testing.T) {
	h, pid := setupWithHistory(t, 90)
	rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/price-history?range=90d")
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw))
	for _, k := range []string{"product_id", "range", "daily", "tail"} {
		_, ok := raw[k]
		require.True(t, ok, "thiếu khóa "+k)
	}
	require.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
}
