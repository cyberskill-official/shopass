package trend

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAggregate_PercentileOrder(t *testing.T) {
	pts := makePoints(60, 100_000, 200_000)
	c := AggregateCell(7, 1, d0, pts)
	require.False(t, c.Suppressed)
	require.LessOrEqual(t, *c.P25P, *c.MedianP)
	require.LessOrEqual(t, *c.MedianP, *c.P75P)
	require.GreaterOrEqual(t, *c.AvgDiscountPct, 0.0)
	require.LessOrEqual(t, *c.AvgDiscountPct, 100.0)
}

func TestAggregate_BelowK_Suppressed(t *testing.T) {
	pts := makePoints(49, 100_000, 200_000)
	c := AggregateCell(7, 1, d0, pts)
	require.True(t, c.Suppressed)
	require.Nil(t, c.MedianP)
	require.EqualValues(t, 49, c.SKUCount)
}

func TestAggregate_NoIdentifierColumns(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	mig := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "0001_market_trend_daily.sql")
	b, err := os.ReadFile(mig)
	require.NoError(t, err)
	// Strip SQL comments before scanning for forbidden identity columns.
	var cleaned strings.Builder
	for _, line := range strings.Split(string(b), "\n") {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}
		cleaned.WriteString(line)
		cleaned.WriteByte('\n')
	}
	sql := strings.ToLower(cleaned.String())
	require.Contains(t, sql, "market_trend_daily")
	require.NotContains(t, sql, "product_id")
	require.NotContains(t, sql, "shop_id")
	require.NotContains(t, sql, "user_id")
}

func TestAggregateSourceSQL_NoUserLevelTables(t *testing.T) {
	sql := strings.ToLower(AggregateSourceSQL)
	require.Contains(t, sql, "price_daily")
	require.Contains(t, sql, "tracked_product")
	for _, bad := range ForbiddenSourceTokens {
		require.NotContains(t, sql, bad, "DEC-B2B-01: must not touch %s", bad)
	}
}

func TestBuildCell_SuppressesBelowK(t *testing.T) {
	c := BuildCell(1, 1, time.Now(), 49, 1, 2, 3, 10)
	require.True(t, c.Suppressed)
	require.Nil(t, c.MedianP)
}

func TestBuildCell_PublishesAtK(t *testing.T) {
	c := BuildCell(1, 1, time.Now(), 50, 10, 20, 30, 12.5)
	require.False(t, c.Suppressed)
	require.Equal(t, int64(20), *c.MedianP)
}

func TestKMinConstant(t *testing.T) {
	require.Equal(t, 50, KMin)
	require.True(t, Publishable(50))
	require.False(t, Publishable(49))
}
