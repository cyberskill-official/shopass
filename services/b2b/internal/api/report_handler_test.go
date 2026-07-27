package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/b2b/internal/report"
	"shopass/services/b2b/internal/trend"
)

type memSubs map[int64]report.Subscription

func (m memSubs) GetByID(_ context.Context, id int64) (report.Subscription, error) {
	s, ok := m[id]
	if !ok {
		return report.Subscription{}, errors.New("not found")
	}
	return s, nil
}

type memTrend struct {
	cells []trend.MarketTrendCell
}

func (m memTrend) QueryCells(_ context.Context, categoryID int64, platformID int16, from, to time.Time) ([]trend.MarketTrendCell, error) {
	var out []trend.MarketTrendCell
	for _, c := range m.cells {
		if c.Suppressed || c.CategoryID != categoryID || c.PlatformID != platformID {
			continue
		}
		d := time.Date(c.Day.Year(), c.Day.Month(), c.Day.Day(), 0, 0, 0, 0, time.UTC)
		if (d.Equal(from) || d.After(from)) && d.Before(to) {
			out = append(out, c)
		}
	}
	return out, nil
}

func setupHandler(t *testing.T, sub report.Subscription, cells []trend.MarketTrendCell) *http.ServeMux {
	t.Helper()
	h := &ReportHandler{
		Subs: memSubs{1: sub},
		Builder: &report.Builder{Trend: memTrend{cells: cells}, Now: func() time.Time {
			return time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
		}},
		Metrics: report.NewMetrics(),
		Now:     func() time.Time { return time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC) },
	}
	mux := http.NewServeMux()
	RegisterRoutes(mux, h)
	return mux
}

func TestReport_InactiveSubscription_402(t *testing.T) {
	mux := setupHandler(t, report.Subscription{
		ID: 1, Tier: "basic", Status: "past_due", MaxCategories: 3, HistoryDays: 30,
		ExpiresAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/b2b/reports?categories=7&platforms=1&from=2026-06-14&to=2026-06-21", nil)
	req.Header.Set("X-B2B-Org-Id", "1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, 402, rec.Code)
}

func TestReport_ScopeExceeded_403(t *testing.T) {
	mux := setupHandler(t, report.Subscription{
		ID: 1, Tier: "basic", Status: "active", MaxCategories: 3, HistoryDays: 30,
		ExpiresAt: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
	}, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/b2b/reports?categories=1,2,3,4&platforms=1&from=2026-06-14&to=2026-06-21", nil)
	req.Header.Set("X-B2B-Org-Id", "1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, 403, rec.Code)
}

func TestReport_OK(t *testing.T) {
	day := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	med := int64(320000)
	mux := setupHandler(t, report.Subscription{
		ID: 1, Tier: "pro", Status: "active", MaxCategories: 10, HistoryDays: 180, CanExport: true,
		ExpiresAt: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
	}, []trend.MarketTrendCell{{
		CategoryID: 7, PlatformID: 1, Day: day, SKUCount: 100,
		MedianP: &med, P25P: &med, P75P: &med, AvgDiscountPct: ptrF(10),
		ComputedAt: time.Date(2026, 6, 21, 1, 0, 0, 0, time.UTC),
	}})
	req := httptest.NewRequest(http.MethodGet, "/v1/b2b/reports?categories=7&platforms=1&from=2026-06-14&to=2026-06-21", nil)
	req.Header.Set("X-B2B-Org-Id", "1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)
	var body report.Report
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Cells, 1)
}

func TestExport_ForbiddenWithoutCanExport(t *testing.T) {
	mux := setupHandler(t, report.Subscription{
		ID: 1, Tier: "basic", Status: "active", MaxCategories: 3, HistoryDays: 30, CanExport: false,
		ExpiresAt: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
	}, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/b2b/reports/export?categories=7&platforms=1&from=2026-06-14&to=2026-06-21&format=csv", nil)
	req.Header.Set("X-B2B-Org-Id", "1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, 403, rec.Code)
}

func ptrF(v float64) *float64 { return &v }
