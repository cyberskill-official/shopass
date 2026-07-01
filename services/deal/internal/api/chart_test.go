package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/deal/internal/chart"
)

type mockRepo struct {
	exists bool
	daily  []chart.DailyPoint
}

func (m *mockRepo) ProductExists(ctx context.Context, productID int64) (bool, error) {
	if productID == 999999 {
		return false, nil
	}
	return m.exists, nil
}

func (m *mockRepo) QueryDaily(ctx context.Context, productID int64, from time.Time) ([]chart.DailyPoint, error) {
	var res []chart.DailyPoint
	for _, p := range m.daily {
		if !p.Day.Before(from) {
			res = append(res, p)
		}
	}
	return res, nil
}

type mockDeal struct {
	mat     string
	verdict string
}

func (m *mockDeal) Maturity(ctx context.Context, productID int64) string {
	if m.mat == "" {
		return "MATURE"
	}
	return m.mat
}

func (m *mockDeal) Verdict(ctx context.Context, productID int64) string {
	return m.verdict
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func doGET(t *testing.T, h *Handler, url string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	// Since we are using Go 1.22 ServeMux for PathValue, we need to pass it through a mux
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/products/{id}/chart", h.HandleChart)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	err := json.NewDecoder(rec.Body).Decode(v)
	require.NoError(t, err)
}

func setupWithHistory(t *testing.T, days int) (*Handler, int64) {
	repo := &mockRepo{exists: true}
	for i := days; i >= 0; i-- {
		repo.daily = append(repo.daily, chart.DailyPoint{
			Day:    time.Now().AddDate(0, 0, -i),
			MinP:   100000,
			MaxP:   100000,
			CloseP: 100000,
		})
	}
	return NewHandler(repo, &mockDeal{mat: "MATURE"}), 1
}

func setupWithRange(t *testing.T, fromStr, toStr string) (*Handler, int64) {
	repo := &mockRepo{exists: true}
	// just to make sure the time covers the mock test
	return NewHandler(repo, &mockDeal{mat: "MATURE"}), 1
}

func setupWithMaturity(t *testing.T, mat string, days int) (*Handler, int64) {
	repo := &mockRepo{exists: true}
	for i := days; i >= 0; i-- {
		repo.daily = append(repo.daily, chart.DailyPoint{
			Day:    time.Now().AddDate(0, 0, -i),
			MinP:   100000,
			MaxP:   100000,
			CloseP: 100000,
		})
	}
	return NewHandler(repo, &mockDeal{mat: mat}), 1
}

func seedDailyCloses(t *testing.T, h *Handler, pid int64, closes []int64) {
	repo := h.repo.(*mockRepo)
	repo.daily = []chart.DailyPoint{}
	for i, c := range closes {
		repo.daily = append(repo.daily, chart.DailyPoint{
			Day:    time.Now().AddDate(0, 0, -len(closes)+i),
			MinP:   c, // Simplified min=close
			MaxP:   c,
			CloseP: c,
		})
	}
}

func TestChart_Default90d(t *testing.T) {
    h, pid := setupWithHistory(t, 120) // 120 ngày dữ liệu
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/chart") // không range
    require.Equal(t, 200, rec.Code)
    var body ChartResponse
    decode(t, rec, &body)
    require.Equal(t, "90d", body.Range)
    require.LessOrEqual(t, len(body.Daily), 91) // ~90 ngày, không phải 120
}

func TestChart_Annotations_MedianTrailingMin(t *testing.T) {
    h, pid := setupWithHistory(t, 90)
    seedDailyCloses(t, h, pid, []int64{120_000, 100_000, 80_000, 100_000}) // đáy 80k
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/chart?range=90d")
    var body ChartResponse
    decode(t, rec, &body)
    require.Equal(t, int64(100_000), body.Annotations.Median90)    // trung vị
    require.Equal(t, int64(80_000), body.Annotations.TrailingMin)  // đáy
}

func TestChart_DoubleDateMarkers(t *testing.T) {
	setupWithHistory(t, 365)
	
	// mock current time as 2026-06-27 so we can hit double dates consistently
	// For actual implementation, doubleDates works on `time.Now().Add(-window)`
	// Wait, we can't easily mock time.Now() globally unless we inject it, 
	// but doubleDates logic just loops from start year to end year.
	// We'll test `chart.Build` directly for this one to use fixed time.

	daily := []chart.DailyPoint{}
	from, _ := time.Parse("2006-01-02", "2026-04-01")
	to, _ := time.Parse("2006-01-02", "2026-06-27")

	ann := chart.Build(daily, from, to)
    require.Contains(t, ann.DoubleDates, "2026-04-04")
    require.Contains(t, ann.DoubleDates, "2026-05-05")
    require.NotContains(t, ann.DoubleDates, "2026-03-03") // ngoài khoảng
}

func TestChart_MaturityFlag_Warming(t *testing.T) {
    h, pid := setupWithMaturity(t, "WARMING", 40) // 40 ngày: đủ vẽ, chưa đủ 90
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/chart?range=90d")
    var body ChartResponse
    decode(t, rec, &body)
    require.Equal(t, "WARMING", body.Maturity)
    require.True(t, body.Annotations.Accumulating) // cờ đang tích lũy
}

func TestChart_UnknownProduct_404(t *testing.T) {
    h, _ := setupWithHistory(t, 30)
    rec := doGET(t, h, "/v1/products/999999/chart?range=30d")
    require.Equal(t, 404, rec.Code)
}

func TestChart_InvalidRange_400(t *testing.T) {
    h, pid := setupWithHistory(t, 30)
    rec := doGET(t, h, "/v1/products/"+itoa(pid)+"/chart?range=5d")
    require.Equal(t, 400, rec.Code)
}
