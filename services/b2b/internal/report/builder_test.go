package report

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/b2b/internal/trend"
)

type stubTrend struct {
	cells []trend.MarketTrendCell
}

func (s stubTrend) QueryCells(_ context.Context, categoryID int64, platformID int16, from, to time.Time) ([]trend.MarketTrendCell, error) {
	var out []trend.MarketTrendCell
	for _, c := range s.cells {
		if c.CategoryID != categoryID || c.PlatformID != platformID || c.Suppressed {
			continue
		}
		d := time.Date(c.Day.Year(), c.Day.Month(), c.Day.Day(), 0, 0, 0, 0, time.UTC)
		if (d.Equal(from) || d.After(from)) && d.Before(to) {
			out = append(out, c)
		}
	}
	return out, nil
}

func TestBuild_AllSuppressed_EmptyWithReason(t *testing.T) {
	// QueryCells already filters suppressed — stub returns empty → insufficient_data.
	b := &Builder{Trend: stubTrend{}, Now: func() time.Time { return dNow }}
	r, err := b.Build(context.Background(), ReportScope{
		CategoryIDs: []int64{991}, PlatformIDs: []int16{1},
		From: d0, To: d0.AddDate(0, 0, 7),
	})
	require.NoError(t, err)
	require.Empty(t, r.Cells)
	require.Equal(t, "insufficient_data", r.Reason)
	require.Nil(t, r.SourceComputedAt)
}

func TestBuild_PublishedCells_Series(t *testing.T) {
	cells := make([]trend.MarketTrendCell, 0, 7)
	for i := 0; i < 7; i++ {
		day := d0.AddDate(0, 0, i)
		med := int64(300_000 + i)
		cells = append(cells, trend.MarketTrendCell{
			CategoryID: 7, PlatformID: 1, Day: day, SKUCount: 60,
			MedianP: &med, P25P: ptr(int64(250_000)), P75P: ptr(int64(400_000)),
			AvgDiscountPct: ptr(14.0), ComputedAt: dNow.Add(-time.Hour),
		})
	}
	b := &Builder{Trend: stubTrend{cells: cells}, Now: func() time.Time { return dNow }}
	r, err := b.Build(context.Background(), ReportScope{
		CategoryIDs: []int64{7}, PlatformIDs: []int16{1},
		From: d0, To: d0.AddDate(0, 0, 7),
	})
	require.NoError(t, err)
	require.Len(t, r.Cells, 7)
	require.NotNil(t, r.SourceComputedAt)
	require.False(t, r.SourceComputedAt.IsZero())
	require.Empty(t, r.Reason)
}

func TestExport_NoIdentifierColumns(t *testing.T) {
	r := Report{Cells: []TrendPoint{{
		CategoryID: 7, PlatformID: 1, Day: "2026-06-20",
		MedianP: 320000, P25P: 250000, P75P: 410000, AvgDiscountPct: 14.3, SKUCount: 100,
	}}}
	csv, err := ExportCSV(r)
	require.NoError(t, err)
	require.NotContains(t, strings.ToLower(csv), "product_id")
	require.NotContains(t, strings.ToLower(csv), "shop_id")
	require.NotContains(t, strings.ToLower(csv), "user_id")
	js, err := ExportJSON(r)
	require.NoError(t, err)
	require.NotContains(t, strings.ToLower(string(js)), "product_id")
}

var d0 = time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)

func ptr[T any](v T) *T { return &v }
